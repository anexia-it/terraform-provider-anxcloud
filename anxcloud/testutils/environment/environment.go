package environment

import (
	"context"
	"os"
	"sync"
)

type Info struct {
	Location string
	VlanID   string
}

var (
	envInfo *Info
	mutex   sync.Mutex
)

type contextKey struct{}

func NewContext(ctx context.Context, environmentInfo Info) context.Context {
	return context.WithValue(ctx, contextKey{}, &environmentInfo)
}

func GetInfo(ctx context.Context) Info {
	return *ctx.Value(contextKey{}).(*Info)
}

func init() {
	if token := os.Getenv("ANEXIA_TOKEN"); token == "" {
		return
	}
	// already initialised
	if envInfo != nil {
		return
	}

	// lock until end of function
	mutex.Lock()
	defer mutex.Unlock()

	if envInfo != nil {
		return
	}
	envInfo = &Info{}
}
