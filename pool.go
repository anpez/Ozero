package ozero

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// ErrorFunc represents an error handling function for panics happening in workers.
// It receives the data failing in the operation and the error occured.
// It is only called after all the retries failed.
type ErrorFunc func(data interface{}, err error)

// ShouldRetryFunc represents an error handling function for panics happening in workers.
// It receives the data failing in the operation, the error occured and the current retry count, eg. 0 if first time, 1 if second.
// Should return whether the op has to be retried or not.
// It is only called if there are remaining retries left.
type ShouldRetryFunc func(data interface{}, err error, retry int) bool

// Pool represents a thread (goroutine) pool. All of his methods are thread-safe.
type Pool struct {
	mutex           sync.RWMutex
	size            int
	workerExitedCh  chan struct{}
	exitCh          chan struct{}
	jobsCh          chan job
	workers         map[string]WorkerFunc
	errorFunc       ErrorFunc
	shouldRetryFunc ShouldRetryFunc
	closed          bool
	totalTryCount   int
	retryTimeout    time.Duration
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
		totalTryCount:  1,
		errorFunc: func(data interface{}, err error) {
			fmt.Fprint(os.Stderr, err)
		},
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
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// If already closed, do nothing
	if pool.closed {
		return
	}

	// Ensure new goroutines aren't spawned
	pool.exitCh <- struct{}{}

	// Signal all goroutines to end
	close(pool.jobsCh)
	for i := 0; i < pool.size; i++ {
		<-pool.workerExitedCh
	}

	pool.closed = true
}

// GetSize returns the number of threads in the pool.
func (pool *Pool) GetSize() int {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()
	return pool.size
}
