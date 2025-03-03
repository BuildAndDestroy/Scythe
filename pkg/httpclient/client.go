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
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

// RequestOptions defines available options for an HTTP request.
type RequestOptions struct {
	Method        string            // HTTP method (GET, POST, etc.)
	URL           string            // Request URL
	Headers       map[string]string // Request headers
	Body          interface{}       // Request body (JSON struct or string)
	Timeout       time.Duration     // Timeout duration
	SkipTLSVerify bool              // Whether to skip TLS verification
}

// Parse flags from user input
func (ro *RequestOptions) SetRequestFlag(fs *flag.FlagSet) {
	fs.StringVar(&ro.Method, "method", "GET", "HTTP method (GET, POST, etc.)")
	fs.StringVar(&ro.URL, "url", "", "Request URL")
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

// Make the request to our endpoint
func MakeRequest(opts *RequestOptions) (*ResponseData, error) {
	// Debugging
	// log.Println(opts.URL)
	// log.Println(opts.Timeout)
	// log.Println(opts.Body)
	// log.Println(opts.Headers)
	// log.Println(opts.SkipTLSVerify)
	if opts.URL == "" {
		return nil, errors.New("[-] URL is required")
	}

	if opts.Method == "GET" && opts.Body != nil {
		return nil, errors.New("[-] GET request and body payload doesn't make sense")
	}

	// Convert request body to JSON if necessary
	var reqBody io.Reader
	if opts.Body != nil {
		switch body := opts.Body.(type) {
		case string:
			reqBody = bytes.NewBufferString(body)
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("[-] Error encoding request body to JSON: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}
	}

	req, err := http.NewRequest(opts.Method, opts.URL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("[-] Error creating request: %w", err)
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
		return nil, fmt.Errorf("[-] Error reading response: %w", err)
	}

	// Parse JSON response if applicable
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err == nil {
		return &ResponseData{
			Status:     resp.StatusCode,
			Body:       string(respBody),
			JSONParsed: jsonResponse,
		}, nil
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
				log.Printf("[+] Response (%d): %s", resp.Status, resp.Body)
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
