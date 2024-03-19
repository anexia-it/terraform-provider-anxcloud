package environment

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/goombaio/namegenerator"
	testutil "go.anx.io/go-anxcloud/pkg/utils/test"
)

type Info struct {
	TestRunName string
	Location    string
	VlanID      string
	Prefix      Prefix
}

var (
	// consumer tracks how many tests are currently using the environment
	envInfo *Info
)

func (i *Info) setup() error {
	log.Printf("Random Test Name: %s", i.TestRunName)
	prefix, err := CreateTestPrefix(context.Background(), *i)
	if err != nil {
		return err
	}
	i.Prefix = prefix

	return nil
}

func (i *Info) CleanUp() error {
	if i == nil {
		return nil
	}

	return deletePrefix(context.Background(), *i)
}

func GetEnvInfo(t *testing.T) Info {
	if envInfo == nil {
		t.Fatalf("test environment is not setup")
	}
	return *envInfo
}

func shouldRunWithTestEnvironment() bool {
	_, runAcceptanceTest := os.LookupEnv("TF_ACC")
	_, anexiaTokenPresent := os.LookupEnv("ANEXIA_TOKEN")

	return anexiaTokenPresent && runAcceptanceTest
}

var initEnvironmentOnce sync.Once

func InitEnvironment() *Info {
	if !shouldRunWithTestEnvironment() {
		return nil
	}

	var locationID, vlanID string
	var isSet bool
	if locationID, isSet = os.LookupEnv("ANEXIA_LOCATION_ID"); !isSet {
		log.Fatal("'ANEXIA_LOCATION_ID' is not set")
	}
	if vlanID, isSet = os.LookupEnv("ANEXIA_VLAN_ID"); !isSet {
		log.Fatal("'ANEXIA_VLAN_ID' is not set")
	}

	log.Println("Setting up new test environment")

	initEnvironmentOnce.Do(func() {
		testutil.Seed(time.Now().UnixNano())
		// we create a new environment
		envInfo = &Info{
			TestRunName: namegenerator.NewNameGenerator(time.Now().UnixNano()).Generate(),
			VlanID:      vlanID,
			Location:    locationID,
		}

		if err := envInfo.setup(); err != nil {
			log.Fatal(err)
		}
	})

	return envInfo
}

func SkipIfNoEnvironment(t *testing.T) {
	if !shouldRunWithTestEnvironment() {
		t.Skip("Skipping test because either ANEXIA_TOKEN or TF_ACC is not set")
	}
}
