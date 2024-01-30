package provider

import (
	"log"
	"os"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
)

func TestMain(m *testing.M) {
	env := environment.InitEnvironment()

	// run tests
	exitCode := m.Run()

	// cleanup

	if err := env.CleanUp(); err != nil {
		log.Fatalf("could not clean up environment: %s", err.Error())
	}
	os.Exit(exitCode)
}
