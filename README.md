# Ozero

[![Build Status](https://travis-ci.org/ANPez/Ozero.png)](https://travis-ci.org/ANPez/Ozero)
[![GoDoc](https://godoc.org/github.com/ANPez/Ozero?status.svg)](http://godoc.org/github.com/ANPez/Ozero)

Ozero is a goroutine pool for Go, focused on simplicity.

## Goroutine pool

When you create a new _ozero_, you can set the pool size, by using _NewPoolN()_, or you can just let it be the default size, being the CPU count, with _NewPool()_.
The interface has been designed to have a user-friendly style, so all you have to do is _.SetWorkerFunc()_ to handle the jobs in the pool.
From that moment on, you have a ready-to-use pool, to which you can _.SendJob(data)_ any job you want to the pool, and it will be processed by the first available goroutine.

## Sending jobs
You have two functions to send jobs to the pool:
- _SendJob(data)_. It will send the job to the pool and return inmediately, no matter how busy the pool is.
- _SendJobSync(data)_. It will send the job to the pool, waiting until one goroutine gets the job and starts working on it.

Usually, there is no big difference on which method you use, however, there is one little _gotcha_ you need to know.

In the following example, the pool will behave randomly, because some of the goroutines started by .SendJob may have not been initialized yet when your main goroutine gets to _.Close()_. Because of that, some of the jobs may not get processed, because they are being sent to a closed pool.
Sending jobs on a closed pool does not cause a panic, they will just get ignored.

```go
func main() {
	nThreads := 10

	taskPool := ozero.NewPoolN(nThreads).SetWorkerFunc(func(data interface{}) error {
		x := data.(int)
		log.Printf("Data: %d\n", x)
		time.Sleep(time.Second)
        return nil
	})

	before := time.Now()
	for i := 0; i < 20; i++ {
		taskPool.SendJob(i) // Here you should use .SendJobSync()
	}
	taskPool.Close()

	log.Printf("Elapsed %.2f seconds", time.Now().Sub(before).Seconds())
}
```

In the previous example, using _SendJobSync_, being that the pool size is 10, and there are 20 jobs to be processed, each lasting 1 second, the expected total time is 2 seconds.

If you are not going to close the pool, the recommended method to send jobs is _SendJob_, because it will let you send jobs, even if the pool is busy.

## Errors
If your WorkerFunc crashes, a new goroutine is spawned, so you don't have to worry about the pool crashing. Everything is built thread-safe for you.
If you want to catch this crashes, or the errors your workerFunc returns, you can just _.SetErrorFunc()_, and you'll get the data and the error caused.

## Retrying jobs
Often you want to retry a job if it fails. To do this, you have the following functions available:
- _SetTries(n)_. Sets the maximum number of times that a job is retried if it crashes. Set to zero to retry indefinitely.
- _SetRetryDelay(duration)_. Set the time between retries.
- _SetShouldRetryFunc(data, error, retry count)_. You can avoid a job being retried for a specficied error by implementing this funcion and returning false. This is useful if your job might fail in a permanent way, like in a HTTP 404 error, or might fail in a temporary way, like in a HTTP 500 error.

One important note, is that your error func is only called **after** all the retries are being executed, and your _ShouldRetryFunc_ is called after every error or crash.

A common way to use the WorkerFunc is to let it panic on error, letting the pool retry the job.

Finally, you can create as many pools as you want!

## Complete usage example

```go
package main

import (
	"log"
	"time"

	"github.com/ANPez/Ozero"
)

func main() {
    nThreads := 10

	taskPool := ozero.NewPoolN(nThreads).SetWorkerFunc(func(data interface{}) error {
		url := data.(string)
		log.Printf("Downloading URL: %s.", url)
		downloadOrPanic(url)
		log.Printf("Job finished OK")
        return nil
	}).SetErrorFunc(func(data interface{}, err error) {
		log.Printf("Error while processing job in queue")
	}).SetShouldRetryFunc(func(data interface{}, err error, retry int) bool {
		switch err := err.(type) {
		case *types.HTTPError:
			return (err.StatusCode < 400) || (err.StatusCode >= 500)
		}
		return true
	}).SetTries(3).SetRetryDelay(time.Second)
}
```

## License
    Copyright 2016 Antonio Nicolás Pina

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
