// Copyright 2024 ecodeclub
//
// 本文件为 pool.go 的测试用例，覆盖对象池、任务池、带超时任务池的主要功能，符合Go测试规范。

package pool

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试对象池基本功能
func TestSimpleObjectPool_Basic(t *testing.T) {
	pool := NewSimpleObjectPool(func() *int { v := 42; return &v })
	obj := pool.Get()
	assert.Equal(t, 42, *obj)
	*obj = 100
	pool.Put(obj)
	obj2 := pool.Get()
	// 由于sync.Pool的特性，obj2可能是新分配的，也可能是上次Put的
	assert.NotNil(t, obj2)
	pool.Put(obj2)
}

// 测试任务池基本功能
func TestFixedTaskPool_Basic(t *testing.T) {
	pool := NewFixedTaskPool(2)
	var sum int32
	task := func() { atomic.AddInt32(&sum, 1) }
	for i := 0; i < 5; i++ {
		_ = pool.Submit(task)
	}
	pool.Shutdown()
	assert.Equal(t, int32(5), sum)
}

// 测试任务池并发提交
func TestFixedTaskPool_Concurrent(t *testing.T) {
	pool := NewFixedTaskPool(4)
	var sum int32
	task := func() { atomic.AddInt32(&sum, 1) }
	n := 100
	for i := 0; i < n; i++ {
		_ = pool.Submit(task)
	}
	pool.Shutdown()
	assert.Equal(t, int32(n), sum)
}

// 测试任务池关闭后不能再提交
func TestFixedTaskPool_Shutdown(t *testing.T) {
	pool := NewFixedTaskPool(1)
	pool.Shutdown()
	err := pool.Submit(func() {})
	assert.Error(t, err)
}

// 测试带超时的任务池
func TestTimeoutTaskPool_Basic(t *testing.T) {
	pool := NewTimeoutTaskPool(1, 50*time.Millisecond)
	// 正常完成
	err := pool.Submit(func() { time.Sleep(10 * time.Millisecond) })
	assert.NoError(t, err)
	// 超时
	err = pool.Submit(func() { time.Sleep(100 * time.Millisecond) })
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("任务执行超时")))
	pool.Shutdown()
}
