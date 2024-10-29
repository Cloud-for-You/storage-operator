package provisioner

type Response struct {
	Status string
	Data   map[string]interface{}
}

type JobParameters struct {
	Limit     string
	ExtraVars ExtraVars
}

type ExtraVars struct {
	K8s           bool   `json:"k8s"`
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
