package provisioner

type Response struct {
	Status string
	Data   map[string]interface{}
}

type JobParameters struct {
	Limit     string    `json:"limit"`
	ExtraVars ExtraVars `json:"extra_vars"`
}

type ExtraVars struct {
	K8s           string `json:"k8s,omitempty"`
	ClusterName   string `json:"cluster-name"`
	NamespaceName string `json:"namespace-name"`
	PvcName       string `json:"pvc-name"`
	PvcSize       string `json:"pvc-size"`
	VgName        string `json:"vg-name,omitempty"`
	ExportCidr    string `json:"export-cidr,omitempty"`
}

type Plugin interface {
	Run(jobId string, jobParameters JobParameters) (*Response, error)
	Validate(params interface{}) (*Response, error)
}
