// Package sunatlib provides functionality for SUNAT electronic document validation
package sunatlib

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DocumentValidationClient handles document validation requests to SUNAT
type DocumentValidationClient struct {
	RUC      string
	Username string
	Password string
	Endpoint string
	Client   *http.Client
}

// ValidationRequest represents a document validation request
type ValidationRequest struct {
	RUC                    string // Issuer RUC
	DocumentType          string // Document type code
	Series                string // Document series
	Number                string // Document number
	IssueDate             string // Issue date (DD/MM/YYYY)
	TotalAmount           string // Total amount
	RecipientDocumentType string // Recipient document type (optional, use "-" for empty)
	RecipientDocument     string // Recipient document number (optional)
	AuthorizationNumber   string // Authorization number (optional)
}

// ValidationResponse represents the response from SUNAT validation service
type ValidationResponse struct {
	Success       bool
	IsValid       bool
	StatusMessage string
	ErrorMessage  string
	ResponseXML   []byte
}

// ValidationSOAPResponse represents the SOAP response structure
type ValidationSOAPResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		ValidaCDPResponse struct {
			StatusCode    string `xml:"statusCode"`
			StatusMessage string `xml:"statusMessage"`
			CDPValidated  string `xml:"cdpvalidado"`
		} `xml:"validaCDPcriteriosResponse"`
		Fault struct {
			FaultCode   string `xml:"faultcode"`
			FaultString string `xml:"faultstring"`
		} `xml:"Fault"`
	} `xml:"Body"`
}

// NewDocumentValidationClientWithCredentials creates a new document validation client with SUNAT credentials (PRODUCTION)
func NewDocumentValidationClientWithCredentials(ruc, username, password string) *DocumentValidationClient {
	return &DocumentValidationClient{
		RUC:      ruc,
		Username: username,
		Password: password,
		Endpoint: GetValidationServiceEndpoint(Production),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewDocumentValidationClientBeta creates a new document validation client for BETA testing
func NewDocumentValidationClientBeta(ruc, username, password string) *DocumentValidationClient {
	return &DocumentValidationClient{
		RUC:      ruc,
		Username: username,
		Password: password,
		Endpoint: GetValidationServiceEndpoint(Beta),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ValidateDocument validates an electronic document with SUNAT using SOAP
func (c *DocumentValidationClient) ValidateDocument(req *ValidationRequest) (*ValidationResponse, error) {
	// Set default values for optional fields
	recipientDocType := req.RecipientDocumentType
	if recipientDocType == "" {
		recipientDocType = "-"
	}

	recipientDoc := req.RecipientDocument
	authNumber := req.AuthorizationNumber

	// Build SOAP envelope based on the PHP example
	soapBody := fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"
xmlns:ser="http://service.sunat.gob.pe"
xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
<soapenv:Header>
<wsse:Security>
<wsse:UsernameToken>
<wsse:Username>%s%s</wsse:Username>
<wsse:Password><![CDATA[%s]]></wsse:Password>
</wsse:UsernameToken>
</wsse:Security>
</soapenv:Header>
<soapenv:Body>
<ser:validaCDPcriterios>
<rucEmisor>%s</rucEmisor>
<tipoCDP>%s</tipoCDP>
<serieCDP>%s</serieCDP>
<numeroCDP>%s</numeroCDP>
<tipoDocIdReceptor>%s</tipoDocIdReceptor>
<numeroDocIdReceptor>%s</numeroDocIdReceptor>
<fechaEmision>%s</fechaEmision>
<importeTotal>%s</importeTotal>
<nroAutorizacion>%s</nroAutorizacion>
</ser:validaCDPcriterios>
</soapenv:Body>
</soapenv:Envelope>`,
		c.RUC, c.Username, c.Password,
		req.RUC,
		req.DocumentType,
		req.Series,
		req.Number,
		recipientDocType,
		recipientDoc,
		req.IssueDate,
		req.TotalAmount,
		authNumber)

	// Send HTTP request
	httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewBuffer([]byte(soapBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpReq.Header.Set("Accept", "text/xml")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Pragma", "no-cache")
	httpReq.Header.Set("SOAPAction", "")
	httpReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(soapBody)))

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return c.parseValidationResponse(responseData, resp.StatusCode)
}

// parseValidationResponse parses the SOAP response from SUNAT
func (c *DocumentValidationClient) parseValidationResponse(responseData []byte, httpCode int) (*ValidationResponse, error) {
	response := &ValidationResponse{
		ResponseXML: responseData,
	}

	if httpCode != 200 {
		response.Success = false
		response.ErrorMessage = "Se ha perdido la comunicación con la SUNAT"
		return response, nil
	}

	responseStr := string(responseData)

	// Check for SOAP fault first
	if strings.Contains(responseStr, "<soap-env:Fault") || strings.Contains(responseStr, "<faultstring>") {
		response.Success = false

		// Extract fault string
		if start := strings.Index(responseStr, "<faultstring>"); start != -1 {
			start += 13
			if end := strings.Index(responseStr[start:], "</faultstring>"); end != -1 {
				response.ErrorMessage = responseStr[start : start+end]
			}
		}

		return response, nil
	}

	// Parse XML response
	var soapResp ValidationSOAPResponse
	err := xml.Unmarshal(responseData, &soapResp)
	if err != nil {
		// Try to extract manually if XML parsing fails
		if strings.Contains(responseStr, "<cdpvalidado>") {
			response.Success = true
			response.IsValid = true

			// Extract status message
			if start := strings.Index(responseStr, "<statusMessage>"); start != -1 {
				start += 15
				if end := strings.Index(responseStr[start:], "</statusMessage>"); end != -1 {
					response.StatusMessage = responseStr[start : start+end]
				}
			}
		} else {
			response.Success = false
			response.ErrorMessage = "Error parsing SUNAT response"
		}

		return response, nil
	}

	// Check if we have a fault in the parsed response
	if soapResp.Body.Fault.FaultCode != "" {
		response.Success = false
		response.ErrorMessage = soapResp.Body.Fault.FaultString
		return response, nil
	}

	// Check for valid response
	if soapResp.Body.ValidaCDPResponse.CDPValidated != "" {
		response.Success = true
		response.IsValid = true
		response.StatusMessage = soapResp.Body.ValidaCDPResponse.StatusMessage
	} else {
		response.Success = false
		response.ErrorMessage = "Documento no encontrado o inválido"
	}

	return response, nil
}

// ValidateInvoice validates an electronic invoice
func (c *DocumentValidationClient) ValidateInvoice(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error) {
	req := &ValidationRequest{
		RUC:          ruc,
		DocumentType: "01", // Invoice
		Series:       series,
		Number:       number,
		IssueDate:    issueDate,
		TotalAmount:  totalAmount,
	}

	return c.ValidateDocument(req)
}

// ValidateReceipt validates an electronic receipt (boleta)
func (c *DocumentValidationClient) ValidateReceipt(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error) {
	req := &ValidationRequest{
		RUC:          ruc,
		DocumentType: "03", // Receipt
		Series:       series,
		Number:       number,
		IssueDate:    issueDate,
		TotalAmount:  totalAmount,
	}

	return c.ValidateDocument(req)
}

// ValidateCreditNote validates an electronic credit note
func (c *DocumentValidationClient) ValidateCreditNote(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error) {
	req := &ValidationRequest{
		RUC:          ruc,
		DocumentType: "07", // Credit Note
		Series:       series,
		Number:       number,
		IssueDate:    issueDate,
		TotalAmount:  totalAmount,
	}

	return c.ValidateDocument(req)
}

// ValidateDebitNote validates an electronic debit note
func (c *DocumentValidationClient) ValidateDebitNote(ruc, series, number, issueDate, totalAmount string) (*ValidationResponse, error) {
	req := &ValidationRequest{
		RUC:          ruc,
		DocumentType: "08", // Debit Note
		Series:       series,
		Number:       number,
		IssueDate:    issueDate,
		TotalAmount:  totalAmount,
	}

	return c.ValidateDocument(req)
}

// CheckDocumentStatus checks the status of a document using basic parameters
func (c *DocumentValidationClient) CheckDocumentStatus(ruc, documentType, series, number string) (*ValidationResponse, error) {
	req := &ValidationRequest{
		RUC:          ruc,
		DocumentType: documentType,
		Series:       series,
		Number:       number,
	}

	return c.ValidateDocument(req)
}

// IsDocumentValid returns true if the document is valid
func (vr *ValidationResponse) IsDocumentValid() bool {
	return vr.Success && vr.IsValid
}

// GetStatusDescription returns a human-readable status description
func (vr *ValidationResponse) GetStatusDescription() string {
	if vr.StatusMessage != "" {
		return vr.StatusMessage
	}
	if vr.ErrorMessage != "" {
		return vr.ErrorMessage
	}
	return "Sin información de estado disponible"
}

// HasError returns true if there was an error in validation
func (vr *ValidationResponse) HasError() bool {
	return !vr.Success || vr.ErrorMessage != ""
}

// GetErrorMessage returns the error message if any
func (vr *ValidationResponse) GetErrorMessage() string {
	return vr.ErrorMessage
}