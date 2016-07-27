package ozero

import (
	"fmt"
	"time"
)

// WorkerFunc defines a function that receives a job and processes it.
type WorkerFunc func(interface{})

// DefaultWorkerID is the default WorkerID when posting jobs without defined WorkerID.
const DefaultWorkerID string = "_DEFAULT"

func (pool *Pool) worker() {
	var job job

	// Catch panics and notify of exit.
	defer func() {
		if r := recover(); nil != r {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			pool.mutex.RLock()
			defer pool.mutex.RUnlock()
			if nil != pool.errorFunc {
				pool.errorFunc(job.Data, err)
			}
		}
		pool.workerExitedCh <- struct{}{}
	}()

	for job = range pool.jobsCh {
		pool.mutex.RLock()
		f := pool.workers[job.WorkerID]
		tries := pool.totalTryCount
		delay := pool.retryTimeout
		shouldRetryFunc := pool.shouldRetryFunc
		pool.mutex.RUnlock()

		if nil != f {
			var err error
			var currentTry = 0
			for {
				err = pool.work(f, job.Data)
				if nil == err { // Executed successfully
					break
				}
				if 0 != tries { // Maximum try count not infinite.
					currentTry++
					if currentTry >= tries { // Maximum try count exceeded.
						break
					}
				}
				if nil != shouldRetryFunc {
					if !shouldRetryFunc(job.Data, err, currentTry-1) {
						break
					}
				}
				time.Sleep(delay)
			}
			if nil != err {
				panic(err)
			}
		}
	}
}

func (pool *Pool) work(f WorkerFunc, data interface{}) (ret error) {
	// Catch panics and notify of exit.
	defer func() {
		if r := recover(); nil != r {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
			ret = err
		}
	}()

	f(data)

	return nil
}

// AddWorkerFunc adds the function to be processed when sending jobs to the default workerId.
func (pool *Pool) AddWorkerFunc(f WorkerFunc) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.workers[DefaultWorkerID] = f
	return pool
}

// AddWorkerFuncForWorkerID adds the function to be processed when sending jobs to the specified workerId.
func (pool *Pool) AddWorkerFuncForWorkerID(workerID string, f WorkerFunc) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.workers[workerID] = f
	return pool
}

// SetErrorFunc sets the function to be executed when a panic occurrs in a worker.
// If nil, nothing gets executed on panic.
func (pool *Pool) SetErrorFunc(f ErrorFunc) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.errorFunc = f
	return pool
}

// SetShouldRetryFunc sets a function to be executed when a panic occurrs, to determine if the job should be retried.
func (pool *Pool) SetShouldRetryFunc(f ShouldRetryFunc) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.shouldRetryFunc = f
	return pool
}
