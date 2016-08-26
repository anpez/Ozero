package ozero

import "time"

type job struct {
	Data interface{}
}

func (pool *Pool) addJob(data interface{}) {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	// If closed, do nothing.
	if !pool.closed {
		pool.jobsCh <- job{data}
	}
}

// SendJob sends a new job to the pool to be processed by the worker.
// It returns inmediately no matter how busy the pool is.
func (pool *Pool) SendJob(data interface{}) {
	go pool.addJob(data)
}

// SendJobSync sends a new job to the pool to be processed by the worker.
// It waits until a worker gets the job and then returns.
func (pool *Pool) SendJobSync(data interface{}) {
	pool.addJob(data)
}

// SetTries sets the default amount of times a failing job gets re-executed before giving up and calling error function.
// The default amount of times is 1. Set to zero to retry indefinitely.
func (pool *Pool) SetTries(count int) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.totalTryCount = count

	return pool
}

// SetRetryDelay sets the default timeout after a failing function gets retried.
// Default is retry inmediately.
func (pool *Pool) SetRetryDelay(d time.Duration) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.retryTimeout = d

	return pool
}
