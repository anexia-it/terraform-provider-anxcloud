package v1

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"go.anx.io/go-anxcloud/pkg/api/types"
)

type E5EFunctionDeployment struct {
	FunctionIdentifier string `json:"-"`
}

func (d *E5EFunctionDeployment) GetIdentifier(ctx context.Context) (string, error) {
	return "", nil
}

func (d *E5EFunctionDeployment) EndpointURL(ctx context.Context) (*url.URL, error) {
	op, err := types.OperationFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if op != types.OperationCreate {
		return nil, errors.New("helper resource 'E5EFunctionDeployment' only supports create operations")
	}

	return url.Parse(fmt.Sprintf("/api/e5e/v1/function.json/%s/deploy", d.FunctionIdentifier))
}
