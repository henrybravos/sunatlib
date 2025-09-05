// Package sunatlib provides XML digital signature functionality for SUNAT Peru
package sunatlib

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/henrybravos/sunatlib/signer"
	"github.com/henrybravos/sunatlib/utils"
)

// SUNATClient handles interactions with SUNAT web services
type SUNATClient struct {
	RUC       string
	Username  string
	Password  string
	Endpoint  string
	signer    *signer.XMLSigner
}

// NewSUNATClient creates a new SUNAT client
func NewSUNATClient(ruc, username, password, endpoint string) *SUNATClient {
	return &SUNATClient{
		RUC:      ruc,
		Username: username,
		Password: password,
		Endpoint: endpoint,
	}
}

// SetCertificate configures the XML signer with certificate files
func (c *SUNATClient) SetCertificate(privateKeyPath, certificatePath string) error {
	var err error
	c.signer, err = signer.NewXMLSigner(privateKeyPath, certificatePath)
	return err
}

// SetCertificateFromPFX extracts and configures certificate from PFX file
func (c *SUNATClient) SetCertificateFromPFX(pfxPath, password, tempDir string) error {
	// Extract PEM files from PFX
	privateKeyPath, certPath, err := utils.ExtractPEMFromPFX(pfxPath, password, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract PEM from PFX: %w", err)
	}

	// Set up signer
	return c.SetCertificate(privateKeyPath, certPath)
}

// SignXML signs an XML document and returns the signed XML
func (c *SUNATClient) SignXML(xmlContent []byte) ([]byte, error) {
	if c.signer == nil {
		return nil, fmt.Errorf("certificate not configured - use SetCertificate() first")
	}

	// Check xmlsec1 availability
	if err := utils.CheckXMLSec1Available(); err != nil {
		return nil, err
	}

	// Sign the XML
	signedXML, err := c.signer.SignXML(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to sign XML: %w", err)
	}

	return signedXML, nil
}

// SendToSUNAT sends a signed XML document to SUNAT
func (c *SUNATClient) SendToSUNAT(signedXML []byte, documentType, seriesNumber string) (*SUNATResponse, error) {
	return c.sendToSUNAT(signedXML, documentType, seriesNumber)
}

// SignAndSendInvoice signs an XML invoice and sends it to SUNAT (convenience method)
func (c *SUNATClient) SignAndSendInvoice(xmlContent []byte, documentType, seriesNumber string) (*SUNATResponse, error) {
	// Sign the XML
	signedXML, err := c.SignXML(xmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to sign XML: %w", err)
	}

	// Send to SUNAT
	return c.SendToSUNAT(signedXML, documentType, seriesNumber)
}

// sendToSUNAT handles the SOAP communication with SUNAT
func (c *SUNATClient) sendToSUNAT(signedXML []byte, documentType, seriesNumber string) (*SUNATResponse, error) {
	// Create ZIP file
	zipData, zipName, err := c.createZIP(signedXML, documentType, seriesNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP: %w", err)
	}

	// Encode to base64
	zipB64 := base64.StdEncoding.EncodeToString(zipData)

	// Build SOAP envelope
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ser="http://service.sunat.gob.pe" xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
  <soapenv:Header>
    <wsse:Security>
      <wsse:UsernameToken>
        <wsse:Username>%s%s</wsse:Username>
        <wsse:Password>%s</wsse:Password>
      </wsse:UsernameToken>
    </wsse:Security>
  </soapenv:Header>
  <soapenv:Body>
    <ser:sendBill>
      <fileName>%s</fileName>
      <contentFile>%s</contentFile>
    </ser:sendBill>
  </soapenv:Body>
</soapenv:Envelope>`, c.RUC, c.Username, c.Password, zipName, zipB64)

	// Send HTTP request
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(soapBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return c.parseResponse(responseData)
}

// createZIP creates a ZIP file with the signed XML
func (c *SUNATClient) createZIP(signedXML []byte, documentType, seriesNumber string) ([]byte, string, error) {
	xmlName := fmt.Sprintf("%s-%s-%s.xml", c.RUC, documentType, seriesNumber)
	zipName := fmt.Sprintf("%s-%s-%s.zip", c.RUC, documentType, seriesNumber)

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	fw, err := zipWriter.Create(xmlName)
	if err != nil {
		return nil, "", err
	}

	if _, err := fw.Write(signedXML); err != nil {
		return nil, "", err
	}

	zipWriter.Close()

	return buf.Bytes(), zipName, nil
}

// SUNATResponse represents the response from SUNAT
type SUNATResponse struct {
	Success          bool
	Message          string
	ResponseXML      []byte
	ApplicationResponse []byte
	Error            error
}

// parseResponse parses SUNAT's SOAP response
func (c *SUNATClient) parseResponse(responseData []byte) (*SUNATResponse, error) {
	responseStr := string(responseData)
	response := &SUNATResponse{
		ResponseXML: responseData,
	}

	// Check for SOAP fault
	if strings.Contains(responseStr, "<soap-env:Fault") {
		response.Success = false
		
		// Extract fault string
		if start := strings.Index(responseStr, "<faultstring>"); start != -1 {
			start += 13
			if end := strings.Index(responseStr[start:], "</faultstring>"); end != -1 {
				response.Message = responseStr[start : start+end]
				// Decode HTML entities
				response.Message = strings.ReplaceAll(response.Message, "&#243;", "ó")
			}
		}
		
		return response, nil
	}

	// Check for successful response
	if strings.Contains(responseStr, "<br:sendBillResponse") {
		response.Success = true
		response.Message = "Documento enviado exitosamente"

		// Extract application response (base64 encoded ZIP)
		if start := strings.Index(responseStr, "<applicationResponse>"); start != -1 {
			start += 21
			if end := strings.Index(responseStr[start:], "</applicationResponse>"); end != -1 {
				b64Data := responseStr[start : start+end]
				appResponse, err := base64.StdEncoding.DecodeString(b64Data)
				if err == nil {
					response.ApplicationResponse = appResponse
				}
			}
		}
		
		return response, nil
	}

	response.Success = false
	response.Message = "Respuesta no reconocida de SUNAT"
	return response, nil
}

// Cleanup cleans up temporary files
func (c *SUNATClient) Cleanup() error {
	if c.signer != nil {
		return c.signer.Cleanup()
	}
	return nil
}

// SaveApplicationResponse saves the CDR (Constancia de Recepción) to a file
func (r *SUNATResponse) SaveApplicationResponse(outputPath string) error {
	if r.ApplicationResponse == nil {
		return fmt.Errorf("no application response data available")
	}

	return os.WriteFile(outputPath, r.ApplicationResponse, 0644)
}