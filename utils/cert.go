// Package utils provides certificate and utility functions for SUNAT XML signing
package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"software.sslmate.com/src/go-pkcs12"
)

// ExtractPEMFromPFX extracts PEM private key and certificate from PFX file
func ExtractPEMFromPFX(pfxPath, password, outputDir string) (privateKeyPath, certPath string, err error) {
	// Read PFX file
	pfxData, err := os.ReadFile(pfxPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read PFX file: %w", err)
	}

	// Decode PFX
	privateKey, cert, caCerts, err := pkcs12.DecodeChain(pfxData, password)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode PFX: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write private key to PEM
	privateKeyPath = filepath.Join(outputDir, "private_key.pem")
	keyPEM, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %w", err)
	}
	
	keyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create private key file: %w", err)
	}
	defer keyFile.Close()
	
	err = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: keyPEM})
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	// Write certificate to PEM
	certPath = filepath.Join(outputDir, "certificate.pem")
	certFile, err := os.Create(certPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certFile.Close()
	
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err != nil {
		return "", "", fmt.Errorf("failed to encode certificate: %w", err)
	}

	// Write CA certificates if present
	if len(caCerts) > 0 {
		caPath := filepath.Join(outputDir, "ca_certificates.pem")
		caFile, err := os.Create(caPath)
		if err == nil {
			defer caFile.Close()
			for _, caCert := range caCerts {
				pem.Encode(caFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})
			}
		}
	}

	return privateKeyPath, certPath, nil
}

// ValidateCertificate validates a certificate file
func ValidateCertificate(certPath string) (*x509.Certificate, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// CheckXMLSec1Available checks if xmlsec1 is available in the system
func CheckXMLSec1Available() error {
	cmd := exec.Command("xmlsec1", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("xmlsec1 not found - please install xmlsec1: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GetCertificateInfo returns basic information about a certificate
func GetCertificateInfo(certPath string) (map[string]string, error) {
	cert, err := ValidateCertificate(certPath)
	if err != nil {
		return nil, err
	}

	info := map[string]string{
		"Subject":    cert.Subject.String(),
		"Issuer":     cert.Issuer.String(),
		"NotBefore":  cert.NotBefore.String(),
		"NotAfter":   cert.NotAfter.String(),
		"SerialNumber": cert.SerialNumber.String(),
	}

	return info, nil
}