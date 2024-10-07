package provisioner

type ProvisionerList []string

var ValidProvisioners = ProvisionerList{
	"kubernetes.io/no-provisioner",
	"storage-operator.cfy.cz/awx",
	"storage-operator.cfy.cz/custom",
}

// Funkce, která kontroluje, zda ProvisionerList obsahuje hodnotu
func (p ProvisionerList) Contains(item string) bool {
	for _, v := range p {
		if v == item {
			return true
		}
	}
	return false
}
