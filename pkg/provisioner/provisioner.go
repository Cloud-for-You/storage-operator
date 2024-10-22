package provisioner

const (
	NoProvisioner = "no-provisioner"
	Awx           = "awx"
)

var supportedProvisionersMap = map[string]bool{
	NoProvisioner: false,
	Awx:           true,
}

func IsSupportProvisioner(provisioner string) bool {
	return supportedProvisionersMap[provisioner]
}
