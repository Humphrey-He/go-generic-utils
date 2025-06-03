package maputils

import (
	"sort"
	"sync"

	"golang.org/x/exp/constraints"
)

// ================== 泛型Map ==================

// GenericMap 泛型Map，适合大多数业务场景
type GenericMap[K comparable, V any] struct {
	data map[K]V
}

// NewGenericMap 创建一个空的泛型Map
func NewGenericMap[K comparable, V any]() *GenericMap[K, V] {
	return &GenericMap[K, V]{data: make(map[K]V)}
}

// Set 设置键值对
func (m *GenericMap[K, V]) Set(key K, value V) {
	m.data[key] = value
}

// Get 获取键对应的值，第二个返回值表示是否存在
func (m *GenericMap[K, V]) Get(key K) (V, bool) {
	val, ok := m.data[key]
	return val, ok
}

// Delete 删除键
func (m *GenericMap[K, V]) Delete(key K) {
	delete(m.data, key)
}

// Keys 返回所有键
func (m *GenericMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回所有值
func (m *GenericMap[K, V]) Values() []V {
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// Len 返回元素数量
func (m *GenericMap[K, V]) Len() int {
	return len(m.data)
}

// Map 对每个键值对应用变换函数，返回新的Map
// 由于Go方法不能有自己的类型参数，改为使用接口类型
func (m *GenericMap[K, V]) Map(mapper interface{}) interface{} {
	switch fn := mapper.(type) {
	case func(K, V) (string, int):
		result := NewGenericMap[string, int]()
		for k, v := range m.data {
			k2, v2 := fn(k, v)
			result.Set(k2, v2)
		}
		return result
	case func(K, V) (string, string):
		result := NewGenericMap[string, string]()
		for k, v := range m.data {
			k2, v2 := fn(k, v)
			result.Set(k2, v2)
		}
		return result
	case func(K, V) (int, string):
		result := NewGenericMap[int, string]()
		for k, v := range m.data {
			k2, v2 := fn(k, v)
			result.Set(k2, v2)
		}
		return result
	case func(K, V) (int, int):
		result := NewGenericMap[int, int]()
		for k, v := range m.data {
			k2, v2 := fn(k, v)
			result.Set(k2, v2)
		}
		return result
	default:
		// 默认情况，返回原map的拷贝
		result := NewGenericMap[K, V]()
		for k, v := range m.data {
			result.Set(k, v)
		}
		return result
	}
}

// Filter 过滤Map，返回符合条件的键值对组成的新Map
func (m *GenericMap[K, V]) Filter(predicate func(K, V) bool) *GenericMap[K, V] {
	result := NewGenericMap[K, V]()
	for k, v := range m.data {
		if predicate(k, v) {
			result.Set(k, v)
		}
	}
	return result
}

// ForEach 对每个键值对执行操作
func (m *GenericMap[K, V]) ForEach(action func(K, V)) {
	for k, v := range m.data {
		action(k, v)
	}
}

// Merge 合并两个Map，如果有重复键，后者会覆盖前者
func (m *GenericMap[K, V]) Merge(other *GenericMap[K, V]) *GenericMap[K, V] {
	result := NewGenericMap[K, V]()

	// 复制第一个map
	for k, v := range m.data {
		result.Set(k, v)
	}

	// 添加或覆盖第二个map的内容
	for k, v := range other.data {
		result.Set(k, v)
	}

	return result
}

// Clear 清空Map
func (m *GenericMap[K, V]) Clear() {
	m.data = make(map[K]V)
}

// ================== 线程安全Map ==================

// SyncMap 线程安全的Map，适合高并发场景
type SyncMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

// NewSyncMap 创建线程安全Map
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{data: make(map[K]V)}
}

func (m *SyncMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SyncMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *SyncMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *SyncMap[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *SyncMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// ================== 链表Map（有序Map） ==================

// LinkedMapEntry 链表Map的节点
type LinkedMapEntry[K comparable, V any] struct {
	Key   K
	Value V
	prev  *LinkedMapEntry[K, V]
	next  *LinkedMapEntry[K, V]
}

// LinkedMap 保持插入顺序的Map
type LinkedMap[K comparable, V any] struct {
	data map[K]*LinkedMapEntry[K, V]
	head *LinkedMapEntry[K, V]
	tail *LinkedMapEntry[K, V]
}

// NewLinkedMap 创建链表Map
func NewLinkedMap[K comparable, V any]() *LinkedMap[K, V] {
	return &LinkedMap[K, V]{data: make(map[K]*LinkedMapEntry[K, V])}
}

// Set 插入或更新键值对
func (m *LinkedMap[K, V]) Set(key K, value V) {
	if entry, ok := m.data[key]; ok {
		entry.Value = value
		return
	}
	entry := &LinkedMapEntry[K, V]{Key: key, Value: value}
	m.data[key] = entry
	if m.tail == nil {
		m.head, m.tail = entry, entry
	} else {
		m.tail.next = entry
		entry.prev = m.tail
		m.tail = entry
	}
}

// Get 获取键对应的值
func (m *LinkedMap[K, V]) Get(key K) (V, bool) {
	entry, ok := m.data[key]
	if !ok {
		var zero V
		return zero, false
	}
	return entry.Value, true
}

// Delete 删除键
func (m *LinkedMap[K, V]) Delete(key K) {
	entry, ok := m.data[key]
	if !ok {
		return
	}
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		m.head = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		m.tail = entry.prev
	}
	delete(m.data, key)
}

// Keys 返回插入顺序的所有键
func (m *LinkedMap[K, V]) Keys() []K {
	var keys []K
	for e := m.head; e != nil; e = e.next {
		keys = append(keys, e.Key)
	}
	return keys
}

// Len 返回元素数量
func (m *LinkedMap[K, V]) Len() int {
	return len(m.data)
}

// ================== TreeMap（有序Map，基于红黑树） ==================

// TreeMap 有序Map，适合需要排序的场景
type TreeMap[K interface {
	constraints.Ordered
	comparable
}, V any] struct {
	data map[K]V
	keys []K
}

// NewTreeMap 创建TreeMap
func NewTreeMap[K interface {
	constraints.Ordered
	comparable
}, V any]() *TreeMap[K, V] {
	return &TreeMap[K, V]{
		data: make(map[K]V),
		keys: make([]K, 0),
	}
}

// Set 插入或更新键值对
func (m *TreeMap[K, V]) Set(key K, value V) {
	if _, ok := m.data[key]; !ok {
		// 新键，需要插入到keys切片并保持有序
		m.keys = append(m.keys, key)
		m.sortKeys()
	}
	m.data[key] = value
}

// sortKeys 对keys进行排序
func (m *TreeMap[K, V]) sortKeys() {
	sort.Slice(m.keys, func(i, j int) bool {
		return m.keys[i] < m.keys[j]
	})
}

// Get 获取键对应的值
func (m *TreeMap[K, V]) Get(key K) (V, bool) {
	val, ok := m.data[key]
	return val, ok
}

// Delete 删除键
func (m *TreeMap[K, V]) Delete(key K) {
	delete(m.data, key)
	// 从keys中删除key
	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

// Keys 返回有序的所有键
func (m *TreeMap[K, V]) Keys() []K {
	// 返回已排序的keys
	result := make([]K, len(m.keys))
	copy(result, m.keys)
	return result
}

// Clear 清空TreeMap
func (m *TreeMap[K, V]) Clear() {
	m.data = make(map[K]V)
	m.keys = make([]K, 0)
}

// Len 返回元素数量
func (m *TreeMap[K, V]) Len() int {
	return len(m.data)
}

// ================== 多值Map ==================

// MultiMap 支持一个key对应多个value
type MultiMap[K comparable, V any] struct {
	data map[K][]V
}

// NewMultiMap 创建多值Map
func NewMultiMap[K comparable, V any]() *MultiMap[K, V] {
	return &MultiMap[K, V]{data: make(map[K][]V)}
}

// Add 添加一个值到key
func (m *MultiMap[K, V]) Add(key K, value V) {
	m.data[key] = append(m.data[key], value)
}

// Get 获取key对应的所有值
func (m *MultiMap[K, V]) Get(key K) []V {
	return m.data[key]
}

// Delete 删除key及其所有值
func (m *MultiMap[K, V]) Delete(key K) {
	delete(m.data, key)
}

// RemoveValue 从指定键中删除特定值，如果找到并删除则返回true
func (m *MultiMap[K, V]) RemoveValue(key K, value V) bool {
	values, ok := m.data[key]
	if !ok {
		return false
	}

	for i, v := range values {
		// 这里使用反射进行值比较，实际应用中可能需要自定义相等性比较函数
		if interface{}(v) == interface{}(value) {
			// 删除该元素
			m.data[key] = append(values[:i], values[i+1:]...)
			return true
		}
	}
	return false
}

// AllValues 返回所有键的所有值的平铺列表
func (m *MultiMap[K, V]) AllValues() []V {
	var result []V
	for _, values := range m.data {
		result = append(result, values...)
	}
	return result
}

// Contains 检查键是否包含特定值
func (m *MultiMap[K, V]) Contains(key K, value V) bool {
	values, ok := m.data[key]
	if !ok {
		return false
	}

	for _, v := range values {
		if interface{}(v) == interface{}(value) {
			return true
		}
	}
	return false
}

// Clear 清空MultiMap
func (m *MultiMap[K, V]) Clear() {
	m.data = make(map[K][]V)
}

// Keys 返回所有有值的键
func (m *MultiMap[K, V]) Keys() []K {
	var keys []K
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Len 返回不同键的数量
func (m *MultiMap[K, V]) Len() int {
	return len(m.data)
}
