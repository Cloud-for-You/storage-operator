package provisioning_plugin

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Cloud-for-You/storage-operator/pkg/httpclient"
	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
	"github.com/tidwall/gjson"
	"sigs.k8s.io/controller-runtime/pkg/log"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
)

// AWXPlugin implementuje Plugin interface
type AWXPlugin struct{}

func (p *AWXPlugin) Run(scp provisioner.StorageClassParameters, po interface{}) (*provisioner.Response, error) {
	log.Log.Info(fmt.Sprintf("%s %s", "Running AWX job with params: ", scp))
	provisionerResponse := &provisioner.Response{}

	host := os.Getenv("AWX_URL")
	username := os.Getenv("AWX_USERNAME")
	password := os.Getenv("AWX_PASSWORD")
	jobId := scp["job-template-id"]

	// Get BearerToken for username/password
	token := getToken(host, username, password)
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	bearerToken := gjson.Get(string(tokenJson), "token").Str
	fmt.Println(bearerToken)

	// Run jobTemplate in AWX
	jobTemplate := runJobTemplate(host, bearerToken, jobId, po)
	fmt.Println(jobTemplate)

	provisionerResponse.Status = storagev1.AutomationRunning
	provisionerResponse.Data = token
	return provisionerResponse, nil
}

func (p *AWXPlugin) Validate(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Validating AWX with params:", params)
	response := &provisioner.Response{}
	return response, nil
}

func getToken(host, username, password string) (responseToken httpclient.APIResponse) {
	params := httpclient.RequestParams{
		URL:    fmt.Sprintf("%s%s", host, "/api/v2/tokens/"),
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Username: username,
		Password: password,
	}

	response, err := httpclient.SendRequest(params)
	if err != nil {
		log.Log.Error(err, "Unable get token")
		return nil
	}

	return response
}

func runJobTemplate(host, token, jobId string, ansibleParams interface{}) (responseTemplate httpclient.APIResponse) {
	fmt.Println(ansibleParams)

	params := httpclient.RequestParams{
		URL:    fmt.Sprintf("%s%s", host, fmt.Sprintf("/api/v2/job_templates/%s/launch/", jobId)),
		Method: "POST",
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Content-Type":  "application/json",
		},
	}

	response, err := httpclient.SendRequest(params)
	if err != nil {
		log.Log.Error(err, "Unable launch template")
		return nil
	}

	return response
}
