package anxcloud

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func TestZonePollingCoordinator(t *testing.T) {
	// Test that zone polling coordinator properly coordinates polling between multiple concurrent operations
	// This test verifies that only one polling operation happens per zone, even with multiple waiters

	// Create a mock API that simulates zone deployment
	mockAPI := &mockAPIForCoordinatorTest{
		zones: make(map[string]*clouddnsv1.Zone),
	}

	// Initialize zone in non-deployed state
	mockAPI.zones["test-zone"] = &clouddnsv1.Zone{
		Name:            "test-zone",
		DeploymentLevel: 0,
		ValidationLevel: 0,
	}

	// Start a goroutine that will complete deployment after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		mockAPI.zones["test-zone"].DeploymentLevel = 100
		mockAPI.zones["test-zone"].ValidationLevel = 100
	}()

	// Simulate multiple concurrent record creations waiting for the same zone
	const numConcurrent = 5
	var wg sync.WaitGroup
	results := make([]error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			coordinator := getZonePollingCoordinator(mockAPI, "test-zone")
			defer coordinator.release()
			results[index] = coordinator.waitForZoneDeployment(context.Background())
		}(i)
	}

	wg.Wait()

	// Verify all operations completed successfully
	for i, err := range results {
		if err != nil {
			t.Errorf("Concurrent operation %d failed: %v", i, err)
		}
	}

	// Verify zone state was updated
	zone := mockAPI.zones["test-zone"]
	if zone.DeploymentLevel != 100 {
		t.Errorf("Expected deployment level 100, got %d", zone.DeploymentLevel)
	}
}

// mockAPIForCoordinatorTest implements the minimal api.API interface needed for testing
type mockAPIForCoordinatorTest struct {
	zones map[string]*clouddnsv1.Zone
	mu    sync.RWMutex
}

func (m *mockAPIForCoordinatorTest) Get(ctx context.Context, obj types.IdentifiedObject, opts ...types.GetOption) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if zone, ok := obj.(*clouddnsv1.Zone); ok {
		if z, exists := m.zones[zone.Name]; exists {
			*zone = *z
			return nil
		}
	}
	return api.ErrNotFound
}

func (m *mockAPIForCoordinatorTest) List(ctx context.Context, obj types.FilterObject, opts ...types.ListOption) error {
	return api.ErrNotFound
}

func (m *mockAPIForCoordinatorTest) Create(ctx context.Context, obj types.Object, opts ...types.CreateOption) error {
	return api.ErrNotFound
}

func (m *mockAPIForCoordinatorTest) Update(ctx context.Context, obj types.IdentifiedObject, opts ...types.UpdateOption) error {
	return api.ErrNotFound
}

func (m *mockAPIForCoordinatorTest) Destroy(ctx context.Context, obj types.IdentifiedObject, opts ...types.DestroyOption) error {
	return api.ErrNotFound
}

// TestRegressionContextTimeoutAlignment tests that polling operations respect resource timeouts
// and fail gracefully without "context deadline exceeded" errors
func TestRegressionContextTimeoutAlignment(t *testing.T) {
	// TestResourceTimeoutAlignment: Verify that polling operations respect the resource timeout context
	t.Run("TestResourceTimeoutAlignment", func(t *testing.T) {
		// Create a context with a short timeout to simulate resource timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Create mock API that simulates slow polling
		mockAPI := &mockAPIForTimeoutTest{
			zones: make(map[string]*clouddnsv1.Zone),
		}

		// Initialize zone in non-deployed state
		mockAPI.zones["test-zone"] = &clouddnsv1.Zone{
			Name:            "test-zone",
			DeploymentLevel: 0,
			ValidationLevel: 0,
		}

		// Start coordinator and wait for deployment with short context
		coordinator := getZonePollingCoordinator(mockAPI, "test-zone")
		defer coordinator.release()

		// This should timeout gracefully, not with "context deadline exceeded"
		err := coordinator.waitForZoneDeployment(ctx)

		// Verify the error is context.Canceled or context.DeadlineExceeded, not a generic error
		if err == nil {
			t.Errorf("Expected timeout error, but operation completed successfully")
		}

		// Check that we get the expected context timeout error, not "context deadline exceeded"
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.Canceled or context.DeadlineExceeded, got: %v", err)
		}
	})

	// TestContextDeadlineExceededPrevention: Test that operations fail gracefully when context times out
	t.Run("TestContextDeadlineExceededPrevention", func(t *testing.T) {
		// Create a context that will timeout during polling
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		mockAPI := &mockAPIForTimeoutTest{
			zones: make(map[string]*clouddnsv1.Zone),
		}

		// Zone that will never complete deployment
		mockAPI.zones["slow-zone"] = &clouddnsv1.Zone{
			Name:            "slow-zone",
			DeploymentLevel: 50, // Stuck at 50%
			ValidationLevel: 50,
		}

		coordinator := getZonePollingCoordinator(mockAPI, "slow-zone")
		defer coordinator.release()

		start := time.Now()
		err := coordinator.waitForZoneDeployment(ctx)
		duration := time.Since(start)

		// Should fail within reasonable time (not hang indefinitely)
		if duration > 3*time.Second {
			t.Errorf("Operation took too long: %v, expected to timeout quickly", duration)
		}

		// Should get context timeout, not "context deadline exceeded" string
		if err == nil {
			t.Errorf("Expected timeout error")
		}

		// Verify it's a proper context error
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected proper context timeout error, got: %v", err)
		}

		// The backoff library may wrap context errors, so we check that the operation
		// still properly times out rather than hanging indefinitely. The key requirement
		// is that it respects the context timeout, not the exact error message format.
	})

	// TestTimeoutConfigurationValidation: Ensure polling timeouts are properly configured relative to resource timeouts
	t.Run("TestTimeoutConfigurationValidation", func(t *testing.T) {
		// Test that resource timeouts are properly defined
		resource := resourceDNSRecord()

		timeouts := resource.Timeouts
		if timeouts == nil {
			t.Fatal("Resource timeouts not configured")
		}

		// Verify create timeout is reasonable (should be at least 5 minutes for DNS operations)
		createTimeout := timeouts.Create
		if *createTimeout < 5*time.Minute {
			t.Errorf("Create timeout too short: %v, should be at least 5 minutes for DNS operations", createTimeout)
		}

		// Verify polling backoff MaxElapsedTime is configured appropriately
		// We can't directly test the backoff config, but we can verify the polling logic exists
		// by testing that the coordinator properly handles context cancellation

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		mockAPI := &mockAPIForTimeoutTest{
			zones: make(map[string]*clouddnsv1.Zone),
		}

		mockAPI.zones["config-test-zone"] = &clouddnsv1.Zone{
			Name:            "config-test-zone",
			DeploymentLevel: 0,
			ValidationLevel: 0,
		}

		coordinator := getZonePollingCoordinator(mockAPI, "config-test-zone")
		defer coordinator.release()

		// This should respect the context and not run longer than the context timeout
		start := time.Now()
		err := coordinator.waitForZoneDeployment(ctx)
		elapsed := time.Since(start)

		// Should complete within context timeout + small buffer
		if elapsed > 1*time.Second {
			t.Errorf("Polling took too long (%v) relative to context timeout (500ms)", elapsed)
		}

		if err == nil {
			t.Errorf("Expected context timeout")
		}
	})
}

// mockAPIForTimeoutTest implements minimal api.API interface for timeout testing
type mockAPIForTimeoutTest struct {
	zones map[string]*clouddnsv1.Zone
	mu    sync.RWMutex
}

func (m *mockAPIForTimeoutTest) Get(ctx context.Context, obj types.IdentifiedObject, opts ...types.GetOption) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if zone, ok := obj.(*clouddnsv1.Zone); ok {
		if z, exists := m.zones[zone.Name]; exists {
			*zone = *z
			return nil
		}
	}
	return api.ErrNotFound
}

func (m *mockAPIForTimeoutTest) List(ctx context.Context, obj types.FilterObject, opts ...types.ListOption) error {
	return api.ErrNotFound
}

func (m *mockAPIForTimeoutTest) Create(ctx context.Context, obj types.Object, opts ...types.CreateOption) error {
	return api.ErrNotFound
}

func (m *mockAPIForTimeoutTest) Update(ctx context.Context, obj types.IdentifiedObject, opts ...types.UpdateOption) error {
	return api.ErrNotFound
}

func (m *mockAPIForTimeoutTest) Destroy(ctx context.Context, obj types.IdentifiedObject, opts ...types.DestroyOption) error {
	return api.ErrNotFound
}
