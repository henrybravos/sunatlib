// Package sunatlib provides document validation functionality for SUNAT Peru
package sunatlib

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ValidationClient handles SUNAT document validation with master credentials
type ValidationClient struct {
	masterRUC      string
	masterUsername string
	masterPassword string
	endpoint       string
	httpClient     *http.Client
}

// NewValidationClient creates a new SUNAT validation client with master credentials
func NewValidationClient(masterRUC, masterUsername, masterPassword string) *ValidationClient {
	return &ValidationClient{
		masterRUC:      masterRUC,
		masterUsername: masterUsername,
		masterPassword: masterPassword,
		endpoint:       "https://e-factura.sunat.gob.pe/ol-it-wsconsvalidcpe/billValidService",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}


// ValidateDocument validates a document with SUNAT using master credentials
func (vc *ValidationClient) ValidateDocument(params *ValidationParams) (*ValidationResult, error) {
	// Format parameters for SUNAT
	formattedParams, err := vc.formatValidationParams(params)
	if err != nil {
		return nil, fmt.Errorf("error formatting parameters: %w", err)
	}

	// Build SOAP request
	soapXML := vc.buildSOAPRequest(formattedParams)

	// Log the request XML for debugging
	fmt.Printf("📤 [SUNATLIB] Request XML being sent to SUNAT:\n%s\n", soapXML)

	// Execute request
	result, err := vc.executeValidationRequest(soapXML, formattedParams)
	if err != nil {
		return nil, fmt.Errorf("validation request failed: %w", err)
	}

	return result, nil
}

// ValidationParams contains the parameters for document validation
type ValidationParams struct {
	IssuerRUC           string  // RUC of the document issuer
	DocumentType        string  // Document type code (01=Invoice, 03=Receipt, etc.)
	SeriesNumber        string  // Series of the document (e.g., F001)
	DocumentNumber      string  // Document number (e.g., 00000001)
	RecipientDocType    string  // Recipient document type ("-" for default)
	RecipientDocNumber  string  // Recipient document number ("" for default)
	IssueDate           string  // Issue date in YYYY-MM-DD format
	TotalAmount         float64 // Total amount of the document
	AuthorizationNumber string  // Authorization number (usually empty)
}

// ValidationResult contains the result of SUNAT validation
type ValidationResult struct {
	Success       bool   `json:"success"`
	IsValid       bool   `json:"is_valid"`
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message"`
	ErrorDetails  string `json:"error_details,omitempty"`
	State         string `json:"state"` // VALIDO, NO_INFORMADO, ANULADO, RECHAZADO
	ResponseXML   string `json:"response_xml,omitempty"` // Raw XML response from SUNAT
}

// formattedValidationParams holds formatted parameters for SUNAT request
type formattedValidationParams struct {
	RucEmisor             string
	TipoCDP               string
	SerieCDP              string
	NumeroCDP             string
	TipoDocIdReceptor     string
	NumeroDocIdReceptor   string
	FechaEmision          string
	ImporteTotal          string
	NroAutorizacion       string
	FullUsername          string
	Password              string
}

// formatValidationParams formats validation parameters according to SUNAT requirements
func (vc *ValidationClient) formatValidationParams(params *ValidationParams) (*formattedValidationParams, error) {
	// Validate required fields
	if params.IssuerRUC == "" {
		return nil, fmt.Errorf("issuer RUC cannot be empty")
	}
	if params.SeriesNumber == "" {
		return nil, fmt.Errorf("series number cannot be empty")
	}
	if params.DocumentNumber == "" {
		return nil, fmt.Errorf("document number cannot be empty")
	}

	// Format issue date from YYYY-MM-DD to DD/MM/YYYY
	formattedDate, err := vc.formatDateForSUNAT(params.IssueDate)
	if err != nil {
		return nil, fmt.Errorf("invalid issue date format: %w", err)
	}

	// Format total amount
	formattedAmount := fmt.Sprintf("%.2f", params.TotalAmount)

	// Set default values for recipient if not provided
	recipientDocType := params.RecipientDocType
	if recipientDocType == "" {
		recipientDocType = "-"
	}

	recipientDocNumber := params.RecipientDocNumber
	if recipientDocNumber == "" {
		recipientDocNumber = ""
	}

	return &formattedValidationParams{
		RucEmisor:           params.IssuerRUC,
		TipoCDP:             params.DocumentType,
		SerieCDP:            params.SeriesNumber,
		NumeroCDP:           params.DocumentNumber,
		TipoDocIdReceptor:   recipientDocType,
		NumeroDocIdReceptor: recipientDocNumber,
		FechaEmision:        formattedDate,
		ImporteTotal:        formattedAmount,
		NroAutorizacion:     params.AuthorizationNumber,
		FullUsername:        fmt.Sprintf("%s%s", vc.masterRUC, vc.masterUsername),
		Password:            vc.masterPassword,
	}, nil
}

// formatDateForSUNAT converts YYYY-MM-DD to DD/MM/YYYY format
func (vc *ValidationClient) formatDateForSUNAT(dateStr string) (string, error) {
	if dateStr == "" {
		return "", fmt.Errorf("date cannot be empty")
	}

	// Parse the date in YYYY-MM-DD format
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("cannot parse date '%s': %w", dateStr, err)
	}

	// Format as DD/MM/YYYY
	return parsedDate.Format("02/01/2006"), nil
}

// buildSOAPRequest creates the SOAP XML request for validation
func (vc *ValidationClient) buildSOAPRequest(params *formattedValidationParams) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:ser="http://service.sunat.gob.pe"
                  xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
    <soapenv:Header>
        <wsse:Security>
            <wsse:UsernameToken>
                <wsse:Username>%s</wsse:Username>
                <wsse:Password>%s</wsse:Password>
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
		params.FullUsername,
		params.Password,
		params.RucEmisor,
		params.TipoCDP,
		params.SerieCDP,
		params.NumeroCDP,
		params.TipoDocIdReceptor,
		params.NumeroDocIdReceptor,
		params.FechaEmision,
		params.ImporteTotal,
		params.NroAutorizacion)
}

// executeValidationRequest executes the SOAP request to SUNAT
func (vc *ValidationClient) executeValidationRequest(soapXML string, params *formattedValidationParams) (*ValidationResult, error) {
	// Create HTTP request
	req, err := http.NewRequest("POST", vc.endpoint, strings.NewReader(soapXML))
	if err != nil {
		return nil, fmt.Errorf("error creating SOAP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "")
	req.Header.Set("User-Agent", "SUNATLib/1.1.0")

	// Execute request
	resp, err := vc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing SOAP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading SOAP response: %w", err)
	}

	// Parse response
	result := vc.parseValidationResponse(string(responseBody), resp.StatusCode)

	// Store raw XML response for logging/debugging
	result.ResponseXML = string(responseBody)

	return result, nil
}

// parseValidationResponse parses the SOAP response from SUNAT
func (vc *ValidationClient) parseValidationResponse(responseBody string, httpStatusCode int) *ValidationResult {
	result := &ValidationResult{
		Success:       httpStatusCode == 200,
		IsValid:       false,
		StatusCode:    "UNKNOWN",
		StatusMessage: "Unable to parse response",
		State:         "UNKNOWN",
	}

	// Extract status code
	if strings.Contains(responseBody, "<statusCode>") {
		start := strings.Index(responseBody, "<statusCode>") + 12
		end := strings.Index(responseBody[start:], "</statusCode>")
		if end != -1 {
			result.StatusCode = responseBody[start : start+end]
		}
	}

	// Extract status message
	if strings.Contains(responseBody, "<statusMessage>") {
		start := strings.Index(responseBody, "<statusMessage>") + 15
		end := strings.Index(responseBody[start:], "</statusMessage>")
		if end != -1 {
			result.StatusMessage = responseBody[start : start+end]
		}
	}

	// Determine validity based on message content (ignoring status codes)
	message := result.StatusMessage

	// Check message content patterns
	aux1 := strings.Contains(message, "no existe en los registros de SUNAT")
	aux2 := strings.Contains(message, "no ha sido informada")
	aux3 := strings.Contains(message, "BAJA")
	aux4 := strings.Contains(message, "RECHAZADO")
	aux5 := strings.Contains(message, "es un comprobante de pago válido")
	aux6 := strings.Contains(message, "ha sido informada")
	aux7 := strings.Contains(message, "rechazada")
	aux8 := strings.Contains(message, "AUTORIZADO (Con autorización de imprenta)")

	// Determine state based on message content
	state := "NO_INFORMADO" // default: no informado
	if aux1 || aux2 {
		state = "NO_INFORMADO" // No informado
	}
	if aux3 {
		state = "ANULADO" // Anulado/Baja
	}
	if aux4 || aux7 {
		state = "RECHAZADO" // Rechazado
	}
	if aux5 || (aux6 && !aux2) || aux8 {
		state = "VALIDO" // Válido
	}

	// Set validity and error details based on state
	switch state {
	case "VALIDO":
		result.IsValid = true
		result.ErrorDetails = "Documento válido en SUNAT"
	case "ANULADO":
		result.IsValid = false
		result.ErrorDetails = "Documento anulado o dado de baja"
	case "RECHAZADO":
		result.IsValid = false
		result.ErrorDetails = "Documento rechazado por SUNAT"
	case "NO_INFORMADO":
		result.IsValid = false
		result.ErrorDetails = "Documento no informado a SUNAT"
	default:
		result.IsValid = false
		result.ErrorDetails = "Estado no determinado"
	}

	result.State = state
	return result
}

// ValidateInvoice is a convenience method for validating invoices
func (vc *ValidationClient) ValidateInvoice(issuerRUC, seriesNumber, documentNumber, issueDate string, totalAmount float64) (*ValidationResult, error) {
	params := &ValidationParams{
		IssuerRUC:           issuerRUC,
		DocumentType:        "01", // Invoice
		SeriesNumber:        seriesNumber,
		DocumentNumber:      documentNumber,
		RecipientDocType:    "-",
		RecipientDocNumber:  "",
		IssueDate:           issueDate,
		TotalAmount:         totalAmount,
		AuthorizationNumber: "",
	}
	return vc.ValidateDocument(params)
}

// ValidateReceipt is a convenience method for validating receipts
func (vc *ValidationClient) ValidateReceipt(issuerRUC, seriesNumber, documentNumber, issueDate string, totalAmount float64) (*ValidationResult, error) {
	params := &ValidationParams{
		IssuerRUC:           issuerRUC,
		DocumentType:        "03", // Receipt
		SeriesNumber:        seriesNumber,
		DocumentNumber:      documentNumber,
		RecipientDocType:    "-",
		RecipientDocNumber:  "",
		IssueDate:           issueDate,
		TotalAmount:         totalAmount,
		AuthorizationNumber: "",
	}
	return vc.ValidateDocument(params)
}