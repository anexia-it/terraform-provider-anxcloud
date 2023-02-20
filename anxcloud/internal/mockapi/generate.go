package mockapi

import (
	_ "github.com/golang/mock/gomock"
)

//go:generate mockgen -package mockapi -destination api.go go.anx.io/go-anxcloud/pkg/api/types API
