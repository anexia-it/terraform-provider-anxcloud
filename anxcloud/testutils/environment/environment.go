package environment

import (
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

type Info struct {
	Location string
	VlanID   string
	Prefix   Prefix
}

var (
	// consumer tracks how many tests are currently using the environment
	consumers uint64
	envInfo   *Info
	mutex     sync.Mutex
)

func (i *Info) setup() error {
	context.Background()
	prefix, err := CreateTestPrefix(context.Background(), *i)
	if err != nil {
		return err
	}
	i.Prefix = prefix

	return nil
}

func (i *Info) cleanup(*testing.T) {

}

func GetEnvInfo(t *testing.T) Info {
	if envInfo == nil {
		initEnvironment(t)
	}
	return *envInfo
}

func initEnvironment(t *testing.T) {
	if token := os.Getenv("ANEXIA_TOKEN"); token == "" {
		t.Fatalf("'ANEXIA_TOKEN must be set in order to setup test environment'")
	}

	t.Cleanup(func() {
		log.Println("unsubscirbe from test environment")
		var swapped bool
		// if we are the last environment clean up
		for !swapped {
			oldVal := atomic.LoadUint64(&consumers)
			if oldVal == 0 {
				return
			}
			newVal := oldVal - 1
			swapped = atomic.CompareAndSwapUint64(&consumers, oldVal, newVal)
		}
		// if we are not the last one return
		if atomic.LoadUint64(&consumers) != 0 {
			log.Println("test environment is still in use. skipping cleanup")
			return
		}

		// we are the last one lock the test environment for cleanup
		mutex.Lock()
		defer mutex.Unlock()
		// we abort the delete here if someone started using the environment as well
		if atomic.LoadUint64(&consumers) != 0 {
			return
		}
		log.Println("clean up test environment")
		envInfo.cleanup(t)
		envInfo = nil
	})

	atomic.AddUint64(&consumers, 1)
	// lock until end of function
	mutex.Lock()
	defer mutex.Unlock()

	// we still have a envInfo and can use it
	if envInfo != nil {
		log.Println("attaching to existing test environment")
		return
	}

	log.Println("Setting up new test environment")
	// we create a new environment
	envInfo = &Info{
		VlanID:   "02f39d20ca0f4adfb5032f88dbc26c39",
		Location: "52b5f6b2fd3a4a7eaaedf1a7c019e9ea",
	}

	err := envInfo.setup()
	if err != nil {
		t.Fatalf("could not setup test environment: %s", err.Error())
	}
}
