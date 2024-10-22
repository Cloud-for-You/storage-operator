package provisioning_plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
)

// AWXPlugin implementuje Plugin interface
type AWXPlugin struct{}

type loginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type tokenResponse struct {
	Token string `json:"token"`
}

func (p *AWXPlugin) Run(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Running AWX with params:", params)
	response := &provisioner.Response{}
	url := os.Getenv("AWX_URL")
	username := os.Getenv("AWX_USERNAME")
	password := os.Getenv("AWX_PASSWORD")

	token, err := getBearerToken(url, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}
	err = callAPIWithBearerToken(url, token)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %v", err)
	}

	return response, nil
}

func (p *AWXPlugin) Validate(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Validating AWX with params:", params)
	response := &provisioner.Response{}
	return response, nil
}

func getBearerToken(url, username, password string) (string, error) {
	// Vytvoření payloadu s přihlašovacími údaji
	loginData := loginPayload{
		Username: username,
		Password: password,
	}

	// Serializace do JSONu
	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return "", err
	}

	// HTTP POST request pro získání tokenu
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Zpracování chyby při zavírání
			log.Printf("error closing response body: %v", cerr)
		}
	}()

	// Kontrola HTTP status kódu
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to login, status code: %d", resp.StatusCode)
	}

	// Čtení odpovědi
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Deserializace odpovědi do struktury TokenResponse
	var tokenResponse tokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.Token, nil
}

func callAPIWithBearerToken(apiUrl, token string) error {
	// Vytvoření HTTP requestu s Bearer tokenem
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Odeslání requestu
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Zpracování chyby při zavírání
			log.Printf("error closing response body: %v", cerr)
		}
	}()

	// Čtení odpovědi
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Výstup odpovědi
	fmt.Println("Response:", string(body))

	return nil
}
