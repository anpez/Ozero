package ozero

import (
	"runtime"
	"sync"
)

// Pool represents a thread (goroutine) pool. All of his methods are thread-safe.
type Pool struct {
	sync.RWMutex
	size           int
	workerExitedCh chan struct{}
	exitCh         chan struct{}
	jobsCh         chan job
	workers        map[string]WorkerFunc
}

// NewPool creates a new pool with predefined size.
// By default, it uses the CPU count.
func NewPool() *Pool {
	return NewPoolN(runtime.NumCPU())
}

// NewPoolN creates a new pool with fixed size.
func NewPoolN(size int) *Pool {
	pool := &Pool{
		size:           size,
		workerExitedCh: make(chan struct{}),
		exitCh:         make(chan struct{}),
		jobsCh:         make(chan job),
		workers:        make(map[string]WorkerFunc),
	}

	go pool.ensureRunning()

	// Launch worker threads
	for i := 0; i < size; i++ {
		pool.launchGoroutine()
	}

	return pool
}

func (pool *Pool) launchGoroutine() {
	go pool.worker()
}

func (pool *Pool) ensureRunning() {
running:
	for {
		select {
		case <-pool.exitCh:
			break running
		case <-pool.workerExitedCh:
			pool.launchGoroutine()
		}
	}
}

// CloseAsync asynchronously closes the pool and waits for the running tasks to end. It returns inmediately.
func (pool *Pool) CloseAsync() {
	go pool.Close()
}

// Close closes the pool inmediately, waiting for the running tasks to end.
func (pool *Pool) Close() {
	pool.Lock()
	defer pool.Unlock()

	// Ensure new goroutines aren't spawned
	pool.exitCh <- struct{}{}

	// Signal all goroutines to end
	close(pool.jobsCh)
	for i := 0; i < pool.size; i++ {
		<-pool.workerExitedCh
	}
}
