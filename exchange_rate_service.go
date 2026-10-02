package sunatlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ExchangeRate represents the exchange rate structure
type ExchangeRate struct {
	Date     string  `json:"fecha"`     // Format: YYYY-MM-DD
	Buy      float64 `json:"compra"`    // Compra
	Sell     float64 `json:"venta"`     // Venta
	Currency string  `json:"moneda"`    // USD
	Origin   string  `json:"origen"`    // SUNAT, BCRP, etc.
}

// ExchangeRateResponse represents the outcome of an exchange rate query
type ExchangeRateResponse struct {
	Success bool          `json:"success"`
	Data    *ExchangeRate `json:"data,omitempty"`
	Message string        `json:"message,omitempty"`
}

// RawSunatExchangeRate matches standard API responses like apis.net.pe
type RawSunatExchangeRate struct {
	Origen string      `json:"origen"`
	Compra interface{} `json:"compra"`
	Venta  interface{} `json:"venta"`
	Moneda string      `json:"moneda"`
	Fecha  string      `json:"fecha"`
}

// BCRPResponse matches the BCRP statistics API response
type BCRPResponse struct {
	Config struct {
		Title  string `json:"title"`
		Series []struct {
			Name string `json:"name"`
		} `json:"series"`
	} `json:"config"`
	Periods []struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	} `json:"periods"`
}

// ExchangeRateService handles querying official Peruvian exchange rates
type ExchangeRateService struct {
	BaseURL     string
	BCRPBaseURL string
	APIToken    string
	HTTPClient  *http.Client
}

// NewExchangeRateService creates a new ExchangeRateService instance
func NewExchangeRateService(apiToken string) *ExchangeRateService {
	return &ExchangeRateService{
		BaseURL:     "https://api.apis.net.pe/v1/tipo-cambio-sunat",
		BCRPBaseURL: "https://estadisticas.bcrp.gob.pe/estadisticas/series/api/PD04637PD-PD04638PD/json",
		APIToken:    apiToken,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetTodayExchangeRate retrieves the exchange rate for the current date
func (s *ExchangeRateService) GetTodayExchangeRate() (*ExchangeRateResponse, error) {
	return s.GetExchangeRate(time.Now().Format("2006-01-02"))
}

// GetExchangeRate retrieves the exchange rate for a given date (YYYY-MM-DD).
// If date is empty, it defaults to today's date.
func (s *ExchangeRateService) GetExchangeRate(date string) (*ExchangeRateResponse, error) {
	if strings.TrimSpace(date) == "" {
		date = time.Now().Format("2006-01-02")
	}

	// 1. Try Primary Provider (SUNAT API / apis.net.pe)
	rate, err := s.fetchFromPrimary(date)
	if err == nil && rate != nil && rate.Buy > 0 && rate.Sell > 0 {
		return &ExchangeRateResponse{
			Success: true,
			Data:    rate,
			Message: "Tipo de cambio obtenido exitosamente",
		}, nil
	}

	// 2. Fallback to Secondary Provider (BCRP Official API)
	bcrpRate, bcrpErr := s.fetchFromBCRP(date)
	if bcrpErr == nil && bcrpRate != nil && bcrpRate.Buy > 0 && bcrpRate.Sell > 0 {
		return &ExchangeRateResponse{
			Success: true,
			Data:    bcrpRate,
			Message: "Tipo de cambio obtenido mediante BCRP (fallback)",
		}, nil
	}

	// If both failed
	errMsg := "No se pudo obtener el tipo de cambio"
	if err != nil {
		errMsg = fmt.Sprintf("%s. Error primario: %v", errMsg, err)
	}
	if bcrpErr != nil {
		errMsg = fmt.Sprintf("%s. Error fallback: %v", errMsg, bcrpErr)
	}

	return &ExchangeRateResponse{
		Success: false,
		Message: errMsg,
	}, fmt.Errorf("fallo la consulta de tipo de cambio: %s", errMsg)
}

func (s *ExchangeRateService) fetchFromPrimary(date string) (*ExchangeRate, error) {
	url := s.BaseURL
	if date != "" {
		if strings.Contains(url, "?") {
			url = fmt.Sprintf("%s&fecha=%s", url, date)
		} else {
			url = fmt.Sprintf("%s?fecha=%s", url, date)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request primario: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	if s.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.APIToken)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request primario: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status HTTP primario no OK: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo cuerpo primario: %w", err)
	}

	var raw RawSunatExchangeRate
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("error parseando JSON primario: %w", err)
	}

	buyRate := parseFlexFloat(raw.Compra)
	sellRate := parseFlexFloat(raw.Venta)

	if buyRate <= 0 || sellRate <= 0 {
		return nil, fmt.Errorf("tasas invalidas en respuesta primaria (compra: %v, venta: %v)", buyRate, sellRate)
	}

	origin := raw.Origen
	if origin == "" {
		origin = "SUNAT"
	}

	currency := raw.Moneda
	if currency == "" {
		currency = "USD"
	}

	resDate := raw.Fecha
	if resDate == "" {
		resDate = date
	}

	return &ExchangeRate{
		Date:     resDate,
		Buy:      buyRate,
		Sell:     sellRate,
		Currency: currency,
		Origin:   origin,
	}, nil
}

func (s *ExchangeRateService) fetchFromBCRP(dateStr string) (*ExchangeRate, error) {
	// Format BCRP date string: YYYY-MM-DD -> start/end dates
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t = time.Now()
	}

	// BCRP accepts URL like /json/YYYY-M-D/YYYY-M-D or range
	// Query a small 5-day window ending on dateStr to account for weekends/holidays
	startDate := t.AddDate(0, 0, -5).Format("2006-1-2")
	endDate := t.Format("2006-1-2")

	url := fmt.Sprintf("%s/%s/%s", s.BCRPBaseURL, startDate, endDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request BCRP: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request BCRP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status HTTP BCRP no OK: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo cuerpo BCRP: %w", err)
	}

	var bcrpResp BCRPResponse
	if err := json.Unmarshal(body, &bcrpResp); err != nil {
		return nil, fmt.Errorf("error parseando JSON BCRP: %w", err)
	}

	if len(bcrpResp.Periods) == 0 {
		return nil, fmt.Errorf("sin datos en respuesta de BCRP")
	}

	// Iterate backwards to find the last valid period with numeric rates (ignoring "n.d.")
	for i := len(bcrpResp.Periods) - 1; i >= 0; i-- {
		period := bcrpResp.Periods[i]
		if len(period.Values) < 2 {
			continue
		}

		buy := parseFlexFloat(period.Values[0])
		sell := parseFlexFloat(period.Values[1])

		if buy > 0 && sell > 0 {
			return &ExchangeRate{
				Date:     dateStr,
				Buy:      buy,
				Sell:     sell,
				Currency: "USD",
				Origin:   "BCRP",
			}, nil
		}
	}

	return nil, fmt.Errorf("no se encontraron tasas validas en BCRP para el periodo")
}

func parseFlexFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		clean := strings.TrimSpace(val)
		if clean == "" || clean == "n.d." {
			return 0
		}
		f, err := strconv.ParseFloat(clean, 64)
		if err == nil {
			return f
		}
	}
	return 0
}
