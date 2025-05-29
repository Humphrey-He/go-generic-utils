// Copyright 2024 Humphrey
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

// 本文件整合了常用切片操作工具，适合电商平台后端高并发、线程安全等场景。
// 包含查找、增删、去重、映射、聚合、集合运算等功能，并提供线程安全版本。

package sliceutils

import (
	"errors"
	"sync"
)

// ================== 类型定义 ==================

// equalFunc 比较两个元素是否相等
// 用于自定义去重、查找等场景
type equalFunc[T any] func(src, dst T) bool

// matchFunc 判断元素是否匹配
// 用于查找、过滤等场景
type matchFunc[T any] func(src T) bool

// ================== 线程安全切片封装 ==================
// ThreadSafeSlice 封装了线程安全的切片操作，适合高并发场景
type ThreadSafeSlice[T any] struct {
	mu    sync.RWMutex
	slice []T
}

// NewThreadSafeSlice 创建线程安全切片
func NewThreadSafeSlice[T any](init []T) *ThreadSafeSlice[T] {
	var slice []T
	if init != nil {
		slice = append([]T(nil), init...)
	} else {
		// 问题：原实现中，当 init 为 nil 时，append([]T(nil), init...) 会返回 nil 而不是空切片
		// 这导致 ThreadSafeSlice.slice 为 nil，违反了测试中的期望（非 nil 的空切片）
		//
		// 修复思路：显式创建一个长度为 0 的切片，确保即使 init 为 nil，返回的也是一个非 nil 的空切片
		// 这样可以避免后续对 nil 切片的操作可能导致的 panic
		slice = make([]T, 0)
	}
	return &ThreadSafeSlice[T]{slice: slice}
}

// Append 线程安全追加元素
func (s *ThreadSafeSlice[T]) Append(vals ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slice = append(s.slice, vals...)
}

// Delete 删除指定下标元素
func (s *ThreadSafeSlice[T]) Delete(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.slice) {
		return errors.New("下标越界")
	}
	s.slice = append(s.slice[:index], s.slice[index+1:]...)
	return nil
}

// Get 获取指定下标元素
func (s *ThreadSafeSlice[T]) Get(index int) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index < 0 || index >= len(s.slice) {
		var zero T
		return zero, errors.New("下标越界")
	}
	return s.slice[index], nil
}

// Set 设置指定下标元素
func (s *ThreadSafeSlice[T]) Set(index int, val T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.slice) {
		return errors.New("下标越界")
	}
	s.slice[index] = val
	return nil
}

// AsSlice 返回切片副本
func (s *ThreadSafeSlice[T]) AsSlice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]T, len(s.slice))
	copy(res, s.slice)
	return res
}

// Len 返回切片长度
func (s *ThreadSafeSlice[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.slice)
}

// ================== 查找相关 ==================

// Find 查找第一个匹配元素，未找到返回false
func Find[T any](src []T, match matchFunc[T]) (T, bool) {
	for _, val := range src {
		if match(val) {
			return val, true
		}
	}
	var t T
	return t, false
}

// FindAll 查找所有匹配元素
func FindAll[T any](src []T, match matchFunc[T]) []T {
	res := make([]T, 0, len(src)>>3+1)
	for _, val := range src {
		if match(val) {
			res = append(res, val)
		}
	}
	return res
}

// Index 返回第一个等于dst的下标，未找到返回-1
func Index[T comparable](src []T, dst T) int {
	return IndexFunc(src, func(src T) bool { return src == dst })
}

// IndexFunc 返回第一个匹配的下标，未找到返回-1
func IndexFunc[T any](src []T, match matchFunc[T]) int {
	for k, v := range src {
		if match(v) {
			return k
		}
	}
	return -1
}

// LastIndex 返回最后一个等于dst的下标，未找到返回-1
func LastIndex[T comparable](src []T, dst T) int {
	return LastIndexFunc(src, func(src T) bool { return src == dst })
}

// LastIndexFunc 返回最后一个匹配的下标，未找到返回-1
func LastIndexFunc[T any](src []T, match matchFunc[T]) int {
	for i := len(src) - 1; i >= 0; i-- {
		if match(src[i]) {
			return i
		}
	}
	return -1
}

// IndexAll 返回所有等于dst的下标
func IndexAll[T comparable](src []T, dst T) []int {
	return IndexAllFunc(src, func(src T) bool { return src == dst })
}

// IndexAllFunc 返回所有匹配的下标
func IndexAllFunc[T any](src []T, match matchFunc[T]) []int {
	var indexes = make([]int, 0, len(src))
	for k, v := range src {
		if match(v) {
			indexes = append(indexes, k)
		}
	}
	return indexes
}

// ================== 增删改 ==================

// Add 在index处插入元素
func Add[T any](src []T, element T, index int) ([]T, error) {
	if index < 0 || index > len(src) {
		return nil, errors.New("下标越界")
	}
	src = append(src, element)
	copy(src[index+1:], src[index:])
	src[index] = element
	return src, nil
}

// Delete 删除index处元素
func Delete[T any](src []T, index int) ([]T, error) {
	if index < 0 || index >= len(src) {
		return nil, errors.New("下标越界")
	}
	return append(src[:index], src[index+1:]...), nil
}

// FilterDelete 删除所有匹配条件的元素
func FilterDelete[T any](src []T, match func(idx int, src T) bool) []T {
	emptyPos := 0
	for idx := range src {
		if match(idx, src[idx]) {
			continue
		}
		src[emptyPos] = src[idx]
		emptyPos++
	}
	return src[:emptyPos]
}

// Set 设置index处元素
func Set[T any](src []T, index int, val T) ([]T, error) {
	if index < 0 || index >= len(src) {
		return nil, errors.New("下标越界")
	}
	src[index] = val
	return src, nil
}

// ================== 包含/去重相关 ==================

// Contains 判断src中是否存在dst
func Contains[T comparable](src []T, dst T) bool {
	return ContainsFunc(src, func(src T) bool { return src == dst })
}

// ContainsFunc 判断src中是否存在满足条件的元素
func ContainsFunc[T any](src []T, equal func(src T) bool) bool {
	for _, v := range src {
		if equal(v) {
			return true
		}
	}
	return false
}

// ContainsAny 判断src中是否存在dst中的任意一个元素
func ContainsAny[T comparable](src, dst []T) bool {
	srcMap := toMap(src)
	for _, v := range dst {
		if _, exist := srcMap[v]; exist {
			return true
		}
	}
	return false
}

// ContainsAll 判断src中是否包含dst中的所有元素
func ContainsAll[T comparable](src, dst []T) bool {
	srcMap := toMap(src)
	for _, v := range dst {
		if _, exist := srcMap[v]; !exist {
			return false
		}
	}
	return true
}

// deduplicate 去重（comparable类型）
func deduplicate[T comparable](data []T) []T {
	dataMap := toMap(data)
	var newData = make([]T, 0, len(dataMap))
	for key := range dataMap {
		newData = append(newData, key)
	}
	return newData
}

// deduplicateFunc 去重（自定义相等）
func deduplicateFunc[T any](data []T, equal equalFunc[T]) []T {
	var newData = make([]T, 0, len(data))
	for k, v := range data {
		if !ContainsFunc(data[k+1:], func(src T) bool { return equal(src, v) }) {
			newData = append(newData, v)
		}
	}
	return newData
}

// ================== 映射/聚合相关 ==================

// Map 映射操作
func Map[Src any, Dst any](src []Src, m func(idx int, src Src) Dst) []Dst {
	dst := make([]Dst, len(src))
	for i, s := range src {
		dst[i] = m(i, s)
	}
	return dst
}

// FilterMap 过滤并映射
func FilterMap[Src any, Dst any](src []Src, m func(idx int, src Src) (Dst, bool)) []Dst {
	res := make([]Dst, 0, len(src))
	for i, s := range src {
		dst, ok := m(i, s)
		if ok {
			res = append(res, dst)
		}
	}
	return res
}

// ToMap 将切片映射为map
func ToMap[Ele any, Key comparable](elements []Ele, fn func(element Ele) Key) map[Key]Ele {
	return ToMapV(elements, func(element Ele) (Key, Ele) { return fn(element), element })
}

// ToMapV 将切片映射为map，支持自定义value
func ToMapV[Ele any, Key comparable, Val any](elements []Ele, fn func(element Ele) (Key, Val)) map[Key]Val {
	resultMap := make(map[Key]Val, len(elements))
	for _, element := range elements {
		k, v := fn(element)
		resultMap[k] = v
	}
	return resultMap
}

// ================== 集合运算相关 ==================

// toMap 辅助函数，将切片转为map
func toMap[T comparable](src []T) map[T]struct{} {
	var dataMap = make(map[T]struct{}, len(src))
	for _, v := range src {
		dataMap[v] = struct{}{}
	}
	return dataMap
}

// UnionSet 并集，已去重
func UnionSet[T comparable](src, dst []T) []T {
	srcMap, dstMap := toMap(src), toMap(dst)
	for key := range srcMap {
		dstMap[key] = struct{}{}
	}
	var ret = make([]T, 0, len(dstMap))
	for key := range dstMap {
		ret = append(ret, key)
	}
	return ret
}

// IntersectSet 交集，已去重
func IntersectSet[T comparable](src, dst []T) []T {
	srcMap := toMap(src)
	var ret = make([]T, 0, len(src))
	for _, val := range dst {
		if _, exist := srcMap[val]; exist {
			ret = append(ret, val)
		}
	}
	return deduplicate(ret)
}

// DiffSet 差集，已去重
func DiffSet[T comparable](src, dst []T) []T {
	srcMap := toMap(src)
	for _, val := range dst {
		delete(srcMap, val)
	}
	var ret = make([]T, 0, len(srcMap))
	for key := range srcMap {
		ret = append(ret, key)
	}
	return ret
}

// SymmetricDiffSet 对称差集，已去重
func SymmetricDiffSet[T comparable](src, dst []T) []T {
	srcMap, dstMap := toMap(src), toMap(dst)
	for k := range dstMap {
		if _, ok := srcMap[k]; ok {
			delete(srcMap, k)
		} else {
			srcMap[k] = struct{}{}
		}
	}
	res := make([]T, 0, len(srcMap))
	for k := range srcMap {
		res = append(res, k)
	}
	return res
}

// ================== 聚合相关 ==================

// Max 返回最大值，假设切片非空
func Max[T interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64
}](ts []T) T {
	res := ts[0]
	for i := 1; i < len(ts); i++ {
		if ts[i] > res {
			res = ts[i]
		}
	}
	return res
}

// Min 返回最小值，假设切片非空
func Min[T interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64
}](ts []T) T {
	res := ts[0]
	for i := 1; i < len(ts); i++ {
		if ts[i] < res {
			res = ts[i]
		}
	}
	return res
}

// Sum 求和
func Sum[T interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64
}](ts []T) T {
	var res T
	for _, n := range ts {
		res += n
	}
	return res
}

// ================== 反转相关 ==================

// Reverse 返回反转后的新切片
func Reverse[T any](src []T) []T {
	var ret = make([]T, 0, len(src))
	for i := len(src) - 1; i >= 0; i-- {
		ret = append(ret, src[i])
	}
	return ret
}

// ReverseSelf 原地反转切片
func ReverseSelf[T any](src []T) {
	for i, j := 0, len(src)-1; i < j; i, j = i+1, j-1 {
		src[i], src[j] = src[j], src[i]
	}
}
