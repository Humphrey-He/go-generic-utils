// Copyright 2024 Humphrey-He
//
// 本文件为 retry.go 的测试用例，覆盖固定间隔、指数退避、自适应、线程安全包装器等策略，符合Go测试规范。

package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFixedIntervalRetryStrategy_Basic(t *testing.T) {
	s, err := NewFixedIntervalRetryStrategy(time.Millisecond*10, 3)
	assert.NoError(t, err)
	var intervals []time.Duration
	var oks []bool
	for i := 0; i < 5; i++ {
		d, ok := s.Next()
		intervals = append(intervals, d)
		oks = append(oks, ok)
	}
	assert.Equal(t, []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 0, 0}, intervals)
	assert.Equal(t, []bool{true, true, true, false, false}, oks)
}

func TestFixedIntervalRetryStrategy_Unlimited(t *testing.T) {
	s, err := NewFixedIntervalRetryStrategy(time.Millisecond, 0)
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		d, ok := s.Next()
		assert.True(t, ok)
		assert.Equal(t, time.Millisecond, d)
	}
}

func TestExponentialBackoffRetryStrategy_Basic(t *testing.T) {
	s, err := NewExponentialBackoffRetryStrategy(time.Millisecond, 8*time.Millisecond, 4)
	assert.NoError(t, err)
	var got []time.Duration
	for i := 0; i < 6; i++ {
		d, ok := s.Next()
		if !ok {
			break
		}
		got = append(got, d)
	}
	// 1,2,4,8
	assert.Equal(t, []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 4 * time.Millisecond, 8 * time.Millisecond}, got)
}

func TestExponentialBackoffRetryStrategy_MaxInterval(t *testing.T) {
	s, err := NewExponentialBackoffRetryStrategy(time.Millisecond, 2*time.Millisecond, 10)
	assert.NoError(t, err)
	var got []time.Duration
	for i := 0; i < 5; i++ {
		d, ok := s.Next()
		if !ok {
			break
		}
		got = append(got, d)
	}
	assert.Equal(t, []time.Duration{1 * time.Millisecond, 2 * time.Millisecond, 2 * time.Millisecond, 2 * time.Millisecond, 2 * time.Millisecond}, got)
}

func TestAdaptiveTimeoutRetryStrategy_Basic(t *testing.T) {
	base, _ := NewFixedIntervalRetryStrategy(time.Millisecond, 10)
	ad := NewAdaptiveTimeoutRetryStrategy(base, 2, 1)
	// 初始失败数为0，允许重试
	d, ok := ad.Next()
	assert.True(t, ok)
	assert.Equal(t, time.Millisecond, d)
	// 模拟失败，超过阈值
	ad.Report(errors.New("fail"))
	d, ok = ad.Next()
	assert.False(t, ok)
}

func TestThreadSafeStrategy_Concurrent(t *testing.T) {
	base, _ := NewFixedIntervalRetryStrategy(time.Millisecond, 100)
	ts := &ThreadSafeStrategy{strategy: base}
	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ts.Next()
			}
		}()
	}
	wg.Wait()
	// 线程安全性校验：不会panic，重试次数累加
}

func TestRetry_SuccessAndExhausted(t *testing.T) {
	// 业务第一次成功
	base, _ := NewFixedIntervalRetryStrategy(time.Millisecond, 3)
	ctx := context.Background()
	callCount := 0
	err := Retry(ctx, base, func() error {
		callCount++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// 业务一直失败，重试耗尽
	base2, _ := NewFixedIntervalRetryStrategy(time.Millisecond, 2)
	callCount2 := 0
	err = Retry(ctx, base2, func() error {
		callCount2++
		return errors.New("fail")
	})
	assert.Error(t, err)
	assert.GreaterOrEqual(t, callCount2, 2)
}

func TestRetry_ContextCancel(t *testing.T) {
	base, _ := NewFixedIntervalRetryStrategy(time.Second, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := Retry(ctx, base, func() error { return errors.New("fail") })
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 200*time.Millisecond)
}
