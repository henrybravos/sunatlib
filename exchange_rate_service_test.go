package sunatlib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetExchangeRate_PrimarySuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"origen":"SUNAT","compra":3.751,"venta":3.758,"moneda":"USD","fecha":"2026-10-01"}`))
	}))
	defer mockServer.Close()

	service := NewExchangeRateService("")
	service.BaseURL = mockServer.URL

	resp, err := service.GetExchangeRate("2026-10-01")
	if err != nil {
		t.Fatalf("se esperaba exito, se obtuvo error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("se esperaba resp.Success == true, mensaje: %s", resp.Message)
	}

	if resp.Data == nil {
		t.Fatal("resp.Data es nil")
	}

	if resp.Data.Buy != 3.751 {
		t.Errorf("Buy = %v, esperaba 3.751", resp.Data.Buy)
	}
	if resp.Data.Sell != 3.758 {
		t.Errorf("Sell = %v, esperaba 3.758", resp.Data.Sell)
	}
	if resp.Data.Currency != "USD" {
		t.Errorf("Currency = %s, esperaba USD", resp.Data.Currency)
	}
	if resp.Data.Date != "2026-10-01" {
		t.Errorf("Date = %s, esperaba 2026-10-01", resp.Data.Date)
	}
}

func TestGetExchangeRate_FallbackBCRP(t *testing.T) {
	// Mock primary server returns error 500
	mockPrimary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockPrimary.Close()

	// Mock BCRP server returns valid BCRP JSON structure
	mockBCRP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		bcrpJSON := `{
			"config": {"title": "Tipo de cambio", "series": [{"name": "Compra"}, {"name": "Venta"}]},
			"periods": [
				{"name": "01.Oct.26", "values": ["3.750", "3.755"]}
			]
		}`
		w.Write([]byte(bcrpJSON))
	}))
	defer mockBCRP.Close()

	service := NewExchangeRateService("")
	service.BaseURL = mockPrimary.URL
	service.BCRPBaseURL = mockBCRP.URL

	resp, err := service.GetExchangeRate("2026-10-01")
	if err != nil {
		t.Fatalf("se esperaba fallback exitoso, se obtuvo error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("se esperaba exito en fallback BCRP, mensaje: %s", resp.Message)
	}

	if resp.Data.Buy != 3.750 || resp.Data.Sell != 3.755 {
		t.Errorf("Valores incorrectos en BCRP: Buy=%v, Sell=%v", resp.Data.Buy, resp.Data.Sell)
	}
	if resp.Data.Origin != "BCRP" {
		t.Errorf("Origin = %s, esperaba BCRP", resp.Data.Origin)
	}
}

func TestGetExchangeRate_DefaultToday(t *testing.T) {
	todayStr := time.Now().Format("2006-01-02")
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"origen": "SUNAT",
			"compra": 3.740,
			"venta":  3.748,
			"moneda": "USD",
			"fecha":  todayStr,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	service := NewExchangeRateService("")
	service.BaseURL = mockServer.URL

	resp, err := service.GetTodayExchangeRate()
	if err != nil {
		t.Fatalf("Error al obtener tipo de cambio de hoy: %v", err)
	}

	if resp.Data.Date != todayStr {
		t.Errorf("Date = %s, esperaba hoy (%s)", resp.Data.Date, todayStr)
	}
}

func TestGetExchangeRate_LiveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live integration test in short mode")
	}

	service := NewExchangeRateService("")
	resp, err := service.GetTodayExchangeRate()
	if err != nil {
		t.Fatalf("Error en consulta real de tipo de cambio: %v", err)
	}

	if !resp.Success {
		t.Fatalf("Consulta real devolvio success=false: %s", resp.Message)
	}

	if resp.Data.Buy <= 0 || resp.Data.Sell <= 0 {
		t.Errorf("Tasas obtenidas invalidas: Compra=%v, Venta=%v", resp.Data.Buy, resp.Data.Sell)
	}

	t.Logf("Tipo de cambio real obtenido: Fecha=%s, Compra=%v, Venta=%v, Origen=%s",
		resp.Data.Date, resp.Data.Buy, resp.Data.Sell, resp.Data.Origin)
}
