package anxcloud

import "go.anx.io/go-anxcloud/pkg/vsphere/provisioning/vm"

// Disk describes the extended disk type holding the API model and the un-modified floating point disk size for comparison.
type Disk struct {
	*vm.Disk

	// ExactDiskSize attribute is used to determine whether or not a change to the disk size is applicable and requires
	// scaling of the corresponding disk.
	ExactDiskSize float64
}
