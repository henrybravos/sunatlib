package gre

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGreClient_GetToken(t *testing.T) {
	// Mock SUNAT OAuth server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/clientessol/client-id/oauth2/token/" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "fake-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	client := &GreClient{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     server.URL + "/v1/clientessol/%s/oauth2/token/",
	}

	token, err := client.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken() error = %v", err)
	}

	if token.AccessToken != "fake-token" {
		t.Errorf("Expected token fake-token, got %s", token.AccessToken)
	}
}

func TestGreClient_SendGuide(t *testing.T) {
	// Mock SUNAT GRE API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"numTicket": "123456789",
			"fecPedido": "2023-10-27T12:00:00Z",
		})
	}))
	defer server.Close()

	client := &GreClient{
		ApiURL: server.URL,
		Token:  &OAuthToken{AccessToken: "valid-token"},
	}

	resp, err := client.SendGuide(context.Background(), "20600000000-09-T001-1.zip", []byte("fake-zip-content"))
	if err != nil {
		t.Fatalf("SendGuide() error = %v", err)
	}

	if resp.NumTicket != "123456789" {
		t.Errorf("Expected ticket 123456789, got %s", resp.NumTicket)
	}
}

func TestGreClient_GetStatus(t *testing.T) {
	// Mock SUNAT GRE Status API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"codRespuesta": "0",
			"indProg":      "0",
			"arcCdr":       "SGVsbG8gQ0RS", // "Hello CDR" in base64
		})
	}))
	defer server.Close()

	client := &GreClient{
		ApiURL: server.URL,
		Token:  &OAuthToken{AccessToken: "valid-token"},
	}

	resp, err := client.GetStatus(context.Background(), "123456789")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}

	if resp.CodRespuesta != "0" {
		t.Errorf("Expected codRespuesta 0, got %s", resp.CodRespuesta)
	}

	if resp.ArcCdr == "" {
		t.Error("Expected ArcCdr to be not empty")
	}
}
