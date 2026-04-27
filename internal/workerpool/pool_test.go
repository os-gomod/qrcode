package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_BasicProcessing(t *testing.T) {
	pool := New[int, int](4)
	jobs := []int{1, 2, 3, 4, 5}

	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		return job * 2, nil
	})
	if err != nil {
		t.Fatalf("Process() unexpected error: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Fatalf("result %d: unexpected error: %v", i, r.Err)
		}
		if r.Value != jobs[i]*2 {
			t.Errorf("result %d: expected %d, got %d", i, jobs[i]*2, r.Value)
		}
	}
}

func TestPool_EmptyJobs(t *testing.T) {
	pool := New[int, int](4)
	results, err := pool.Process(context.Background(), nil, func(ctx context.Context, job int) (int, error) {
		return job, nil
	})
	if err != nil {
		t.Error("Process(nil) should not error")
	}
	if results != nil {
		t.Fatalf("expected nil results for empty jobs, got %v", results)
	}
}

func TestPool_SingleJob(t *testing.T) {
	pool := New[int, string](2)
	results, err := pool.Process(context.Background(), []int{42}, func(ctx context.Context, job int) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Value != "ok" {
		t.Fatalf("expected [ok], got %v", results)
	}
}

func TestPool_WorkerBound(t *testing.T) {
	var maxConcurrent atomic.Int32
	pool := New[int, int](3)

	jobs := make([]int, 20)
	for i := range jobs {
		jobs[i] = i
	}

	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		cur := maxConcurrent.Add(1)
		time.Sleep(10 * time.Millisecond) // hold the slot
		maxConcurrent.Store(cur - 1)
		return job, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 20 {
		t.Fatalf("expected 20 results, got %d", len(results))
	}
	// Peak concurrency should not exceed 3.
	peak := maxConcurrent.Load()
	if peak > 3 {
		t.Errorf("peak concurrency %d exceeded worker limit of 3", peak)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pool := New[int, int](2)
	jobs := make([]int, 100)
	for i := range jobs {
		jobs[i] = i
	}

	// Cancel after a short delay.
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	results, err := pool.Process(ctx, jobs, func(ctx context.Context, job int) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return job, nil
		}
	})

	if err == nil {
		t.Error("expected aggregate error due to context cancellation")
	}
	// Verify it's an AggregateError.
	var aggErr *AggregateError
	if !errors.As(err, &aggErr) {
		t.Fatalf("expected AggregateError, got %T: %v", err, err)
	}
	if len(aggErr.Errors) == 0 {
		t.Error("expected some errors due to context cancellation")
	}
	errorCount := 0
	for _, r := range results {
		if r.Err != nil {
			errorCount++
		}
	}
	if errorCount == 0 {
		t.Error("expected some result errors due to context cancellation")
	}
}

func TestPool_JobError(t *testing.T) {
	pool := New[int, int](2)
	jobs := []int{1, 2, 3}

	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		if job == 2 {
			return 0, errors.New("job 2 failed")
		}
		return job, nil
	})

	if err == nil {
		t.Fatal("expected aggregate error for partial failures")
	}
	var aggErr *AggregateError
	if !errors.As(err, &aggErr) {
		t.Fatalf("expected AggregateError, got %T", err)
	}
	if len(aggErr.Errors) != 1 {
		t.Errorf("expected 1 error in aggregate, got %d", len(aggErr.Errors))
	}
	if aggErr.Total != 3 {
		t.Errorf("expected Total=3, got %d", aggErr.Total)
	}

	if results[0].Err != nil || results[0].Value != 1 {
		t.Errorf("result 0: expected value=1 nil-err, got value=%d err=%v", results[0].Value, results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("result 1: expected error for job 2")
	}
	if results[2].Err != nil || results[2].Value != 3 {
		t.Errorf("result 2: expected value=3 nil-err, got value=%d err=%v", results[2].Value, results[2].Err)
	}
}

func TestPool_WithZeroWorkers(t *testing.T) {
	pool := New[int, int](0) // should default to 1
	jobs := []int{1, 2}
	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		return job, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestPool_DurationTracking(t *testing.T) {
	pool := New[int, int](2, WithDurationTracking())
	jobs := []int{1, 2, 3}

	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		time.Sleep(20 * time.Millisecond)
		return job, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, r := range results {
		if r.Elapsed == 0 {
			t.Errorf("result %d: expected non-zero elapsed time", i)
		}
		if r.Elapsed < 15*time.Millisecond {
			t.Errorf("result %d: elapsed %v is suspiciously low", i, r.Elapsed)
		}
	}
}

func TestPool_Workers(t *testing.T) {
	pool := New[int, int](7)
	if pool.Workers() != 7 {
		t.Errorf("expected Workers()=7, got %d", pool.Workers())
	}
}

func TestPool_NegativeWorkers(t *testing.T) {
	pool := New[int, int](-5)
	if pool.Workers() != 1 {
		t.Errorf("expected Workers()=1 for negative input, got %d", pool.Workers())
	}
}

// ---------------------------------------------------------------------------
// New: Error aggregation tests
// ---------------------------------------------------------------------------

func TestAggregateError_Basic(t *testing.T) {
	agg := &AggregateError{
		Errors: map[int]error{0: errors.New("err0"), 2: errors.New("err2")},
		Total:  5,
	}
	str := agg.Error()
	if str == "" {
		t.Error("Error() should not be empty")
	}
	if agg.Total != 5 {
		t.Errorf("Total = %d, want 5", agg.Total)
	}
	if len(agg.Errors) != 2 {
		t.Errorf("len(Errors) = %d, want 2", len(agg.Errors))
	}
}

func TestAggregateError_Empty(t *testing.T) {
	agg := &AggregateError{
		Errors: map[int]error{},
		Total:  3,
	}
	str := agg.Error()
	if str == "" {
		t.Error("Error() should not be empty even with 0 errors")
	}
}

func TestAggregateError_Unwrap(t *testing.T) {
	inner := errors.New("root cause")
	agg := &AggregateError{
		Errors: map[int]error{1: inner},
		Total:  5,
	}
	// Unwrap returns the first error.
	if !errors.Is(agg, inner) {
		t.Error("errors.Is should find the root cause via Unwrap")
	}
}

func TestAggregateError_First(t *testing.T) {
	agg := &AggregateError{
		Errors: map[int]error{2: errors.New("second"), 0: errors.New("first")},
		Total:  3,
	}
	// First returns the first error in map iteration order, which may vary,
	// but should always be non-nil when there are errors.
	if agg.First() == nil {
		t.Error("First() should not return nil when Errors is non-empty")
	}

	emptyAgg := &AggregateError{Errors: map[int]error{}, Total: 2}
	if emptyAgg.First() != nil {
		t.Error("First() should return nil when Errors is empty")
	}
}

func TestProcess_AggregateError_NoFailures(t *testing.T) {
	pool := New[int, int](2)
	jobs := []int{1, 2, 3}
	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		return job * 10, nil
	})
	if err != nil {
		t.Fatalf("expected no error when all succeed, got: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestProcess_AggregateError_AllFailures(t *testing.T) {
	pool := New[int, int](2)
	jobs := []int{1, 2, 3}
	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		return 0, fmt.Errorf("job %d failed", job)
	})
	if err == nil {
		t.Fatal("expected error when all jobs fail")
	}
	var aggErr *AggregateError
	if !errors.As(err, &aggErr) {
		t.Fatalf("expected AggregateError, got %T", err)
	}
	if len(aggErr.Errors) != 3 {
		t.Errorf("expected 3 errors, got %d", len(aggErr.Errors))
	}
	for i, r := range results {
		if r.Err == nil {
			t.Errorf("result %d: expected error", i)
		}
	}
}

// ---------------------------------------------------------------------------
// New: Backpressure tests
// ---------------------------------------------------------------------------

func TestPool_Backpressure(t *testing.T) {
	// Buffer size of 1 means at most 1 job is buffered ahead of workers.
	// With 1 worker, this means the producer will block after enqueueing 1 item
	// (1 buffered + 1 being processed = 2 in flight).
	pool := New[int, int](2, WithBufferSize(1))

	jobs := make([]int, 50)
	for i := range jobs {
		jobs[i] = i
	}

	var processed atomic.Int32
	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		processed.Add(1)
		time.Sleep(time.Millisecond) // slow down workers to create backpressure
		return job, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed.Load() != 50 {
		t.Errorf("expected 50 processed, got %d", processed.Load())
	}
	if len(results) != 50 {
		t.Fatalf("expected 50 results, got %d", len(results))
	}
}

func TestPool_BackpressureCancellation(t *testing.T) {
	// With a tiny buffer and a cancelled context, backpressure should not deadlock.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	pool := New[int, int](1, WithBufferSize(1))
	jobs := make([]int, 100)
	for i := range jobs {
		jobs[i] = i
	}

	results, err := pool.Process(ctx, jobs, func(ctx context.Context, job int) (int, error) {
		// Each job checks context — since context is already cancelled,
		// the job itself should return an error or the pool should aggregate
		// context errors from jobs that were pre-cancelled.
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return job, nil
		}
	})

	// Should return quickly, not deadlock. Either we get errors from the
	// cancelled context or results are produced. The critical requirement is
	// that the function returns without deadlocking.
	_ = results
	_ = err // may or may not have errors depending on timing
}

// ---------------------------------------------------------------------------
// New: Interface compliance
// ---------------------------------------------------------------------------

func TestPool_ImplementsWorkerPool(t *testing.T) {
	var _ WorkerPool[int, int] = New[int, int](4)
}

func TestPool_InterfaceProcess(t *testing.T) {
	var wp WorkerPool[int, string] = New[int, string](3)
	jobs := []int{1, 2}
	results, err := wp.Process(context.Background(), jobs, func(ctx context.Context, job int) (string, error) {
		return fmt.Sprintf("item-%d", job), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Value != "item-1" || results[1].Value != "item-2" {
		t.Errorf("unexpected values: %v", results)
	}
}

// ---------------------------------------------------------------------------
// New: Concurrency correctness tests
// ---------------------------------------------------------------------------

func TestPool_ConcurrentAccess(t *testing.T) {
	pool := New[int, int](4)
	jobs := []int{1, 2, 3}

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
				return job, nil
			})
			if err != nil {
				t.Errorf("concurrent Process error: %v", err)
				return
			}
			if len(results) != 3 {
				t.Errorf("expected 3 results, got %d", len(results))
			}
		}()
	}

	wg.Wait()
}

func TestPool_HighLoad(t *testing.T) {
	pool := New[int, int](8, WithDurationTracking())
	jobs := make([]int, 1000)
	for i := range jobs {
		jobs[i] = i
	}

	start := time.Now()
	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		return job * 2, nil
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1000 {
		t.Fatalf("expected 1000 results, got %d", len(results))
	}
	// Verify order preservation.
	for i, r := range results {
		if r.Value != i*2 {
			t.Errorf("result %d: expected %d, got %d", i, i*2, r.Value)
		}
	}
	// 1000 trivial jobs with 8 workers should complete well under 5 seconds.
	if elapsed > 5*time.Second {
		t.Errorf("high load took too long: %v", elapsed)
	}
}

func TestPool_PartialFailures(t *testing.T) {
	pool := New[int, int](4)
	jobs := []int{1, 2, 3, 4, 5}

	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		if job%2 == 0 {
			return 0, fmt.Errorf("even job %d rejected", job)
		}
		return job, nil
	})

	if err == nil {
		t.Fatal("expected aggregate error for partial failures")
	}
	var aggErr *AggregateError
	if !errors.As(err, &aggErr) {
		t.Fatalf("expected AggregateError, got %T", err)
	}
	if len(aggErr.Errors) != 2 {
		t.Errorf("expected 2 errors (jobs 2,4), got %d", len(aggErr.Errors))
	}
	// Verify successful jobs still have correct values.
	if results[0].Value != 1 || results[0].Err != nil {
		t.Errorf("result 0: expected value=1 nil-err, got value=%d err=%v", results[0].Value, results[0].Err)
	}
	if results[2].Value != 3 || results[2].Err != nil {
		t.Errorf("result 2: expected value=3 nil-err, got value=%d err=%v", results[2].Value, results[2].Err)
	}
	if results[4].Value != 5 || results[4].Err != nil {
		t.Errorf("result 4: expected value=5 nil-err, got value=%d err=%v", results[4].Value, results[4].Err)
	}
	// Verify failed jobs have errors.
	if results[1].Err == nil {
		t.Error("result 1 (job 2) should have error")
	}
	if results[3].Err == nil {
		t.Error("result 3 (job 4) should have error")
	}
}

func TestPool_BufferSize_ClampedToJobCount(t *testing.T) {
	// Buffer size larger than job count should be clamped.
	pool := New[int, int](2, WithBufferSize(100))
	jobs := []int{1, 2, 3}

	results, err := pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
		return job, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}
