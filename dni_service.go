package sunatlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DNIResponse represents the response from EsSalud DNI validation service
type DNIResponse struct {
	Success       bool      `json:"success"`
	Data          *DNIData  `json:"data,omitempty"`
	Message       string    `json:"message,omitempty"`
}

// DNIData contains personal information from RENIEC
type DNIData struct {
	DNI           string `json:"dni"`
	NombreCompleto string `json:"nombre_completo"`
	Nombres       string `json:"nombres"`
	ApellidoPaterno string `json:"apellido_paterno"`
	ApellidoMaterno string `json:"apellido_materno"`
	FechaNacimiento string `json:"fecha_nacimiento,omitempty"`
	Sexo          string `json:"sexo,omitempty"`
	EstadoCivil   string `json:"estado_civil,omitempty"`
}

// EsSaludResponse represents the raw response from EsSalud service
type EsSaludResponse struct {
	Datos     string `json:"datos"`     // Nombre completo
	Apellidos string `json:"apellidos"` // Apellidos completos
	Nombres   string `json:"nombres"`   // Solo nombres
	// Campos alternativos por si cambia el formato
	NumeroDocumento string `json:"numeroDocumento"`
	TipoDocumento   string `json:"tipoDocumento"`
	ApellidoPaterno string `json:"apellidoPaterno"`
	ApellidoMaterno string `json:"apellidoMaterno"`
	NombreCompleto  string `json:"nombreCompleto"`
}

// DNIService handles DNI consultation operations
type DNIService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewDNIService creates a new DNI service instance
func NewDNIService() *DNIService {
	return &DNIService{
		BaseURL: "https://viva.essalud.gob.pe/viva/validar-ws-reniec",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ConsultDNI performs a DNI consultation using EsSalud service
func (ds *DNIService) ConsultDNI(dni string) (*DNIResponse, error) {
	if !IsValidDNI(dni) {
		return &DNIResponse{
			Success: false,
			Message: "DNI debe tener 8 dígitos",
		}, fmt.Errorf("DNI inválido: debe tener 8 dígitos")
	}

	// EsSalud uses tipoDoc=01 for DNI
	url := fmt.Sprintf("%s?numero=%s&tipoDoc=01", ds.BaseURL, dni)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	// Set realistic headers to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")
	req.Header.Set("Referer", "https://viva.essalud.gob.pe/")

	resp, err := ds.HTTPClient.Do(req)
	if err != nil {
		return &DNIResponse{
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
		return &DNIResponse{
			Success: false,
			Message: fmt.Sprintf("Error HTTP %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	// Try to parse as EsSalud response format
	var essaludResp EsSaludResponse
	if err := json.Unmarshal(body, &essaludResp); err != nil {
		return &DNIResponse{
			Success: false,
			Message: "Error parseando respuesta del servicio",
		}, fmt.Errorf("error parseando JSON: %w", err)
	}

	// Check if we got valid data (either new format or old format)
	if essaludResp.Datos == "" && essaludResp.NombreCompleto == "" {
		return &DNIResponse{
			Success: false,
			Message: "DNI no encontrado o inválido",
		}, fmt.Errorf("DNI no encontrado")
	}

	// Convert to our standard format
	result := &DNIResponse{
		Success: true,
		Data: &DNIData{
			DNI:             dni, // Use the DNI we queried since it's not in response
			NombreCompleto:  essaludResp.Datos,    // "datos" field has full name
			Nombres:         essaludResp.Nombres,   // "nombres" field
			ApellidoPaterno: "",                    // Not available in this format
			ApellidoMaterno: "",                    // Not available in this format
		},
		Message: "Consulta exitosa",
	}

	// If we have apellidos field, try to split it
	if essaludResp.Apellidos != "" {
		// Simple split by space to get paterno/materno
		apellidosParts := strings.Fields(essaludResp.Apellidos)
		if len(apellidosParts) >= 1 {
			result.Data.ApellidoPaterno = apellidosParts[0]
		}
		if len(apellidosParts) >= 2 {
			result.Data.ApellidoMaterno = apellidosParts[1]
		}
	}

	return result, nil
}

// ConsultCE performs a Carnet de Extranjería consultation
func (ds *DNIService) ConsultCE(ce string) (*DNIResponse, error) {
	if len(ce) < 9 || len(ce) > 12 {
		return &DNIResponse{
			Success: false,
			Message: "Carnet de Extranjería debe tener entre 9 y 12 caracteres",
		}, fmt.Errorf("CE inválido: longitud incorrecta")
	}

	// EsSalud uses tipoDoc=04 for Carnet de Extranjería
	url := fmt.Sprintf("%s?numero=%s&tipoDoc=04", ds.BaseURL, ce)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "es-ES,es;q=0.9")
	req.Header.Set("Referer", "https://viva.essalud.gob.pe/")

	resp, err := ds.HTTPClient.Do(req)
	if err != nil {
		return &DNIResponse{
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
		return &DNIResponse{
			Success: false,
			Message: fmt.Sprintf("Error HTTP %d", resp.StatusCode),
		}, fmt.Errorf("error HTTP %d", resp.StatusCode)
	}

	var essaludResp EsSaludResponse
	if err := json.Unmarshal(body, &essaludResp); err != nil {
		return &DNIResponse{
			Success: false,
			Message: "Error parseando respuesta del servicio",
		}, fmt.Errorf("error parseando JSON: %w", err)
	}

	if essaludResp.NumeroDocumento == "" || essaludResp.NombreCompleto == "" {
		return &DNIResponse{
			Success: false,
			Message: "Carnet de Extranjería no encontrado o inválido",
		}, fmt.Errorf("CE no encontrado")
	}

	result := &DNIResponse{
		Success: true,
		Data: &DNIData{
			DNI:             essaludResp.NumeroDocumento,
			NombreCompleto:  essaludResp.NombreCompleto,
			Nombres:         essaludResp.Nombres,
			ApellidoPaterno: essaludResp.ApellidoPaterno,
			ApellidoMaterno: essaludResp.ApellidoMaterno,
		},
		Message: "Consulta exitosa",
	}

	return result, nil
}

// IsValidDNI validates if a DNI number has the correct format
func IsValidDNI(dni string) bool {
	if len(dni) != 8 {
		return false
	}
	
	// Verify it's all digits
	for _, char := range dni {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}

// IsValidCE validates if a CE number has the correct format
func IsValidCE(ce string) bool {
	if len(ce) < 9 || len(ce) > 12 {
		return false
	}
	
	// CE can contain letters and numbers
	return true
}