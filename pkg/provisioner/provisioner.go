package provisioner

import (
	"fmt"
	"strconv"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
	"github.com/Cloud-for-You/storage-operator/pkg/provisioner/awx"
)

type ProvisionerResult struct {
	Data map[string]interface{} `json:"result,omitempty"`
}

func RunAutomation(provider string, parameters map[string]string) (*ProvisionerResult, error) {
	// Provolame AWX automatizaci
	switch provider {
	case "awx":
		fmt.Println("Spustime automatizaci pres AWX")
		jobTemplateId, exists := parameters["job_template_id"]
		if !exists {
			return nil, fmt.Errorf("not found 'parameters.hosts' in storageclass")
		}
		jtid, err := strconv.Atoi(jobTemplateId)
		if err != nil {
			return nil, fmt.Errorf("failed to convert job_template_id to int: %v", err)
		}
		hosts, exists := parameters["hosts"]
		if !exists {
			return nil, fmt.Errorf("not found 'parameters.hosts' in storageclass")
		}
		data := map[string]interface{}{
			"Limit": hosts,
		}
		jobLaunch, err := awx.LaunchJobTemplate(jtid, data)
		if err != nil {
			return nil, fmt.Errorf("failed to run job in awx: %v", err)
		}
		resultData := map[string]interface{}{
			"status": storagev1.AutomationRunning,
			"job_id": jobLaunch.ID,
		}
		return &ProvisionerResult{Data: resultData}, nil
	default:
		err := fmt.Errorf("not found automation plugin for %s", provider)
		return nil, err
	}
}

func ValidateAutomation(provider string, parameters map[string]string) (*ProvisionerResult, error) {
	resultData := map[string]interface{}{
		"status": storagev1.AutomationCompleted,
	}
	return &ProvisionerResult{Data: resultData}, nil
}
