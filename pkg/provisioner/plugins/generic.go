package provisioning_plugin

import (
	"fmt"

	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
	"sigs.k8s.io/controller-runtime/pkg/log"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
)

// GenericPlugin implementuje Plugin interface
type GenericPlugin struct{}

func (p *GenericPlugin) Run(
	jobId string,
	jobParameters provisioner.JobParameters,
) (*provisioner.Response, error) {
	log.Log.Info("Running Generic job")
	provisionerResponse := &provisioner.Response{}
	provisionerResponse.Data = "{}"
	return provisionerResponse, nil
}

func (p *GenericPlugin) Validate(status storagev1.NfsStatus) (*provisioner.Response, error) {
	fmt.Println("Validating Generic with params:", status)
	response := &provisioner.Response{}
	return response, nil
}
