package ozero

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPoolRunsOk(t *testing.T) {
	c := make(chan struct{})

	pool := NewPool()
	pool.AddWorkerFunc(func(data interface{}) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		c <- struct{}{}
	})
	pool.SendJob(1)
	assert.NotNil(t, <-c)
	pool.Close()
}

func TestDoubleCloseDoesNotCrash(t *testing.T) {
	pool := NewPool()
	pool.Close()
	pool.Close()
}

func TestErrorFuncGetsCalledOnPanic(t *testing.T) {
	c := make(chan struct{})

	pool := NewPool()
	pool.AddWorkerFunc(func(data interface{}) {
		panic("an error")
	}).SetErrorFunc(func(data interface{}, err error) {
		assert.IsType(t, 1, data)
		assert.Equal(t, 1, data)
		c <- struct{}{}
	})
	pool.SendJob(1)
	assert.NotNil(t, <-c)
	pool.Close()
}
