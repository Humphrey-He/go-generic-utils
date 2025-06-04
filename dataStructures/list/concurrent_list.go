// Copyright 2024 Humphrey-He
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package list

import (
	"sync"
)

var (
	_ List[any] = &ConcurrentList[any]{}
)

// ConcurrentList 线程安全的列表实现
// 使用读写锁保护对列表的操作，适合多线程访问场景
// 例如电商中的商品库存管理、订单并发处理等
type ConcurrentList[T any] struct {
	List[T]
	lock sync.RWMutex
}

// NewConcurrentList 创建一个基于ArrayList的并发安全列表
func NewConcurrentList[T any](cap int) *ConcurrentList[T] {
	return &ConcurrentList[T]{
		List: NewArrayList[T](cap),
	}
}

// NewConcurrentLinkedList 创建一个基于LinkedList的并发安全列表
func NewConcurrentLinkedList[T any]() *ConcurrentList[T] {
	return &ConcurrentList[T]{
		List: NewLinkedList[T](),
	}
}

// Get 返回对应下标的元素
func (c *ConcurrentList[T]) Get(index int) (T, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// 直接使用内部List的长度
	l := c.List.Len()
	if index < 0 || index >= l {
		var t T
		return t, NewIndexOutOfRangeError(l, index)
	}

	return c.List.Get(index)
}

// Append 线程安全地追加元素
func (c *ConcurrentList[T]) Append(ts ...T) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.List.Append(ts...)
}

// Add 在特定下标处增加一个新元素
func (c *ConcurrentList[T]) Add(index int, t T) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 直接使用内部List的长度，避免再次获取锁
	l := c.List.Len()
	if index < 0 || index > l {
		return NewIndexOutOfRangeError(l, index)
	}

	return c.List.Add(index, t)
}

// Set 重置 index 位置的值
func (c *ConcurrentList[T]) Set(index int, t T) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 直接使用内部List的长度
	l := c.List.Len()
	if index < 0 || index >= l {
		return NewIndexOutOfRangeError(l, index)
	}

	return c.List.Set(index, t)
}

// Delete 删除目标元素的位置，并且返回该位置的值
func (c *ConcurrentList[T]) Delete(index int) (T, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 直接使用内部List的长度
	l := c.List.Len()
	if index < 0 || index >= l {
		var t T
		return t, NewIndexOutOfRangeError(l, index)
	}

	return c.List.Delete(index)
}

// DeleteValue 线程安全地删除指定值的元素
func (c *ConcurrentList[T]) DeleteValue(t T, equals func(src T, dst T) bool) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.List.DeleteValue(t, equals)
}

// Len 线程安全地获取列表长度
func (c *ConcurrentList[T]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.List.Len()
}

// Cap 线程安全地获取列表容量
func (c *ConcurrentList[T]) Cap() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.List.Cap()
}

// Range 线程安全地遍历列表
func (c *ConcurrentList[T]) Range(fn func(index int, t T) error) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.List.Range(fn)
}

// ReverseRange 线程安全地逆序遍历列表
func (c *ConcurrentList[T]) ReverseRange(fn func(index int, t T) error) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.List.ReverseRange(fn)
}

// AsSlice 线程安全地将列表转换为切片
func (c *ConcurrentList[T]) AsSlice() []T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.List.AsSlice()
}

// Sort 线程安全地对列表进行排序
func (c *ConcurrentList[T]) Sort(less func(a, b T) bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.List.Sort(less)
}

// Filter 线程安全地过滤列表
func (c *ConcurrentList[T]) Filter(predicate func(t T) bool) List[T] {
	c.lock.RLock()
	// 获取当前列表的快照
	snapshot := c.List.AsSlice()
	c.lock.RUnlock()

	// 使用非并发列表做过滤操作
	result := NewArrayList[T](0)
	for _, v := range snapshot {
		if predicate(v) {
			result.Append(v)
		}
	}
	return result
}

// Map 线程安全地转换列表
func (c *ConcurrentList[T]) Map(mapper func(t T) T) List[T] {
	c.lock.RLock()
	// 获取当前列表的快照
	snapshot := c.List.AsSlice()
	c.lock.RUnlock()

	// 使用非并发列表做映射操作
	result := NewArrayList[T](len(snapshot))
	for _, v := range snapshot {
		result.Append(mapper(v))
	}
	return result
}

// Clear 线程安全地清空列表
func (c *ConcurrentList[T]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.List.Clear()
}

// ------- 电商专用并发工具 -------

// ConcurrentInventory 并发安全的库存管理系统
type ConcurrentInventory[T any] struct {
	ConcurrentList[T]
	inventoryLocks map[string]*sync.Mutex // 单品锁，用于精细化控制库存操作
	locksMutex     sync.RWMutex           // 保护锁映射表的锁
}

// NewConcurrentInventory 创建并发库存管理系统
func NewConcurrentInventory[T any](cap int) *ConcurrentInventory[T] {
	return &ConcurrentInventory[T]{
		ConcurrentList: ConcurrentList[T]{
			List: NewArrayList[T](cap),
		},
		inventoryLocks: make(map[string]*sync.Mutex),
	}
}

// LockItem 锁定特定商品的库存操作
// 用于确保同一商品的库存操作串行执行
func (c *ConcurrentInventory[T]) LockItem(itemID string) *sync.Mutex {
	c.locksMutex.RLock()
	lock, exists := c.inventoryLocks[itemID]
	c.locksMutex.RUnlock()

	if exists {
		return lock
	}

	// 创建新锁
	c.locksMutex.Lock()
	defer c.locksMutex.Unlock()

	// 双重检查，防止在获取写锁期间被其他协程创建
	lock, exists = c.inventoryLocks[itemID]
	if exists {
		return lock
	}

	lock = &sync.Mutex{}
	c.inventoryLocks[itemID] = lock
	return lock
}

// UpdateInventory 安全地更新商品库存
// 此方法会同时锁定整个列表和特定商品
func (c *ConcurrentInventory[T]) UpdateInventory(
	itemID string,
	updateFn func(item T) (T, error),
	finder func(t T) bool) error {

	// 先锁定特定商品
	itemLock := c.LockItem(itemID)
	itemLock.Lock()
	defer itemLock.Unlock()

	// 再锁定整个列表
	c.lock.Lock()
	defer c.lock.Unlock()

	// 查找商品
	var itemIndex = -1
	err := c.List.Range(func(i int, t T) error {
		if finder(t) {
			itemIndex = i
			return ErrInvalidArgument // 使用错误中断查找
		}
		return nil
	})

	if err != nil && err != ErrInvalidArgument {
		return err
	}

	if itemIndex == -1 {
		return ErrIndexOutOfRange
	}

	// 获取商品
	item, err := c.List.Get(itemIndex)
	if err != nil {
		return err
	}

	// 更新商品
	updatedItem, err := updateFn(item)
	if err != nil {
		return err
	}

	// 保存更新后的商品
	return c.List.Set(itemIndex, updatedItem)
}

// BatchUpdate 批量更新多个商品
// 此方法仅锁定整个列表，适用于需要原子更新多个商品的场景
func (c *ConcurrentInventory[T]) BatchUpdate(updates map[string]func(T) (T, bool), getID func(T) string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 获取所有商品
	items := c.List.AsSlice()

	// 跟踪需要更新的商品索引和新值
	toUpdate := make(map[int]T)

	// 遍历所有商品，找出需要更新的
	for i, item := range items {
		id := getID(item)
		if updateFn, exists := updates[id]; exists {
			if updated, shouldUpdate := updateFn(item); shouldUpdate {
				toUpdate[i] = updated
			}
		}
	}

	// 应用更新
	for idx, newVal := range toUpdate {
		if err := c.List.Set(idx, newVal); err != nil {
			return err
		}
	}

	return nil
}
