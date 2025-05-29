// Copyright 2021 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syncx

import (
	"context"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ===================== Cond 条件变量 =====================
// Cond 实现了一个条件变量，是等待或宣布一个事件发生的goroutines的汇合点。
// 绝大多数简单用例, 最好使用 channels 而不是 Cond。
type Cond struct {
	noCopy     noCopy
	L          sync.Locker
	notifyList *notifyList
	checker    unsafe.Pointer
	once       sync.Once
}

// NewCond 返回 关联了 l 的新 Cond。
func NewCond(l sync.Locker) *Cond {
	return &Cond{L: l}
}

// Wait 自动解锁 c.L 并挂起当前调用的 goroutine，直到被唤醒或超时。
func (c *Cond) Wait(ctx context.Context) error {
	c.checkCopy()
	c.checkFirstUse()
	t := c.notifyList.add()
	c.L.Unlock()
	defer c.L.Lock()
	return c.notifyList.wait(ctx, t)
}

// Signal 唤醒一个等待在 c 上的goroutine。
func (c *Cond) Signal() {
	c.checkCopy()
	c.checkFirstUse()
	c.notifyList.notifyOne()
}

// Broadcast 唤醒所有等待在 c 上的goroutine。
func (c *Cond) Broadcast() {
	c.checkCopy()
	c.checkFirstUse()
	c.notifyList.notifyAll()
}

func (c *Cond) checkCopy() {
	if c.checker != unsafe.Pointer(c) &&
		!atomic.CompareAndSwapPointer(&c.checker, nil, unsafe.Pointer(c)) &&
		c.checker != unsafe.Pointer(c) {
		panic("syncx.Cond is copied")
	}
}

func (c *Cond) checkFirstUse() {
	c.once.Do(func() {
		if c.notifyList == nil {
			c.notifyList = newNotifyList()
		}
	})
}

type notifyList struct {
	mu   sync.Mutex
	list *chanList
}

func newNotifyList() *notifyList {
	return &notifyList{
		mu:   sync.Mutex{},
		list: newChanList(),
	}
}

func (l *notifyList) add() *node {
	l.mu.Lock()
	defer l.mu.Unlock()
	el := l.list.alloc()
	l.list.pushBack(el)
	return el
}

func (l *notifyList) wait(ctx context.Context, elem *node) error {
	ch := elem.Value
	defer l.list.free(elem)
	select {
	case <-ctx.Done():
		l.mu.Lock()
		defer l.mu.Unlock()
		select {
		case <-ch:
			if l.list.len() != 0 {
				l.notifyNext()
			}
		default:
			l.list.remove(elem)
		}
		return ctx.Err()
	case <-ch:
		return nil
	}
}

func (l *notifyList) notifyOne() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.list.len() == 0 {
		return
	}
	l.notifyNext()
}

func (l *notifyList) notifyNext() {
	front := l.list.front()
	ch := front.Value
	l.list.remove(front)
	ch <- struct{}{}
}

func (l *notifyList) notifyAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.list.len() != 0 {
		l.notifyNext()
	}
}

type node struct {
	prev  *node
	next  *node
	Value chan struct{}
}

type chanList struct {
	sentinel *node
	size     int
	pool     *sync.Pool
}

func newChanList() *chanList {
	sentinel := &node{}
	sentinel.prev = sentinel
	sentinel.next = sentinel
	return &chanList{
		sentinel: sentinel,
		size:     0,
		pool: &sync.Pool{
			New: func() any {
				return &node{
					Value: make(chan struct{}, 1),
				}
			},
		},
	}
}

func (l *chanList) len() int {
	return l.size
}

func (l *chanList) front() *node {
	return l.sentinel.next
}

func (l *chanList) alloc() *node {
	return l.pool.Get().(*node)
}

func (l *chanList) pushBack(elem *node) {
	elem.next = l.sentinel
	elem.prev = l.sentinel.prev
	l.sentinel.prev.next = elem
	l.sentinel.prev = elem
	l.size++
}

func (l *chanList) remove(elem *node) {
	elem.prev.next = elem.next
	elem.next.prev = elem.prev
	elem.prev = nil
	elem.next = nil
	l.size--
}

func (l *chanList) free(elem *node) {
	l.pool.Put(elem)
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// ===================== Map 泛型并发安全Map =====================
type Map[K comparable, V any] struct {
	m sync.Map
}

// Load 加载键值对
func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	var anyVal any
	anyVal, ok = m.m.Load(key)
	if anyVal != nil {
		value = anyVal.(V)
	}
	return
}

// Store 存储键值对
func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// LoadOrStore 加载或者存储一个键值对
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	var anyVal any
	anyVal, loaded = m.m.LoadOrStore(key, value)
	if anyVal != nil {
		actual = anyVal.(V)
	}
	return
}

// LoadOrStoreFunc 避免无意义的创建实例，适合高性能场景。
func (m *Map[K, V]) LoadOrStoreFunc(key K, fn func() (V, error)) (actual V, loaded bool, err error) {
	val, ok := m.Load(key)
	if ok {
		return val, true, nil
	}
	val, err = fn()
	if err != nil {
		return
	}
	actual, loaded = m.LoadOrStore(key, val)
	return
}

// LoadAndDelete 加载并且删除一个键值对
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	var anyVal any
	anyVal, loaded = m.m.LoadAndDelete(key)
	if anyVal != nil {
		value = anyVal.(V)
	}
	return
}

// Delete 删除键值对
func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Range 遍历, f 不能为 nil
func (m *Map[K, V]) Range(f func(key K, value V) bool) {
	m.m.Range(func(key, value any) bool {
		var (
			k K
			v V
		)
		if value != nil {
			v = value.(V)
		}
		if key != nil {
			k = key.(K)
		}
		return f(k, v)
	})
}

// ===================== Pool 泛型对象池 =====================
type Pool[T any] struct {
	p sync.Pool
}

// NewPool 创建一个 Pool 实例
func NewPool[T any](factory func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return factory()
			},
		},
	}
}

// Get 取出一个元素
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put 放回去一个元素
func (p *Pool[T]) Put(t T) {
	p.p.Put(t)
}

// ===================== LimitPool 限流对象池 =====================
type LimitPool[T any] struct {
	pool   *Pool[T]
	tokens *atomic.Int32
}

// NewLimitPool 创建一个 LimitPool 实例
func NewLimitPool[T any](maxTokens int, factory func() T) *LimitPool[T] {
	var tokens atomic.Int32
	tokens.Add(int32(maxTokens))
	return &LimitPool[T]{
		pool:   NewPool[T](factory),
		tokens: &tokens,
	}
}

// Get 取出一个元素，true 代表从池中取出，false 代表新建
func (l *LimitPool[T]) Get() (T, bool) {
	if l.tokens.Add(-1) < 0 {
		l.tokens.Add(1)
		var zero T
		return zero, false
	}
	return l.pool.Get(), true
}

// Put 放回去一个元素
func (l *LimitPool[T]) Put(t T) {
	l.pool.Put(t)
	l.tokens.Add(1)
}

// ===================== SegmentKeysLock 分段Key锁 =====================
type SegmentKeysLock struct {
	locks []*sync.RWMutex
	size  uint32
}

// NewSegmentKeysLock 创建 SegmentKeysLock 示例
func NewSegmentKeysLock(size uint32) *SegmentKeysLock {
	locks := make([]*sync.RWMutex, size)
	for i := range locks {
		locks[i] = &sync.RWMutex{}
	}
	return &SegmentKeysLock{
		locks: locks,
		size:  size,
	}
}

func (s *SegmentKeysLock) hash(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

// RLock 读锁加锁
func (s *SegmentKeysLock) RLock(key string) {
	s.getLock(key).RLock()
}

// TryRLock 试着加读锁
func (s *SegmentKeysLock) TryRLock(key string) bool {
	return s.getLock(key).TryRLock()
}

// RUnlock 读锁解锁
func (s *SegmentKeysLock) RUnlock(key string) {
	s.getLock(key).RUnlock()
}

// Lock 写锁加锁
func (s *SegmentKeysLock) Lock(key string) {
	s.getLock(key).Lock()
}

// TryLock 试着加锁
func (s *SegmentKeysLock) TryLock(key string) bool {
	return s.getLock(key).TryLock()
}

// Unlock 写锁解锁
func (s *SegmentKeysLock) Unlock(key string) {
	s.getLock(key).Unlock()
}

func (s *SegmentKeysLock) getLock(key string) *sync.RWMutex {
	hash := s.hash(key)
	return s.locks[hash%s.size]
}

// ===================== AtomicValue 泛型原子操作 =====================
// Value 是对 atomic.Value 的泛型封装，适合高并发场景下的安全数据交换。
type Value[T any] struct {
	val atomic.Value
}

// NewValue 创建一个 Value 对象，里面存放着 T 的零值
func NewValue[T any]() *Value[T] {
	var t T
	return NewValueOf[T](t)
}

// NewValueOf 使用传入的值来创建一个 Value 对象
func NewValueOf[T any](t T) *Value[T] {
	val := atomic.Value{}
	val.Store(t)
	return &Value[T]{
		val: val,
	}
}

// Load 加载当前值
func (v *Value[T]) Load() (val T) {
	data := v.val.Load()
	val = data.(T)
	return
}

// Store 存储新值
func (v *Value[T]) Store(val T) {
	v.val.Store(val)
}

// Swap 交换值，返回旧值
func (v *Value[T]) Swap(new T) (old T) {
	data := v.val.Swap(new)
	old = data.(T)
	return
}

// CompareAndSwap CAS操作
func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return v.val.CompareAndSwap(old, new)
}
