// Package workerpool provides a generic, bounded-concurrency worker pool.
// It replaces unbounded goroutine spawning patterns with a controlled pool
// that respects context cancellation, enforces worker limits, provides
// error aggregation, and supports backpressure for high-throughput scenarios.
package workerpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// JobFunc processes a single job of type T and returns a result of type R.
type JobFunc[T any, R any] func(ctx context.Context, job T) (R, error)

// Result holds the outcome of processing a single job.
type Result[T any, R any] struct {
	Index   int
	Job     T
	Value   R
	Err     error
	Elapsed time.Duration // set when created with WithDurationTracking
}

// AggregateError collects errors from a worker pool run.
// It implements the error interface and supports errors.Is / errors.As.
type AggregateError struct {
	Errors map[int]error
	Total  int
}

// Error returns a human-readable summary of all errors.
func (ae *AggregateError) Error() string {
	if len(ae.Errors) == 0 {
		return "workerpool: 0 of 0 jobs failed"
	}
	return fmt.Sprintf("workerpool: %d of %d jobs failed", len(ae.Errors), ae.Total)
}

// Unwrap supports errors.Is / errors.As chaining through the first error.
func (ae *AggregateError) Unwrap() error {
	for _, e := range ae.Errors {
		return e
	}
	return nil
}

// First returns the first error encountered, or nil if no errors occurred.
func (ae *AggregateError) First() error {
	for _, e := range ae.Errors {
		return e
	}
	return nil
}

// WorkerPool defines the contract for processing jobs with bounded concurrency.
// Implementations must be safe for concurrent use.
type WorkerPool[T any, R any] interface {
	// Process executes fn for each job, returning ordered results.
	// Context cancellation aborts pending and in-flight work.
	// Returns aggregated error if any job fails.
	Process(ctx context.Context, jobs []T, fn JobFunc[T, R]) ([]Result[T, R], error)

	// Workers returns the configured maximum concurrency.
	Workers() int
}

// Option configures the worker pool.
type Option func(*poolConfig)

type poolConfig struct {
	trackDurations bool
	bufferSize     int // 0 = unbuffered (backpressure via channel send)
}

// WithDurationTracking enables per-job elapsed time tracking in results.
func WithDurationTracking() Option {
	return func(c *poolConfig) {
		c.trackDurations = true
	}
}

// WithBufferSize sets the job channel buffer size. A value of 0 (default)
// means the channel is fully buffered to len(jobs). A value of 1 or more
// applies backpressure: the producer blocks when the buffer is full,
// preventing unbounded memory growth for very large job sets.
func WithBufferSize(size int) Option {
	return func(c *poolConfig) {
		c.bufferSize = size
	}
}

// Pool processes jobs of type T using a bounded number of goroutines.
// It implements the WorkerPool interface.
// It is safe for concurrent use.
type Pool[T any, R any] struct {
	workers int
	cfg     poolConfig
}

// New creates a new worker pool with the given maximum number of concurrent workers.
// If workers <= 0, it defaults to 1.
func New[T, R any](workers int, opts ...Option) *Pool[T, R] {
	if workers <= 0 {
		workers = 1
	}
	cfg := poolConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Pool[T, R]{
		workers: workers,
		cfg:     cfg,
	}
}

// Workers returns the configured maximum concurrency.
func (p *Pool[T, R]) Workers() int {
	return p.workers
}

// Process executes fn for each job using the bounded worker pool.
// Results are returned in the same order as the input jobs.
// Context cancellation aborts pending and in-flight work.
// Returns an AggregateError if any job failed.
//
//nolint:funlen,cyclop // worker pool orchestration requires setup + dispatch + aggregation
func (p *Pool[T, R]) Process(ctx context.Context, jobs []T, fn JobFunc[T, R]) ([]Result[T, R], error) {
	n := len(jobs)
	if n == 0 {
		return nil, nil
	}

	results := make([]Result[T, R], n)

	// Determine channel buffer size.
	// bufferSize=0 means fully buffered (old behavior).
	// bufferSize>0 applies backpressure.
	bufSize := n
	if p.cfg.bufferSize > 0 {
		bufSize = p.cfg.bufferSize
		if bufSize > n {
			bufSize = n
		}
	}

	// Use a channel to distribute work to bounded workers.
	type workItem struct {
		idx int
		job T
	}
	work := make(chan workItem, bufSize)

	// Feed jobs into the channel. This goroutine allows backpressure:
	// if the buffer is full, the producer blocks, preventing unbounded
	// memory allocation for large job sets.
	var feedErr atomic.Value
	var feedWg sync.WaitGroup
	feedWg.Add(1)
	go func() {
		defer feedWg.Done()
		defer close(work)
		for i, job := range jobs {
			select {
			case <-ctx.Done():
				feedErr.Store(ctx.Err())
				return
			case work <- workItem{idx: i, job: job}:
			}
		}
	}()

	actualWorkers := p.workers
	if actualWorkers > n {
		actualWorkers = n
	}

	track := p.cfg.trackDurations
	var wg sync.WaitGroup

	for range actualWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range work {
				// Check context before processing each job.
				select {
				case <-ctx.Done():
					results[item.idx] = Result[T, R]{
						Index: item.idx,
						Job:   item.job,
						Err:   ctx.Err(),
					}
					continue
				default:
				}

				if track {
					start := time.Now()
					val, err := fn(ctx, item.job)
					results[item.idx] = Result[T, R]{
						Index:   item.idx,
						Job:     item.job,
						Value:   val,
						Err:     err,
						Elapsed: time.Since(start),
					}
				} else {
					val, err := fn(ctx, item.job)
					results[item.idx] = Result[T, R]{
						Index: item.idx,
						Job:   item.job,
						Value: val,
						Err:   err,
					}
				}
			}
		}()
	}

	wg.Wait()
	feedWg.Wait()

	// Aggregate errors from results.
	aggErr := newAggregateError(results)
	return results, aggErr
}

// newAggregateError scans results and returns an AggregateError if any failed.
func newAggregateError[T, R any](results []Result[T, R]) error {
	var agg AggregateError
	agg.Total = len(results)
	agg.Errors = make(map[int]error, len(results))
	for i, r := range results {
		if r.Err != nil {
			agg.Errors[i] = r.Err
		}
	}
	if len(agg.Errors) == 0 {
		return nil
	}
	return &agg
}

// Compile-time interface compliance check.
var _ WorkerPool[int, int] = (*Pool[int, int])(nil)
