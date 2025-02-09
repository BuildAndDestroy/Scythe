package proxies

import (
	"crypto/tls"
	"fmt"
	"net"

	"golang.org/x/net/proxy"
)

func DialThroughSOCKS5(socks5Addr, targetAddr string, tlsConfig *tls.Config, tlsBoolean bool) (net.Conn, error) {
	// Step 1: Create a SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("[-] Failed to create SOCKS5 dialer: %w", err)
	}

	// Step 2: Open a raw TCP connection via SOCKS5 proxy
	rawConn, err := dialer.Dial("tcp", targetAddr)
	if err != nil {
		return nil, fmt.Errorf("[-] Failed to dial target via SOCKS5: %w", err)
	}

	if !tlsBoolean {
		return rawConn, nil
	}
	// Step 3: Wrap the raw TCP connection in TLS
	tlsConn := tls.Client(rawConn, tlsConfig)

	// Step 4: Perform TLS handshake
	if err := tlsConn.Handshake(); err != nil {
		rawConn.Close() // Clean up if handshake fails
		return nil, fmt.Errorf("[-] TLS handshake failed: %w", err)
	}

	return tlsConn, nil
}
