package filetransfer

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/BuildAndDestroy/backdoorBoi/pkg/encryption"
	"github.com/BuildAndDestroy/backdoorBoi/pkg/proxies"
)

type FileTransfer struct {
	Port     int
	FileName string
	Listen   bool
	Download bool
	Send     bool
	Hostname string
	Tls      bool
	Proxy    string
}

func (ft *FileTransfer) FileTransferInput(fs *flag.FlagSet) {
	fs.IntVar(&ft.Port, "port", 8080, "Port to bind or connect to")
	fs.StringVar(&ft.FileName, "filename", "", "Filename to transfer or request")
	fs.BoolVar(&ft.Listen, "listen", false, "Act as server to serve or receive files")
	fs.BoolVar(&ft.Download, "download", false, "Act as client to download a file")
	fs.BoolVar(&ft.Send, "send", false, "Act as client to send a file")
	fs.StringVar(&ft.Hostname, "hostname", "127.0.0.1", "Server hostname or IP address")
	fs.BoolVar(&ft.Tls, "tls", false, "Use encryption. RECOMMENDED")
	fs.StringVar(&ft.Proxy, "proxy", "", "Use a SOCKS5 proxy between us and target, example 127.0.0.1:9050")
}

// Logic check. Make sure user input is usable
func FileTransferLogic(fti *FileTransfer) {
	if fti.Listen && fti.Send {
		log.Fatalln("[*] Either Send or Listen, unable to do both.")
	}
	if fti.Listen && !fti.Tls {
		RunServer(fti.Port)
	}
	if fti.Listen && fti.Tls {
		tlsConfig := encryption.SetupTLSServer()
		TlsRunServer(fti.Port, tlsConfig)
	}
	if fti.Download && !fti.Tls && fti.Proxy != "" {
		ProxyRunDownloadClient(fti.Hostname, fti.Port, fti.FileName, fti.Proxy)
	}
	if fti.Download && fti.Tls && fti.Proxy != "" {
		tlsConfig := encryption.SetupTLSClient()
		ProxyTlsRunDownloadClient(fti.Hostname, fti.Port, fti.FileName, tlsConfig, fti.Proxy)
	}
	if fti.Download && !fti.Tls {
		RunDownloadClient(fti.Hostname, fti.Port, fti.FileName)
	}
	if fti.Download && fti.Tls {
		tlsConfig := encryption.SetupTLSClient()
		TlsRunDownloadClient(fti.Hostname, fti.Port, fti.FileName, tlsConfig)
	}
	if fti.Send && !fti.Tls && fti.Proxy != "" { // UPDATE ME
		ProxyRunSendClient(fti.Hostname, fti.Port, fti.FileName, fti.Proxy)
	}
	if fti.Send && fti.Tls && fti.Proxy != "" { // UPDATE ME
		tlsConfig := encryption.SetupTLSClient()
		ProxyTlsRunSendClient(fti.Hostname, fti.Port, fti.FileName, tlsConfig, fti.Proxy)
	}
	if fti.Send && !fti.Tls {
		RunSendClient(fti.Hostname, fti.Port, fti.FileName)
	}
	if fti.Send && fti.Tls {
		tlsConfig := encryption.SetupTLSClient()
		TlsRunSendClient(fti.Hostname, fti.Port, fti.FileName, tlsConfig)
	}
	if fti.Listen && fti.Download && fti.Send {
		log.Fatalln("You must specify one of --listen, --download, or --send")
	}
}

// Server logic but with TLS
func TlsRunServer(port int, tlsConfig *tls.Config) {
	// Create a TLS listener
	address := fmt.Sprintf(":%d", port)
	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		log.Fatalf("[-] Server: listen: %s", err)
	}
	defer listener.Close()

	log.Printf("[*] Server listening on %s with TLS\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		go handleClient(conn)
	}
}

// Server Logic
func RunServer(port int) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
	defer listener.Close()

	log.Printf("Server listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		go handleClient(conn)
	}
}

// Handle Client connection from RunServer
func handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read command type: "DOWNLOAD" or "SEND"
	command, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading command: %s", err)
		return
	}
	command = strings.TrimSpace(command)

	switch command {
	case "DOWNLOAD":
		handleDownloadRequest(conn, reader)
	case "SEND":
		handleFileReceive(conn, reader)
	default:
		log.Printf("Unknown command: %s", command)
		conn.Write([]byte("ERROR: Unknown command\n"))
	}
}

// Handle file download request from client
func handleDownloadRequest(conn net.Conn, reader *bufio.Reader) {
	// Read requested file name
	fileName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading file name: %s", err)
		return
	}
	fileName = strings.TrimSpace(fileName)
	log.Printf("Client requested file: %s", fileName)

	// Open the requested file
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error opening file %s: %s", fileName, err)
		conn.Write([]byte("ERROR: File not found\n"))
		return
	}
	defer file.Close()

	// Send confirmation to client
	conn.Write([]byte("OK\n"))

	// Send file data
	_, err = io.Copy(conn, file)
	if err != nil {
		log.Printf("Error sending file: %s", err)
		return
	}

	log.Printf("File %s sent successfully", fileName)
}

// Handle file receive from client
func handleFileReceive(conn net.Conn, reader *bufio.Reader) {
	// Read file name from the client
	fileName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading file name: %s", err)
		return
	}
	fileName = strings.TrimSpace(fileName)
	log.Printf("Receiving file: %s", fileName)

	// Create the file to save the received data
	outputFile, err := os.Create(fileName)
	if err != nil {
		log.Printf("Error creating file %s: %s", fileName, err)
		conn.Write([]byte("ERROR: Could not create file\n"))
		return
	}
	defer outputFile.Close()

	// Copy data from connection to file
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		log.Printf("Error receiving file: %s", err)
		return
	}

	log.Printf("File %s received successfully", fileName)
}

func ProxyTlsRunDownloadClient(hostname string, port int, fileName string, tlsConfig *tls.Config, proxyAddress string) {
	var tlsBoolean bool = true
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := proxies.DialThroughSOCKS5(proxyAddress, address, tlsConfig, tlsBoolean)

	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Send command and file name
	fmt.Fprintf(conn, "DOWNLOAD\n")
	fmt.Fprintf(conn, "%s\n", fileName)

	// Read server response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading server response: %s", err)
	}
	response = strings.TrimSpace(response)

	if response != "OK" {
		log.Fatalf("Server responded with error: %s", response)
	}

	// Create local file to save the downloaded content
	outputFileName := filepath.Base(fileName)
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer outputFile.Close()

	// Receive file data
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		log.Fatalf("Error downloading file: %s", err)
	}

	log.Printf("File %s downloaded successfully", outputFileName)
}

// Client Logic: Download file over TLS
func TlsRunDownloadClient(hostname string, port int, fileName string, tlsConfig *tls.Config) {
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Send command and file name
	fmt.Fprintf(conn, "DOWNLOAD\n")
	fmt.Fprintf(conn, "%s\n", fileName)

	// Read server response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading server response: %s", err)
	}
	response = strings.TrimSpace(response)

	if response != "OK" {
		log.Fatalf("Server responded with error: %s", response)
	}

	// Create local file to save the downloaded content
	outputFileName := filepath.Base(fileName)
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer outputFile.Close()

	// Receive file data
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		log.Fatalf("Error downloading file: %s", err)
	}

	log.Printf("File %s downloaded successfully", outputFileName)
}

// Download a file over a SOCKS5 proxy connection
func ProxyRunDownloadClient(hostname string, port int, fileName, proxyAddress string) {
	var tlsBoolean bool = false
	var tlsConfig *tls.Config
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := proxies.DialThroughSOCKS5(proxyAddress, address, tlsConfig, tlsBoolean)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Send command and file name
	fmt.Fprintf(conn, "DOWNLOAD\n")
	fmt.Fprintf(conn, "%s\n", fileName)

	// Read server response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading server response: %s", err)
	}
	response = strings.TrimSpace(response)

	if response != "OK" {
		log.Fatalf("Server responded with error: %s", response)
	}

	// Create local file to save the downloaded content
	outputFileName := filepath.Base(fileName)
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer outputFile.Close()

	// Receive file data
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		log.Fatalf("Error downloading file: %s", err)
	}

	log.Printf("File %s downloaded successfully", outputFileName)
}

// Client Logic: Download file
func RunDownloadClient(hostname string, port int, fileName string) {
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Send command and file name
	fmt.Fprintf(conn, "DOWNLOAD\n")
	fmt.Fprintf(conn, "%s\n", fileName)

	// Read server response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading server response: %s", err)
	}
	response = strings.TrimSpace(response)

	if response != "OK" {
		log.Fatalf("Server responded with error: %s", response)
	}

	// Create local file to save the downloaded content
	outputFileName := filepath.Base(fileName)
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer outputFile.Close()

	// Receive file data
	_, err = io.Copy(outputFile, reader)
	if err != nil {
		log.Fatalf("Error downloading file: %s", err)
	}

	log.Printf("File %s downloaded successfully", outputFileName)
}

func ProxyTlsRunSendClient(hostname string, port int, fileName string, tlsConfig *tls.Config, proxyAddress string) {
	var tlsBoolean bool = true
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := proxies.DialThroughSOCKS5(proxyAddress, address, tlsConfig, tlsBoolean)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Open the file to be sent
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err)
	}
	defer file.Close()

	// Send command and file name
	fmt.Fprintf(conn, "SEND\n")
	fmt.Fprintf(conn, "%s\n", filepath.Base(fileName))

	// Send file data
	_, err = io.Copy(conn, file)
	if err != nil {
		log.Fatalf("Error sending file: %s", err)
	}

	log.Printf("File %s sent successfully", fileName)
}

// Client Logic: Send file over TLS
func TlsRunSendClient(hostname string, port int, fileName string, tlsConfig *tls.Config) {
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := tls.Dial("tcp", address, tlsConfig)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Open the file to be sent
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err)
	}
	defer file.Close()

	// Send command and file name
	fmt.Fprintf(conn, "SEND\n")
	fmt.Fprintf(conn, "%s\n", filepath.Base(fileName))

	// Send file data
	_, err = io.Copy(conn, file)
	if err != nil {
		log.Fatalf("Error sending file: %s", err)
	}

	log.Printf("File %s sent successfully", fileName)
}

func ProxyRunSendClient(hostname string, port int, fileName, proxyAddress string) {
	var tlsBoolean bool = false
	var tlsConfig *tls.Config
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := proxies.DialThroughSOCKS5(proxyAddress, address, tlsConfig, tlsBoolean)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Open the file to be sent
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err)
	}
	defer file.Close()

	// Send command and file name
	fmt.Fprintf(conn, "SEND\n")
	fmt.Fprintf(conn, "%s\n", filepath.Base(fileName))

	// Send file data
	_, err = io.Copy(conn, file)
	if err != nil {
		log.Fatalf("Error sending file: %s", err)
	}

	log.Printf("File %s sent successfully", fileName)
}

// Client Logic: Send file
func RunSendClient(hostname string, port int, fileName string) {
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
	}
	defer conn.Close()

	// Open the file to be sent
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %s", err)
	}
	defer file.Close()

	// Send command and file name
	fmt.Fprintf(conn, "SEND\n")
	fmt.Fprintf(conn, "%s\n", filepath.Base(fileName))

	// Send file data
	_, err = io.Copy(conn, file)
	if err != nil {
		log.Fatalf("Error sending file: %s", err)
	}

	log.Printf("File %s sent successfully", fileName)
}
