package provisioner

import (
	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
)

type Response struct {
	Data interface{} `json:"data,omitempty"`
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
	Validate(status storagev1.NfsStatus) (*Response, error)
}
