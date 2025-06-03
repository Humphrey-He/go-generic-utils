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

package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 准备基准测试数据
func prepareItems(n int) []int {
	items := make([]int, n)
	for i := 0; i < n; i++ {
		items[i] = i
	}
	return items
}

// 对所有Queue实现进行测试的工具函数
func benchmarkQueueEnqueue(b *testing.B, createQueue func() Queue[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareItems(size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				queue := createQueue()
				b.StartTimer()

				for _, item := range data {
					_ = queue.Enqueue(item)
				}
			}
		})
	}
}

func benchmarkQueueDequeue(b *testing.B, createQueue func() Queue[int], sizes []int) {
	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				queue := createQueue()
				data := prepareItems(size)
				for _, item := range data {
					queue.Enqueue(item)
				}
				b.StartTimer()

				// 出队所有元素
				for j := 0; j < size; j++ {
					_, _ = queue.Dequeue()
				}
			}
		})
	}
}

// ArrayQueue 基准测试
func BenchmarkArrayQueue_Enqueue(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}
	benchmarkQueueEnqueue(b, func() Queue[int] {
		return &ConcurrentArrayBlockingQueue[int]{
			data: make([]int, 100000),
			cap:  100000,
			cond: sync.NewCond(&sync.Mutex{}),
		}
	}, sizes)
}

func BenchmarkArrayQueue_Dequeue(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}
	benchmarkQueueDequeue(b, func() Queue[int] {
		return &ConcurrentArrayBlockingQueue[int]{
			data: make([]int, 100000),
			cap:  100000,
			cond: sync.NewCond(&sync.Mutex{}),
		}
	}, sizes)
}

// LinkedQueue 基准测试
func BenchmarkLinkedQueue_Enqueue(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}
	benchmarkQueueEnqueue(b, func() Queue[int] {
		dummy := &node[int]{}
		return &ConcurrentLinkedBlockingQueue[int]{
			head: dummy,
			tail: dummy,
			cond: sync.NewCond(&sync.Mutex{}),
		}
	}, sizes)
}

func BenchmarkLinkedQueue_Dequeue(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000, 100000}
	benchmarkQueueDequeue(b, func() Queue[int] {
		dummy := &node[int]{}
		return &ConcurrentLinkedBlockingQueue[int]{
			head: dummy,
			tail: dummy,
			cond: sync.NewCond(&sync.Mutex{}),
		}
	}, sizes)
}

// ConcurrentQueue 基准测试
func BenchmarkConcurrentQueue_Enqueue(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	benchmarkQueueEnqueue(b, func() Queue[int] {
		dummy := &node[int]{}
		return &ConcurrentLinkedBlockingQueue[int]{
			head: dummy,
			tail: dummy,
			cond: sync.NewCond(&sync.Mutex{}),
		}
	}, sizes)
}

func BenchmarkConcurrentQueue_Dequeue(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	benchmarkQueueDequeue(b, func() Queue[int] {
		dummy := &node[int]{}
		return &ConcurrentLinkedBlockingQueue[int]{
			head: dummy,
			tail: dummy,
			cond: sync.NewCond(&sync.Mutex{}),
		}
	}, sizes)
}

// 并发队列 基准测试
func BenchmarkConcurrentQueue_ConcurrentOperations(b *testing.B) {
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
					dummy := &node[int]{}
					queue := &ConcurrentLinkedBlockingQueue[int]{
						head: dummy,
						tail: dummy,
						cond: sync.NewCond(&sync.Mutex{}),
					}
					wg := sync.WaitGroup{}
					b.StartTimer()

					// 一半协程进行入队操作
					wg.Add(concurrency / 2)
					for j := 0; j < concurrency/2; j++ {
						start := j * itemsPerGoroutine
						end := start + itemsPerGoroutine

						go func(startIdx, endIdx int) {
							defer wg.Done()
							for k := startIdx; k < endIdx; k++ {
								queue.Enqueue(k)
							}
						}(start, end)
					}

					// 等待入队操作完成一部分
					time.Sleep(time.Millisecond * 10)

					// 另一半协程进行出队操作
					wg.Add(concurrency / 2)
					for j := 0; j < concurrency/2; j++ {
						go func() {
							defer wg.Done()
							for k := 0; k < itemsPerGoroutine; k++ {
								_, _ = queue.Dequeue()
							}
						}()
					}

					wg.Wait()
				}
			})
		}
	}
}

// DelayQueue 基准测试
func BenchmarkDelayQueue_EnqueueWithDelay(b *testing.B) {
	sizes := []int{10, 100, 1000}
	delayMs := 100 // 100毫秒延迟

	for _, size := range sizes {
		name := fmt.Sprintf("size=%d_delay=%dms", size, delayMs)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				queue := NewDelayQueue[int]()
				b.StartTimer()

				for j := 0; j < size; j++ {
					delay := time.Duration(delayMs) * time.Millisecond
					expireAt := time.Now().Add(delay)
					_ = queue.EnqueueWithDelay(j, expireAt)
				}
			}
		})
	}
}

// PriorityQueue 基准测试
func BenchmarkPriorityQueue_EnqueueWithPriority(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				queue := NewConcurrentPriorityQueue[int]()

				// 使用随机优先级入队
				for j := 0; j < size; j++ {
					priority := j % 10 // 0-9的优先级
					_ = queue.EnqueueWithPriority(j, priority)
				}
			}
		})
	}
}

func BenchmarkPriorityQueue_DequeueHighestPriority(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				queue := NewConcurrentPriorityQueue[int]()

				// 使用随机优先级入队
				for j := 0; j < size; j++ {
					priority := j % 10 // 0-9的优先级
					queue.EnqueueWithPriority(j, priority)
				}
				b.StartTimer()

				// 出队所有元素
				for j := 0; j < size; j++ {
					_, _ = queue.Dequeue()
				}
			}
		})
	}
}

// 队列类型比较 (ArrayQueue vs LinkedQueue vs ConcurrentQueue)
func BenchmarkQueueTypes_Comparison(b *testing.B) {
	size := 1000
	operations := []string{"Enqueue", "Dequeue"}

	for _, op := range operations {
		b.Run(op, func(b *testing.B) {
			data := prepareItems(size)

			b.Run("ArrayQueue", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					queue := &ConcurrentArrayBlockingQueue[int]{
						data: make([]int, size*2),
						cap:  size * 2,
						cond: sync.NewCond(&sync.Mutex{}),
					}

					switch op {
					case "Enqueue":
						for _, val := range data {
							queue.Enqueue(val)
						}
					case "Dequeue":
						// 先入队
						for _, val := range data {
							queue.Enqueue(val)
						}
						// 然后出队
						for j := 0; j < size; j++ {
							_, _ = queue.Dequeue()
						}
					}
				}
			})

			b.Run("LinkedQueue", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					dummy := &node[int]{}
					queue := &ConcurrentLinkedBlockingQueue[int]{
						head: dummy,
						tail: dummy,
						cond: sync.NewCond(&sync.Mutex{}),
					}

					switch op {
					case "Enqueue":
						for _, val := range data {
							queue.Enqueue(val)
						}
					case "Dequeue":
						// 先入队
						for _, val := range data {
							queue.Enqueue(val)
						}
						// 然后出队
						for j := 0; j < size; j++ {
							_, _ = queue.Dequeue()
						}
					}
				}
			})

			b.Run("ConcurrentQueue", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					dummy := &node[int]{}
					queue := &ConcurrentLinkedBlockingQueue[int]{
						head: dummy,
						tail: dummy,
						cond: sync.NewCond(&sync.Mutex{}),
					}

					switch op {
					case "Enqueue":
						for _, val := range data {
							queue.Enqueue(val)
						}
					case "Dequeue":
						// 先入队
						for _, val := range data {
							queue.Enqueue(val)
						}
						// 然后出队
						for j := 0; j < size; j++ {
							_, _ = queue.Dequeue()
						}
					}
				}
			})
		})
	}
}

// 以下基准测试因为缺少相应实现暂时注释掉
/*
// 循环队列测试
func BenchmarkCircularQueue_Operations(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		capacity := size // 设置容量等于大小
		name := fmt.Sprintf("size=%d_capacity=%d", size, capacity)
		b.Run(name, func(b *testing.B) {
			data := prepareItems(size)

			b.Run("EnqueueDequeue", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					queue := NewCircularQueue[int](capacity)

					// 填满队列
					for j := 0; j < capacity; j++ {
						_ = queue.Enqueue(data[j])
					}

					// 出队一半
					for j := 0; j < capacity/2; j++ {
						_, _ = queue.Dequeue()
					}

					// 再入队一半
					for j := 0; j < capacity/2; j++ {
						_ = queue.Enqueue(data[j])
					}

					// 清空队列
					for !queue.IsEmpty() {
						_, _ = queue.Dequeue()
					}
				}
			})
		})
	}
}

// 有界队列测试
func BenchmarkBoundedQueue_CapacityLimit(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		name := fmt.Sprintf("size=%d", size)
		b.Run(name, func(b *testing.B) {
			data := prepareItems(size * 2) // 准备两倍大小的数据

			b.Run("EnqueueUntilFull", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					queue := NewBoundedQueue[int](size)

					// 尝试入队超过容量的元素
					for j := 0; j < size*2; j++ {
						_ = queue.Enqueue(data[j])
						// 忽略错误，我们预期队列会被填满
					}

					// 验证队列长度
					if queue.Len() != size {
						b.Fatalf("预期队列长度为 %d, 实际为 %d", size, queue.Len())
					}
				}
			})
		})
	}
}
*/
