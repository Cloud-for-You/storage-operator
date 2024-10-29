package provisioner

type Response struct {
	ProvisioningPlugin string      `json:"provisioning_plugin"`
	State              string      `json:"state"`
	Data               interface{} `json:"k8s,omitempty"`
}

type JobParameters struct {
	Limit     string    `json:"limit"`
	ExtraVars ExtraVars `json:"extra_vars"`
}

type ExtraVars struct {
	K8s           string `json:"k8s,omitempty"`
	ClusterName   string `json:"cluster_name"`
	NamespaceName string `json:"namespace_name"`
	PvcName       string `json:"pvc_name"`
	PvcSize       string `json:"pvc_size"`
	VgName        string `json:"vg_name,omitempty"`
	ExportCidr    string `json:"export_cidr,omitempty"`
}

type Plugin interface {
	Run(jobId string, jobParameters JobParameters) (*Response, error)
	Validate(params interface{}) (*Response, error)
}
