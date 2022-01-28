package recorder

import (
	"context"
	"errors"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
)

type prefixCleanupHandler func(ctx context.Context) error

type PrefixRecorder struct {
	handler []prefixCleanupHandler
}

func (p PrefixRecorder) Cleanup(ctx context.Context) []error {
	var cleanUpErrors []error
	for _, handler := range p.handler {
		err := handler(ctx)
		if err != nil {
			cleanUpErrors = append(cleanUpErrors, err)
		}
	}
	return nil
}

func (p *PrefixRecorder) RecordPrefixByID(identifier string) {
	p.handler = append(p.handler, func(ctx context.Context) error {
		c, err := client.New(client.TokenFromEnv(false))
		if err != nil {
			return err
		}

		prefixAPI := prefix.NewAPI(c)
		_, err = prefixAPI.Get(ctx, identifier)

		var responseError *client.ResponseError
		isResponseError := errors.As(err, &responseError)
		if isResponseError && responseError.Response.StatusCode == 404 {
			return nil
		} else if err != nil {
			return err
		}

		return prefixAPI.Delete(ctx, identifier)
	})
}
