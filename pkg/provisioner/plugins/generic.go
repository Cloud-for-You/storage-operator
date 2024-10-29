package provisioning_plugin

import (
	"fmt"

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
	return provisionerResponse, nil
}

func (p *GenericPlugin) Validate(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Validating Generic with params:", params)
	response := &provisioner.Response{}
	return response, nil
}
