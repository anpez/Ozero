package ozero

// WorkerFunc defines a function that receives a job and processes it.
type WorkerFunc func(interface{})

// DefaultWorkerID is the default WorkerID when posting jobs without defined WorkerID.
const DefaultWorkerID string = "_DEFAULT"

func (pool *Pool) worker() {
	// Catch panics and notify of exit.
	defer func() {
		recover()
		pool.workerExitedCh <- struct{}{}
	}()

	for job := range pool.jobsCh {
		pool.RLock()
		f := pool.workers[job.WorkerID]
		pool.RUnlock()
		if nil != f {
			f(job.Data)
		}
	}
}

// AddWorkerFunc adds the function to be processed when sending jobs to the default workerId.
func (pool *Pool) AddWorkerFunc(f WorkerFunc) *Pool {
	pool.Lock()
	defer pool.Unlock()

	pool.workers["_DEFAULT"] = f
	return pool
}

// AddWorkerFuncForWorkerID adds the function to be processed when sending jobs to the specified workerId.
func (pool *Pool) AddWorkerFuncForWorkerID(workerID string, f WorkerFunc) *Pool {
	pool.Lock()
	defer pool.Unlock()

	pool.workers[workerID] = f
	return pool
}
