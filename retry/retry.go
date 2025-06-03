// Copyright 2024 Humphrey-He
//
// 本文件整合了常用重试工具，适合电商平台后端高并发、线程安全等场景。
// 包含固定间隔、指数退避、自适应等重试策略，并支持线程安全扩展。

package retry

import (
	"context"
	"errors"
	"math"
	"math/bits"
	"sync"
	"sync/atomic"
	"time"
)

// ================== 重试策略接口 ==================

// Strategy 定义重试策略接口
type Strategy interface {
	// Next 返回下一次重试的间隔，如果不需要继续重试，第二参数返回 false
	Next() (time.Duration, bool)
	// Report 用于上报本次重试的结果（可选实现）
	Report(err error) Strategy
}

// ================== 固定间隔重试 ==================

// FixedIntervalRetryStrategy 固定间隔重试策略，线程安全
type FixedIntervalRetryStrategy struct {
	maxRetries int32         // 最大重试次数，<=0 表示无限重试
	interval   time.Duration // 重试间隔
	retries    int32         // 当前重试次数
}

func NewFixedIntervalRetryStrategy(interval time.Duration, maxRetries int32) (*FixedIntervalRetryStrategy, error) {
	if interval <= 0 {
		return nil, errors.New("无效的间隔时间，必须大于0")
	}
	return &FixedIntervalRetryStrategy{
		maxRetries: maxRetries,
		interval:   interval,
	}, nil
}

func (s *FixedIntervalRetryStrategy) Next() (time.Duration, bool) {
	retries := atomic.AddInt32(&s.retries, 1)
	if s.maxRetries <= 0 || retries <= s.maxRetries {
		return s.interval, true
	}
	return 0, false
}

func (s *FixedIntervalRetryStrategy) Report(err error) Strategy {
	return s
}

// ================== 指数退避重试 ==================

// ExponentialBackoffRetryStrategy 指数退避重试策略，线程安全
type ExponentialBackoffRetryStrategy struct {
	initialInterval    time.Duration
	maxInterval        time.Duration
	maxRetries         int32
	retries            int32
	maxIntervalReached atomic.Value
}

func NewExponentialBackoffRetryStrategy(initialInterval, maxInterval time.Duration, maxRetries int32) (*ExponentialBackoffRetryStrategy, error) {
	if initialInterval <= 0 {
		return nil, errors.New("无效的间隔时间，必须大于0")
	} else if initialInterval > maxInterval {
		return nil, errors.New("最大重试间隔应大于等于初始重试间隔")
	}
	return &ExponentialBackoffRetryStrategy{
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		maxRetries:      maxRetries,
	}, nil
}

func (s *ExponentialBackoffRetryStrategy) Next() (time.Duration, bool) {
	retries := atomic.AddInt32(&s.retries, 1)
	if s.maxRetries <= 0 || retries <= s.maxRetries {
		if reached, ok := s.maxIntervalReached.Load().(bool); ok && reached {
			return s.maxInterval, true
		}
		interval := s.initialInterval * time.Duration(math.Pow(2, float64(retries-1)))
		if interval <= 0 || interval > s.maxInterval {
			s.maxIntervalReached.Store(true)
			return s.maxInterval, true
		}
		return interval, true
	}
	return 0, false
}

func (s *ExponentialBackoffRetryStrategy) Report(err error) Strategy {
	return s
}

// ================== 自适应超时重试 ==================

// AdaptiveTimeoutRetryStrategy 自适应超时重试策略，线程安全
type AdaptiveTimeoutRetryStrategy struct {
	strategy   Strategy // 基础重试策略
	threshold  int      // 超时比率阈值
	ringBuffer []uint64 // 比特环（滑动窗口存储超时信息）
	reqCount   uint64   // 请求数量
	bufferLen  int      // 滑动窗口长度
	bitCnt     uint64   // 比特位总数
}

func NewAdaptiveTimeoutRetryStrategy(strategy Strategy, bufferLen, threshold int) *AdaptiveTimeoutRetryStrategy {
	return &AdaptiveTimeoutRetryStrategy{
		strategy:   strategy,
		threshold:  threshold,
		bufferLen:  bufferLen,
		ringBuffer: make([]uint64, bufferLen),
		bitCnt:     uint64(64) * uint64(bufferLen),
	}
}

func (s *AdaptiveTimeoutRetryStrategy) Next() (time.Duration, bool) {
	failCount := s.getFailed()
	if failCount >= s.threshold {
		return 0, false
	}
	return s.strategy.Next()
}

func (s *AdaptiveTimeoutRetryStrategy) Report(err error) Strategy {
	if err == nil {
		s.markSuccess()
	} else {
		s.markFail()
	}
	return s
}

func (s *AdaptiveTimeoutRetryStrategy) markSuccess() {
	count := atomic.AddUint64(&s.reqCount, 1)
	count = count % s.bitCnt
	idx := count >> 6
	bitPos := count & 63
	old := atomic.LoadUint64(&s.ringBuffer[idx])
	atomic.StoreUint64(&s.ringBuffer[idx], old&^(uint64(1)<<bitPos))
}

func (s *AdaptiveTimeoutRetryStrategy) markFail() {
	count := atomic.AddUint64(&s.reqCount, 1)
	count = count % s.bitCnt
	idx := count >> 6
	bitPos := count & 63
	old := atomic.LoadUint64(&s.ringBuffer[idx])
	atomic.StoreUint64(&s.ringBuffer[idx], old|(uint64(1)<<bitPos))
}

func (s *AdaptiveTimeoutRetryStrategy) getFailed() int {
	var failCount int
	for i := 0; i < len(s.ringBuffer); i++ {
		v := atomic.LoadUint64(&s.ringBuffer[i])
		failCount += bits.OnesCount64(v)
	}
	return failCount
}

// ================== 线程安全包装器（装饰器）==================

// ThreadSafeStrategy 装饰器，为任意Strategy增加互斥锁，保证多协程安全
type ThreadSafeStrategy struct {
	mu       sync.Mutex
	strategy Strategy
}

func NewThreadSafeStrategy(s Strategy) *ThreadSafeStrategy {
	return &ThreadSafeStrategy{strategy: s}
}

func (t *ThreadSafeStrategy) Next() (time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.strategy.Next()
}

func (t *ThreadSafeStrategy) Report(err error) Strategy {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.strategy.Report(err)
}

// ================== 通用重试入口 ==================

// Retry 通用重试函数，支持上下文取消、超时、最大重试次数等
// bizFunc 返回 nil 表示成功，否则会根据策略重试
func Retry(ctx context.Context, s Strategy, bizFunc func() error) error {
	var ticker *time.Ticker
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()
	for {
		err := bizFunc()
		if err == nil {
			return nil
		}
		duration, ok := s.Next()
		if !ok {
			return errors.New("重试耗尽")
		}
		if ticker == nil {
			ticker = time.NewTicker(duration)
		} else {
			ticker.Reset(duration)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
