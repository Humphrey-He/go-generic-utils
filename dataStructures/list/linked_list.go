// Copyright 2023 ecodeclub
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
	"math"
)

var (
	_ List[any] = &LinkedList[any]{}
)

// node 双向循环链表结点
type node[T any] struct {
	prev *node[T]
	next *node[T]
	val  T
}

// LinkedList 双向循环链表
// 适合频繁插入和删除的场景，如购物车操作、订单处理等
type LinkedList[T any] struct {
	head   *node[T] // 头哨兵节点
	tail   *node[T] // 尾哨兵节点
	length int      // 链表长度
}

// NewLinkedList 创建一个双向循环链表
func NewLinkedList[T any]() *LinkedList[T] {
	head := &node[T]{}
	tail := &node[T]{next: head, prev: head}
	head.next, head.prev = tail, tail
	return &LinkedList[T]{
		head: head,
		tail: tail,
	}
}

// NewLinkedListOf 将切片转换为双向循环链表, 直接使用了切片元素的值，而没有进行复制
func NewLinkedListOf[T any](ts []T) *LinkedList[T] {
	list := NewLinkedList[T]()
	if err := list.Append(ts...); err != nil {
		panic(err)
	}
	return list
}

// findNode 查找指定索引的节点
// 优化：从头或尾开始查找，选择较近的方向
func (l *LinkedList[T]) findNode(index int) *node[T] {
	var cur *node[T]
	if index <= l.Len()/2 {
		// 从头开始查找
		cur = l.head
		for i := -1; i < index; i++ {
			cur = cur.next
		}
	} else {
		// 从尾开始查找
		cur = l.tail
		for i := l.Len(); i > index; i-- {
			cur = cur.prev
		}
	}

	return cur
}

// Get 获取指定索引的元素
func (l *LinkedList[T]) Get(index int) (T, error) {
	if !l.checkIndex(index) {
		var zeroValue T
		return zeroValue, ErrIndexOutOfRange
	}
	n := l.findNode(index)
	return n.val, nil
}

// checkIndex 检查索引是否有效
func (l *LinkedList[T]) checkIndex(index int) bool {
	return 0 <= index && index < l.Len()
}

// Append 往链表最后添加元素
func (l *LinkedList[T]) Append(ts ...T) error {
	for _, t := range ts {
		node := &node[T]{prev: l.tail.prev, next: l.tail, val: t}
		node.prev.next, node.next.prev = node, node
		l.length++
	}
	return nil
}

// Add 在 LinkedList 下标为 index 的位置插入一个元素
// 当 index 等于 LinkedList 长度等同于 Append
func (l *LinkedList[T]) Add(index int, t T) error {
	if index < 0 || index > l.length {
		return ErrIndexOutOfRange
	}
	if index == l.length {
		return l.Append(t)
	}
	next := l.findNode(index)
	node := &node[T]{prev: next.prev, next: next, val: t}
	node.prev.next, node.next.prev = node, node
	l.length++
	return nil
}

// Set 设置链表中index索引处的值为t
func (l *LinkedList[T]) Set(index int, t T) error {
	if !l.checkIndex(index) {
		return ErrIndexOutOfRange
	}
	node := l.findNode(index)
	node.val = t
	return nil
}

// Delete 删除指定位置的元素
func (l *LinkedList[T]) Delete(index int) (T, error) {
	if !l.checkIndex(index) {
		var zeroValue T
		return zeroValue, ErrIndexOutOfRange
	}
	node := l.findNode(index)
	node.prev.next = node.next
	node.next.prev = node.prev
	node.prev, node.next = nil, nil
	l.length--
	return node.val, nil
}

// DeleteValue 删除指定值的元素，如果找到并删除则返回true
func (l *LinkedList[T]) DeleteValue(t T, equals func(src T, dst T) bool) bool {
	// 从头开始查找，跳过哨兵节点
	for cur := l.head.next; cur != l.tail; cur = cur.next {
		if equals(cur.val, t) {
			// 断开连接
			cur.prev.next = cur.next
			cur.next.prev = cur.prev
			// 清空指针帮助GC
			cur.prev, cur.next = nil, nil
			l.length--
			return true
		}
	}
	return false
}

// Len 返回链表长度
func (l *LinkedList[T]) Len() int {
	return l.length
}

// Cap 返回链表容量（与长度相同）
func (l *LinkedList[T]) Cap() int {
	return l.Len()
}

// Range 遍历链表元素
func (l *LinkedList[T]) Range(fn func(index int, t T) error) error {
	for cur, i := l.head.next, 0; i < l.length; i++ {
		err := fn(i, cur.val)
		if err != nil {
			return err
		}
		cur = cur.next
	}
	return nil
}

// ReverseRange 逆序遍历链表元素
func (l *LinkedList[T]) ReverseRange(fn func(index int, t T) error) error {
	for cur, i := l.tail.prev, l.length-1; i >= 0; i-- {
		err := fn(i, cur.val)
		if err != nil {
			return err
		}
		cur = cur.prev
	}
	return nil
}

// AsSlice 将链表转换为切片
func (l *LinkedList[T]) AsSlice() []T {
	slice := make([]T, l.length)
	for cur, i := l.head.next, 0; i < l.length; i++ {
		slice[i] = cur.val
		cur = cur.next
	}
	return slice
}

// Sort 对链表元素进行排序
// 使用简单的插入排序，对于小型列表效率较高
func (l *LinkedList[T]) Sort(less func(a, b T) bool) {
	if l.length <= 1 {
		return
	}

	// 将链表转为切片，使用系统排序后再重建链表
	slice := l.AsSlice()

	// 冒泡排序实现，可根据需要改为其他算法
	for i := 0; i < len(slice)-1; i++ {
		for j := i + 1; j < len(slice); j++ {
			if less(slice[j], slice[i]) {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}

	// 清空原链表
	l.Clear()

	// 将排序后的切片重新添加到链表
	_ = l.Append(slice...)
}

// Filter 过滤链表，返回符合条件的元素组成的新链表
func (l *LinkedList[T]) Filter(predicate func(t T) bool) List[T] {
	result := NewLinkedList[T]()

	for cur := l.head.next; cur != l.tail; cur = cur.next {
		if predicate(cur.val) {
			result.Append(cur.val)
		}
	}

	return result
}

// Map 对链表中的每个元素应用转换函数，返回转换后的新链表
func (l *LinkedList[T]) Map(mapper func(t T) T) List[T] {
	result := NewLinkedList[T]()

	for cur := l.head.next; cur != l.tail; cur = cur.next {
		result.Append(mapper(cur.val))
	}

	return result
}

// Clear 清空链表
func (l *LinkedList[T]) Clear() {
	// 重置链表结构
	l.head.next = l.tail
	l.tail.prev = l.head
	l.length = 0
}

// LinkedListPaged 支持分页的链表
type LinkedListPaged[T any] struct {
	LinkedList[T]
}

// NewLinkedListPaged 创建一个支持分页的链表
func NewLinkedListPaged[T any]() *LinkedListPaged[T] {
	return &LinkedListPaged[T]{
		LinkedList: *NewLinkedList[T](),
	}
}

// Page 返回指定页的元素
func (l *LinkedListPaged[T]) Page(page, pageSize int) ([]T, int, error) {
	if page < 1 || pageSize < 1 {
		return nil, 0, ErrInvalidArgument
	}

	totalItems := l.Len()
	totalPages := l.TotalPages(pageSize)

	if totalItems == 0 {
		return []T{}, 0, nil
	}

	if page > totalPages {
		return nil, totalPages, ErrInvalidArgument
	}

	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize
	if endIndex > totalItems {
		endIndex = totalItems
	}

	result := make([]T, 0, endIndex-startIndex)

	// 找到起始节点
	cur := l.head
	for i := 0; i <= startIndex; i++ {
		cur = cur.next
	}

	// 收集当前页的节点
	for i := startIndex; i < endIndex; i++ {
		result = append(result, cur.val)
		cur = cur.next
	}

	return result, totalPages, nil
}

// TotalPages 计算总页数
func (l *LinkedListPaged[T]) TotalPages(pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(l.Len()) / float64(pageSize)))
}

// --- 为电商场景提供的专用链表 ---

// OrderNode 订单节点
type OrderNode[T any] struct {
	Value    T
	Priority int    // 优先级，可用于订单处理优先级
	Status   string // 状态，如"待支付"、"待发货"等
	Time     int64  // 时间戳，可用于订单创建时间
}

// PriorityLinkedList 优先级链表，用于订单队列管理
type PriorityLinkedList[T any] struct {
	LinkedList[OrderNode[T]]
}

// NewPriorityLinkedList 创建优先级链表
func NewPriorityLinkedList[T any]() *PriorityLinkedList[T] {
	return &PriorityLinkedList[T]{
		LinkedList: *NewLinkedList[OrderNode[T]](),
	}
}

// AddWithPriority 添加带优先级的元素
func (p *PriorityLinkedList[T]) AddWithPriority(val T, priority int, status string, timestamp int64) error {
	node := OrderNode[T]{
		Value:    val,
		Priority: priority,
		Status:   status,
		Time:     timestamp,
	}

	// 找到合适的插入位置（按优先级降序）
	// 优先级相同时按时间戳升序（先进先出）
	cur := p.head.next
	index := 0

	for ; cur != p.tail; cur = cur.next {
		orderNode := cur.val
		if node.Priority > orderNode.Priority ||
			(node.Priority == orderNode.Priority && node.Time < orderNode.Time) {
			break
		}
		index++
	}

	return p.LinkedList.Add(index, node)
}

// PopHighestPriority 弹出最高优先级元素
func (p *PriorityLinkedList[T]) PopHighestPriority() (T, error) {
	var zero T
	if p.LinkedList.Len() == 0 {
		return zero, ErrEmptyList
	}

	node, err := p.LinkedList.Delete(0)
	if err != nil {
		return zero, err
	}

	return node.Value, nil
}

// FilterByStatus 根据状态过滤元素
func (p *PriorityLinkedList[T]) FilterByStatus(status string) []T {
	result := make([]T, 0)

	_ = p.LinkedList.Range(func(_ int, node OrderNode[T]) error {
		if node.Status == status {
			result = append(result, node.Value)
		}
		return nil
	})

	return result
}

// UpdateStatus 更新指定元素的状态
func (p *PriorityLinkedList[T]) UpdateStatus(equals func(T) bool, newStatus string) bool {
	found := false

	_ = p.LinkedList.Range(func(i int, node OrderNode[T]) error {
		if equals(node.Value) {
			node.Status = newStatus
			_ = p.LinkedList.Set(i, node)
			found = true
			return ErrInvalidArgument // 使用错误中断遍历
		}
		return nil
	})

	return found
}
