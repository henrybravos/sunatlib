// Package signer provides XML digital signature functionality for SUNAT Peru
package signer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// XMLSigner handles XML digital signatures using xmlsec1
type XMLSigner struct {
	privateKeyPath   string
	certificatePath  string
	tempDir         string
}

// NewXMLSigner creates a new XML signer with private key and certificate paths
func NewXMLSigner(privateKeyPath, certificatePath string) (*XMLSigner, error) {
	// Verify files exist
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("private key file not found: %s", privateKeyPath)
	}
	if _, err := os.Stat(certificatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate file not found: %s", certificatePath)
	}

	// Create temp directory for operations
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("sunatlib_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &XMLSigner{
		privateKeyPath:  privateKeyPath,
		certificatePath: certificatePath,
		tempDir:        tempDir,
	}, nil
}

// SignXML signs an XML document and returns the signed XML bytes
func (s *XMLSigner) SignXML(xmlContent []byte) ([]byte, error) {
	// Create template with signature placeholder
	template, err := s.createSignatureTemplate(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature template: %w", err)
	}

	// Write template to temp file
	templateFile := filepath.Join(s.tempDir, "template.xml")
	if err := os.WriteFile(templateFile, template, 0644); err != nil {
		return nil, fmt.Errorf("failed to write template file: %w", err)
	}

	// Sign using xmlsec1
	outputFile := filepath.Join(s.tempDir, "signed.xml")
	cmd := exec.Command("xmlsec1", "sign",
		"--lax-key-search",
		"--privkey-pem", fmt.Sprintf("%s,%s", s.privateKeyPath, s.certificatePath),
		"--output", outputFile,
		templateFile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("xmlsec1 signing failed: %w\nOutput: %s", err, string(output))
	}

	// Check if signing was successful
	if !strings.Contains(string(output), "Signature status: OK") {
		return nil, fmt.Errorf("signing failed - xmlsec1 output: %s", string(output))
	}

	// Read signed XML
	signedXML, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read signed XML: %w", err)
	}

	return signedXML, nil
}

// createSignatureTemplate creates a UBL Invoice template with signature placeholder
func (s *XMLSigner) createSignatureTemplate(xmlContent []byte) ([]byte, error) {
	// Parse the input XML and inject signature template
	xmlStr := string(xmlContent)
	
	// Find ExtensionContent and inject signature template
	signatureTemplate := `    <ds:Signature Id="SignatureSP">
        <ds:SignedInfo>
            <ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
            <ds:SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
            <ds:Reference URI="">
                <ds:Transforms>
                    <ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
                </ds:Transforms>
                <ds:DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/>
                <ds:DigestValue/>
            </ds:Reference>
        </ds:SignedInfo>
        <ds:SignatureValue/>
        <ds:KeyInfo>
            <ds:X509Data>
                <ds:X509Certificate/>
            </ds:X509Data>
        </ds:KeyInfo>
    </ds:Signature>`

	// Ensure xmlns:ds is present
	if !strings.Contains(xmlStr, `xmlns:ds="http://www.w3.org/2000/09/xmldsig#"`) {
		xmlStr = strings.Replace(xmlStr, "<Invoice ", `<Invoice xmlns:ds="http://www.w3.org/2000/09/xmldsig#" `, 1)
	}

	// Find empty ExtensionContent and inject signature
	if strings.Contains(xmlStr, "<ext:ExtensionContent>\n    </ext:ExtensionContent>") {
		xmlStr = strings.Replace(xmlStr, "<ext:ExtensionContent>\n    </ext:ExtensionContent>",
			"<ext:ExtensionContent>\n"+signatureTemplate+"\n    </ext:ExtensionContent>", 1)
	} else if strings.Contains(xmlStr, "<ext:ExtensionContent></ext:ExtensionContent>") {
		xmlStr = strings.Replace(xmlStr, "<ext:ExtensionContent></ext:ExtensionContent>",
			"<ext:ExtensionContent>\n"+signatureTemplate+"\n    </ext:ExtensionContent>", 1)
	} else if strings.Contains(xmlStr, "<ext:ExtensionContent/>") {
		xmlStr = strings.Replace(xmlStr, "<ext:ExtensionContent/>",
			"<ext:ExtensionContent>\n"+signatureTemplate+"\n    </ext:ExtensionContent>", 1)
	} else {
		return nil, fmt.Errorf("no suitable ExtensionContent found for signature injection")
	}

	return []byte(xmlStr), nil
}

// Cleanup removes temporary files
func (s *XMLSigner) Cleanup() error {
	if s.tempDir != "" {
		return os.RemoveAll(s.tempDir)
	}
	return nil
}