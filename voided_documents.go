// Package sunatlib provides functionality for SUNAT voided documents (comunicación de baja)
package sunatlib

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/henrybravos/sunatlib/utils"
)

// VoidedDocument represents a document to be voided
type VoidedDocument struct {
	DocumentTypeCode string // Document type code (01=Invoice, 03=Receipt, etc.)
	DocumentSeries   string // Document series (e.g., "F001", "B001")
	DocumentNumber   string // Document correlative number
	VoidedReason     string // Reason for voiding the document
}

// VoidedDocumentsRequest represents a voided documents communication request
type VoidedDocumentsRequest struct {
	RUC             string           // Company RUC
	CompanyName     string           // Company name/reason social
	SeriesNumber    string           // Voided document series number (RA-YYYYMMDD-###)
	IssueDate       time.Time        // Issue date
	ReferenceDate   time.Time        // Reference date (date of voided documents)
	Documents       []VoidedDocument // List of documents to void
	Description     string           // Description of the voiding communication
}

// VoidedDocumentsResponse represents the response from SUNAT
type VoidedDocumentsResponse struct {
	Success         bool
	Message         string
	Ticket          string // Ticket number for async status checking
	ResponseXML     []byte
	Error           error
}


// GenerateVoidedDocumentsXML generates the XML for voided documents communication
func (c *SUNATClient) GenerateVoidedDocumentsXML(request *VoidedDocumentsRequest) ([]byte, error) {
	if len(request.Documents) == 0 {
		return nil, fmt.Errorf("no documents to void")
	}

	// Generate XML content based on SUNAT VoidedDocuments schema (following PHP example format)
	xmlContent := fmt.Sprintf(`<?xml version="1.0" encoding="ISO-8859-1" standalone="no"?>
<VoidedDocuments xmlns="urn:sunat:names:specification:ubl:peru:schema:xsd:VoidedDocuments-1"
xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
xmlns:ds="http://www.w3.org/2000/09/xmldsig#"
xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2"
xmlns:sac="urn:sunat:names:specification:ubl:peru:schema:xsd:SunatAggregateComponents-1"
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
<ext:UBLExtensions><ext:UBLExtension>
<ext:ExtensionContent>
</ext:ExtensionContent>
</ext:UBLExtension></ext:UBLExtensions>
<cbc:UBLVersionID>2.0</cbc:UBLVersionID>
<cbc:CustomizationID>1.0</cbc:CustomizationID>
<cbc:ID>%s</cbc:ID>
<cbc:ReferenceDate>%s</cbc:ReferenceDate>
<cbc:IssueDate>%s</cbc:IssueDate>
<cac:Signature>
<cbc:ID>IDSignKG</cbc:ID>
<cac:SignatoryParty>
<cac:PartyIdentification>
<cbc:ID>%s</cbc:ID>
</cac:PartyIdentification>
<cac:PartyName>
<cbc:Name><![CDATA[%s]]></cbc:Name>
</cac:PartyName>
</cac:SignatoryParty>
<cac:DigitalSignatureAttachment>
<cac:ExternalReference>
<cbc:URI>#signatureKG</cbc:URI>
</cac:ExternalReference>
</cac:DigitalSignatureAttachment>
</cac:Signature>
<cac:AccountingSupplierParty>
<cbc:CustomerAssignedAccountID>%s</cbc:CustomerAssignedAccountID>
<cbc:AdditionalAccountID>6</cbc:AdditionalAccountID>
<cac:Party>
<cac:PartyLegalEntity>
<cbc:RegistrationName><![CDATA[%s]]></cbc:RegistrationName>
</cac:PartyLegalEntity>
</cac:Party>
</cac:AccountingSupplierParty>`,
		request.SeriesNumber,
		request.ReferenceDate.Format("2006-01-02"),
		request.IssueDate.Format("2006-01-02"),
		request.RUC,
		utils.ValidateSpecialCharacters(request.CompanyName),
		request.RUC,
		utils.ValidateSpecialCharacters(request.CompanyName))

	// Add voided document lines
	for i, doc := range request.Documents {
		line := fmt.Sprintf(`
<sac:VoidedDocumentsLine>
<cbc:LineID>%d</cbc:LineID>
<cbc:DocumentTypeCode>%s</cbc:DocumentTypeCode>
<sac:DocumentSerialID>%s</sac:DocumentSerialID>
<sac:DocumentNumberID>%s</sac:DocumentNumberID>
<sac:VoidReasonDescription>%s</sac:VoidReasonDescription>
</sac:VoidedDocumentsLine>`,
			i+1,
			doc.DocumentTypeCode,
			doc.DocumentSeries,
			doc.DocumentNumber,
			utils.ValidateSpecialCharacters(doc.VoidedReason))
		xmlContent += line
	}

	xmlContent += `
</VoidedDocuments>`

	return []byte(xmlContent), nil
}

// SendVoidedDocuments sends voided documents communication to SUNAT
func (c *SUNATClient) SendVoidedDocuments(request *VoidedDocumentsRequest) (*VoidedDocumentsResponse, error) {
	// Validate request first
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Generate XML
	xmlContent, err := c.GenerateVoidedDocumentsXML(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate XML: %w", err)
	}

	// Sign XML if signer is available
	var signedXML []byte
	if c.signer != nil {
		signedXML, err = c.SignXML(xmlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to sign XML: %w", err)
		}
	} else {
		signedXML = xmlContent
	}

	// Create ZIP file
	zipData, zipName, err := c.createVoidedDocumentsZIP(signedXML, request.SeriesNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZIP: %w", err)
	}

	// Encode to base64
	zipB64 := base64.StdEncoding.EncodeToString(zipData)

	// Build SOAP envelope for sendSummary
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
    <ser:sendSummary>
      <fileName>%s</fileName>
      <contentFile>%s</contentFile>
    </ser:sendSummary>
  </soapenv:Body>
</soapenv:Envelope>`, c.RUC, c.Username, c.Password, zipName, zipB64)

	// Send HTTP request
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(soapBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "urn:sendSummary")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return c.parseVoidedDocumentsResponse(responseData)
}

// createVoidedDocumentsZIP creates a ZIP file for voided documents
func (c *SUNATClient) createVoidedDocumentsZIP(signedXML []byte, seriesNumber string) ([]byte, string, error) {
	xmlName := fmt.Sprintf("%s-%s.xml", c.RUC, seriesNumber)
	zipName := fmt.Sprintf("%s-%s.zip", c.RUC, seriesNumber)

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

// parseVoidedDocumentsResponse parses SUNAT's response for voided documents
func (c *SUNATClient) parseVoidedDocumentsResponse(responseData []byte) (*VoidedDocumentsResponse, error) {
	responseStr := string(responseData)
	response := &VoidedDocumentsResponse{
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

	// Check for successful response - sendSummary returns a ticket
	if strings.Contains(responseStr, "<br:sendSummaryResponse") {
		response.Success = true
		response.Message = "Comunicación de baja enviada exitosamente"

		// Extract ticket
		if start := strings.Index(responseStr, "<ticket>"); start != -1 {
			start += 8
			if end := strings.Index(responseStr[start:], "</ticket>"); end != -1 {
				response.Ticket = responseStr[start : start+end]
			}
		}

		return response, nil
	}

	response.Success = false
	response.Message = "Respuesta no reconocida de SUNAT"
	return response, nil
}

// GetVoidedDocumentsStatus checks the status of a voided documents communication using the ticket
func (c *SUNATClient) GetVoidedDocumentsStatus(ticket string) (*SUNATResponse, error) {
	// Build SOAP envelope for getStatus
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
    <ser:getStatus>
      <ticket>%s</ticket>
    </ser:getStatus>
  </soapenv:Body>
</soapenv:Envelope>`, c.RUC, c.Username, c.Password, ticket)

	// Send HTTP request
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(soapBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "urn:getStatus")

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

// Validate validates the voided documents request
func (req *VoidedDocumentsRequest) Validate() error {
	if req.RUC == "" {
		return fmt.Errorf("RUC is required")
	}

	if !utils.ValidateRUC(req.RUC) {
		return fmt.Errorf("invalid RUC format: %s", req.RUC)
	}

	if req.CompanyName == "" {
		return fmt.Errorf("company name is required")
	}

	if req.SeriesNumber == "" {
		return fmt.Errorf("series number is required")
	}

	if len(req.Documents) == 0 {
		return fmt.Errorf("at least one document is required")
	}

	// Validate each document
	for i, doc := range req.Documents {
		if err := doc.Validate(); err != nil {
			return fmt.Errorf("document %d: %w", i+1, err)
		}
	}

	return nil
}

// Validate validates a single voided document
func (doc *VoidedDocument) Validate() error {
	if doc.DocumentTypeCode == "" {
		return fmt.Errorf("document type code is required")
	}

	if !utils.ValidateDocumentType(doc.DocumentTypeCode) {
		return fmt.Errorf("invalid document type code: %s", doc.DocumentTypeCode)
	}

	if doc.DocumentSeries == "" {
		return fmt.Errorf("document series is required")
	}

	if !utils.ValidateDocumentSeries(doc.DocumentSeries) {
		return fmt.Errorf("invalid document series format: %s", doc.DocumentSeries)
	}

	if doc.DocumentNumber == "" {
		return fmt.Errorf("document number is required")
	}

	if !utils.ValidateDocumentNumber(doc.DocumentNumber) {
		return fmt.Errorf("invalid document number format: %s", doc.DocumentNumber)
	}

	if doc.VoidedReason == "" {
		return fmt.Errorf("voided reason is required")
	}

	return nil
}

// TicketStatusResponse represents the response from ticket status query
type TicketStatusResponse struct {
	Success           bool
	Message           string
	Ticket            string
	StatusCode        string      // SUNAT status code
	StatusDescription string      // SUNAT status description
	ProcessDate       time.Time   // Date when the document was processed
	ResponseXML       []byte      // Full SOAP response
	ApplicationResponse []byte    // CDR ZIP content if available
	Error             error
}

// GetTicketStatusDescription returns a human-readable description of the ticket status
func (r *TicketStatusResponse) GetTicketStatusDescription() string {
	switch r.StatusCode {
	case "0":
		return "Procesado correctamente"
	case "98":
		return "En proceso"
	case "99":
		return "Procesado con errores"
	default:
		return r.StatusDescription
	}
}

// IsProcessed returns true if the ticket has been processed (successfully or with errors)
func (r *TicketStatusResponse) IsProcessed() bool {
	return r.StatusCode == "0" || r.StatusCode == "99"
}

// IsSuccessful returns true if the ticket was processed successfully
func (r *TicketStatusResponse) IsSuccessful() bool {
	return r.StatusCode == "0"
}

// IsInProgress returns true if the ticket is still being processed
func (r *TicketStatusResponse) IsInProgress() bool {
	return r.StatusCode == "98"
}

// HasErrors returns true if the ticket was processed but with errors
func (r *TicketStatusResponse) HasErrors() bool {
	return r.StatusCode == "99"
}

// GetApplicationResponse returns the CDR response data if available
func (r *TicketStatusResponse) GetApplicationResponse() []byte {
	return r.ApplicationResponse
}

// HasApplicationResponse returns true if CDR response data is available
func (r *TicketStatusResponse) HasApplicationResponse() bool {
	return len(r.ApplicationResponse) > 0
}

// QueryVoidedDocumentsTicket queries the status of a voided documents communication ticket
// This is a more specific and enhanced version of GetVoidedDocumentsStatus
func (c *SUNATClient) QueryVoidedDocumentsTicket(ticket string) (*TicketStatusResponse, error) {
	if ticket == "" {
		return nil, fmt.Errorf("ticket number is required")
	}

	// Build SOAP envelope for getStatus
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
    <ser:getStatus>
      <ticket>%s</ticket>
    </ser:getStatus>
  </soapenv:Body>
</soapenv:Envelope>`, c.RUC, c.Username, c.Password, ticket)

	// Send HTTP request
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(soapBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "urn:getStatus")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return c.parseTicketStatusResponse(responseData, ticket)
}

// parseTicketStatusResponse parses SUNAT's response for ticket status queries
func (c *SUNATClient) parseTicketStatusResponse(responseData []byte, ticket string) (*TicketStatusResponse, error) {
	responseStr := string(responseData)
	response := &TicketStatusResponse{
		ResponseXML: responseData,
		Ticket:      ticket,
	}

	// Check for SOAP fault
	if strings.Contains(responseStr, "<soap-env:Fault") || strings.Contains(responseStr, "<soap:Fault") {
		response.Success = false

		// Extract fault string
		if start := strings.Index(responseStr, "<faultstring>"); start != -1 {
			start += 13
			if end := strings.Index(responseStr[start:], "</faultstring>"); end != -1 {
				response.Message = responseStr[start : start+end]
				// Decode HTML entities
				response.Message = strings.ReplaceAll(response.Message, "&#243;", "ó")
				response.Message = strings.ReplaceAll(response.Message, "&lt;", "<")
				response.Message = strings.ReplaceAll(response.Message, "&gt;", ">")
				response.Message = strings.ReplaceAll(response.Message, "&amp;", "&")
			}
		}

		return response, nil
	}

	// Check for successful response
	if strings.Contains(responseStr, "<br:getStatusResponse") || strings.Contains(responseStr, "getStatusResponse") {
		response.Success = true

		// Extract status code
		if start := strings.Index(responseStr, "<statusCode>"); start != -1 {
			start += 12
			if end := strings.Index(responseStr[start:], "</statusCode>"); end != -1 {
				response.StatusCode = responseStr[start : start+end]
			}
		}

		// Set status description based on code
		response.StatusDescription = response.GetTicketStatusDescription()

		// Extract content (CDR) if available and status is successful
		if response.StatusCode == "0" {
			if start := strings.Index(responseStr, "<content>"); start != -1 {
				start += 9
				if end := strings.Index(responseStr[start:], "</content>"); end != -1 {
					contentB64 := responseStr[start : start+end]
					if decodedContent, err := base64.StdEncoding.DecodeString(contentB64); err == nil {
						response.ApplicationResponse = decodedContent
					}
				}
			}
			response.Message = "Comunicación de baja procesada exitosamente"
		} else if response.StatusCode == "98" {
			response.Message = "Comunicación de baja en proceso de validación"
		} else if response.StatusCode == "99" {
			response.Message = "Comunicación de baja procesada con errores"
			// Try to extract error content for more details
			if start := strings.Index(responseStr, "<content>"); start != -1 {
				start += 9
				if end := strings.Index(responseStr[start:], "</content>"); end != -1 {
					contentB64 := responseStr[start : start+end]
					if decodedContent, err := base64.StdEncoding.DecodeString(contentB64); err == nil {
						response.ApplicationResponse = decodedContent
					}
				}
			}
		}

		return response, nil
	}

	response.Success = false
	response.Message = "Respuesta no reconocida de SUNAT para consulta de ticket"
	return response, nil
}

// WaitForTicketProcessing waits for a ticket to be processed, polling every interval
// Returns the final status response when processing is complete or timeout is reached
func (c *SUNATClient) WaitForTicketProcessing(ticket string, maxWaitTime time.Duration, pollInterval time.Duration) (*TicketStatusResponse, error) {
	if pollInterval <= 0 {
		pollInterval = 30 * time.Second // Default to 30 seconds
	}

	startTime := time.Now()

	for {
		response, err := c.QueryVoidedDocumentsTicket(ticket)
		if err != nil {
			return nil, fmt.Errorf("error querying ticket: %w", err)
		}

		// Return immediately if there's an error in the response
		if !response.Success {
			return response, nil
		}

		// Return if processing is complete (success or error)
		if response.IsProcessed() {
			return response, nil
		}

		// Check timeout
		if time.Since(startTime) >= maxWaitTime {
			response.Message = "Timeout esperando procesamiento del ticket"
			return response, nil
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}

// BatchQueryTickets queries multiple tickets and returns their status
func (c *SUNATClient) BatchQueryTickets(tickets []string) ([]*TicketStatusResponse, error) {
	if len(tickets) == 0 {
		return nil, fmt.Errorf("no tickets provided")
	}

	responses := make([]*TicketStatusResponse, 0, len(tickets))

	for _, ticket := range tickets {
		response, err := c.QueryVoidedDocumentsTicket(ticket)
		if err != nil {
			// Create error response for this ticket
			errorResponse := &TicketStatusResponse{
				Success: false,
				Ticket:  ticket,
				Message: fmt.Sprintf("Error querying ticket: %v", err),
				Error:   err,
			}
			responses = append(responses, errorResponse)
		} else {
			responses = append(responses, response)
		}

		// Small delay to avoid overwhelming SUNAT servers
		time.Sleep(100 * time.Millisecond)
	}

	return responses, nil
}

// GenerateVoidedDocumentsSeries generates a series number for voided documents
// Format: RA-YYYYMMDD-### where ### is a sequential number
func GenerateVoidedDocumentsSeries(referenceDate time.Time, sequential int) string {
	return fmt.Sprintf("RA-%s-%03d", referenceDate.Format("20060102"), sequential)
}