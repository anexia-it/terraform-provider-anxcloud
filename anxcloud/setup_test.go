package anxcloud

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"go.anx.io/go-anxcloud/pkg/client"
	"go.anx.io/go-anxcloud/pkg/ipam/prefix"
	testutil "go.anx.io/go-anxcloud/pkg/utils/test"
)

func TestMain(m *testing.M) {
	testutil.Seed(time.Now().UnixNano())

	//pre cleanup
	cleanupEnvironment()

	// setup test environment
	var env *environment.Info
	var err error

	env, err = environment.InitEnvironment()
	if err != nil {
		log.Fatalf("could not setup environment: %s", err.Error())
	}

	// run tests
	exitCode := m.Run()

	// post cleanup
	err = env.CleanUp()
	if err != nil {
		log.Fatalf("could not clean up environment: %s", err.Error())
	}
	os.Exit(exitCode)
}

func cleanupEnvironment() {
	ctx := context.Background()

	// Init client
	c, err := client.New(client.TokenFromEnv(false))
	if err != nil {
		log.Fatalf("could not int client: %s", err.Error())
		return
	}
	prefixAPI := prefix.NewAPI(c)

	//get all
	prefixes, err := prefixAPI.List(ctx, 1, 100, "")
	if err != nil {
		log.Fatalf("could not List prefixes: %s", err.Error())
		return
	}

	//delete
	for _, p := range prefixes {
		deletePrefix(ctx, p, prefixAPI)
	}
}

func deletePrefix(ctx context.Context, prefix prefix.Summary,
	prefixAPI prefix.API) {

	inf, err := prefixAPI.Get(ctx, prefix.ID)
	if err != nil {
		return
	}

	//only consider active prefixes
	if inf.Status != "Active" {
		return
	}

	wl := map[string]any{
		"tf-acc-test":    struct{}{},
		"tf-acc-tags":    struct{}{},
		"terraform-test": struct{}{},
	}

	for k := range wl {
		if strings.HasPrefix(prefix.CustomerDescription, k) {
			err := prefixAPI.Delete(ctx, prefix.ID)
			if err == nil {
				log.Printf("deleted prefix %s", prefix.CustomerDescription)
			} else {
				log.Fatalf("could not delete prefix '%s'. %s",
					prefix.CustomerDescription, err.Error())
			}
			return
		}
	}
}
