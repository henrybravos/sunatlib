package sunatlib

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestExchangeService apunta las tres fuentes a servidores de prueba y fija
// "ahora" en el 1/10/2026 10:00 hora de Lima.
func newTestExchangeService(sunat, bcrp, apis http.HandlerFunc) (*ExchangeRateService, func()) {
	srvS := httptest.NewServer(sunat)
	srvB := httptest.NewServer(bcrp)
	srvA := httptest.NewServer(apis)
	s := NewExchangeRateService("")
	s.SunatURL = srvS.URL
	s.BCRPBaseURL = srvB.URL
	s.BaseURL = srvA.URL
	s.RetryDelay = time.Millisecond
	s.Now = func() time.Time { return time.Date(2026, 10, 1, 15, 0, 0, 0, time.UTC) }
	return s, func() { srvS.Close(); srvB.Close(); srvA.Close() }
}

func respond(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func failIfCalled(t *testing.T, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("%s no debía consultarse", name)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

const bcrpUntil30Sep = `{"periods":[
	{"name":"29.Set.26","values":["3.441","3.45"]},
	{"name":"30.Set.26","values":["3.429","3.437"]}]}`

func TestExchangeRate_TodayFromSunat(t *testing.T) {
	s, done := newTestExchangeService(
		respond(200, "01/10/2026|3.429|3.437|\n"),
		failIfCalled(t, "BCRP"),
		failIfCalled(t, "apis.net.pe"))
	defer done()

	resp, err := s.GetTodayExchangeRate()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Date != "2026-10-01" || resp.Data.Buy != 3.429 || resp.Data.Sell != 3.437 || resp.Data.Origin != ExchangeRateOriginSunat {
		t.Fatalf("obtuve %+v", resp.Data)
	}
}

func TestExchangeRate_SunatStillShowsYesterday_UsesBCRP(t *testing.T) {
	s, done := newTestExchangeService(
		respond(200, "30/09/2026|3.441|3.45|"),
		respond(200, bcrpUntil30Sep),
		failIfCalled(t, "apis.net.pe"))
	defer done()

	resp, err := s.GetExchangeRate("2026-10-01")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Sell != 3.437 || resp.Data.Origin != ExchangeRateOriginBCRP || resp.Data.CloseDate != "2026-09-30" {
		t.Fatalf("obtuve %+v", resp.Data)
	}
}

func TestExchangeRate_BCRPAsksForPreviousDaysOnly(t *testing.T) {
	var path string
	s, done := newTestExchangeService(
		failIfCalled(t, "SUNAT"),
		func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			w.Write([]byte(`{"periods":[{"name":"25.Set.26","values":["3.416","3.425"]}]}`))
		},
		failIfCalled(t, "apis.net.pe"))
	defer done()

	// Domingo 27/09: corresponde el cierre del viernes 25.
	resp, err := s.GetExchangeRate("2026-09-27")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(path, "/2026-9-17/2026-9-26") {
		t.Fatalf("el rango debe terminar el día anterior, se pidió %s", path)
	}
	if resp.Data.Date != "2026-09-27" || resp.Data.Sell != 3.425 || resp.Data.CloseDate != "2026-09-25" {
		t.Fatalf("obtuve %+v", resp.Data)
	}
}

func TestExchangeRate_BCRPRetriesOnChallengePage(t *testing.T) {
	calls := 0
	s, done := newTestExchangeService(
		failIfCalled(t, "SUNAT"),
		func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls < 3 {
				w.Write([]byte(`<!DOCTYPE html><html><head><script src="/x"></script></head></html>`))
				return
			}
			w.Write([]byte(bcrpUntil30Sep))
		},
		failIfCalled(t, "apis.net.pe"))
	defer done()

	resp, err := s.GetExchangeRate("2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || resp.Data.Origin != ExchangeRateOriginBCRP {
		t.Fatalf("esperaba éxito al tercer intento; intentos=%d origen=%s", calls, resp.Data.Origin)
	}
}

func TestExchangeRate_FallbackApisNetPe(t *testing.T) {
	s, done := newTestExchangeService(
		respond(500, ""),
		respond(200, `<!DOCTYPE html>`),
		respond(200, `{"origen":"SUNAT","compra":3.429,"venta":3.437,"moneda":"USD","fecha":"2026-10-01"}`))
	defer done()

	resp, err := s.GetTodayExchangeRate()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Sell != 3.437 || resp.Data.Origin != ExchangeRateOriginApisNetPe {
		t.Fatalf("obtuve %+v", resp.Data)
	}
}

func TestExchangeRate_ApisNetPeOtherDateRejected(t *testing.T) {
	s, done := newTestExchangeService(
		respond(500, ""),
		respond(500, ""),
		respond(200, `{"compra":3.44,"venta":3.45,"fecha":"2026-09-30"}`))
	defer done()

	resp, err := s.GetTodayExchangeRate()
	if err == nil || resp.Success {
		t.Fatal("un tipo de cambio de otro día no puede pasar como el pedido")
	}
}

func TestExchangeRate_AllFail(t *testing.T) {
	s, done := newTestExchangeService(respond(500, ""), respond(500, ""), respond(429, ""))
	defer done()

	resp, err := s.GetExchangeRate("2026-10-01")
	if err == nil || resp.Success {
		t.Fatal("sin fuentes debía fallar")
	}
	for _, src := range []string{"SUNAT", "BCRP", "apis.net.pe"} {
		if !strings.Contains(resp.Message, src) {
			t.Errorf("el mensaje debe decir qué falló en %s: %s", src, resp.Message)
		}
	}
}

func TestExchangeRate_TodayIsLimaDate(t *testing.T) {
	s := NewExchangeRateService("")
	// 2/10 03:00 UTC = 1/10 22:00 en Lima.
	s.Now = func() time.Time { return time.Date(2026, 10, 2, 3, 0, 0, 0, time.UTC) }
	if got := s.today(); got != "2026-10-01" {
		t.Fatalf("hoy en Lima es 2026-10-01, obtuve %s", got)
	}
}

func TestExchangeRate_InvalidDate(t *testing.T) {
	s := NewExchangeRateService("")
	if _, err := s.GetExchangeRate("01/10/2026"); err == nil {
		t.Fatal("formato inválido debía fallar")
	}
}

func TestGetExchangeRate_LiveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live integration test in short mode")
	}
	service := NewExchangeRateService("")
	for _, d := range []string{"", "2026-09-27"} {
		resp, err := service.GetExchangeRate(d)
		if err != nil {
			t.Fatalf("%q: %v", d, err)
		}
		t.Logf("%q → %s compra %.3f venta %.3f (%s, cierre %s)",
			d, resp.Data.Date, resp.Data.Buy, resp.Data.Sell, resp.Data.Origin, resp.Data.CloseDate)
	}
}
