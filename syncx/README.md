# syncx - 同步原语增强

`syncx`包提供了对Go标准库`sync`包的增强，提供更易用的同步原语和并发工具。该包充分利用Go 1.18+的泛型特性，提供类型安全的并发工具。

## 核心特性

- **增强的互斥锁和读写锁**：支持带返回值的锁定函数
- **泛型Map**：提供类型安全的并发Map实现
- **泛型Pool**：提供类型安全的对象池实现，包括限制大小的对象池
- **条件变量**：增强的Cond实现，支持超时控制
- **分段锁**：基于键的分段锁实现，减少锁竞争
- **原子值**：泛型的原子值实现，支持任意类型的原子操作

## 使用示例

### 泛型Map

```go
// 创建一个string->int的并发安全Map
m := new(syncx.Map[string, int])

// 存储键值对
m.Store("one", 1)
m.Store("two", 2)

// 加载键值对
val, ok := m.Load("one")
if ok {
    fmt.Println(val) // 输出: 1
}

// 遍历所有键值对
m.Range(func(key string, value int) bool {
    fmt.Printf("%s: %d\n", key, value)
    return true
})
```

### 泛型Pool

```go
// 创建一个bytes.Buffer对象池
bufferPool := syncx.NewPool(func() *bytes.Buffer {
    return &bytes.Buffer{}
})

// 获取一个Buffer
buf := bufferPool.Get()

// 使用Buffer
buf.WriteString("Hello, World!")

// 归还Buffer
bufferPool.Put(buf)

// 创建限制大小的对象池
limitedPool := syncx.NewLimitPool(100, func() *bytes.Buffer {
    return &bytes.Buffer{}
})

// 尝试获取对象，如果池已满则返回false
buf, ok := limitedPool.Get()
if ok {
    defer limitedPool.Put(buf)
    // 使用buf
}
```

### 分段锁

```go
// 创建一个有64个分段的锁
segmentLock := syncx.NewSegmentKeysLock(64)

// 锁定特定键
segmentLock.Lock("user:123")
// 临界区操作
segmentLock.Unlock("user:123")

// 读锁
segmentLock.RLock("user:123")
// 只读操作
segmentLock.RUnlock("user:123")
```

### 泛型原子值

```go
// 创建一个Config类型的原子值
configValue := syncx.NewValue[Config]()

// 存储值
configValue.Store(Config{Timeout: 30})

// 加载值
config := configValue.Load()

// 原子交换
oldConfig := configValue.Swap(Config{Timeout: 60})

// 比较并交换
swapped := configValue.CompareAndSwap(oldConfig, Config{Timeout: 90})
```

## 性能考量

- 分段锁设计减少了高并发场景下的锁竞争
- 泛型实现避免了类型断言和反射带来的性能开销
- 对象池实现减少了频繁创建临时对象的GC压力

## 最佳实践

- 对于需要频繁创建和销毁的对象，使用`Pool`或`LimitPool`
- 对于并发访问的共享数据，使用`Map`而非手动加锁的标准map
- 对于基于键的并发访问场景，使用`SegmentKeysLock`减少锁竞争
- 对于需要原子更新的复杂类型，使用`Value`而非手动加锁 