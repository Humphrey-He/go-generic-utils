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

package maputils

import (
	"strconv"
	"sync"
	"testing"
)

// 准备基准测试数据
func prepareIntStringMap(n int) map[int]string {
	m := make(map[int]string, n)
	for i := 0; i < n; i++ {
		m[i] = "value" + strconv.Itoa(i)
	}
	return m
}

func prepareStringIntMap(n int) map[string]int {
	m := make(map[string]int, n)
	for i := 0; i < n; i++ {
		m["key"+strconv.Itoa(i)] = i
	}
	return m
}

// 基准测试: 键值转换函数
func BenchmarkKeys(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareIntStringMap(size)
			gm := NewGenericMap[int, string]()
			for k, v := range m {
				gm.Set(k, v)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = gm.Keys()
			}
		})
	}
}

func BenchmarkValues(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareIntStringMap(size)
			gm := NewGenericMap[int, string]()
			for k, v := range m {
				gm.Set(k, v)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = gm.Values()
			}
		})
	}
}

// 基准测试: 映射转换函数 - 实现简化版用于测试
func MapKeys[K1 comparable, V any, K2 comparable](m map[K1]V, f func(K1) K2) map[K2]V {
	result := make(map[K2]V, len(m))
	for k, v := range m {
		result[f(k)] = v
	}
	return result
}

func BenchmarkMapKeys(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareIntStringMap(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = MapKeys(m, func(key int) string {
					return strconv.Itoa(key)
				})
			}
		})
	}
}

// 实现简化版用于测试
func MapValues[K comparable, V1 any, V2 any](m map[K]V1, f func(V1) V2) map[K]V2 {
	result := make(map[K]V2, len(m))
	for k, v := range m {
		result[k] = f(v)
	}
	return result
}

func BenchmarkMapValues(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareIntStringMap(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = MapValues(m, func(val string) int {
					// 简单的转换操作
					return len(val)
				})
			}
		})
	}
}

// 基准测试: 映射合并函数 - 实现简化版用于测试
func Merge[K comparable, V any](m1, m2 map[K]V) map[K]V {
	result := make(map[K]V, len(m1)+len(m2))
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

func BenchmarkMerge(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m1 := prepareIntStringMap(size / 2)
			m2 := make(map[int]string, size/2)
			// 创建一个部分重叠的第二个映射
			for i := size / 4; i < size*3/4; i++ {
				m2[i] = "newvalue" + strconv.Itoa(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Merge(m1, m2)
			}
		})
	}
}

// 实现简化版用于测试
func MergeWithStrategy[K comparable, V any](m1, m2 map[K]V, strategy func(K, V, V) V) map[K]V {
	result := make(map[K]V, len(m1)+len(m2))

	// 先复制m1
	for k, v := range m1 {
		result[k] = v
	}

	// 处理m2
	for k, v2 := range m2 {
		if v1, exists := result[k]; exists {
			result[k] = strategy(k, v1, v2)
		} else {
			result[k] = v2
		}
	}

	return result
}

func BenchmarkMergeWithStrategy(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m1 := prepareIntStringMap(size / 2)
			m2 := make(map[int]string, size/2)
			// 创建一个部分重叠的第二个映射
			for i := size / 4; i < size*3/4; i++ {
				m2[i] = "newvalue" + strconv.Itoa(i)
			}

			// 使用优先保留第一个映射值的策略
			strategy := func(key int, v1, v2 string) string {
				return v1
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = MergeWithStrategy(m1, m2, strategy)
			}
		})
	}
}

// 基准测试: 过滤函数 - 实现简化版用于测试
func Filter[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

func BenchmarkFilter(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareIntStringMap(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 过滤偶数键
				_ = Filter(m, func(key int, val string) bool {
					return key%2 == 0
				})
			}
		})
	}
}

// 基准测试: 转换函数 - 实现简化版用于测试
func Transform[K1 comparable, V1 any, K2 comparable, V2 any](m map[K1]V1, transformer func(K1, V1) (K2, V2)) map[K2]V2 {
	result := make(map[K2]V2, len(m))
	for k, v := range m {
		newK, newV := transformer(k, v)
		result[newK] = newV
	}
	return result
}

func BenchmarkTransform(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareStringIntMap(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 转换为 int -> string 映射
				_ = Transform(m, func(key string, val int) (int, string) {
					return val, key
				})
			}
		})
	}
}

// 基准测试: 遍历函数 - 实现简化版用于测试
func ForEach[K comparable, V any](m map[K]V, action func(K, V)) {
	for k, v := range m {
		action(k, v)
	}
}

func BenchmarkForEach(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			m := prepareIntStringMap(size)
			sum := 0

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sum = 0 // 重置累加器
				ForEach(m, func(key int, val string) {
					sum += key
				})
			}
		})
	}
}

// 基准测试: 安全并发Map
func BenchmarkSafeMap(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			b.Run("Set", func(b *testing.B) {
				safeMap := NewSyncMap[int, string]()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for j := 0; j < size; j++ {
						safeMap.Set(j, "value"+strconv.Itoa(j))
					}
				}
			})

			b.Run("Get", func(b *testing.B) {
				safeMap := NewSyncMap[int, string]()
				for j := 0; j < size; j++ {
					safeMap.Set(j, "value"+strconv.Itoa(j))
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for j := 0; j < size; j++ {
						_, _ = safeMap.Get(j)
					}
				}
			})

			b.Run("Delete", func(b *testing.B) {
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					b.StopTimer()
					safeMap := NewSyncMap[int, string]()
					for j := 0; j < size; j++ {
						safeMap.Set(j, "value"+strconv.Itoa(j))
					}
					b.StartTimer()

					for j := 0; j < size; j++ {
						safeMap.Delete(j)
					}
				}
			})
		})
	}
}

// 基准测试: 并发操作对比
func BenchmarkConcurrentMapOperations(b *testing.B) {
	sizes := []int{1000, 10000}
	concurrencyLevels := []int{2, 4, 8, 16}

	for _, size := range sizes {
		for _, concurrency := range concurrencyLevels {
			name := "size=" + strconv.Itoa(size) + "_concurrency=" + strconv.Itoa(concurrency)
			b.Run(name, func(b *testing.B) {
				b.Run("SafeMap", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						safeMap := NewSyncMap[int, string]()
						wg := sync.WaitGroup{}
						wg.Add(concurrency)

						// 每个goroutine处理的元素数量
						itemsPerGoroutine := size / concurrency

						for j := 0; j < concurrency; j++ {
							start := j * itemsPerGoroutine
							end := start + itemsPerGoroutine
							if j == concurrency-1 {
								end = size // 确保处理所有元素
							}

							go func(startIdx, endIdx int) {
								defer wg.Done()
								// 写入
								for k := startIdx; k < endIdx; k++ {
									safeMap.Set(k, "value"+strconv.Itoa(k))
								}

								// 读取
								for k := startIdx; k < endIdx; k++ {
									_, _ = safeMap.Get(k)
								}

								// 删除一部分
								for k := startIdx; k < startIdx+itemsPerGoroutine/2; k++ {
									safeMap.Delete(k)
								}
							}(start, end)
						}

						wg.Wait()
					}
				})

				b.Run("SyncMap", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						var syncMap sync.Map
						wg := sync.WaitGroup{}
						wg.Add(concurrency)

						// 每个goroutine处理的元素数量
						itemsPerGoroutine := size / concurrency

						for j := 0; j < concurrency; j++ {
							start := j * itemsPerGoroutine
							end := start + itemsPerGoroutine
							if j == concurrency-1 {
								end = size // 确保处理所有元素
							}

							go func(startIdx, endIdx int) {
								defer wg.Done()
								// 写入
								for k := startIdx; k < endIdx; k++ {
									syncMap.Store(k, "value"+strconv.Itoa(k))
								}

								// 读取
								for k := startIdx; k < endIdx; k++ {
									_, _ = syncMap.Load(k)
								}

								// 删除一部分
								for k := startIdx; k < startIdx+itemsPerGoroutine/2; k++ {
									syncMap.Delete(k)
								}
							}(start, end)
						}

						wg.Wait()
					}
				})

				b.Run("StdMapWithMutex", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						stdMap := make(map[int]string, size)
						var mutex sync.RWMutex
						wg := sync.WaitGroup{}
						wg.Add(concurrency)

						// 每个goroutine处理的元素数量
						itemsPerGoroutine := size / concurrency

						for j := 0; j < concurrency; j++ {
							start := j * itemsPerGoroutine
							end := start + itemsPerGoroutine
							if j == concurrency-1 {
								end = size // 确保处理所有元素
							}

							go func(startIdx, endIdx int) {
								defer wg.Done()
								// 写入
								for k := startIdx; k < endIdx; k++ {
									mutex.Lock()
									stdMap[k] = "value" + strconv.Itoa(k)
									mutex.Unlock()
								}

								// 读取
								for k := startIdx; k < endIdx; k++ {
									mutex.RLock()
									_ = stdMap[k]
									mutex.RUnlock()
								}

								// 删除一部分
								for k := startIdx; k < startIdx+itemsPerGoroutine/2; k++ {
									mutex.Lock()
									delete(stdMap, k)
									mutex.Unlock()
								}
							}(start, end)
						}

						wg.Wait()
					}
				})
			})
		}
	}
}

// 定义SafeMap简单版本
type safeMapCompute struct {
	sync.RWMutex
	data map[string]int
}

func newSafeMapCompute() *safeMapCompute {
	return &safeMapCompute{
		data: make(map[string]int),
	}
}

func (m *safeMapCompute) Get(key string) (int, bool) {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *safeMapCompute) Set(key string, value int) {
	m.Lock()
	defer m.Unlock()
	m.data[key] = value
}

func (m *safeMapCompute) ComputeIfAbsent(key string, compute func() int) int {
	m.Lock()
	defer m.Unlock()
	if val, ok := m.data[key]; ok {
		return val
	}
	val := compute()
	m.data[key] = val
	return val
}

func (m *safeMapCompute) GetOrDefault(key string, defaultVal int) int {
	m.RLock()
	defer m.RUnlock()
	if val, ok := m.data[key]; ok {
		return val
	}
	return defaultVal
}

// 基准测试: 特殊Map功能
func BenchmarkSpecialMapFeatures(b *testing.B) {
	size := 1000

	b.Run("ComputeIfAbsent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			safeMap := newSafeMapCompute()

			for j := 0; j < size; j++ {
				key := "key" + strconv.Itoa(j%100) // 使用循环键，促使ComputeIfAbsent功能被测试

				// 如果键不存在，计算并添加
				safeMap.ComputeIfAbsent(key, func() int {
					return j * 2
				})
			}
		}
	})

	b.Run("GetOrDefault", func(b *testing.B) {
		safeMap := newSafeMapCompute()

		// 填充一部分数据
		for j := 0; j < size/2; j++ {
			safeMap.Set("key"+strconv.Itoa(j), j)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < size; j++ {
				// 一半存在，一半不存在，使用默认值
				_ = safeMap.GetOrDefault("key"+strconv.Itoa(j), -1)
			}
		}
	})
}
