# GGU DataStructures - 泛型数据结构库

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-Apache_2.0-blue?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Type-Library-green?style=for-the-badge" alt="Type">
</p>

## 项目描述

GGU DataStructures 是一个基于 Go 泛型的高性能数据结构库，提供了丰富的数据结构实现，满足各种开发场景需求。

### 核心特性

- **全泛型支持**: 所有数据结构均基于 Go 1.18+ 泛型实现，类型安全且代码复用
- **高性能实现**: 经过基准测试优化，确保各种操作的高效执行
- **丰富的数据结构**: 提供集合、元组、列表、队列、映射等多种基础数据结构
- **并发安全选项**: 大多数数据结构提供并发安全的实现版本
- **功能扩展**: 包含排序、过滤、映射等丰富的功能扩展

### 适用场景

- 需要泛型数据结构的 Go 项目
- 需要高效数据操作的业务系统
- 构建复杂数据处理逻辑的场景
- 追求代码简洁和类型安全的开发环境

## 快速开始

### 安装

```bash
go get github.com/Humphrey-He/go-generic-utils
```

### 基础用法

```go
package main

import (
	"fmt"
	
	"github.com/Humphrey-He/go-generic-utils/dataStructures/set"
	"github.com/Humphrey-He/go-generic-utils/dataStructures/list"
	"github.com/Humphrey-He/go-generic-utils/dataStructures/queue"
	"github.com/Humphrey-He/go-generic-utils/dataStructures/tuple"
	"github.com/Humphrey-He/go-generic-utils/dataStructures/maputils"
)

func main() {
	// 使用集合
	intSet := set.NewMapSet[int](0)
	intSet.Add(1, 2, 3, 4, 5)
	fmt.Println("Set contains 3:", intSet.Exist(3))
	
	// 使用列表
	intList := list.NewArrayList[int](0)
	intList.Append(10, 20, 30)
	intList.Add(1, 15) // 在索引1处添加元素
	fmt.Println("List:", intList.AsSlice())
	
	// 使用队列
	intQueue := queue.NewArrayQueue[int]()
	intQueue.Enqueue(100, 200, 300)
	value, _ := intQueue.Dequeue()
	fmt.Println("Dequeued value:", value)
	
	// 使用元组
	pair := tuple.NewPair("key", 100)
	fmt.Println("Pair:", pair.First, pair.Second)
	
	// 使用映射工具
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := maputils.Keys(m)
	fmt.Println("Map keys:", keys)
}
```

## 包结构与功能

### 1. 集合包 (set)

提供各种集合实现，支持集合运算、并发操作和元素过期。

#### 主要实现

- **MapSet**: 基于 map 的高效集合实现
- **TreeSet**: 有序集合实现，基于比较器排序
- **ConcurrentSet**: 线程安全的集合实现
- **ExpirableSet**: 带元素过期功能的集合

#### 示例

```go
// 创建集合并添加元素
set := set.NewMapSet[string](0)
set.Add("apple", "banana", "orange")

// 检查元素是否存在
exists := set.Exist("apple") // true

// 集合操作
otherSet := set.NewMapSet[string](0)
otherSet.Add("banana", "grape", "pear")

// 并集
unionSet := set.Union(otherSet) // [apple, banana, orange, grape, pear]

// 交集
intersectSet := set.Intersect(otherSet) // [banana]

// 差集
diffSet := set.Difference(otherSet) // [apple, orange]
```

### 2. 元组包 (tuple)

提供键值对和三元组的数据结构，支持元组操作和转换。

#### 主要实现

- **Pair**: 二元组，表示键值对
- **Triple**: 三元组，包含三个相关值
- **Pairs**: 键值对切片及相关操作

#### 示例

```go
// 创建键值对
pair := tuple.NewPair("id", 12345)
fmt.Println(pair.First, pair.Second) // id 12345

// 创建三元组
triple := tuple.NewTriple("product", 29.99, true)
fmt.Println(triple.First, triple.Second, triple.Third) // product 29.99 true

// 键值对批量操作
keys := []string{"name", "age", "city"}
values := []interface{}{"John", 30, "New York"}
pairs, _ := tuple.NewPairs(keys, values)

// 映射转换
result := tuple.Map(pairs, func(k string, v interface{}) (string, interface{}) {
    if k == "age" {
        return k, v.(int) + 1
    }
    return k, v
})
```

### 3. 列表包 (list)

提供各种列表实现，支持元素增删改查、排序、分页等功能。

#### 主要实现

- **ArrayList**: 基于切片的列表实现，随机访问高效
- **LinkedList**: 基于双向链表的实现，插入删除高效
- **ConcurrentList**: 线程安全的列表实现
- **ArrayListPaged**: 支持分页的列表
- **ArrayListSorted**: 保持元素有序的列表

#### 示例

```go
// 创建并填充列表
list := list.NewArrayList[int](0)
list.Append(1, 2, 3, 4, 5)

// 在特定位置添加元素
list.Add(2, 10) // [1, 2, 10, 3, 4, 5]

// 获取元素
val, _ := list.Get(0) // 1

// 删除元素
list.Delete(1) // [1, 10, 3, 4, 5]

// 排序
list.Sort(func(a, b int) bool {
    return a > b
})
// [10, 5, 4, 3, 1]

// 分页列表
pagedList := list.NewArrayListPaged[string](0)
// 添加多项...
page, total, _ := pagedList.Page(1, 10) // 获取第一页，每页10项
```

### 4. 队列包 (queue)

提供各种队列实现，包括普通队列、优先队列、延迟队列等。

#### 主要实现

- **ArrayQueue**: 基于数组的队列实现
- **LinkedQueue**: 基于链表的队列实现
- **PriorityQueue**: 优先队列，根据优先级出队
- **DelayQueue**: 延迟队列，元素在指定延迟后可用
- **CircularQueue**: 循环队列，固定容量循环使用
- **ConcurrentQueue**: 线程安全的队列实现

#### 示例

```go
// 基本队列操作
queue := queue.NewArrayQueue[string]()
queue.Enqueue("first", "second", "third")

value, _ := queue.Dequeue() // "first"
size := queue.Len() // 2

// 优先队列
pq := queue.NewPriorityQueue[string]()
pq.EnqueueWithPriority("normal", 0)
pq.EnqueueWithPriority("important", 10)
pq.EnqueueWithPriority("critical", 20)

value, _ = pq.Dequeue() // "critical"

// 延迟队列
delayQ := queue.NewDelayQueue[int]()
delayQ.EnqueueWithDelay(100, 5*time.Second) // 5秒后可取
```

### 5. 映射工具包 (maputils)

提供各种映射操作工具函数和并发安全的映射实现。

#### 主要功能

- **键值操作**: 提取键值、键值映射转换
- **映射转换**: 过滤、合并、转换等操作
- **安全映射**: 并发安全的映射实现
- **遍历工具**: 便捷的映射遍历函数

#### 示例

```go
// 提取键值
m := map[string]int{"a": 1, "b": 2, "c": 3}
keys := maputils.Keys(m) // ["a", "b", "c"]
values := maputils.Values(m) // [1, 2, 3]

// 映射转换
newMap := maputils.MapKeys(m, func(key string) string {
    return key + "_new"
}) // {"a_new": 1, "b_new": 2, "c_new": 3}

// 过滤
filteredMap := maputils.Filter(m, func(key string, val int) bool {
    return val > 1
}) // {"b": 2, "c": 3}

// 安全映射
safeMap := maputils.NewSafeMap[string, int]()
safeMap.Set("counter", 0)
safeMap.ComputeIfPresent("counter", func(oldVal int) int {
    return oldVal + 1
})
```

## 性能基准测试

每个数据结构包都包含完整的基准测试，用于评估不同操作的性能。运行基准测试:

```bash
cd dataStructures/set
go test -bench=. -benchmem
```

## FAQ

### 为什么使用 GGU DataStructures 而不是标准库?

- 标准库缺少泛型数据结构
- GGU 提供更丰富的数据结构和算法
- 提供并发安全的实现选项
- 针对不同场景提供多种实现

### 如何选择合适的数据结构实现?

- 随机访问频繁: ArrayList, MapSet
- 频繁插入删除: LinkedList, LinkedQueue
- 并发环境: ConcurrentList, ConcurrentSet, ConcurrentQueue
- 需要排序: TreeSet, ArrayListSorted
- 需要过期机制: ExpirableSet, DelayQueue

### 泛型约束和限制?

大多数数据结构支持任意类型 (any)，但有些特定实现可能有约束:
- TreeSet 要求元素可比较
- 某些操作可能需要自定义比较器
- 对于特定场景，可能需要实现特定接口

### 线程安全与性能平衡?

- 非并发版本: 单线程性能最佳
- 并发版本: 适合多线程访问，但有同步开销
- 根据实际场景选择适当实现，避免不必要的同步开销

## 联系方式

- 作者：Humphrey
- 电子邮箱：steve1484121793@gmail.com

若有任何问题或建议，欢迎通过上述方式联系我们。
