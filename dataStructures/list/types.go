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
	"errors"
)

// 常用错误定义
var (
	ErrIndexOutOfRange = errors.New("ggu: 索引超出范围")
	ErrEmptyList       = errors.New("ggu: 列表为空")
	ErrInvalidArgument = errors.New("ggu: 无效参数")
)

// List 通用列表接口
// 定义各种列表类型的通用操作
type List[T any] interface {
	// Get 返回对应下标的元素
	// 在下标超出范围的情况下，返回错误
	Get(index int) (T, error)

	// Append 在末尾追加元素
	Append(ts ...T) error

	// Add 在特定下标处增加一个新元素
	// 如果下标不在[0, Len()]范围之内
	// 应该返回错误
	// 如果index == Len()则表示往List末端增加一个值
	Add(index int, t T) error

	// Set 重置 index 位置的值
	// 如果下标超出范围，应该返回错误
	Set(index int, t T) error

	// Delete 删除目标元素的位置，并且返回该位置的值
	// 如果 index 超出下标，应该返回错误
	Delete(index int) (T, error)

	// DeleteValue 删除指定值，如果找到并删除成功则返回true
	// equals是用于比较两个元素是否相等的函数
	DeleteValue(t T, equals func(src T, dst T) bool) bool

	// Len 返回长度
	Len() int

	// Cap 返回容量
	Cap() int

	// Range 遍历 List 的所有元素
	Range(fn func(index int, t T) error) error

	// ReverseRange 逆序遍历 List 的所有元素
	ReverseRange(fn func(index int, t T) error) error

	// AsSlice 将 List 转化为一个切片
	// 不允许返回nil，在没有元素的情况下，
	// 必须返回一个长度和容量都为 0 的切片
	// AsSlice 每次调用都必须返回一个全新的切片
	AsSlice() []T

	// Sort 对列表中的元素进行排序
	// less 函数用于比较两个元素的大小
	// 对 ConcurrentList 进行 Sort 会锁住整个列表直到排序完成
	Sort(less func(a, b T) bool)

	// Filter 过滤列表，返回符合条件的元素组成的新列表
	// predicate 返回 true 表示保留该元素
	Filter(predicate func(t T) bool) List[T]

	// Map 对列表中的每个元素应用转换函数，返回转换后的新列表
	Map(mapper func(t T) T) List[T]

	// Clear 清空列表
	Clear()
}

// SortedList 有序列表接口
// 保持列表元素按某种顺序排列
type SortedList[T any] interface {
	List[T]

	// InsertSorted 插入元素并保持列表有序
	// less函数用于比较元素大小，如果a < b则返回true
	InsertSorted(t T, less func(a, b T) bool) error

	// Find 查找元素，返回其索引
	// 如果不存在，返回-1
	Find(t T, equals func(a, b T) bool) int

	// BinarySearch 二分查找元素，返回其索引
	// 仅当列表已排序时才能使用
	// 如果存在多个满足条件的元素，不保证返回哪一个
	// 如果不存在，返回-1
	BinarySearch(t T, comparator func(a, b T) int) int
}

// PagedList 分页列表接口
// 支持分页操作，适用于电商后台管理等场景
type PagedList[T any] interface {
	List[T]

	// Page 返回指定页的元素
	// page 从1开始计数
	// pageSize 为每页元素个数
	// 返回指定页的元素切片和总页数
	Page(page, pageSize int) ([]T, int, error)

	// TotalPages 根据给定的每页大小计算总页数
	TotalPages(pageSize int) int
}
