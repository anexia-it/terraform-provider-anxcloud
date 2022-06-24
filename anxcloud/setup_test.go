package anxcloud

import (
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())

	// setup test environment
	var env *environment.Info
	var err error

	env, err = environment.InitEnvironment()
	if err != nil {
		log.Fatalf("could not setup environment: %s", err.Error())
	}

	// run tests
	exitCode := m.Run()

	// cleanup
	err = env.CleanUp()
	if err != nil {
		log.Fatalf("could not clean up environment: %s", err.Error())
	}
	os.Exit(exitCode)
}
