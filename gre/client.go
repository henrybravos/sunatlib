package gre

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GreClient is the client for SUNAT's New GRE REST API
type GreClient struct {
	ClientID     string
	ClientSecret string
	Username     string // RUC + SOL User
	Password     string // SOL Password
	TokenURL     string // URL template with %s for ClientID
	ApiURL       string
	Token        *OAuthToken
	HttpClient   *http.Client
}

// GetToken requests a new OAuth token from SUNAT
func (c *GreClient) GetToken(ctx context.Context) (*OAuthToken, error) {
	tokenURL := fmt.Sprintf(c.TokenURL, c.ClientID)
	
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("scope", "https://api-cpe.sunat.gob.pe")
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("username", c.Username)
	data.Set("password", c.Password)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := c.HttpClient
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OAuth error: %d - %s", resp.StatusCode, string(body))
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	token.IssuedAt = time.Now()
	c.Token = &token
	return &token, nil
}

// SendGuide sends a ZIP containing the signed XML to SUNAT
func (c *GreClient) SendGuide(ctx context.Context, fileName string, zipContent []byte) (*GreResponse, error) {
	if c.Token == nil || c.Token.AccessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	// Calculate SHA256 of the ZIP
	hash := sha256.Sum256(zipContent)
	hashHex := fmt.Sprintf("%x", hash)

	reqPayload := GreRequest{}
	reqPayload.Archivo.NomArchivo = fileName + ".zip"
	reqPayload.Archivo.ArcGreZip = base64.StdEncoding.EncodeToString(zipContent)
	reqPayload.Archivo.HashZip = hashHex

	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.ApiURL+"/"+fileName, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := c.HttpClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GRE API error: %d - %s", resp.StatusCode, string(body))
	}

	var greResp GreResponse
	if err := json.NewDecoder(resp.Body).Decode(&greResp); err != nil {
		return nil, err
	}

	return &greResp, nil
}

// GetStatus queries the status of a previously submitted ticket
func (c *GreClient) GetStatus(ctx context.Context, ticket string) (*GreStatusResponse, error) {
	if c.Token == nil || c.Token.AccessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.ApiURL+"/"+ticket, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token.AccessToken)

	client := c.HttpClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GRE API error: %d - %s", resp.StatusCode, string(body))
	}

	var statusResp GreStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, err
	}

	return &statusResp, nil
}
