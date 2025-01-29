package proxies

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

// Flags struct for user input
type Flags struct {
	Port int
}

// ParseFlags parses the user input flags and returns a Flags instance
func (ft *Flags) ProxyFlagInput(fs *flag.FlagSet) {
	fs.IntVar(&ft.Port, "port", 8181, "Port for the SOCKS5 proxy (default: 8181)")
}

// Address returns the port as a combined string
func (f *Flags) Address() string {
	return fmt.Sprintf(":%d", f.Port)
}

func ProxyLogic(pf *Flags) {
	serverPort := fmt.Sprintf(":%d", pf.Port)
	server := NewServer(serverPort)
	if err := server.Start(); err != nil {
		log.Fatalf("[-] Failed to start SOCKS5 server: %v", err)
	}
}

// Server represents a SOCKS5 proxy server.
type Server struct {
	Addr string
}

// NewServer creates a new SOCKS5 server.
func NewServer(addr string) *Server {
	return &Server{Addr: addr}
}

// Start starts the SOCKS5 server.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("[-] Failed to listen on %s: %v", s.Addr, err)
	}
	defer listener.Close()

	log.Printf("[+] SOCKS5 proxy server listening on %s\n", s.Addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[-] Failed to accept connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a new client connection.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// SOCKS5 authentication negotiation
	if err := s.authenticate(conn); err != nil {
		log.Printf("[-] Authentication failed: %v\n", err)
		return
	}

	// Read the client's request
	request, err := s.readRequest(conn)
	if err != nil {
		log.Printf("[-] Failed to read request: %v\n", err)
		return
	}

	// Connect to the target server
	targetConn, err := net.Dial("tcp", request.DestAddr)
	if err != nil {
		log.Printf("[-] Failed to connect to target: %v\n", err)
		return
	}
	defer targetConn.Close()

	// Send a success response to the client
	if err := s.sendSuccessResponse(conn, targetConn.LocalAddr().(*net.TCPAddr)); err != nil {
		log.Printf("[-] Failed to send success response: %v\n", err)
		return
	}

	// Forward data between client and target
	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}

// authenticate performs SOCKS5 authentication.
func (s *Server) authenticate(conn net.Conn) error {
	// Read the authentication method selection message
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("[-] Failed to read auth header: %v", err)
	}

	// Check SOCKS version (should be 5)
	if buf[0] != 0x05 {
		return errors.New("[-] Unsupported SOCKS version")
	}

	// Read the number of authentication methods supported
	numMethods := int(buf[1])
	methods := make([]byte, numMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("[-] Failed to read auth methods: %v", err)
	}

	// Check if "no authentication" is supported
	noAuthSupported := false
	for _, method := range methods {
		if method == 0x00 { // 0x00 = no authentication
			noAuthSupported = true
			break
		}
	}

	if !noAuthSupported {
		return errors.New("[-] No supported authentication methods")
	}

	// Send the selected authentication method (0x00 = no authentication)
	conn.Write([]byte{0x05, 0x00})
	return nil
}

// Request represents a SOCKS5 request.
type Request struct {
	Version  byte
	Cmd      byte
	DestAddr string
}

// readRequest reads the SOCKS5 request from the client.
func (s *Server) readRequest(conn net.Conn) (*Request, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, fmt.Errorf("[-] Failed to read request header: %v", err)
	}

	// This is for logging, if authentication fails, uncomment below.
	// We should be receiving [5 1 0 1],
	// If the first slice is not a 5, we have failed
	// log.Printf("received request header: %v", buf)

	version, cmd := buf[0], buf[1]
	if version != 0x05 {
		return nil, errors.New("[-] Unsupported SOCKS version")
	}

	// Read the destination address
	addrType := buf[3]
	var destAddr string
	switch addrType {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, fmt.Errorf("[-] Failed to read IPv4 address: %v", err)
		}
		destAddr = net.IP(ip).String()
	case 0x03: // Domain name
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, fmt.Errorf("[-] Failed to read domain length: %v", err)
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, fmt.Errorf("[-] Failed to read domain: %v", err)
		}
		destAddr = string(domain)
	default:
		return nil, fmt.Errorf("[-] Unsupported address type: %d", addrType)
	}

	// Read the destination port
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return nil, fmt.Errorf("[-] Failed to read port: %v", err)
	}
	port := int(portBuf[0])<<8 | int(portBuf[1])
	destAddr = fmt.Sprintf("%s:%d", destAddr, port)

	return &Request{
		Version:  version,
		Cmd:      cmd,
		DestAddr: destAddr,
	}, nil
}

// sendSuccessResponse sends a success response to the client.
func (s *Server) sendSuccessResponse(conn net.Conn, addr *net.TCPAddr) error {
	response := []byte{0x05, 0x00, 0x00, 0x01}
	ip := addr.IP.To4()
	response = append(response, ip...)
	port := []byte{byte(addr.Port >> 8), byte(addr.Port & 0xff)}
	response = append(response, port...)

	_, err := conn.Write(response)
	return err
}
