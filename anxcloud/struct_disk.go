package anxcloud

import "github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"

type Disk struct {
	*vm.Disk
	ExactDiskSize float32
}
