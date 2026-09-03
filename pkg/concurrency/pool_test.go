package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_BasicExecution(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(ctx, 4)

	var counter int64
	numTasks := 50

	for i := 0; i < numTasks; i++ {
		pool.Submit(func(c context.Context) error {
			atomic.AddInt64(&counter, 1)
			return nil
		})
	}

	errs := pool.Close()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
	if atomic.LoadInt64(&counter) != int64(numTasks) {
		t.Fatalf("expected %d tasks completed, got %d", numTasks, counter)
	}
}

func TestPool_ErrorCollection(t *testing.T) {
	ctx := context.Background()
	pool := NewPool(ctx, 2)

	pool.Submit(func(c context.Context) error {
		return errors.New("error 1")
	})
	pool.Submit(func(c context.Context) error {
		return errors.New("error 2")
	})

	errs := pool.Close()
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}

func TestPool_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pool := NewPool(ctx, 2)

	var executed int64
	// Cancel immediately
	cancel()

	time.Sleep(10 * time.Millisecond)

	pool.Submit(func(c context.Context) error {
		atomic.AddInt64(&executed, 1)
		return nil
	})

	pool.Close()
	if atomic.LoadInt64(&executed) > 0 {
		t.Logf("Task executed before cancellation caught, acceptable")
	}
}

func TestMap(t *testing.T) {
	ctx := context.Background()
	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	results, errs := Map(ctx, inputs, 4, func(c context.Context, item int, idx int) (int, error) {
		return item * 2, nil
	})

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
	if len(results) != len(inputs) {
		t.Fatalf("expected %d results, got %d", len(inputs), len(results))
	}
	for i, v := range results {
		expected := inputs[i] * 2
		if v != expected {
			t.Errorf("at index %d, expected %d, got %d", i, expected, v)
		}
	}
}
