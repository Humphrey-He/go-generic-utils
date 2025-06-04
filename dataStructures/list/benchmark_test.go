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

package list

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

// 准备基准测试数据
func prepareInts(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i
	}
	return result
}

// 对所有List实现进行测试的工具函数
func benchmarkListAppend(b *testing.B, createList func() List[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareInts(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				list := createList()
				for _, val := range data {
					_ = list.Append(val)
				}
			}
		})
	}
}

func benchmarkListGet(b *testing.B, createList func() List[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareInts(size)
			list := createList()
			for _, val := range data {
				_ = list.Append(val)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < size; j++ {
					index := j % size // 循环访问所有元素
					_, _ = list.Get(index)
				}
			}
		})
	}
}

func benchmarkListAdd(b *testing.B, createList func() List[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareInts(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				list := createList()
				b.StartTimer()

				// 在随机位置插入元素
				for j := 0; j < size; j++ {
					index := rand.Intn(j + 1) // 随机选择0到j之间的位置
					_ = list.Add(index, data[j])
				}
			}
		})
	}
}

func benchmarkListDelete(b *testing.B, createList func() List[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// 创建并填充列表
				list := createList()
				data := prepareInts(size)
				for _, val := range data {
					list.Append(val)
				}
				b.StartTimer()

				// 从列表中删除元素
				// 从最后往前删除，以避免索引变化问题
				for j := size - 1; j >= 0; j-- {
					_, _ = list.Delete(j)
				}
			}
		})
	}
}

func benchmarkListRange(b *testing.B, createList func() List[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			// 创建并填充列表
			list := createList()
			data := prepareInts(size)
			for _, val := range data {
				list.Append(val)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = list.Range(func(index int, val int) error {
					return nil
				})
			}
		})
	}
}

func benchmarkListSort(b *testing.B, createList func() List[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// 创建并填充列表（逆序填充，确保需要排序）
				list := createList()
				for j := size - 1; j >= 0; j-- {
					list.Append(j)
				}
				b.StartTimer()

				// 排序
				list.Sort(func(a, b int) bool {
					return a < b
				})
			}
		})
	}
}

// 运行基准测试并打印结果
func RunBenchmarkAndPrintResults(b *testing.B, name string, f func(b *testing.B)) {
	result := testing.Benchmark(func(b *testing.B) {
		f(b)
	})

	fmt.Printf("基准测试 %s:\n", name)
	fmt.Printf("  操作次数: %d\n", result.N)
	fmt.Printf("  每次操作: %.2f ns/op\n", float64(result.NsPerOp()))
	fmt.Printf("  内存分配: %d 次, %.2f bytes/op\n", result.MemAllocs, float64(result.AllocsPerOp()))
	fmt.Printf("  内存总量: %d bytes\n", result.AllocedBytesPerOp())
	fmt.Println()
}

// 测试不同列表实现的性能比较
func TestListImplementationsBenchmark(t *testing.T) {
	t.Run("列表实现性能比较", func(t *testing.T) {
		// 跳过自动测试，只在手动请求时运行
		if testing.Short() {
			t.Skip("跳过基准测试")
		}

		// 测试Append操作
		RunBenchmarkAndPrintResults(nil, "ArrayList_Append_10000", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				list := NewArrayList[int](0)
				for j := 0; j < 10000; j++ {
					_ = list.Append(j)
				}
			}
		})

		RunBenchmarkAndPrintResults(nil, "LinkedList_Append_10000", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				list := NewLinkedList[int]()
				for j := 0; j < 10000; j++ {
					_ = list.Append(j)
				}
			}
		})

		RunBenchmarkAndPrintResults(nil, "ConcurrentList_Append_10000", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				list := NewConcurrentList[int](0)
				for j := 0; j < 10000; j++ {
					_ = list.Append(j)
				}
			}
		})

		// 测试Get操作
		RunBenchmarkAndPrintResults(nil, "ArrayList_Get_10000", func(b *testing.B) {
			list := NewArrayList[int](10000)
			for j := 0; j < 10000; j++ {
				_ = list.Append(j)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < 100; j++ {
					index := j * 100 // 均匀分布的访问
					_, _ = list.Get(index)
				}
			}
		})

		RunBenchmarkAndPrintResults(nil, "LinkedList_Get_10000", func(b *testing.B) {
			list := NewLinkedList[int]()
			for j := 0; j < 10000; j++ {
				_ = list.Append(j)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < 100; j++ {
					index := j * 100 // 均匀分布的访问
					_, _ = list.Get(index)
				}
			}
		})

		RunBenchmarkAndPrintResults(nil, "ConcurrentList_Get_10000", func(b *testing.B) {
			list := NewConcurrentList[int](10000)
			for j := 0; j < 10000; j++ {
				_ = list.Append(j)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < 100; j++ {
					index := j * 100 // 均匀分布的访问
					_, _ = list.Get(index)
				}
			}
		})
	})
}

// 标准基准测试函数，用于go test -bench命令
func BenchmarkArrayList_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list := NewArrayList[int](0)
		for j := 0; j < 1000; j++ {
			_ = list.Append(j)
		}
	}
}

func BenchmarkLinkedList_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list := NewLinkedList[int]()
		for j := 0; j < 1000; j++ {
			_ = list.Append(j)
		}
	}
}

func BenchmarkConcurrentList_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list := NewConcurrentList[int](0)
		for j := 0; j < 1000; j++ {
			_ = list.Append(j)
		}
	}
}

func BenchmarkArrayList_Get(b *testing.B) {
	list := NewArrayList[int](1000)
	for j := 0; j < 1000; j++ {
		_ = list.Append(j)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			index := j * 10 % 1000
			_, _ = list.Get(index)
		}
	}
}

func BenchmarkLinkedList_GetItems(b *testing.B) {
	list := NewLinkedList[int]()
	for j := 0; j < 1000; j++ {
		_ = list.Append(j)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			index := j * 10 % 1000
			_, _ = list.Get(index)
		}
	}
}

func BenchmarkConcurrentList_Get(b *testing.B) {
	list := NewConcurrentList[int](1000)
	for j := 0; j < 1000; j++ {
		_ = list.Append(j)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			index := j * 10 % 1000
			_, _ = list.Get(index)
		}
	}
}

// 并发添加测试
func BenchmarkConcurrentList_ConcurrentAppend(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	concurrencyLevels := []int{2, 4, 8, 16}

	for _, size := range sizes {
		for _, concurrency := range concurrencyLevels {
			name := fmt.Sprintf("size=%d_concurrency=%d", size, concurrency)
			b.Run(name, func(b *testing.B) {
				itemsPerGoroutine := size / concurrency

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					b.StopTimer()
					list := NewConcurrentList[int](0)
					b.StartTimer()

					wg := sync.WaitGroup{}
					wg.Add(concurrency)

					for j := 0; j < concurrency; j++ {
						start := j * itemsPerGoroutine
						end := start + itemsPerGoroutine
						if j == concurrency-1 {
							end = size // 确保最后一个协程处理所有剩余的元素
						}

						go func(startIdx, endIdx int) {
							defer wg.Done()
							for k := startIdx; k < endIdx; k++ {
								list.Append(k)
							}
						}(start, end)
					}

					wg.Wait()
				}
			})
		}
	}
}

// 分页列表测试
func BenchmarkArrayListPaged_Page(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	pageSizes := []int{10, 20, 50, 100}

	for _, size := range sizes {
		for _, pageSize := range pageSizes {
			name := fmt.Sprintf("size=%d_pageSize=%d", size, pageSize)
			b.Run(name, func(b *testing.B) {
				// 创建并填充分页列表
				list := NewArrayListPaged[int](size)
				data := prepareInts(size)
				for _, val := range data {
					list.Append(val)
				}

				// 计算总页数
				totalPages := list.TotalPages(pageSize)

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// 遍历所有页
					for page := 1; page <= totalPages; page++ {
						_, _, _ = list.Page(page, pageSize)
					}
				}
			})
		}
	}
}

// 有序列表测试
func BenchmarkArrayListSorted_InsertSorted(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareInts(size)
			// 打乱数据顺序，以测试有序插入
			rand.Shuffle(len(data), func(i, j int) {
				data[i], data[j] = data[j], data[i]
			})

			less := func(a, b int) bool {
				return a < b
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list := NewArrayListSorted[int](0, less)

				for _, val := range data {
					list.InsertSorted(val, less)
				}
			}
		})
	}
}

// 列表类型比较 (ArrayList vs LinkedList)
func BenchmarkListTypes_Comparison(b *testing.B) {
	size := 1000
	operations := []string{"Append", "Get", "Add", "Delete"}

	for _, op := range operations {
		b.Run(op, func(b *testing.B) {
			data := prepareInts(size)

			b.Run("ArrayList", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					list := NewArrayList[int](0)

					switch op {
					case "Append":
						for _, val := range data {
							list.Append(val)
						}
					case "Get":
						// 先添加数据
						for _, val := range data {
							list.Append(val)
						}
						// 然后获取
						for j := 0; j < size; j++ {
							_, _ = list.Get(j % size)
						}
					case "Add":
						for j := 0; j < size; j++ {
							list.Add(j%10, data[j]) // 在前10个位置循环插入
						}
					case "Delete":
						// 先添加数据
						for _, val := range data {
							list.Append(val)
						}
						// 然后删除
						for j := size - 1; j >= 0; j-- {
							list.Delete(j)
						}
					}
				}
			})

			b.Run("LinkedList", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					list := NewLinkedList[int]()

					switch op {
					case "Append":
						for _, val := range data {
							list.Append(val)
						}
					case "Get":
						// 先添加数据
						for _, val := range data {
							list.Append(val)
						}
						// 然后获取
						for j := 0; j < size; j++ {
							_, _ = list.Get(j % size)
						}
					case "Add":
						for j := 0; j < size; j++ {
							list.Add(j%10, data[j]) // 在前10个位置循环插入
						}
					case "Delete":
						// 先添加数据
						for _, val := range data {
							list.Append(val)
						}
						// 然后删除
						for j := size - 1; j >= 0; j-- {
							list.Delete(j)
						}
					}
				}
			})
		})
	}
}

// 测试字符串列表性能
func BenchmarkStringList(b *testing.B) {
	sizes := []int{10, 100, 1000}

	// 准备字符串数据
	prepareStrings := func(n int) []string {
		result := make([]string, n)
		for i := 0; i < n; i++ {
			result[i] = "string" + strconv.Itoa(i)
		}
		return result
	}

	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareStrings(size)

			b.Run("ArrayList", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					list := NewArrayList[string](0)

					// 添加
					for _, val := range data {
						list.Append(val)
					}

					// 获取
					for j := 0; j < size; j++ {
						_, _ = list.Get(j)
					}

					// 排序
					list.Sort(func(a, b string) bool {
						return a < b
					})
				}
			})

			b.Run("LinkedList", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					list := NewLinkedList[string]()

					// 添加
					for _, val := range data {
						list.Append(val)
					}

					// 获取
					for j := 0; j < size; j++ {
						_, _ = list.Get(j)
					}

					// 排序
					list.Sort(func(a, b string) bool {
						return a < b
					})
				}
			})
		})
	}
}
