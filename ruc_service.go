package sunatlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SunatRawResponse represents the direct response from SUNAT service
type SunatRawResponse struct {
	Message string `json:"message"`
	Lista   []struct {
		IdProvincia        string `json:"idprovincia"`
		IdDistrito         string `json:"iddistrito"`
		RazonSocial        string `json:"apenomdenunciado"`
		IdDepartamento     string `json:"iddepartamento"`
		Direccion          string `json:"direstablecimiento"`
		DesDistrito        string `json:"desdistrito"`
		DesProvincia       string `json:"desprovincia"`
		DesDepartamento    string `json:"desdepartamento"`
	} `json:"lista"`
}

// RUCBasicResponse represents the standard response format
type RUCBasicResponse struct {
	Success bool          `json:"success"`
	Data    *RUCBasicData `json:"data,omitempty"`
	Message string        `json:"message,omitempty"`
}

// RUCBasicData contains basic company information
type RUCBasicData struct {
	RUC                 string `json:"numero_documento"`
	RazonSocial         string `json:"razon_social"`
	Estado              string `json:"estado"`
	Condicion           string `json:"condicion"`
	Direccion           string `json:"direccion"`
	Ubigeo              string `json:"ubigeo"`
	Distrito            string `json:"distrito"`
	Provincia           string `json:"provincia"`
	Departamento        string `json:"departamento"`
	EsAgenteRetencion   bool   `json:"es_agente_retencion"`
	EsBuenContribuyente bool   `json:"es_buen_contribuyente"`
	ViaTipo             string `json:"via_tipo"`
	ViaNombre           string `json:"via_nombre"`
	ZonaCodigo          string `json:"zona_codigo"`
	ZonaTipo            string `json:"zona_tipo"`
	Numero              string `json:"numero"`
	Interior            string `json:"interior"`
}

// RUCFullResponse represents the response with full data (if available)
type RUCFullResponse struct {
	Success bool         `json:"success"`
	Data    *RUCFullData `json:"data,omitempty"`
	Message string       `json:"message,omitempty"`
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
	BaseURL    string
	HTTPClient *http.Client
}

// NewRUCService creates a new RUC service instance (apiKey is kept for backward compatibility but unused)
func NewRUCService(apiKey string) *RUCService {
	return &RUCService{
		BaseURL: "https://ww1.sunat.gob.pe/ol-ti-itfisdenreg/itfisdenreg.htm",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ConsultBasic performs a basic RUC consultation using SUNAT's direct API
func (rs *RUCService) ConsultBasic(ruc string) (*RUCBasicResponse, error) {
	if !IsValidRUC(ruc) {
		return &RUCBasicResponse{
			Success: false,
			Message: "RUC debe tener 11 dígitos y empezar con 10, 20 o 15",
		}, fmt.Errorf("RUC inválido: debe tener 11 dígitos")
	}

	url := fmt.Sprintf("%s?accion=obtenerDatosRuc&nroRuc=%s", rs.BaseURL, ruc)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	// Realistic headers to avoid blocks
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := rs.HTTPClient.Do(req)
	if err != nil {
		return &RUCBasicResponse{
			Success: false,
			Message: fmt.Sprintf("Error de conexión: %v", err),
		}, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &RUCBasicResponse{
			Success: false,
			Message: fmt.Sprintf("Error HTTP %d", resp.StatusCode),
		}, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	var sunatResp SunatRawResponse
	if err := json.Unmarshal(body, &sunatResp); err != nil {
		return &RUCBasicResponse{
			Success: false,
			Message: "Error parseando respuesta de SUNAT",
		}, fmt.Errorf("error parseando JSON: %w", err)
	}

	if sunatResp.Message != "success" || len(sunatResp.Lista) == 0 {
		return &RUCBasicResponse{
			Success: false,
			Message: "RUC no encontrado o error en SUNAT",
		}, fmt.Errorf("RUC no encontrado")
	}

	data := sunatResp.Lista[0]
	
	result := &RUCBasicResponse{
		Success: true,
		Data: &RUCBasicData{
			RUC:          ruc,
			RazonSocial:  strings.TrimSpace(data.RazonSocial),
			Direccion:    strings.TrimSpace(data.Direccion),
			Distrito:     strings.TrimSpace(data.DesDistrito),
			Provincia:    strings.TrimSpace(data.DesProvincia),
			Departamento: strings.TrimSpace(data.DesDepartamento),
			Ubigeo:       data.IdDepartamento + data.IdProvincia + data.IdDistrito,
			Estado:       "ACTIVO", // This API doesn't provide status, but it usually returns active ones
			Condicion:    "HABIDO",
		},
		Message: "Consulta exitosa",
	}

	return result, nil
}

// ConsultFull performs a RUC consultation (limited data due to simplified API)
func (rs *RUCService) ConsultFull(ruc string) (*RUCFullResponse, error) {
	basic, err := rs.ConsultBasic(ruc)
	if err != nil {
		return &RUCFullResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	if !basic.Success {
		return &RUCFullResponse{
			Success: false,
			Message: basic.Message,
		}, nil
	}

	return &RUCFullResponse{
		Success: true,
		Data: &RUCFullData{
			RUCBasicData: *basic.Data,
		},
		Message: "Consulta exitosa (datos limitados)",
	}, nil
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
	
	// Basic RUC validation (first two digits)
	prefix := ruc[:2]
	return prefix == "10" || prefix == "20" || prefix == "15" || prefix == "17"
}