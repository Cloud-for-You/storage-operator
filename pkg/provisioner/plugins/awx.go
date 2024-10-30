package provisioning_plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Cloud-for-You/storage-operator/pkg/httpclient"
	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
	"github.com/tidwall/gjson"
	"sigs.k8s.io/controller-runtime/pkg/log"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
)

// AWXPlugin implementuje Plugin interface
type AWXPlugin struct{}

func (p *AWXPlugin) Run(
	jobId string,
	jobParameters provisioner.JobParameters,
) (*provisioner.Response, error) {
	log.Log.Info("Running AWX job")
	provisionerResponse := &provisioner.Response{}

	host := os.Getenv("AWX_URL")
	username := os.Getenv("AWX_USERNAME")
	password := os.Getenv("AWX_PASSWORD")

	// Get BearerToken for username/password
	token := getToken(host, username, password)
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	bearerToken := gjson.Get(string(tokenJson), "token").Str

	// Call runJobTemplate()
	jobTemplate := runJobTemplate(host, bearerToken, jobId, jobParameters)
	jobTemplateJson, err := json.Marshal(jobTemplate)
	if err != nil {
		return nil, err
	}
	responseData := map[string]interface{}{
		"job_id": strconv.Itoa(int(gjson.Get(string(jobTemplateJson), "job").Int())),
		"status": gjson.Get(string(jobTemplateJson), "status").Str,
	}
	provisionerResponse.Data = responseData
	return provisionerResponse, nil
}

func (p *AWXPlugin) Validate(status storagev1.NfsStatus) (*provisioner.Response, error) {
	log.Log.Info("Validate AWX job")
	provisionerResponse := &provisioner.Response{}

	host := os.Getenv("AWX_URL")
	username := os.Getenv("AWX_USERNAME")
	password := os.Getenv("AWX_PASSWORD")

	var message provisioner.Response
	err := json.Unmarshal([]byte(status.Message), &message)
	if err != nil {
		return nil, err
	}

	if dataMap, ok := message.Data.(map[string]interface{}); ok {
		if jobId, ok := dataMap["job_id"].(string); ok {
			// Get BearerToken for username/password
			token := getToken(host, username, password)
			tokenJson, err := json.Marshal(token)
			if err != nil {
				return nil, err
			}
			bearerToken := gjson.Get(string(tokenJson), "token").Str

			// Call validateJobTemplate()
			jobTemplate := validateJobTemplate(host, bearerToken, jobId)
			jobTemplateJson, err := json.Marshal(jobTemplate)
			if err != nil {
				return nil, err
			}
			jobStatus := gjson.Get(string(jobTemplateJson), "status").Str

			switch jobStatus {
			case "pending":
				return nil, nil
			case "running":
				return nil, nil
			case "waiting":
				return nil, nil
			case "successful":
				responseData := map[string]interface{}{
					"job_id": gjson.Get(string(jobTemplateJson), "job").Str,
					"status": jobStatus,
				}
				provisionerResponse.Data = responseData
				return provisionerResponse, nil
			default:
				return nil, fmt.Errorf("unexpected status code: %s", jobStatus)
			}
		}
		return nil, nil
	}
	return nil, nil
}

func getToken(
	host string,
	username string,
	password string,
) (responseToken httpclient.APIResponse) {
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

func runJobTemplate(
	host string,
	token string,
	jobId string,
	ansibleParams provisioner.JobParameters,
) (responseTemplate httpclient.APIResponse) {
	var ansibleParamsInterface interface{} = ansibleParams

	params := httpclient.RequestParams{
		URL:    fmt.Sprintf("%s%s", host, fmt.Sprintf("/api/v2/job_templates/%s/launch/", jobId)),
		Method: "POST",
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Content-Type":  "application/json",
		},
		Body: ansibleParamsInterface,
	}

	response, err := httpclient.SendRequest(params)
	if err != nil {
		log.Log.Error(err, "Unable launch template")
		return nil
	}

	return response
}

func validateJobTemplate(
	host string,
	token string,
	jobId string,
) (responseTemplate httpclient.APIResponse) {
	params := httpclient.RequestParams{
		URL:    fmt.Sprintf("%s%s", host, fmt.Sprintf("/api/v2/jobs/%s/", jobId)),
		Method: "POST",
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
			"Content-Type":  "application/json",
		},
	}

	response, err := httpclient.SendRequest(params)
	if err != nil {
		log.Log.Error(err, "Unable validate job")
		return nil
	}

	return response
}
