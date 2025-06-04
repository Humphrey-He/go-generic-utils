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
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// RunBenchmarkAndPrintResults 运行基准测试并打印结果
func RunBenchmarkAndPrintResults(b *testing.B, name string, f func(b *testing.B)) testing.BenchmarkResult {
	result := testing.Benchmark(func(b *testing.B) {
		f(b)
	})

	fmt.Printf("基准测试 %s:\n", name)
	fmt.Printf("  操作次数: %d\n", result.N)
	fmt.Printf("  每次操作: %.2f ns/op\n", float64(result.NsPerOp()))
	fmt.Printf("  内存分配: %d 次, %.2f bytes/op\n", result.MemAllocs, float64(result.AllocsPerOp()))
	fmt.Printf("  内存总量: %d bytes\n", result.AllocedBytesPerOp())
	fmt.Println()

	return result
}

// TestSetBenchmarks 运行多次基准测试并分析结果
func TestSetBenchmarks(t *testing.T) {
	t.Run("Set性能测试", func(t *testing.T) {
		// 跳过自动测试，只在手动请求时运行
		if testing.Short() {
			t.Skip("跳过基准测试")
		}

		// 每种测试运行3次以确保稳定性
		results := make(map[string][]testing.BenchmarkResult)

		// 测试MapSet添加元素
		for i := 0; i < 3; i++ {
			fmt.Printf("=== 运行 #%d ===\n", i+1)
			result := RunBenchmarkAndPrintResults(nil, "MapSet_Add_1000", func(b *testing.B) {
				items := prepareIntItems(1000)
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					set := NewMapSet[int](0)
					for _, item := range items {
						set.Add(item)
					}
				}
			})
			results["MapSet_Add"] = append(results["MapSet_Add"], result)

			result = RunBenchmarkAndPrintResults(nil, "MapSet_Exist_1000", func(b *testing.B) {
				set := NewMapSet[int](1000)
				items := prepareIntItems(1000)

				// 先添加所有元素
				for _, item := range items {
					set.Add(item)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// 查询存在的元素
					for j := 0; j < 100; j++ { // 查询100个元素以避免测试时间过长
						index := rand.Intn(1000)
						_ = set.Exist(index)
					}
				}
			})
			results["MapSet_Exist"] = append(results["MapSet_Exist"], result)

			result = RunBenchmarkAndPrintResults(nil, "MapSet_Delete_1000", func(b *testing.B) {
				items := prepareIntItems(1000)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					set := NewMapSet[int](1000)
					for _, item := range items {
						set.Add(item)
					}
					b.StartTimer()

					// 删除所有元素
					for j := 0; j < 1000; j++ {
						set.Delete(j)
					}
				}
			})
			results["MapSet_Delete"] = append(results["MapSet_Delete"], result)

			// 测试TreeSet添加元素
			result = RunBenchmarkAndPrintResults(nil, "TreeSet_Add_1000", func(b *testing.B) {
				items := prepareIntItems(1000)
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
			results["TreeSet_Add"] = append(results["TreeSet_Add"], result)

			result = RunBenchmarkAndPrintResults(nil, "TreeSet_Exist_1000", func(b *testing.B) {
				intComparator := ComparatorRealNumber[int]()
				set, _ := NewTreeSet[int](intComparator)
				items := prepareIntItems(1000)

				// 先添加所有元素
				for _, item := range items {
					set.Add(item)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// 查询存在的元素
					for j := 0; j < 100; j++ { // 查询100个元素以避免测试时间过长
						index := rand.Intn(1000)
						_ = set.Exist(index)
					}
				}
			})
			results["TreeSet_Exist"] = append(results["TreeSet_Exist"], result)

			// 测试ConcurrentSet在并发环境下添加元素
			result = RunBenchmarkAndPrintResults(nil, "ConcurrentSet_Add_1000_Concurrency_8", func(b *testing.B) {
				items := prepareIntItems(1000)
				concurrency := 8

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					set := NewConcurrentSet[int](1000)
					b.StartTimer()

					done := make(chan bool)
					itemsPerGoroutine := 1000 / concurrency

					for j := 0; j < concurrency; j++ {
						start := j * itemsPerGoroutine
						end := start + itemsPerGoroutine
						if j == concurrency-1 {
							end = 1000 // 确保最后一个协程处理所有剩余的元素
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
			results["ConcurrentSet_Add"] = append(results["ConcurrentSet_Add"], result)
		}

		// 分析结果
		fmt.Println("\n=== 性能测试结果分析 ===")
		for testName, testResults := range results {
			var totalNsPerOp float64
			var totalAllocsPerOp float64
			var totalBytesPerOp int64

			for _, result := range testResults {
				totalNsPerOp += float64(result.NsPerOp())
				totalAllocsPerOp += float64(result.AllocsPerOp())
				totalBytesPerOp += result.AllocedBytesPerOp()
			}

			avgNsPerOp := totalNsPerOp / float64(len(testResults))
			avgAllocsPerOp := totalAllocsPerOp / float64(len(testResults))
			avgBytesPerOp := totalBytesPerOp / int64(len(testResults))

			fmt.Printf("%s 平均性能:\n", testName)
			fmt.Printf("  每次操作: %.2f ns/op\n", avgNsPerOp)
			fmt.Printf("  内存分配: %.2f allocs/op\n", avgAllocsPerOp)
			fmt.Printf("  内存总量: %d bytes/op\n\n", avgBytesPerOp)
		}
	})

	t.Run("EcomSet性能测试", func(t *testing.T) {
		// 跳过自动测试，只在手动请求时运行
		if testing.Short() {
			t.Skip("跳过基准测试")
		}

		// 每种测试运行3次以确保稳定性
		results := make(map[string][]testing.BenchmarkResult)

		// 测试ExpirableSet添加带过期时间的元素
		for i := 0; i < 3; i++ {
			fmt.Printf("=== 运行 #%d ===\n", i+1)
			result := RunBenchmarkAndPrintResults(nil, "ExpirableSet_AddWithTTL_1000", func(b *testing.B) {
				items := prepareIntItems(1000)

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
			results["ExpirableSet_AddWithTTL"] = append(results["ExpirableSet_AddWithTTL"], result)

			// 测试TagSet添加元素
			result = RunBenchmarkAndPrintResults(nil, "TagSet_Add_1000", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					set := NewTagSet(1000)
					b.StartTimer()

					for j := 0; j < 1000; j++ {
						id := strconv.Itoa(j)
						name := "tag-" + id
						_ = set.AddTag(id, name)
					}
				}
			})
			results["TagSet_Add"] = append(results["TagSet_Add"], result)

			// 测试ShoppingCart添加元素
			result = RunBenchmarkAndPrintResults(nil, "ShoppingCart_Add_1000", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					cart := NewShoppingCart("user1", 1000)
					b.StartTimer()

					for j := 0; j < 1000; j++ {
						productID := ProductID(strconv.Itoa(j))
						attrs := map[string]string{"color": "red", "size": "M"}
						_ = cart.AddItem(productID, 1, attrs)
					}
				}
			})
			results["ShoppingCart_Add"] = append(results["ShoppingCart_Add"], result)

			// 测试WishList添加元素
			result = RunBenchmarkAndPrintResults(nil, "WishList_Add_1000", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					wishlist := NewWishList("user1", 1000)
					b.StartTimer()

					for j := 0; j < 1000; j++ {
						productID := ProductID(strconv.Itoa(j))
						_ = wishlist.AddProduct(productID)
					}
				}
			})
			results["WishList_Add"] = append(results["WishList_Add"], result)
		}

		// 分析结果
		fmt.Println("\n=== 性能测试结果分析 ===")
		for testName, testResults := range results {
			var totalNsPerOp float64
			var totalAllocsPerOp float64
			var totalBytesPerOp int64

			for _, result := range testResults {
				totalNsPerOp += float64(result.NsPerOp())
				totalAllocsPerOp += float64(result.AllocsPerOp())
				totalBytesPerOp += result.AllocedBytesPerOp()
			}

			avgNsPerOp := totalNsPerOp / float64(len(testResults))
			avgAllocsPerOp := totalAllocsPerOp / float64(len(testResults))
			avgBytesPerOp := totalBytesPerOp / int64(len(testResults))

			fmt.Printf("%s 平均性能:\n", testName)
			fmt.Printf("  每次操作: %.2f ns/op\n", avgNsPerOp)
			fmt.Printf("  内存分配: %.2f allocs/op\n", avgAllocsPerOp)
			fmt.Printf("  内存总量: %d bytes/op\n\n", avgBytesPerOp)
		}
	})
}

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
