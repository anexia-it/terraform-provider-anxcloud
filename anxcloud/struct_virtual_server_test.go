package anxcloud

import (
	"fmt"
	"testing"

	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
	"github.com/google/go-cmp/cmp"
)

func TestExpanderNetworks(t *testing.T) {
	cases := []struct {
		Input          []interface{}
		ExpectedOutput []vm.Network
	}{
		{
			[]interface{}{
				map[string]interface{}{
					"vlan_id":  "38f8561acfe34qc49c336d2af31a5cc3",
					"nic_type": "vmxnet3",
					"ips": []interface{}{
						"identifier1",
						"identifier2",
						"10.11.12.13",
						"1.0.0.1",
					},
				},
			},
			[]vm.Network{
				{
					VLAN:    "38f8561acfe34qc49c336d2af31a5cc3",
					NICType: "vmxnet3",
					IPs: []string{
						"identifier1",
						"identifier2",
						"10.11.12.13",
						"1.0.0.1",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		output := expandNetworks(tc.Input)
		fmt.Println(output)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
