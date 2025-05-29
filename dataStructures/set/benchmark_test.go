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

package set

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// 准备基准测试数据
func prepareIntItems(n int) []int {
	items := make([]int, n)
	for i := 0; i < n; i++ {
		items[i] = i
	}
	return items
}

func prepareStringItems(n int) []string {
	items := make([]string, n)
	for i := 0; i < n; i++ {
		items[i] = strconv.Itoa(i)
	}
	return items
}

// 基准测试: MapSet 添加整数元素
func BenchmarkMapSet_Add_Int(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			items := prepareIntItems(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				set := NewMapSet[int](0)
				for _, item := range items {
					set.Add(item)
				}
			}
		})
	}
}

// 基准测试: MapSet 检查整数元素是否存在
func BenchmarkMapSet_Exist_Int(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			set := NewMapSet[int](size)
			items := prepareIntItems(size)

			// 先添加所有元素
			for _, item := range items {
				set.Add(item)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 查询存在的元素
				for j := 0; j < size; j++ {
					_ = set.Exist(j)
				}
			}
		})
	}
}

// 基准测试: MapSet 删除整数元素
func BenchmarkMapSet_Delete_Int(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			items := prepareIntItems(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				set := NewMapSet[int](size)
				for _, item := range items {
					set.Add(item)
				}
				b.StartTimer()

				// 删除所有元素
				for j := 0; j < size; j++ {
					set.Delete(j)
				}
			}
		})
	}
}

// 基准测试: TreeSet 添加整数元素
func BenchmarkTreeSet_Add_Int(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			items := prepareIntItems(size)
			intComparator := ComparatorRealNumber[int]()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				set, _ := NewTreeSet[int](intComparator)
				b.StartTimer()

				for _, item := range items {
					set.Add(item)
				}
			}
		})
	}
}

// 基准测试: TreeSet 检查整数元素是否存在
func BenchmarkTreeSet_Exist_Int(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			intComparator := ComparatorRealNumber[int]()
			set, _ := NewTreeSet[int](intComparator)
			items := prepareIntItems(size)

			// 先添加所有元素
			for _, item := range items {
				set.Add(item)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 查询存在的元素
				for j := 0; j < size; j++ {
					_ = set.Exist(j)
				}
			}
		})
	}
}

// 基准测试: ConcurrentSet 在并发环境下添加元素
func BenchmarkConcurrentSet_Add_Concurrent(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	concurrencyLevels := []int{1, 4, 8, 16}

	for _, size := range sizes {
		for _, concurrency := range concurrencyLevels {
			name := "size=" + strconv.Itoa(size) + "_concurrency=" + strconv.Itoa(concurrency)
			b.Run(name, func(b *testing.B) {
				items := prepareIntItems(size)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					set := NewConcurrentSet[int](size)
					b.StartTimer()

					done := make(chan bool)
					itemsPerGoroutine := size / concurrency

					for j := 0; j < concurrency; j++ {
						start := j * itemsPerGoroutine
						end := start + itemsPerGoroutine
						if j == concurrency-1 {
							end = size // 确保最后一个协程处理所有剩余的元素
						}

						go func(startIndex, endIndex int) {
							for k := startIndex; k < endIndex; k++ {
								set.Add(items[k])
							}
							done <- true
						}(start, end)
					}

					// 等待所有协程完成
					for j := 0; j < concurrency; j++ {
						<-done
					}
				}
			})
		}
	}
}

// 基准测试: ExpirableSet 添加带过期时间的元素
func BenchmarkExpirableSet_AddWithTTL(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			items := prepareIntItems(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				set := NewExpirableSet[int](10 * time.Second)
				b.StartTimer()

				for _, item := range items {
					ttl := time.Duration(rand.Int63n(10)+1) * time.Second
					set.AddWithTTL(item, ttl)
				}

				b.StopTimer()
				set.Close() // 关闭清理协程
				b.StartTimer()
			}
		})
	}
}

// 基准测试: 集合操作(Union, Intersect, Difference)
func BenchmarkMapSet_Operations(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			itemsA := prepareIntItems(size)
			itemsB := prepareIntItems(size)

			// 让B集合与A有一半重叠，一半不同
			for i := size / 2; i < size; i++ {
				itemsB[i] = i + size
			}

			setA := NewMapSet[int](size)
			setB := NewMapSet[int](size)

			for _, item := range itemsA {
				setA.Add(item)
			}

			for _, item := range itemsB {
				setB.Add(item)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 并集
				_ = setA.Union(setB)

				// 交集
				_ = setA.Intersect(setB)

				// 差集
				_ = setA.Difference(setB)
			}
		})
	}
}

// 基准测试: 不同类型集合的性能比较 (MapSet vs TreeSet vs ConcurrentSet)
func BenchmarkSetTypes_Comparison(b *testing.B) {
	size := 1000
	items := prepareIntItems(size)

	b.Run("MapSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			set := NewMapSet[int](size)
			for _, item := range items {
				set.Add(item)
			}
			for _, item := range items {
				_ = set.Exist(item)
			}
		}
	})

	b.Run("TreeSet", func(b *testing.B) {
		intComparator := ComparatorRealNumber[int]()
		for i := 0; i < b.N; i++ {
			set, _ := NewTreeSet[int](intComparator)
			for _, item := range items {
				set.Add(item)
			}
			for _, item := range items {
				_ = set.Exist(item)
			}
		}
	})

	b.Run("ConcurrentSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			set := NewConcurrentSet[int](size)
			for _, item := range items {
				set.Add(item)
			}
			for _, item := range items {
				_ = set.Exist(item)
			}
		}
	})
}

// 基准测试: 字符串集合操作
func BenchmarkMapSet_String(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			items := prepareStringItems(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				set := NewMapSet[string](size)

				// 添加
				for _, item := range items {
					set.Add(item)
				}

				// 查找
				for _, item := range items {
					_ = set.Exist(item)
				}

				// 删除
				for _, item := range items {
					set.Delete(item)
				}
			}
		})
	}
}
