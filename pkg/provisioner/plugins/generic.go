package provisioning_plugin

import (
	"fmt"

	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
)

// GenericPlugin implementuje Plugin interface
type GenericPlugin struct{}

func (p *GenericPlugin) Run(
	jobId string,
	jobParameters provisioner.JobParameters,
) (*provisioner.Response, error) {
	fmt.Println("Running Generic job with params:", jobParameters)
	response := &provisioner.Response{}
	return response, nil
}

func (p *GenericPlugin) Validate(params interface{}) (*provisioner.Response, error) {
	fmt.Println("Validating Generic with params:", params)
	response := &provisioner.Response{}
	return response, nil
}
