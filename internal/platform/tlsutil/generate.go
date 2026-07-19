// Package tlsutil provides minimal TLS material helpers for tests and live runtime prep.
package tlsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// WriteSelfSignedFiles creates a self-signed certificate and key for the given IPs.
func WriteSelfSignedFiles(dir string, ips ...string) (certPath, keyPath string, err error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", "", fmt.Errorf("mkdir tls dir: %w", err)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "vyntrio-appliance"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, ipStr := range ips {
		if ip := net.ParseIP(ipStr); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return "", "", fmt.Errorf("create certificate: %w", err)
	}

	certPath = filepath.Join(dir, "dashboard-cert.pem")
	keyPath = filepath.Join(dir, "dashboard-key.pem")

	certOut, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", "", fmt.Errorf("write cert: %w", err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		_ = certOut.Close()
		return "", "", fmt.Errorf("encode cert: %w", err)
	}
	if err := certOut.Close(); err != nil {
		return "", "", fmt.Errorf("close cert: %w", err)
	}

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", "", fmt.Errorf("write key: %w", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		_ = keyOut.Close()
		return "", "", fmt.Errorf("encode key: %w", err)
	}
	if err := keyOut.Close(); err != nil {
		return "", "", fmt.Errorf("close key: %w", err)
	}

	return certPath, keyPath, nil
}
