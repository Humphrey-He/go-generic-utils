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
	"math"
	"sort"
)

var (
	_ List[any] = &ArrayList[any]{}
)

// ArrayList 基于切片的列表实现
// 适合频繁随机访问的场景，如电商商品列表展示、商品库存管理等
type ArrayList[T any] struct {
	vals []T
}

// NewArrayList 初始化一个长度为0，容量为cap的ArrayList
func NewArrayList[T any](cap int) *ArrayList[T] {
	return &ArrayList[T]{vals: make([]T, 0, cap)}
}

// NewArrayListOf 基于已有切片创建ArrayList，不会复制切片内容
func NewArrayListOf[T any](ts []T) *ArrayList[T] {
	return &ArrayList[T]{
		vals: ts,
	}
}

// Get 获取指定索引的元素
func (a *ArrayList[T]) Get(index int) (t T, e error) {
	l := a.Len()
	if index < 0 || index >= l {
		return t, NewIndexOutOfRangeError(l, index)
	}
	return a.vals[index], e
}

// Append 往ArrayList末尾追加数据
func (a *ArrayList[T]) Append(ts ...T) error {
	a.vals = append(a.vals, ts...)
	return nil
}

// Add 在ArrayList下标为index的位置插入一个元素
// 当index等于ArrayList长度等同于append
func (a *ArrayList[T]) Add(index int, t T) error {
	l := a.Len()
	if index < 0 || index > l {
		return NewIndexOutOfRangeError(l, index)
	}

	if index == l {
		return a.Append(t)
	}

	// 扩容
	a.vals = append(a.vals, *new(T))
	// 移动元素腾出位置
	copy(a.vals[index+1:], a.vals[index:])
	// 设置值
	a.vals[index] = t
	return nil
}

// Set 设置ArrayList里index位置的值为t
func (a *ArrayList[T]) Set(index int, t T) error {
	length := len(a.vals)
	if index >= length || index < 0 {
		return NewIndexOutOfRangeError(length, index)
	}
	a.vals[index] = t
	return nil
}

// Delete 删除指定位置的元素并返回该元素
// 方法会在必要的时候引起缩容，其缩容规则是：
// - 如果容量 > 2048，并且长度小于容量一半，那么就会缩容为原本的 5/8
// - 如果容量 (64, 2048]，如果长度是容量的 1/4，那么就会缩容为原本的一半
// - 如果此时容量 <= 64，那么我们将不会执行缩容。在容量很小的情况下，浪费的内存很少，所以没必要消耗 CPU去执行缩容
func (a *ArrayList[T]) Delete(index int) (T, error) {
	var t T
	length := len(a.vals)

	if index < 0 || index >= length {
		return t, NewIndexOutOfRangeError(length, index)
	}

	// 保存要删除的元素
	t = a.vals[index]

	// 删除元素
	a.vals = append(a.vals[:index], a.vals[index+1:]...)

	// 执行缩容操作
	a.shrink()

	return t, nil
}

// DeleteValue 删除指定值的元素，如果找到并删除则返回true
func (a *ArrayList[T]) DeleteValue(t T, equals func(src T, dst T) bool) bool {
	for i, v := range a.vals {
		if equals(v, t) {
			_, err := a.Delete(i)
			return err == nil
		}
	}
	return false
}

// shrink 数组缩容
func (a *ArrayList[T]) shrink() {
	length := len(a.vals)
	capacity := cap(a.vals)

	// 容量很小时不进行缩容
	if capacity <= 64 {
		return
	}

	// 计算缩容后的新容量
	var newCap int
	if capacity > 2048 && length < capacity/2 {
		newCap = int(float64(capacity) * 0.625) // 缩为原来的 5/8
	} else if capacity <= 2048 && length <= capacity/4 {
		newCap = capacity / 2 // 缩为原来的一半
	} else {
		return
	}

	// 创建新切片并复制数据
	newSlice := make([]T, length, newCap)
	copy(newSlice, a.vals)
	a.vals = newSlice
}

// Len 返回列表长度
func (a *ArrayList[T]) Len() int {
	return len(a.vals)
}

// Cap 返回列表容量
func (a *ArrayList[T]) Cap() int {
	return cap(a.vals)
}

// Range 遍历列表元素
func (a *ArrayList[T]) Range(fn func(index int, t T) error) error {
	for i, v := range a.vals {
		if err := fn(i, v); err != nil {
			return err
		}
	}
	return nil
}

// ReverseRange 逆序遍历列表元素
func (a *ArrayList[T]) ReverseRange(fn func(index int, t T) error) error {
	for i := len(a.vals) - 1; i >= 0; i-- {
		if err := fn(i, a.vals[i]); err != nil {
			return err
		}
	}
	return nil
}

// AsSlice 将列表转换为切片
func (a *ArrayList[T]) AsSlice() []T {
	res := make([]T, len(a.vals))
	copy(res, a.vals)
	return res
}

// Sort 对列表元素进行排序
func (a *ArrayList[T]) Sort(less func(a, b T) bool) {
	sort.Slice(a.vals, func(i, j int) bool {
		return less(a.vals[i], a.vals[j])
	})
}

// Filter 过滤列表，返回符合条件的元素组成的新列表
func (a *ArrayList[T]) Filter(predicate func(t T) bool) List[T] {
	result := NewArrayList[T](0)
	for _, v := range a.vals {
		if predicate(v) {
			result.Append(v)
		}
	}
	return result
}

// Map 对列表中的每个元素应用转换函数，返回转换后的新列表
func (a *ArrayList[T]) Map(mapper func(t T) T) List[T] {
	result := NewArrayList[T](len(a.vals))
	for _, v := range a.vals {
		result.Append(mapper(v))
	}
	return result
}

// Clear 清空列表
func (a *ArrayList[T]) Clear() {
	a.vals = make([]T, 0, cap(a.vals))
}

// ArrayListPaged 支持分页的ArrayList
type ArrayListPaged[T any] struct {
	ArrayList[T]
}

// NewArrayListPaged 创建一个支持分页的ArrayList
func NewArrayListPaged[T any](cap int) *ArrayListPaged[T] {
	return &ArrayListPaged[T]{
		ArrayList: ArrayList[T]{vals: make([]T, 0, cap)},
	}
}

// Page 返回指定页的元素
// page 从1开始计数
// pageSize 为每页元素个数
// 返回指定页的元素切片和总页数
func (a *ArrayListPaged[T]) Page(page, pageSize int) ([]T, int, error) {
	if page < 1 || pageSize < 1 {
		return nil, 0, ErrInvalidArgument
	}

	totalItems := a.Len()
	totalPages := a.TotalPages(pageSize)

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

	result := make([]T, endIndex-startIndex)
	copy(result, a.vals[startIndex:endIndex])

	return result, totalPages, nil
}

// TotalPages 计算总页数
func (a *ArrayListPaged[T]) TotalPages(pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(a.Len()) / float64(pageSize)))
}

// ArrayListSorted 有序ArrayList实现
type ArrayListSorted[T any] struct {
	ArrayList[T]
	less func(a, b T) bool // 默认的比较函数
}

// NewArrayListSorted 创建一个有序ArrayList
func NewArrayListSorted[T any](cap int, less func(a, b T) bool) *ArrayListSorted[T] {
	return &ArrayListSorted[T]{
		ArrayList: ArrayList[T]{vals: make([]T, 0, cap)},
		less:      less,
	}
}

// InsertSorted 插入元素并保持列表有序
func (a *ArrayListSorted[T]) InsertSorted(t T, less func(a, b T) bool) error {
	if less == nil {
		less = a.less
	}

	// 找到插入位置
	index := sort.Search(len(a.vals), func(i int) bool {
		return !less(a.vals[i], t)
	})

	return a.Add(index, t)
}

// Find 查找元素，返回其索引
func (a *ArrayListSorted[T]) Find(t T, equals func(a, b T) bool) int {
	for i, v := range a.vals {
		if equals(v, t) {
			return i
		}
	}
	return -1
}

// BinarySearch 二分查找元素，返回其索引
func (a *ArrayListSorted[T]) BinarySearch(t T, comparator func(a, b T) int) int {
	low, high := 0, len(a.vals)-1

	for low <= high {
		mid := (low + high) / 2
		cmp := comparator(a.vals[mid], t)

		if cmp < 0 {
			low = mid + 1
		} else if cmp > 0 {
			high = mid - 1
		} else {
			return mid // 找到元素
		}
	}

	return -1 // 没有找到
}

// Append 重写Append方法，保持有序
func (a *ArrayListSorted[T]) Append(ts ...T) error {
	for _, t := range ts {
		if err := a.InsertSorted(t, a.less); err != nil {
			return err
		}
	}
	return nil
}

// Add 重写Add方法，确保只能在正确的位置插入
func (a *ArrayListSorted[T]) Add(index int, t T) error {
	// 计算正确的插入位置
	correctIndex := sort.Search(len(a.vals), func(i int) bool {
		return !a.less(a.vals[i], t)
	})

	// 如果指定的位置不是正确的排序位置，返回错误
	if index != correctIndex {
		return ErrInvalidArgument
	}

	return a.ArrayList.Add(index, t)
}

// Sort 对有序列表排序会使用已设置的比较函数
func (a *ArrayListSorted[T]) Sort(less func(a, b T) bool) {
	if less == nil {
		less = a.less
	}
	a.ArrayList.Sort(less)
}
