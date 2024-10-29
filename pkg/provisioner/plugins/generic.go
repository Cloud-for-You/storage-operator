package provisioning_plugin

import (
	"fmt"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GenericPlugin implementuje Plugin interface
type GenericPlugin struct{}

func (p *GenericPlugin) Run(
	jobId string,
	jobParameters provisioner.JobParameters,
) (*provisioner.Response, error) {
	log.Log.Info("Running Generic job")
	provisionerResponse := &provisioner.Response{}
	provisionerResponse.ProvisioningPlugin = "generic"
	provisionerResponse.State = storagev1.AutomationRunning
	provisionerResponse.Data = "{}"
	return provisionerResponse, nil
}

func (p *GenericPlugin) Validate(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Validating Generic with params:", params)
	response := &provisioner.Response{}
	return response, nil
}
