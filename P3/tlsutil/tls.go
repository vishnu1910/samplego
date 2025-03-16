package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// LoadServerTLSConfig loads the server's certificate, key and CA certificate for mutual TLS.
func LoadServerTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// Load server certificate and key.
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert/key: %v", err)
	}

	// Load CA cert.
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
		// Require client certificate verification.
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	return tlsConfig, nil
}

// LoadClientTLSConfig loads a TLS configuration for a client.
// This is used by the Payment Gateway when dialing bank servers.
func LoadClientTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// Load client certificate.
	clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key: %v", err)
	}

	// Load CA cert.
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		// Ensure the ServerName matches the SAN in the server certificate.
		ServerName: "localhost",
	}, nil
}

