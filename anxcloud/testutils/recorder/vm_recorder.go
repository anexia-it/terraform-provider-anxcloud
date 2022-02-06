package recorder

import (
	"context"
	"errors"
	"github.com/anexia-it/go-anxcloud/pkg/api"
	"github.com/anexia-it/go-anxcloud/pkg/client"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/provisioning/vm"
	"github.com/anexia-it/go-anxcloud/pkg/vsphere/search"
	"net/http"
)

type vmCleanUpFunc func(ctx context.Context) error
type VMRecoder struct {
	handlers []vmCleanUpFunc
	client   api.API
}

func (v VMRecoder) Cleanup(ctx context.Context) []error {
	var cleanUpErrors []error
	for _, handler := range v.handlers {
		err := handler(ctx)
		if err != nil {
			cleanUpErrors = append(cleanUpErrors, err)
		}
	}
	return nil
}

func (v *VMRecoder) RecordVMByName(name string) {
	v.handlers = append(v.handlers, v.createCleanupHandlerByName(name))
}

func (v *VMRecoder) RecordVMByID(id string) {
	v.handlers = append(v.handlers, v.createCleanupHandlerByName(id))
}

func (v VMRecoder) createCleanupHandlerByID(id string) vmCleanUpFunc {
	return func(ctx context.Context) error {
		anxClient, err := client.New(client.TokenFromEnv(false))
		if err != nil {
			return err
		}

		vmAPI := vm.NewAPI(anxClient)
		_, err = vmAPI.Deprovision(ctx, id, false)

		// it's not an error when it's gone
		var responseErr client.ResponseError
		if errors.As(err, &responseErr) && responseErr.Response.StatusCode != http.StatusNotFound {
			return err
		}

		return nil
	}
}

func (v VMRecoder) createCleanupHandlerByName(name string) vmCleanUpFunc {
	return func(ctx context.Context) error {
		client, err := client.New(client.TokenFromEnv(false))
		if err != nil {
			return err
		}

		res, err := search.NewAPI(client).ByName(ctx, name)
		if err != nil {
			return err
		}

		vmAPI := vm.NewAPI(client)
		for _, machine := range res {
			_, err := vmAPI.Deprovision(ctx, machine.Identifier, false)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
