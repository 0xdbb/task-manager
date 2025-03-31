package worker

import (
	"context"
	"sync"
	"time"
)

// Worker must be implemented by types that want to use
// the work pool.
type Worker interface {
	Task(ctx context.Context)
}

// Pool provides a pool of goroutines that can execute any Worker
// tasks that are submitted.
type Pool struct {
	work    chan Worker
	wg      sync.WaitGroup
	timeout time.Duration
}

// New creates a new work pool with a timeout for tasks.
func New(maxGoroutines int, timeout time.Duration) *Pool {
	p := Pool{
		work:    make(chan Worker),
		timeout: timeout,
	}

	p.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.work {
				// Create a context with a timeout
				ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
				w.Task(ctx) // Pass context to worker task
				cancel()    // Ensure we release resources
			}
			p.wg.Done()
		}()
	}

	return &p
}

// Run submits work to the pool.
func (p *Pool) Run(w Worker) {
	p.work <- w
}

// Shutdown waits for all the goroutines to shut down.
func (p *Pool) Shutdown() {
	close(p.work)
	p.wg.Wait()
}

