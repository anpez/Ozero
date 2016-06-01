package ozero

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewPoolRunsOk(t *testing.T) {
	c := make(chan struct{})

	pool := NewPool()
	defer pool.Close()

	pool.AddWorkerFunc(func(data interface{}) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		c <- struct{}{}
	})
	pool.SendJob(1)
	assert.NotNil(t, <-c)
}

func TestDoubleCloseDoesNotCrash(t *testing.T) {
	pool := NewPool()
	pool.Close()
	pool.Close()
}

func TestErrorFuncGetsCalledOnPanic(t *testing.T) {
	c := make(chan struct{})

	pool := NewPool()
	defer pool.Close()

	pool.AddWorkerFunc(func(data interface{}) {
		panic("an error")
	}).SetErrorFunc(func(data interface{}, err error) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		c <- struct{}{}
	})
	pool.SendJob(1)

	go func() {
		<-time.After(time.Second)
		assert.Fail(t, "Test got stuck")
		// Force the test to end
		c <- struct{}{}
	}()
	assert.NotNil(t, <-c)
}

func TestFuncGetsRetriedExactly3Times(t *testing.T) {
	const RETRY = 3

	wch := make(chan struct{}, RETRY)
	ech := make(chan struct{})

	pool := NewPool()
	defer pool.Close()

	pool.AddWorkerFunc(func(data interface{}) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		wch <- struct{}{}
		panic("an error")
	}).SetErrorFunc(func(data interface{}, err error) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		ech <- struct{}{}
	}).SetTries(RETRY)

	pool.SendJob(1)

	go func() {
		<-time.After(time.Second)
		assert.Fail(t, "Test got stuck")
		// Force the test to end
		for i := 0; i < RETRY; i++ {
			wch <- struct{}{}
		}
		ech <- struct{}{}
	}()

	// Assert function gets called RETRY times
	for i := 0; i < RETRY; i++ {
		assert.NotNil(t, <-wch)
	}
	// Assert error function got called after retries
	assert.NotNil(t, <-ech)
}
