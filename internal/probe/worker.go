package probe

import (
	"context"
	"sync"
	"time"

	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// WorkerPool manages concurrent probe workers
type WorkerPool struct {
	parallel    int
	timeout     int
	jobs        chan func() ProbeResult
	results     chan ProbeResult
	wg          sync.WaitGroup
	stopTimeout time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(parallel, timeout, jobCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	// Use jobCount for buffer size to avoid blocking on Submit
	bufferSize := jobCount
	if bufferSize < parallel*2 {
		bufferSize = parallel * 2
	}
	return &WorkerPool{
		parallel:    parallel,
		timeout:     timeout,
		jobs:        make(chan func() ProbeResult, bufferSize),
		results:     make(chan ProbeResult, bufferSize),
		stopTimeout: time.Duration(timeout) * time.Second,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.parallel; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker is the worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for job := range wp.jobs {
		// Check if context is cancelled
		select {
		case <-wp.ctx.Done():
			return
		default:
		}

		result := job()

		// Try to send result, but respect context cancellation
		select {
		case wp.results <- result:
		case <-wp.ctx.Done():
			return
		}
	}
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job func() ProbeResult) {
	wp.jobs <- job
}

// Wait waits for all jobs to complete
func (wp *WorkerPool) Wait() {
	close(wp.jobs)

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workers finished normally
		close(wp.results)
	case <-time.After(wp.stopTimeout):
		// Timeout occurred
		logger.Warn("Worker pool 超时，已等待 %d 秒", wp.timeout)
		wp.cancel() // Cancel all workers

		// Wait a bit for workers to finish
		go func() {
			wp.wg.Wait()
			close(wp.results)
		}()
	}
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan ProbeResult {
	return wp.results
}
