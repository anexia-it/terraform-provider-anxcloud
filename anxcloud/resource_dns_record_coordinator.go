// Package anxcloud provides Terraform resources for Anexia Cloud services.
// This file contains the zone polling coordinator for DNS record operations.
package anxcloud

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.anx.io/go-anxcloud/pkg/api"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

// zonePollingCoordinator manages polling coordination for a specific zone
type zonePollingCoordinator struct {
	zoneName   string
	api        api.API
	mu         sync.Mutex
	pollingCh  chan struct{} // closed when polling completes
	pollingErr error
	refCount   int                // number of records waiting for this coordinator
	cancelPoll context.CancelFunc // allows cancellation of background polling
}

// zonePollingCoordinatorMap tracks active polling coordinators per zone
var zonePollingCoordinatorMap sync.Map

// getZonePollingCoordinator returns or creates a polling coordinator for the given zone
func getZonePollingCoordinator(a api.API, zoneName string) *zonePollingCoordinator {
	anyCoordinator, _ := zonePollingCoordinatorMap.LoadOrStore(zoneName, &zonePollingCoordinator{
		zoneName:  zoneName,
		api:       a,
		pollingCh: make(chan struct{}),
	})

	coordinator := anyCoordinator.(*zonePollingCoordinator)
	coordinator.mu.Lock()
	coordinator.refCount++
	coordinator.mu.Unlock()

	return coordinator
}

// waitForZoneDeployment coordinates polling for zone deployment completion.
// Uses an independent background context for polling to prevent race conditions
// where the first waiter's context cancellation would affect all other waiters.
func (c *zonePollingCoordinator) waitForZoneDeployment(ctx context.Context) error {
	c.mu.Lock()
	// If polling channel is already closed, polling is complete
	select {
	case <-c.pollingCh:
		c.mu.Unlock()
		return c.pollingErr
	default:
	}

	// Check if we need to start polling
	if c.refCount == 1 {
		// Use background context with independent timeout for polling
		// This prevents the first waiter's context from affecting all others
		pollCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		c.cancelPoll = cancel
		go c.pollZoneDeployment(pollCtx)
	}
	c.mu.Unlock()

	// Wait for polling to complete with caller's context
	// Each waiter can timeout independently
	select {
	case <-c.pollingCh:
		return c.pollingErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// pollZoneDeployment performs the actual zone deployment polling
func (c *zonePollingCoordinator) pollZoneDeployment(ctx context.Context) {
	defer close(c.pollingCh)

	// Initial delay to allow validation to start
	time.Sleep(5 * time.Second)

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 10 * time.Second
	b.MaxInterval = 30 * time.Second
	b.MaxElapsedTime = 10 * time.Minute

	c.pollingErr = backoff.Retry(func() error {
		zone := clouddnsv1.Zone{Name: c.zoneName}
		if err := c.api.Get(ctx, &zone); err != nil {
			return backoff.Permanent(err)
		}

		if zone.DeploymentLevel < 100 {
			return fmt.Errorf("waiting for zone deployment to complete: %d%%", zone.DeploymentLevel)
		}

		log.Printf("[DEBUG] Zone polling coordinator: zone %s deployment complete (validation: %d%%, deployment: %d%%)", c.zoneName, zone.ValidationLevel, zone.DeploymentLevel)
		return nil
	}, backoff.WithContext(b, ctx))

	if c.pollingErr != nil {
		log.Printf("[ERROR] Zone polling coordinator: zone %s deployment failed: %v", c.zoneName, c.pollingErr)
	}
}

// release decrements the reference count and cleans up if no more references.
// If this is the last reference, cancels background polling and removes from map.
func (c *zonePollingCoordinator) release() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.refCount--
	if c.refCount <= 0 {
		// Cancel background polling if still running
		if c.cancelPoll != nil {
			c.cancelPoll()
		}
		// Remove from the map when no more references
		zonePollingCoordinatorMap.Delete(c.zoneName)
		log.Printf("[DEBUG] Zone polling coordinator: cleaned up coordinator for zone %s", c.zoneName)
	}
}
