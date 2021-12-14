package recorder

import (
	"context"
	"sync"
)

type Recorder interface {
	Cleanup(ctx context.Context) []error
}

var DefaultRecorder = ResourceRecorder{
	lock:      &sync.Mutex{},
	recorders: nil,
}

type ResourceRecorder struct {
	lock      *sync.Mutex
	recorders []Recorder
}

func (r *ResourceRecorder) RegisterRecorder(recorder Recorder) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.recorders = append(r.recorders, recorder)
}

func (r *ResourceRecorder) CleanupResources(ctx context.Context) []error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if len(r.recorders) == 0 {
		return nil
	}

	var cleanupErrors []error
	// cleanup resource in the reverse order they have been created
	for i := len(r.recorders) - 1; i >= 0; i-- {
		err := r.recorders[i].Cleanup(ctx)
		if len(err) != 0 {
			cleanupErrors = append(cleanupErrors, err...)
		}
	}

	return cleanupErrors
}
