package provisioner

type Response struct {
	Status string
	Data   map[string]interface{}
}

type StorageClassParameters map[string]string

type Plugin interface {
	Run(scp StorageClassParameters, provisionObject interface{}) (*Response, error)
	Validate(params interface{}) (*Response, error)
}
