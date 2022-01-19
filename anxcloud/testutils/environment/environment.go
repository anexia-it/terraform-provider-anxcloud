package environment

import (
	"context"
)

type Info struct {
	Location string
	VlanID   string
}

type contextKey struct{}

func NewContext(ctx context.Context, environmentInfo Info) context.Context {
	return context.WithValue(ctx, contextKey{}, &environmentInfo)
}

func GetInfo(ctx context.Context) Info {
	return *ctx.Value(contextKey{}).(*Info)
}
