package anxcloud

import (
	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())

	// setup test environment
	var env *environment.Info
	var err error
	if env, err = environment.InitEnvironment(); err != nil {
		log.Fatalf("could not setup environment: %s", err.Error())
	}

	// run tests
	exitCode := m.Run()

	// cleanup
	if err := env.CleanUp(); err != nil {
		log.Fatalf("could not clean up environment: %s", err.Error())
	}
	os.Exit(exitCode)
}
