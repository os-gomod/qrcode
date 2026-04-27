package workerpool

import (
	"context"
	"testing"
)

// ---------------------------------------------------------------------------
// Benchmarks: small batch
// ---------------------------------------------------------------------------

func BenchmarkPool_SmallBatch(b *testing.B) {
	jobs := make([]int, 10)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](4)
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}

func BenchmarkPool_SmallBatch_DurationTracking(b *testing.B) {
	jobs := make([]int, 10)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](4, WithDurationTracking())
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: large batch
// ---------------------------------------------------------------------------

func BenchmarkPool_LargeBatch(b *testing.B) {
	jobs := make([]int, 1000)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](8)
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}

func BenchmarkPool_LargeBatch_DurationTracking(b *testing.B) {
	jobs := make([]int, 1000)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](8, WithDurationTracking())
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: varying worker counts
// ---------------------------------------------------------------------------

func BenchmarkPool_1Worker(b *testing.B) {
	benchmarkWorkers(b, 1)
}

func BenchmarkPool_2Workers(b *testing.B) {
	benchmarkWorkers(b, 2)
}

func BenchmarkPool_4Workers(b *testing.B) {
	benchmarkWorkers(b, 4)
}

func BenchmarkPool_8Workers(b *testing.B) {
	benchmarkWorkers(b, 8)
}

func BenchmarkPool_16Workers(b *testing.B) {
	benchmarkWorkers(b, 16)
}

func benchmarkWorkers(b *testing.B, workers int) {
	jobs := make([]int, 500)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](workers)
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}

// ---------------------------------------------------------------------------
// Benchmarks: backpressure
// ---------------------------------------------------------------------------

func BenchmarkPool_Backpressure(b *testing.B) {
	jobs := make([]int, 500)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](4, WithBufferSize(1))
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}

func BenchmarkPool_NoBackpressure(b *testing.B) {
	jobs := make([]int, 500)
	for i := range jobs {
		jobs[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := New[int, int](4) // default: fully buffered
		_, _ = pool.Process(context.Background(), jobs, func(ctx context.Context, job int) (int, error) {
			return job * 2, nil
		})
	}
}
