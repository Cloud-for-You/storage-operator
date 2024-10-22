package provisioner

type Response struct {
	Status string
	Data   map[string]interface{}
}

type Plugin interface {
	Run(params interface{}) (*Response, error)
	Validate(params interface{}) (*Response, error)
}
