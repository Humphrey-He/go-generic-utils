package set

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// 集合相关错误定义
var (
	ErrNilComparator = errors.New("ggu: 比较器不能为nil")
	ErrKeyNotFound   = errors.New("ggu: 键不存在")
)

// Set 表示集合接口，支持基本的集合操作
type Set[T any] interface {
	// Add 添加元素
	Add(key T)

	// AddIfNotExist 仅当元素不存在时添加，返回是否添加成功
	AddIfNotExist(key T) bool

	// Delete 删除元素
	Delete(key T)

	// Exist 判断元素是否存在
	Exist(key T) bool

	// Keys 返回集合中的所有元素
	Keys() []T

	// Len 返回集合大小
	Len() int

	// Clear 清空集合
	Clear()

	// ForEach 遍历集合中的每个元素
	ForEach(fn func(key T) (cont bool))
}

// MapSet 基于map实现的集合，适用于可比较类型
type MapSet[T comparable] struct {
	m map[T]struct{}
}

// NewMapSet 创建基于map的集合实现
func NewMapSet[T comparable](size int) *MapSet[T] {
	return &MapSet[T]{
		m: make(map[T]struct{}, size),
	}
}

// Add 添加元素
func (s *MapSet[T]) Add(key T) {
	s.m[key] = struct{}{}
}

// AddIfNotExist 仅当元素不存在时添加，返回是否添加成功
func (s *MapSet[T]) AddIfNotExist(key T) bool {
	if _, ok := s.m[key]; ok {
		return false
	}
	s.m[key] = struct{}{}
	return true
}

// Delete 删除元素
func (s *MapSet[T]) Delete(key T) {
	delete(s.m, key)
}

// Exist 判断元素是否存在
func (s *MapSet[T]) Exist(key T) bool {
	_, ok := s.m[key]
	return ok
}

// Keys 返回集合中的所有元素，顺序不保证
func (s *MapSet[T]) Keys() []T {
	res := make([]T, 0, len(s.m))
	for key := range s.m {
		res = append(res, key)
	}
	return res
}

// Len 返回集合大小
func (s *MapSet[T]) Len() int {
	return len(s.m)
}

// Clear 清空集合
func (s *MapSet[T]) Clear() {
	s.m = make(map[T]struct{})
}

// ForEach 遍历集合中的每个元素
func (s *MapSet[T]) ForEach(fn func(key T) (cont bool)) {
	for key := range s.m {
		if !fn(key) {
			break
		}
	}
}

// Union 返回两个集合的并集
func (s *MapSet[T]) Union(other *MapSet[T]) *MapSet[T] {
	result := NewMapSet[T](s.Len() + other.Len())
	for key := range s.m {
		result.Add(key)
	}
	for key := range other.m {
		result.Add(key)
	}
	return result
}

// Intersect 返回两个集合的交集
func (s *MapSet[T]) Intersect(other *MapSet[T]) *MapSet[T] {
	result := NewMapSet[T](min(s.Len(), other.Len()))
	// 遍历较小的集合以提高效率
	if s.Len() <= other.Len() {
		for key := range s.m {
			if other.Exist(key) {
				result.Add(key)
			}
		}
	} else {
		for key := range other.m {
			if s.Exist(key) {
				result.Add(key)
			}
		}
	}
	return result
}

// Difference 返回在当前集合中但不在other集合中的元素
func (s *MapSet[T]) Difference(other *MapSet[T]) *MapSet[T] {
	result := NewMapSet[T](s.Len())
	for key := range s.m {
		if !other.Exist(key) {
			result.Add(key)
		}
	}
	return result
}

// IsSubsetOf 判断当前集合是否为other的子集
func (s *MapSet[T]) IsSubsetOf(other *MapSet[T]) bool {
	if s.Len() > other.Len() {
		return false
	}
	for key := range s.m {
		if !other.Exist(key) {
			return false
		}
	}
	return true
}

// ToSlice 将集合转换为切片
func (s *MapSet[T]) ToSlice() []T {
	return s.Keys()
}

// ToSortedSlice 将集合转换为排序的切片（仅适用于可排序类型）
func ToSortedSlice[T constraints](s *MapSet[T]) []T {
	res := s.ToSlice()
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

// TreeSet 基于比较器的有序集合实现
type TreeSet[T any] struct {
	compare Comparator[T]
	data    []T
	m       map[any]int // 用于快速查找，键是对象的哈希值
}

// NewTreeSet 创建有序集合
func NewTreeSet[T any](compare Comparator[T]) (*TreeSet[T], error) {
	if compare == nil {
		return nil, ErrNilComparator
	}
	return &TreeSet[T]{
		compare: compare,
		data:    make([]T, 0),
		m:       make(map[any]int),
	}, nil
}

// Add 添加元素
func (s *TreeSet[T]) Add(key T) {
	// 检查是否已存在
	for i, item := range s.data {
		cmp := s.compare(key, item)
		if cmp == 0 {
			// 元素已存在，替换
			s.data[i] = key
			return
		} else if cmp < 0 {
			// 找到插入位置
			s.data = append(s.data, key) // 先添加到末尾
			copy(s.data[i+1:], s.data[i:len(s.data)-1])
			s.data[i] = key
			return
		}
	}
	// 添加到末尾
	s.data = append(s.data, key)
}

// AddIfNotExist 仅当元素不存在时添加，返回是否添加成功
func (s *TreeSet[T]) AddIfNotExist(key T) bool {
	for i, item := range s.data {
		cmp := s.compare(key, item)
		if cmp == 0 {
			return false
		} else if cmp < 0 {
			s.data = append(s.data, key)
			copy(s.data[i+1:], s.data[i:len(s.data)-1])
			s.data[i] = key
			return true
		}
	}
	s.data = append(s.data, key)
	return true
}

// Delete 删除元素
func (s *TreeSet[T]) Delete(key T) {
	for i, item := range s.data {
		if s.compare(key, item) == 0 {
			// 找到元素，删除
			s.data = append(s.data[:i], s.data[i+1:]...)
			return
		}
	}
}

// Exist 判断元素是否存在
func (s *TreeSet[T]) Exist(key T) bool {
	for _, item := range s.data {
		if s.compare(key, item) == 0 {
			return true
		}
	}
	return false
}

// Keys 返回有序集合中的所有元素（已排序）
func (s *TreeSet[T]) Keys() []T {
	result := make([]T, len(s.data))
	copy(result, s.data)
	return result
}

// Len 返回集合大小
func (s *TreeSet[T]) Len() int {
	return len(s.data)
}

// Clear 清空集合
func (s *TreeSet[T]) Clear() {
	s.data = make([]T, 0)
}

// ForEach 遍历集合中的每个元素
func (s *TreeSet[T]) ForEach(fn func(key T) (cont bool)) {
	for _, key := range s.data {
		if !fn(key) {
			break
		}
	}
}

// ConcurrentSet 线程安全的集合实现
type ConcurrentSet[T comparable] struct {
	set  *MapSet[T]
	lock sync.RWMutex
}

// NewConcurrentSet 创建线程安全的集合
func NewConcurrentSet[T comparable](size int) *ConcurrentSet[T] {
	return &ConcurrentSet[T]{
		set: NewMapSet[T](size),
	}
}

// Add 添加元素（线程安全）
func (s *ConcurrentSet[T]) Add(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set.Add(key)
}

// AddIfNotExist 仅当元素不存在时添加，返回是否添加成功（线程安全）
func (s *ConcurrentSet[T]) AddIfNotExist(key T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.set.AddIfNotExist(key)
}

// Delete 删除元素（线程安全）
func (s *ConcurrentSet[T]) Delete(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set.Delete(key)
}

// Exist 判断元素是否存在（线程安全）
func (s *ConcurrentSet[T]) Exist(key T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.set.Exist(key)
}

// Keys 返回集合中的所有元素（线程安全）
func (s *ConcurrentSet[T]) Keys() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.set.Keys()
}

// Len 返回集合大小（线程安全）
func (s *ConcurrentSet[T]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.set.Len()
}

// Clear 清空集合（线程安全）
func (s *ConcurrentSet[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set.Clear()
}

// ForEach 遍历集合中的每个元素（线程安全）
func (s *ConcurrentSet[T]) ForEach(fn func(key T) (cont bool)) {
	// 先复制一份数据再遍历，避免长时间持有锁
	keys := s.Keys()
	for _, key := range keys {
		if !fn(key) {
			break
		}
	}
}

// ExpirableSet 带过期时间的集合实现
type ExpirableSet[T comparable] struct {
	data     map[T]time.Time // 值到过期时间的映射
	lock     sync.RWMutex
	interval time.Duration // 清理间隔
	stopCh   chan struct{}
}

// NewExpirableSet 创建带过期时间的集合
func NewExpirableSet[T comparable](cleanInterval time.Duration) *ExpirableSet[T] {
	es := &ExpirableSet[T]{
		data:     make(map[T]time.Time),
		interval: cleanInterval,
		stopCh:   make(chan struct{}),
	}

	// 启动清理协程
	go es.cleanExpired()

	return es
}

// cleanExpired 定期清理过期元素
func (s *ExpirableSet[T]) cleanExpired() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpired()
		case <-s.stopCh:
			return
		}
	}
}

// removeExpired 清理过期元素
func (s *ExpirableSet[T]) removeExpired() {
	now := time.Now()
	s.lock.Lock()
	defer s.lock.Unlock()

	for k, expireTime := range s.data {
		if now.After(expireTime) {
			delete(s.data, k)
		}
	}
}

// Add 添加元素，使用默认过期时间（永不过期）
func (s *ExpirableSet[T]) Add(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 使用最大时间表示永不过期
	s.data[key] = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}

// AddIfNotExist 仅当元素不存在时添加，返回是否添加成功
func (s *ExpirableSet[T]) AddIfNotExist(key T) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, exists := s.data[key]
	if exists && time.Now().Before(s.data[key]) {
		return false
	}

	// 使用最大时间表示永不过期
	s.data[key] = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	return true
}

// AddWithTTL 添加带过期时间的元素
func (s *ExpirableSet[T]) AddWithTTL(key T, ttl time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data[key] = time.Now().Add(ttl)
}

// Delete 删除元素
func (s *ExpirableSet[T]) Delete(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.data, key)
}

// Exist 判断元素是否存在且未过期
func (s *ExpirableSet[T]) Exist(key T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	expireTime, ok := s.data[key]
	return ok && time.Now().Before(expireTime)
}

// Keys 返回所有未过期的元素
func (s *ExpirableSet[T]) Keys() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()

	now := time.Now()
	result := make([]T, 0, len(s.data))

	for k, expireTime := range s.data {
		if now.Before(expireTime) {
			result = append(result, k)
		}
	}

	return result
}

// Len 返回未过期元素的数量
func (s *ExpirableSet[T]) Len() int {
	keys := s.Keys() // 已经过滤掉过期元素
	return len(keys)
}

// Clear 清空集合
func (s *ExpirableSet[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = make(map[T]time.Time)
}

// ForEach 遍历所有未过期的元素
func (s *ExpirableSet[T]) ForEach(fn func(key T) (cont bool)) {
	keys := s.Keys() // 已经过滤掉过期元素
	for _, key := range keys {
		if !fn(key) {
			break
		}
	}
}

// Close 关闭集合，停止清理协程
func (s *ExpirableSet[T]) Close() {
	close(s.stopCh)
}

// GetTTL 获取元素的剩余生存时间，如果元素不存在或已过期，返回-1
func (s *ExpirableSet[T]) GetTTL(key T) time.Duration {
	s.lock.RLock()
	defer s.lock.RUnlock()

	expireTime, ok := s.data[key]
	if !ok {
		return -1
	}

	now := time.Now()
	if now.After(expireTime) {
		return -1
	}

	return expireTime.Sub(now)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// type constraints for sorting
type constraints interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

// Comparator 比较器类型定义
type Comparator[T any] func(a, b T) int

// ComparatorRealNumber 实数比较器
func ComparatorRealNumber[T constraints]() Comparator[T] {
	return func(a, b T) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
}

// ComparatorString 字符串比较器
func ComparatorString() Comparator[string] {
	return func(a, b string) int {
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
}

// 确保所有实现满足Set接口
var (
	_ Set[string] = (*MapSet[string])(nil)
	_ Set[int]    = (*TreeSet[int])(nil)
	_ Set[string] = (*ConcurrentSet[string])(nil)
	_ Set[string] = (*ExpirableSet[string])(nil)
)
