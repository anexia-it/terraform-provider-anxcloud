package utils

import (
	"context"
	"sync"
	"time"
)

type BatchUnitResult[T any] struct {
	Data  T
	Error error
}

type batchUnitRequest[T, U any] struct {
	in T
	ch chan BatchUnitResult[U]
}

type Batcher[T, U any] struct {
	sync.Mutex
	batch     []batchUnitRequest[T, U]
	startWait sync.Once
	Wait      time.Duration
	BatchFunc func(context.Context, []T) []BatchUnitResult[U]
}

func (b *Batcher[T, U]) Process(ctx context.Context, in T) (U, error) {
	b.Lock()
	ch := make(chan BatchUnitResult[U], 1)
	b.batch = append(b.batch, batchUnitRequest[T, U]{in, ch})
	b.startWait.Do(func() { time.AfterFunc(b.Wait, func() { b.processBatch(ctx) }) })
	b.Unlock()

	res := <-ch
	return res.Data, res.Error
}

func (b *Batcher[T, U]) processBatch(ctx context.Context) {
	b.Lock()
	defer b.Unlock()

	in := make([]T, len(b.batch))
	for i, elem := range b.batch {
		in[i] = elem.in
	}

	out := b.BatchFunc(ctx, in)
	for i, elem := range b.batch {
		elem.ch <- out[i]
	}

	b.batch = nil
	b.startWait = sync.Once{}
}
