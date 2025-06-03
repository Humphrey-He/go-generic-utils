# pool - 对象池与任务池

`pool`包提供了泛型对象池和任务池（协程池）的实现，帮助减少GC压力、提高内存使用效率，并简化并发任务处理。

## 核心特性

- **泛型对象池**：基于Go 1.18+泛型特性，提供类型安全的对象池实现
- **任务池（协程池）**：管理和复用goroutine，避免频繁创建销毁带来的开销
- **超时控制**：支持任务执行超时控制
- **线程安全**：所有实现都保证并发安全
- **低内存占用**：优化的内存管理，减少GC压力

## 使用示例

### 泛型对象池

对象池用于复用频繁创建和销毁的对象，减少GC压力。

```go
// 创建一个bytes.Buffer对象池
bufferPool := pool.NewSimpleObjectPool(func() *bytes.Buffer {
    return &bytes.Buffer{}
})

// 从池中获取对象
buf := bufferPool.Get()

// 使用对象
buf.WriteString("Hello, World!")

// 使用完毕后归还对象
bufferPool.Put(buf)
```

### 固定大小的任务池

任务池用于控制并发任务的执行，避免创建过多goroutine。

```go
// 创建一个容量为10的任务池
taskPool := pool.NewFixedTaskPool(10)

// 提交任务
for i := 0; i < 100; i++ {
    i := i // 捕获变量
    _ = taskPool.Submit(func() {
        // 任务逻辑
        fmt.Printf("处理任务 %d\n", i)
        time.Sleep(100 * time.Millisecond)
    })
}

// 等待所有任务完成并关闭池
taskPool.Shutdown()
```

### 带超时的任务池

超时任务池可以为每个任务设置最大执行时间，避免任务阻塞。

```go
// 创建一个容量为5，超时时间为2秒的任务池
timeoutPool := pool.NewTimeoutTaskPool(5, 2*time.Second)

// 提交可能超时的任务
err := timeoutPool.Submit(func() {
    // 模拟耗时操作
    time.Sleep(1 * time.Second)
    fmt.Println("任务完成")
})

if err != nil {
    fmt.Println("任务执行出错:", err)
}

// 提交一个会超时的任务
err = timeoutPool.Submit(func() {
    time.Sleep(3 * time.Second) // 会超时
    fmt.Println("这条消息不会打印")
})

if err != nil {
    fmt.Println("任务执行出错:", err) // 会输出超时错误
}

// 关闭池
timeoutPool.Shutdown()
```

## 性能优化

### 对象池最佳实践

1. **对象重置**：归还对象前重置其状态，避免状态泄露
   ```go
   buf.Reset() // 重置buffer状态
   bufferPool.Put(buf)
   ```

2. **适用场景**：对于创建成本高或频繁创建销毁的对象使用对象池
   - 适合：`bytes.Buffer`, `[]byte`缓冲区, 数据库连接等
   - 不适合：简单的小对象，如基本类型

3. **避免池泄漏**：确保所有从池中获取的对象最终都被归还

### 任务池最佳实践

1. **合理设置容量**：根据CPU核心数和任务特性设置合适的池容量
   ```go
   taskPool := pool.NewFixedTaskPool(runtime.NumCPU())
   ```

2. **任务粒度**：避免提交过于细粒度的任务，增加调度开销
   - 推荐将相关的小任务合并为一个较大的任务提交

3. **错误处理**：在任务内部妥善处理错误，避免panic导致工作协程退出

## 高级用法

### 自定义对象生命周期

```go
type MyConnection struct {
    // ...
}

// 创建连接池
connectionPool := pool.NewSimpleObjectPool(func() *MyConnection {
    // 创建新连接
    conn := &MyConnection{}
    conn.Connect("localhost:8080")
    return conn
})

// 获取连接
conn := connectionPool.Get()

// 使用连接
conn.Execute("SELECT * FROM users")

// 重置连接状态后归还
conn.Reset()
connectionPool.Put(conn)
```

### 批量任务处理

```go
// 创建任务池
taskPool := pool.NewFixedTaskPool(10)

// 等待组用于同步
var wg sync.WaitGroup

// 批量提交任务
for i := 0; i < 100; i++ {
    i := i
    wg.Add(1)
    _ = taskPool.Submit(func() {
        defer wg.Done()
        processItem(i)
    })
}

// 等待所有任务完成
wg.Wait()

// 关闭池
taskPool.Shutdown()
``` 