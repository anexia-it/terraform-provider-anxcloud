package environment

import (
	"context"
	"fmt"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
	"net"
	"sync"
	"time"
)

type Prefix struct {
	ID      string
	CIDR    net.IPNet
	counter uint8
	mutex   *sync.Mutex
}

func (p *Prefix) GetNextIP() string {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var newIP net.IP
	copy(newIP, p.CIDR.IP)

	newIP[3] += p.counter + 2
	p.counter += 1

	return newIP.String()
}

func deletePrefix(ctx context.Context, environment Info) error {
	c, err := client.New(client.TokenFromEnv(false))
	if err != nil {
		return err
	}
	prefixAPI := prefix.NewAPI(c)
	err = prefixAPI.Delete(ctx, environment.Prefix.ID)
	if err != nil {
		return err
	}

	return nil
}

func CreateTestPrefix(ctx context.Context, environment Info) (Prefix, error) {
	c, err := client.New(client.TokenFromEnv(false))
	if err != nil {
		return Prefix{}, err
	}

	create := prefix.Create{
		Location:             environment.Location,
		IPVersion:            4,
		Type:                 1,
		NetworkMask:          24,
		CreateEmpty:          true,
		VLANID:               environment.VlanID,
		EnableVMProvisioning: false,
		CustomerDescription:  fmt.Sprintf("terraform-test: %s", environment.TestRunName),
	}

	prefixAPI := prefix.NewAPI(c)
	summary, err := prefixAPI.Create(ctx, create)
	if err != nil {
		return Prefix{}, err
	}

	for {
		fetchedPrefix, err := prefixAPI.Get(ctx, summary.ID)
		if err != nil {
			panic(err)
		}
		if fetchedPrefix.Status == "Active" {
			_, network, err := net.ParseCIDR(fetchedPrefix.Name)
			if err != nil || network == nil {
				return Prefix{}, fmt.Errorf("could not parse CIDR '%s': %w", fetchedPrefix.Name, err)
			}

			return Prefix{
				ID:    fetchedPrefix.ID,
				CIDR:  *network,
				mutex: &sync.Mutex{},
			}, nil
		}
		if fetchedPrefix.Status == "Failed" {
			err := prefixAPI.Delete(ctx, summary.ID)
			if err != nil {
				panic(fmt.Sprintf("setting up test prefix failed and err occurred when deleting: %s", err.Error()))
			}
			panic("setting up test prefix failed")
		}
		time.Sleep(5 * time.Second)
	}
}
