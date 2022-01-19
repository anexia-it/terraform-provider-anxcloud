package components

import (
	"context"
	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/ipam/prefix"
	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
)

func CreateTestPrefix(ctx context.Context) (string, error) {
	environment := environment.GetInfo(ctx)
	c, err := client.New(client.TokenFromEnv(false))
	if err != nil {
		return "", err
	}

	create := prefix.Create{
		Location:             environment.Location,
		IPVersion:            4,
		Type:                 0,
		NetworkMask:          24,
		CreateEmpty:          true,
		VLANID:               environment.VlanID,
		EnableVMProvisioning: false,
		CustomerDescription:  "A prefix used for testing",
	}

	summary, err := prefix.NewAPI(c).Create(ctx, create)
	if err != nil {
		return "", err
	}

	return summary.ID, err
}
