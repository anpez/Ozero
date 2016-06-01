package ozero

import "fmt"

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
		pool.mutex.RUnlock()
		if nil != f {
			f(job.Data)
		}
	}
}

// AddWorkerFunc adds the function to be processed when sending jobs to the default workerId.
func (pool *Pool) AddWorkerFunc(f WorkerFunc) *Pool {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	pool.workers["_DEFAULT"] = f
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
