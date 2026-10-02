package sunatlib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Tipo de cambio oficial (S/ por US$).
//
// El tipo de cambio que SUNAT publica para el día D es el cierre SBS del último
// día hábil anterior a D ("corresponde a la cotización de cierre de la SBS del
// día anterior", dice la propia SUNAT). Comprobado contra datos reales:
// SUNAT 01/10/2026 = 3.429/3.437 = BCRP 30/09; SUNAT domingo 27/09 =
// 3.416/3.425 = BCRP viernes 25/09.
//
// Orden de fuentes:
//  1. SUNAT, sunat.gob.pe/a/txt/tipoCambio.txt: solo publica el día de hoy.
//  2. BCRP, series SBS PD04639PD (compra) y PD04640PD (venta): el último
//     cierre anterior a D. No las interbancarias PD04637PD/PD04638PD, que dan
//     otro valor.
//  3. apis.net.pe: tercero, limita a pocas consultas (429). Último recurso.

// ExchangeRate represents the exchange rate structure
type ExchangeRate struct {
	Date     string  `json:"fecha"`  // Día al que corresponde el tipo de cambio (YYYY-MM-DD)
	Buy      float64 `json:"compra"` // Compra
	Sell     float64 `json:"venta"`  // Venta
	Currency string  `json:"moneda"` // USD
	Origin   string  `json:"origen"` // SUNAT, BCRP, APISNETPE
	// CloseDate: fecha del cierre SBS usado (solo BCRP). Si no es el día
	// anterior a Date puede ser un fin de semana o feriado, o que el BCRP aún
	// no publicó el último cierre.
	CloseDate string `json:"fecha_cierre,omitempty"`
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

// Orígenes posibles de ExchangeRate.Origin.
const (
	ExchangeRateOriginSunat     = "SUNAT"
	ExchangeRateOriginBCRP      = "BCRP"
	ExchangeRateOriginApisNetPe = "APISNETPE"
)

// limaLocation: Perú no tiene horario de verano; FixedZone evita depender de
// tzdata en el sistema.
var limaLocation = time.FixedZone("America/Lima", -5*3600)

// ExchangeRateService handles querying official Peruvian exchange rates
type ExchangeRateService struct {
	SunatURL    string // tipoCambio.txt de SUNAT
	BaseURL     string // apis.net.pe (último recurso)
	BCRPBaseURL string
	APIToken    string // token opcional para apis.net.pe
	HTTPClient  *http.Client
	// BCRPRetries / RetryDelay: el BCRP a veces responde una página HTML de
	// verificación en vez del JSON (medido: ~la mitad de las consultas
	// seguidas). Se reintenta con pausa creciente.
	BCRPRetries int
	RetryDelay  time.Duration
	// Now permite fijar la hora en tests; por defecto time.Now.
	Now func() time.Time
}

// NewExchangeRateService creates a new ExchangeRateService instance
func NewExchangeRateService(apiToken string) *ExchangeRateService {
	return &ExchangeRateService{
		SunatURL:    "https://www.sunat.gob.pe/a/txt/tipoCambio.txt",
		BaseURL:     "https://api.apis.net.pe/v1/tipo-cambio-sunat",
		BCRPBaseURL: "https://estadisticas.bcrp.gob.pe/estadisticas/series/api/PD04639PD-PD04640PD/json",
		APIToken:    apiToken,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		BCRPRetries: 3,
		RetryDelay:  2 * time.Second,
		Now:         time.Now,
	}
}

func (s *ExchangeRateService) today() string {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	return now().In(limaLocation).Format("2006-01-02")
}

// GetTodayExchangeRate retrieves the exchange rate for the current date (Lima)
func (s *ExchangeRateService) GetTodayExchangeRate() (*ExchangeRateResponse, error) {
	return s.GetExchangeRate(s.today())
}

// GetExchangeRate retrieves the exchange rate for a given date (YYYY-MM-DD).
// If date is empty, it defaults to today's date (Lima).
func (s *ExchangeRateService) GetExchangeRate(date string) (*ExchangeRateResponse, error) {
	date = strings.TrimSpace(date)
	if date == "" {
		date = s.today()
	}
	d, err := time.ParseInLocation("2006-01-02", date, limaLocation)
	if err != nil {
		return &ExchangeRateResponse{Success: false, Message: "fecha invalida, se esperaba YYYY-MM-DD"},
			fmt.Errorf("fecha invalida %q: %w", date, err)
	}

	var errs []string
	ok := func(rate *ExchangeRate, msg string) (*ExchangeRateResponse, error) {
		return &ExchangeRateResponse{Success: true, Data: rate, Message: msg}, nil
	}

	// 1. SUNAT: solo sirve para hoy, y solo si ya publicó el de hoy.
	if date == s.today() {
		rate, err := s.fetchFromSunat()
		switch {
		case err != nil:
			errs = append(errs, "SUNAT: "+err.Error())
		case rate.Date != date:
			errs = append(errs, fmt.Sprintf("SUNAT publica el %s, no el %s", rate.Date, date))
		default:
			return ok(rate, "Tipo de cambio obtenido exitosamente")
		}
	}

	// 2. BCRP: último cierre SBS anterior a la fecha.
	rate, err := s.fetchFromBCRP(d)
	if err == nil {
		return ok(rate, "Tipo de cambio obtenido mediante BCRP")
	}
	errs = append(errs, "BCRP: "+err.Error())

	// 3. apis.net.pe
	rate, err = s.fetchFromPrimary(date)
	if err == nil {
		return ok(rate, "Tipo de cambio obtenido mediante apis.net.pe")
	}
	errs = append(errs, "apis.net.pe: "+err.Error())

	errMsg := "No se pudo obtener el tipo de cambio. " + strings.Join(errs, "; ")
	return &ExchangeRateResponse{Success: false, Message: errMsg},
		fmt.Errorf("fallo la consulta de tipo de cambio: %s", errMsg)
}

// fetchFromSunat lee "01/10/2026|3.429|3.437|" (fecha|compra|venta).
func (s *ExchangeRateService) fetchFromSunat() (*ExchangeRate, error) {
	body, err := s.get(s.SunatURL, "text/plain")
	if err != nil {
		return nil, err
	}
	return parseSunatTipoCambio(body)
}

func parseSunatTipoCambio(body []byte) (*ExchangeRate, error) {
	parts := strings.Split(strings.TrimSpace(string(body)), "|")
	if len(parts) < 3 {
		return nil, fmt.Errorf("respuesta con formato inesperado: %q", truncateString(string(body), 80))
	}
	d, err := time.ParseInLocation("02/01/2006", strings.TrimSpace(parts[0]), limaLocation)
	if err != nil {
		return nil, fmt.Errorf("fecha invalida %q: %w", parts[0], err)
	}
	buy := parseFlexFloat(parts[1])
	sell := parseFlexFloat(parts[2])
	if buy <= 0 || sell <= 0 {
		return nil, fmt.Errorf("tasas invalidas (compra: %q, venta: %q)", parts[1], parts[2])
	}
	return &ExchangeRate{
		Date:     d.Format("2006-01-02"),
		Buy:      buy,
		Sell:     sell,
		Currency: "USD",
		Origin:   ExchangeRateOriginSunat,
	}, nil
}

// fetchFromBCRP devuelve, para el día d, el último cierre SBS anterior a d.
func (s *ExchangeRateService) fetchFromBCRP(d time.Time) (*ExchangeRate, error) {
	// Diez días hacia atrás cubren fines de semana largos y feriados.
	from := d.AddDate(0, 0, -10).Format("2006-1-2")
	to := d.AddDate(0, 0, -1).Format("2006-1-2")
	url := fmt.Sprintf("%s/%s/%s", s.BCRPBaseURL, from, to)

	attempts := s.BCRPRetries
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * s.RetryDelay)
		}
		body, err := s.get(url, "application/json")
		if err != nil {
			lastErr = err
			continue
		}
		closeDate, buy, sell, err := parseBCRP(body)
		if err != nil {
			lastErr = err
			continue
		}
		return &ExchangeRate{
			Date:      d.Format("2006-01-02"),
			Buy:       buy,
			Sell:      sell,
			Currency:  "USD",
			Origin:    ExchangeRateOriginBCRP,
			CloseDate: closeDate.Format("2006-01-02"),
		}, nil
	}
	return nil, lastErr
}

// parseBCRP toma el último periodo con valores numéricos ("n.d." = sin dato).
func parseBCRP(body []byte) (time.Time, float64, float64, error) {
	var resp BCRPResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return time.Time{}, 0, 0, fmt.Errorf("respuesta no es JSON valido: %w", err)
	}
	for i := len(resp.Periods) - 1; i >= 0; i-- {
		p := resp.Periods[i]
		if len(p.Values) < 2 {
			continue
		}
		buy := parseFlexFloat(p.Values[0])
		sell := parseFlexFloat(p.Values[1])
		if buy <= 0 || sell <= 0 {
			continue
		}
		closeDate, err := parseBCRPPeriod(p.Name)
		if err != nil {
			return time.Time{}, 0, 0, err
		}
		return closeDate, buy, sell, nil
	}
	return time.Time{}, 0, 0, fmt.Errorf("sin tipo de cambio en el rango consultado")
}

var bcrpMonths = map[string]time.Month{
	"Ene": time.January, "Feb": time.February, "Mar": time.March, "Abr": time.April,
	"May": time.May, "Jun": time.June, "Jul": time.July, "Ago": time.August,
	"Set": time.September, "Sep": time.September, "Oct": time.October,
	"Nov": time.November, "Dic": time.December,
}

// parseBCRPPeriod interpreta "30.Set.26".
func parseBCRPPeriod(name string) (time.Time, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("periodo inesperado: %q", name)
	}
	day, errD := strconv.Atoi(parts[0])
	month, okM := bcrpMonths[parts[1]]
	year, errY := strconv.Atoi(parts[2])
	if errD != nil || !okM || errY != nil {
		return time.Time{}, fmt.Errorf("periodo inesperado: %q", name)
	}
	return time.Date(2000+year, month, day, 0, 0, 0, 0, limaLocation), nil
}

// fetchFromPrimary consulta apis.net.pe.
func (s *ExchangeRateService) fetchFromPrimary(date string) (*ExchangeRate, error) {
	url := s.BaseURL
	if strings.Contains(url, "?") {
		url = fmt.Sprintf("%s&fecha=%s", url, date)
	} else {
		url = fmt.Sprintf("%s?fecha=%s", url, date)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")
	if s.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.APIToken)
	}
	body, err := s.do(req)
	if err != nil {
		return nil, err
	}

	var raw RawSunatExchangeRate
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}
	buy := parseFlexFloat(raw.Compra)
	sell := parseFlexFloat(raw.Venta)
	if buy <= 0 || sell <= 0 {
		return nil, fmt.Errorf("tasas invalidas (compra: %v, venta: %v)", buy, sell)
	}
	if raw.Fecha != "" && raw.Fecha != date {
		return nil, fmt.Errorf("respondio el %s, no el %s", raw.Fecha, date)
	}
	currency := raw.Moneda
	if currency == "" {
		currency = "USD"
	}
	return &ExchangeRate{
		Date:     date,
		Buy:      buy,
		Sell:     sell,
		Currency: currency,
		Origin:   ExchangeRateOriginApisNetPe,
	}, nil
}

func (s *ExchangeRateService) get(url, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", accept)
	return s.do(req)
}

func (s *ExchangeRateService) do(req *http.Request) ([]byte, error) {
	client := s.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status HTTP %d", resp.StatusCode)
	}
	return body, nil
}

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
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
