package ozero

type job struct {
	WorkerID string
	Data     interface{}
}

func (pool *Pool) addJob(workerID string, data interface{}) {
	pool.jobsCh <- job{workerID, data}
}

// SendJob sends a new job to the pool to be processed by the default worker.
// It returns inmediately no matter how busy the pool is.
func (pool *Pool) SendJob(data interface{}) {
	go pool.addJob(DefaultWorkerID, data)
}

// SendJobForWorkerID sends a new job to the pool to be processed by the specified worker.
// Does nothing if the worker is not specified.
// It returns inmediately no matter how busy the pool is.
func (pool *Pool) SendJobForWorkerID(workerID string, data interface{}) {
	go pool.addJob(workerID, data)
}
