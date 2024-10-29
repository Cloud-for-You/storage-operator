package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIResponse represents a generic structure for the JSON response
type APIResponse map[string]interface{}

// RequestParams defines the parameters for the HTTP request
type RequestParams struct {
	URL      string
	Method   string
	Body     interface{}
	Headers  map[string]string
	Username string
	Password string
}

// SendRequest sends a generic HTTP request based on the provided parameters
func SendRequest(params RequestParams) (APIResponse, error) {
	// Serializujeme tělo požadavku, pokud existuje
	var reqBody []byte
	var err error
	if params.Body != nil {
		reqBody, err = json.Marshal(params.Body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
	}

	// Vytvoříme nový HTTP požadavek s metodou a URL
	req, err := http.NewRequest(params.Method, params.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Nastavíme hlavičky, pokud byly předány
	for key, value := range params.Headers {
		req.Header.Set(key, value)
	}

	// Pokud je uvedena základní autentizace, nastavíme ji
	if params.Username != "" && params.Password != "" {
		req.SetBasicAuth(params.Username, params.Password)
	}

	// Provedeme požadavek
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// ošetření chyby při zavírání těla odpovědi
			fmt.Println("Error closing response body:", cerr)
		}
	}()

	// Zkontrolujeme stavový kód odpovědi
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received non-OK response: %s", resp.Status)
	}

	// Zpracujeme tělo odpovědi
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Ověříme syrový JSON (můžete odstranit, pokud nepotřebujete výpis)
	fmt.Printf("Raw JSON: %s\n", string(body))

	// Parsujeme JSON do struktury APIResponse
	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return apiResponse, nil
}

func ReadDataFromResponse(data APIResponse, key string) (string, error) {
	if t, ok := data[key].(string); ok {
		return t, nil
	}
	return "", fmt.Errorf("token field not found or is not a string")
}
