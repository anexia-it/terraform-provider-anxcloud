package environment

import (
	"context"
	"os"
	"sync"
)

type Info struct {
	Location string
	VlanID   string
	Prefix   Prefix
}

var (
	envInfo *Info
	mutex   sync.Mutex
)

func (i *Info) setup() error {
	context.Background()
	prefix, err := CreateTestPrefix(context.Background())
	if err != nil {
		return err
	}
	i.Prefix = prefix

	return nil
}

func GetEnvInfo() Info {
	if envInfo == nil {
		panic("envInfo is only supported when ANEXIA_TOKEN is set")
	}
	return *envInfo
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
	envInfo = &Info{
		VlanID:   "02f39d20ca0f4adfb5032f88dbc26c39",
		Location: "52b5f6b2fd3a4a7eaaedf1a7c019e9ea",
	}
	err := envInfo.setup()
	if err != nil {
		panic(err)
	}
}
