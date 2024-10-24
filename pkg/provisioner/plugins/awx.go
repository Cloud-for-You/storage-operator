package provisioning_plugin

import (
	"fmt"
	"os"

	"github.com/Cloud-for-You/storage-operator/pkg/httpclient"
	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
)

// AWXPlugin implementuje Plugin interface
type AWXPlugin struct{}

func (p *AWXPlugin) Run(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Running AWX with params:", params)
	response := &provisioner.Response{}
	url := os.Getenv("AWX_URL")
	username := os.Getenv("AWX_USERNAME")
	password := os.Getenv("AWX_PASSWORD")

	token, err := getToken(url, username, password)
	if err != nil {
		return nil, fmt.Errorf("error automation")
	}

	fmt.Println("TEST", token)

	return response, nil
}

func (p *AWXPlugin) Validate(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Validating AWX with params:", params)
	response := &provisioner.Response{}
	return response, nil
}

func getToken(url, username, password string) (*string, error) {
	params := httpclient.RequestParams{
		URL:    fmt.Sprintf("%s%s", url, "/api/v2/tokens/"),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Username: username,
		Password: password,
	}

	response, err := httpclient.SendRequest(params)
	if err != nil {
		return nil, err
	}

	if t, ok := response["token"].(string); ok {
		return &t, nil // Vracíme ukazatel na token
	}

	// Pokud není token nalezen, vracíme nil
	return nil, fmt.Errorf("token not found in response")
}
