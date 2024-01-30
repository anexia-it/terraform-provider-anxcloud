package utils

import (
	"errors"
	"net/http"

	"go.anx.io/go-anxcloud/pkg/client"
)

func IsLegacyClientNotFound(err error) bool {
	var respErr *client.ResponseError
	if errors.As(err, &respErr) && respErr.ErrorData.Code == http.StatusNotFound {
		return true
	}
	return false
}
