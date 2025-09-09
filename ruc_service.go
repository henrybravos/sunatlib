package sunatlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RUCBasicResponse represents the response from DeColecta basic RUC service
type RUCBasicResponse struct {
	Success bool `json:"success"`
	Data    *RUCBasicData `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// RUCBasicData contains basic company information
type RUCBasicData struct {
	RUC                   string `json:"numero_documento"`
	RazonSocial           string `json:"razon_social"`
	Estado                string `json:"estado"`
	Condicion             string `json:"condicion"`
	Direccion             string `json:"direccion"`
	Ubigeo                string `json:"ubigeo"`
	Distrito              string `json:"distrito"`
	Provincia             string `json:"provincia"`
	Departamento          string `json:"departamento"`
	EsAgenteRetencion     bool   `json:"es_agente_retencion"`
	EsBuenContribuyente   bool   `json:"es_buen_contribuyente"`
	ViaTipo               string `json:"via_tipo"`
	ViaNombre             string `json:"via_nombre"`
	ZonaCodigo            string `json:"zona_codigo"`
	ZonaTipo              string `json:"zona_tipo"`
	Numero                string `json:"numero"`
	Interior              string `json:"interior"`
}

// RUCFullResponse represents the response from DeColecta advanced RUC service
type RUCFullResponse struct {
	Success bool `json:"success"`
	Data    *RUCFullData `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// RUCFullData contains complete company information
type RUCFullData struct {
	RUCBasicData
	ActividadEconomica string `json:"actividad_economica"`
	NumeroTrabajadores string `json:"numero_trabajadores"`
	TipoFacturacion    string `json:"tipo_facturacion"`
	TipoContabilidad   string `json:"tipo_contabilidad"`
	ComercioExterior   string `json:"comercio_exterior"`
	FechaInscripcion   string `json:"fecha_inscripcion"`
}

// RUCService handles RUC consultation operations
type RUCService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewRUCService creates a new RUC service instance
func NewRUCService(apiKey string) *RUCService {
	return &RUCService{
		APIKey:  apiKey,
		BaseURL: "https://api.decolecta.com/v1/sunat/ruc",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ConsultBasic performs a basic RUC consultation
func (rs *RUCService) ConsultBasic(ruc string) (*RUCBasicResponse, error) {
	if len(ruc) != 11 {
		return &RUCBasicResponse{
			Success: false,
			Message: "RUC debe tener 11 dígitos",
		}, fmt.Errorf("RUC inválido: debe tener 11 dígitos")
	}

	url := fmt.Sprintf("%s?numero=%s", rs.BaseURL, ruc)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rs.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SunatLib/1.0")

	resp, err := rs.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}


	if resp.StatusCode != http.StatusOK {
		return &RUCBasicResponse{
			Success: false,
			Message: fmt.Sprintf("Error HTTP %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	// Try to parse directly as the data structure instead of wrapped response
	var directData RUCBasicData
	if err := json.Unmarshal(body, &directData); err == nil && directData.RUC != "" {
		return &RUCBasicResponse{
			Success: true,
			Data:    &directData,
			Message: "Consulta exitosa",
		}, nil
	}

	// Try original parsing
	var result RUCBasicResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return &RUCBasicResponse{
			Success: false,
			Message: fmt.Sprintf("Error parseando JSON: %v. Raw response: %s", err, string(body)),
		}, fmt.Errorf("error parseando JSON: %w", err)
	}

	return &result, nil
}

// ConsultFull performs a complete RUC consultation
func (rs *RUCService) ConsultFull(ruc string) (*RUCFullResponse, error) {
	if len(ruc) != 11 {
		return &RUCFullResponse{
			Success: false,
			Message: "RUC debe tener 11 dígitos",
		}, fmt.Errorf("RUC inválido: debe tener 11 dígitos")
	}

	url := fmt.Sprintf("%s/full?numero=%s", rs.BaseURL, ruc)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rs.APIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "SunatLib/1.0")

	resp, err := rs.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &RUCFullResponse{
			Success: false,
			Message: fmt.Sprintf("Error HTTP %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	// Try to parse directly as the data structure instead of wrapped response
	var directData RUCFullData
	if err := json.Unmarshal(body, &directData); err == nil && directData.RUC != "" {
		return &RUCFullResponse{
			Success: true,
			Data:    &directData,
			Message: "Consulta exitosa",
		}, nil
	}

	// Try original parsing
	var result RUCFullResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return &RUCFullResponse{
			Success: false,
			Message: fmt.Sprintf("Error parseando JSON: %v. Raw response: %s", err, string(body)),
		}, fmt.Errorf("error parseando JSON: %w", err)
	}

	return &result, nil
}

// IsValidRUC validates if a RUC number has the correct format
func IsValidRUC(ruc string) bool {
	if len(ruc) != 11 {
		return false
	}
	
	// Verify it's all digits
	for _, char := range ruc {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	// Basic RUC validation (first digit should be 1 or 2)
	firstDigit := ruc[0]
	return firstDigit == '1' || firstDigit == '2'
}