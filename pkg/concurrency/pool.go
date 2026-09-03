package concurrency

import (
	"context"
	"runtime"
	"sync"
)

// Task represents a unit of work that can return an error.
type Task func(ctx context.Context) error

// Pool manages a bounded set of worker goroutines executing tasks.
type Pool struct {
	workers int
	tasks   chan Task
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	errors  []error
	errMu   sync.Mutex
}

// NewPool creates a new worker pool with the specified number of workers.
// If workers <= 0, runtime.NumCPU() is used.
func NewPool(ctx context.Context, workers int) *Pool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	poolCtx, cancel := context.WithCancel(ctx)
	p := &Pool{
		workers: workers,
		tasks:   make(chan Task, workers*2),
		ctx:     poolCtx,
		cancel:  cancel,
		errors:  make([]error, 0),
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			if err := task(p.ctx); err != nil {
				p.addError(err)
			}
		}
	}
}

func (p *Pool) addError(err error) {
	if err == nil {
		return
	}
	p.errMu.Lock()
	defer p.errMu.Unlock()
	p.errors = append(p.errors, err)
}

// Submit queues a task for execution. It blocks if task buffer is full.
func (p *Pool) Submit(task Task) {
	select {
	case <-p.ctx.Done():
		return
	case p.tasks <- task:
	}
}

// Close closes the task queue and waits for all workers to finish.
// It returns all aggregated errors encountered during execution.
func (p *Pool) Close() []error {
	close(p.tasks)
	p.wg.Wait()
	p.cancel()

	p.errMu.Lock()
	defer p.errMu.Unlock()
	errs := make([]error, len(p.errors))
	copy(errs, p.errors)
	return errs
}

// Map runs a mapper function across a slice of items in parallel, returning results in an array.
func Map[T any, R any](ctx context.Context, items []T, workers int, fn func(ctx context.Context, item T, index int) (R, error)) ([]R, []error) {
	if len(items) == 0 {
		return nil, nil
	}

	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(items) {
		workers = len(items)
	}

	results := make([]R, len(items))
	var errMu sync.Mutex
	var errs []error

	pool := NewPool(ctx, workers)
	for i, item := range items {
		idx := i
		it := item
		pool.Submit(func(c context.Context) error {
			res, err := fn(c, it, idx)
			if err != nil {
				errMu.Lock()
				errs = append(errs, err)
				errMu.Unlock()
				return err
			}
			results[idx] = res
			return nil
		})
	}

	poolErrors := pool.Close()
	_ = poolErrors

	return results, errs
}
