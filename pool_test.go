package ozero

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewPoolRunsOk(t *testing.T) {
	c := make(chan struct{})

	pool := NewPool()
	defer pool.Close()

	pool.SetWorkerFunc(func(data interface{}) {
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
	c := make(chan struct{}, 1)

	pool := NewPool()
	defer pool.Close()

	pool.SetWorkerFunc(func(data interface{}) {
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

	wch := make(chan struct{}, RETRY*2)
	ech := make(chan struct{}, 2)

	pool := NewPool()
	defer pool.Close()

	pool.SetWorkerFunc(func(data interface{}) {
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

func TestFuncGetsRetriedAfterADelay(t *testing.T) {
	const RETRY = 2
	const DELAY = 100 * time.Millisecond

	wch := make(chan time.Time, RETRY*2)
	ech := make(chan struct{}, 2)

	pool := NewPool()
	defer pool.Close()

	pool.SetWorkerFunc(func(data interface{}) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		wch <- time.Now()
		panic("an error")
	}).SetErrorFunc(func(data interface{}, err error) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		ech <- struct{}{}
	}).SetTries(RETRY).SetRetryDelay(DELAY)

	pool.SendJob(1)

	go func() {
		<-time.After(time.Second)
		assert.Fail(t, "Test got stuck")
		// Force the test to end
		for i := 0; i < RETRY; i++ {
			wch <- time.Now()
		}
		ech <- struct{}{}
	}()

	// Assert error function got called after retries
	assert.NotNil(t, <-ech)

	// Assert function gets called RETRY times
	t1 := <-wch
	t2 := <-wch
	msg := fmt.Sprintf("Expecting retry delay of at least %dms, got %dms", DELAY/time.Millisecond, t2.Sub(t1)/time.Millisecond)
	assert.True(t, t2.After(t1.Add(DELAY)), msg)
}
