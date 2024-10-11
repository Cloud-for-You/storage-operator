package awx

import (
	"fmt"
	"net/http"
	"os"
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
func LaunchJobTemplate(jtid int, data map[string]interface{}) (*tower.JobLaunch, error) {
	endpoint := os.Getenv("AWX_URL")
	username := os.Getenv("AWX_USERNAME")
	password := os.Getenv("AWX_PASSWORD")

	params := make(map[string]string)

	client, err := NewAWXClient(endpoint, username, password)
	if err != nil {
		return nil, err
	}

	jobLaunch, err := client.Client.JobTemplateService.Launch(jtid, data, params)
	if err != nil {
		return nil, err
	}

	return jobLaunch, nil
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
