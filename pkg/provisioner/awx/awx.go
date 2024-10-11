package awx

import (
	"fmt"
	"net/http"
	"time"

	tower "github.com/Kaginari/ansible-tower-sdk/client"
)

type AWXClient struct {
	Client *tower.AWX
}

func NewAWXClient(baseURL, username, password string) (*AWXClient, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	client, err := tower.NewAWX(baseURL, username, password, httpClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWX client: %w", err)
	}

	return &AWXClient{
		Client: client,
	}, nil
}

// LaunchJobTemplate launches a job template by its ID
func (a *AWXClient) LaunchJobTemplate(jobTemplateID int, data map[string]interface{}, params map[string]string) (int, error) {
	jobTemplateService := a.Client.JobTemplateService

	launch, err := jobTemplateService.Launch(jobTemplateID, data, params)
	if err != nil {
		return 0, fmt.Errorf("failed to launch job template: %w", err)
	}

	return launch.ID, nil
}

/*
// GetJobStatus retrieves the status of a job by its ID
func (a *AWXClient) GetJobStatus(jobID int) (string, error) {
    jobService := a.Client.JobService

    job, err := jobService.GetJobByID(context.Background(), jobID)
    if err != nil {
        return "", fmt.Errorf("failed to get job status: %w", err)
    }

    return job.Status, nil
}
*/
