package pool

import (
	"errors"
	"sync"
	"time"
)

///////////////////// 泛型对象池 /////////////////////

// ObjectPool 泛型对象池接口
// 适用于高效复用对象，减少GC压力
type ObjectPool[T any] interface {
	Get() T    // 获取对象
	Put(obj T) // 放回对象
	Len() int  // 当前池中对象数量
	Cap() int  // 池容量
}

// SimpleObjectPool 基于sync.Pool的简单对象池
type SimpleObjectPool[T any] struct {
	pool *sync.Pool
	new  func() T
}

// NewSimpleObjectPool 创建一个简单对象池
func NewSimpleObjectPool[T any](newFunc func() T) *SimpleObjectPool[T] {
	return &SimpleObjectPool[T]{
		pool: &sync.Pool{New: func() any { return newFunc() }},
		new:  newFunc,
	}
}

func (p *SimpleObjectPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *SimpleObjectPool[T]) Put(obj T) {
	p.pool.Put(obj)
}

func (p *SimpleObjectPool[T]) Len() int { return 0 } // sync.Pool 无法获取长度
func (p *SimpleObjectPool[T]) Cap() int { return 0 } // sync.Pool 无法获取容量

///////////////////// 任务池（协程池） /////////////////////

// Task 定义任务类型
type Task func()

// TaskPool 任务池接口
type TaskPool interface {
	Submit(task Task) error // 提交任务
	Running() int           // 当前运行中的任务数
	Cap() int               // 池容量
	Shutdown()              // 关闭池
}

// FixedTaskPool 固定容量的任务池
type FixedTaskPool struct {
	tasks   chan Task
	wg      sync.WaitGroup
	cap     int
	closed  chan struct{}
	mu      sync.Mutex
	running int
}

// NewFixedTaskPool 创建固定容量的任务池
func NewFixedTaskPool(capacity int) *FixedTaskPool {
	pool := &FixedTaskPool{
		tasks:  make(chan Task),
		cap:    capacity,
		closed: make(chan struct{}),
	}
	for i := 0; i < capacity; i++ {
		go pool.worker()
	}
	return pool
}

func (p *FixedTaskPool) worker() {
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			p.mu.Lock()
			p.running++
			p.mu.Unlock()
			task()
			p.mu.Lock()
			p.running--
			p.mu.Unlock()
			p.wg.Done()
		case <-p.closed:
			return
		}
	}
}

// Submit 提交任务到池
func (p *FixedTaskPool) Submit(task Task) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case <-p.closed:
		return errors.New("任务池已关闭")
	default:
	}
	p.wg.Add(1)
	p.tasks <- task
	return nil
}

// Running 返回当前运行中的任务数
func (p *FixedTaskPool) Running() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// Cap 返回池容量
func (p *FixedTaskPool) Cap() int {
	return p.cap
}

// Shutdown 关闭任务池，等待所有任务完成
func (p *FixedTaskPool) Shutdown() {
	p.mu.Lock()
	select {
	case <-p.closed:
		p.mu.Unlock()
		return
	default:
		close(p.closed)
		close(p.tasks)
	}
	p.mu.Unlock()
	p.wg.Wait()
}

///////////////////// 扩展：带超时的任务池 /////////////////////

// TimeoutTaskPool 支持任务超时的任务池
type TimeoutTaskPool struct {
	*FixedTaskPool
	timeout time.Duration
}

// NewTimeoutTaskPool 创建带超时的任务池
func NewTimeoutTaskPool(capacity int, timeout time.Duration) *TimeoutTaskPool {
	return &TimeoutTaskPool{
		FixedTaskPool: NewFixedTaskPool(capacity),
		timeout:       timeout,
	}
}

// Submit 提交带超时的任务
func (p *TimeoutTaskPool) Submit(task Task) error {
	done := make(chan struct{})
	wrapped := func() {
		defer close(done)
		task()
	}
	err := p.FixedTaskPool.Submit(wrapped)
	if err != nil {
		return err
	}
	select {
	case <-done:
		return nil
	case <-time.After(p.timeout):
		return errors.New("任务执行超时")
	}
}
