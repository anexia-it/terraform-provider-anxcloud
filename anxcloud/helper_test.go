package anxcloud

import (
	"testing"

	"go.anx.io/go-anxcloud/pkg/client"
)

func integrationTestClientFromEnv(t *testing.T) client.Client {
	c, err := client.New(client.AuthFromEnv(false))
	if err != nil {
		t.Errorf("failed to initialize integration test client from env: %s", err)
	}
	return c
}
