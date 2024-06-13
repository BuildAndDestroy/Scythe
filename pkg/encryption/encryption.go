package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

// func GenerateCaCertificate() (tls.Certificate, error) {
func GenerateCaCertificate() {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalln("Unable to generate Private Key")
	}

	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			// This needs to converted to something unique
			// Maybe automated or user input
			Organization: []string{"My CA"},
			CommonName:   "my-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	// Create CA certificate
	caCert, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		log.Fatalln(err)
	}

	// Save CA certificate
	certOut, err := os.Create("ca.crt")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caCert})
	certOut.Close()

	// Save CA private key
	keyOut, err := os.Create("ca.key")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)})
	keyOut.Close()

}

func GenerateServerCertificateWithCA() {
	// Load CA certificate
	caCertPEM, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatalln(err)
	}
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil || caCertBlock.Type != "CERTIFICATE" {
		log.Fatalln("failed to decode CA certificate")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		log.Fatalln(err)
	}

	// Load CA private key
	caKeyPEM, err := ioutil.ReadFile("ca.key")
	if err != nil {
		log.Fatalln(err)
	}
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil || caKeyBlock.Type != "RSA PRIVATE KEY" {
		log.Fatalln("failed to decode CA private key")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		log.Fatalln(err)
	}

	// Generate server private key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	// Create server certificate template
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"My Server"},
			CommonName:   "my-server",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create server certificate
	serverCert, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		log.Fatalln(err)
	}

	// Save server certificate
	certOut, err := os.Create("server.crt")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: serverCert})
	certOut.Close()

	// Save server private key
	keyOut, err := os.Create("server.key")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	keyOut.Close()
}

func GenerateClientCertificateWithCA() {
	// Load CA certificate
	caCertPEM, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatalln(err)
	}
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil || caCertBlock.Type != "CERTIFICATE" {
		log.Fatalln("failed to decode CA certificate")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		log.Fatalln(err)
	}

	// Load CA private key
	caKeyPEM, err := ioutil.ReadFile("ca.key")
	if err != nil {
		log.Fatalln(err)
	}
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil || caKeyBlock.Type != "RSA PRIVATE KEY" {
		log.Fatalln("failed to decode CA private key")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		log.Fatalln(err)
	}

	// Generate client private key
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	// Create client certificate template
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			Organization: []string{"My Client"},
			CommonName:   "my-client",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	// Create client certificate
	clientCert, err := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		log.Fatalln(err)
	}

	// Save client certificate
	certOut, err := os.Create("client.crt")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: clientCert})
	certOut.Close()

	// Save client private key
	keyOut, err := os.Create("client.key")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})
	keyOut.Close()
}

func GenerateSelfSignedServerCert() ([]byte, []byte, error) {
	// Generate server private key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	// Create server certificate template
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My Server"},
			CommonName:   "my-server",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		IsCA:        true,
	}

	// Self-sign the server certificate
	serverCert, err := x509.CreateCertificate(rand.Reader, serverTemplate, serverTemplate, &serverKey.PublicKey, serverKey)
	if err != nil {
		log.Fatalln(err)
	}

	// PEM encode the server certificate
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCert})

	// PEM encode the server private key
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return certPEM, keyPEM, nil

	// Use the below code to write to disk. We don't really want that.
	// Save server certificate
	/* certOut, err := os.Create("server.crt")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: serverCert})
	certOut.Close()

	// Save server private key
	keyOut, err := os.Create("server.key")
	if err != nil {
		log.Fatalln(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	keyOut.Close()*/
}

func GenerateSelfSignedClientCert() ([]byte, []byte, error) {
	// Generate client private key
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	// Create client certificate template
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"My Client"},
			CommonName:   "my-client",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IsCA:        true,
	}

	// Self-sign the client certificate
	clientCert, err := x509.CreateCertificate(rand.Reader, clientTemplate, clientTemplate, &clientKey.PublicKey, clientKey)
	if err != nil {
		log.Fatalln(err)
	}

	// PEM encode the client certificate
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCert})

	// PEM encode the client private key
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	return certPEM, keyPEM, nil

	// Use below code if you want to write to disk. Would not recommend.
	/* // Save client certificate
	certOut, err := os.Create("client.crt")
	if err != nil {
		panic(err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: clientCert})
	certOut.Close()

	// Save client private key
	keyOut, err := os.Create("client.key")
	if err != nil {
		panic(err)
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})
	keyOut.Close() */
}
