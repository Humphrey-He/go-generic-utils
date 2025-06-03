# GGU (Go Generic Utils) 技术使用说明文档

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-Apache_2.0-blue?style=for-the-badge" alt="License">
  <img src="https://img.shields.io/badge/Type-Library-green?style=for-the-badge" alt="Type">
</p>

## 目录

- [简介](#简介)
- [安装方法](#安装方法)
- [模块使用指南](#模块使用指南)
  - [数据结构 (dataStructures)](#数据结构-datastructures)
  - [切片工具 (sliceutils)](#切片工具-sliceutils)
  - [树结构 (tree)](#树结构-tree)
  - [同步工具 (syncx)](#同步工具-syncx)
  - [重试机制 (retry)](#重试机制-retry)
  - [对象池 (pool)](#对象池-pool)
  - [网络工具 (net)](#网络工具-net)
  - [Gin框架增强 (ginutil)](#gin框架增强-ginutil)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)
- [性能考量](#性能考量)
- [调试技巧](#调试技巧)
- [贡献指南](#贡献指南)

## 简介

GGU (Go Generic Utils) 是一个基于 Go 1.18+ 泛型特性开发的工具库，为 Go 开发者提供了丰富的数据结构、算法和实用工具，帮助您更高效地构建应用程序。GGU 专注于类型安全、高性能和易用性，适用于各种规模的项目。

### 主要特点

- **泛型支持**：利用 Go 1.18+ 的泛型特性提供类型安全的 API
- **模块化设计**：各模块可独立使用，也可组合使用
- **高性能实现**：所有数据结构和算法都经过优化，适合高并发场景
- **全面测试**：完整的单元测试和基准测试保证质量
- **实用工具集**：涵盖日常开发中常用的功能和数据结构
- **Web开发支持**：与Gin框架深度集成

## 安装方法

使用 Go 模块安装 GGU：

```bash
go get github.com/Humphrey-He/go-generic-utils
```

在您的代码中导入所需的包：

```go
import (
    "github.com/Humphrey-He/go-generic-utils/dataStructures/set"
    "github.com/Humphrey-He/go-generic-utils/sliceutils"
    "github.com/Humphrey-He/go-generic-utils/tree"
    // 根据需要导入其他包
)
```

## 模块使用指南

### 数据结构 (dataStructures)

GGU提供了丰富的数据结构实现，包括集合、列表、队列、映射等。

#### 集合 (set)

```go
// 创建集合
intSet := set.NewMapSet[int](0)
intSet.Add(1, 2, 3, 4, 5)

// 检查元素
exists := intSet.Exist(3) // true

// 集合操作
otherSet := set.NewMapSet[int](0)
otherSet.Add(3, 4, 5, 6, 7)

// 交集
intersection := intSet.Intersect(otherSet) // [3, 4, 5]

// 并集
union := intSet.Union(otherSet) // [1, 2, 3, 4, 5, 6, 7]

// 差集
difference := intSet.Difference(otherSet) // [1, 2]
```

#### 列表 (list)

```go
// 创建列表
strList := list.NewArrayList[string](0)
strList.Append("Go", "Java", "Python")

// 插入元素
strList.Add(1, "Rust") // ["Go", "Rust", "Java", "Python"]

// 获取元素
value, _ := strList.Get(0) // "Go"

// 删除元素
strList.Delete(2) // ["Go", "Rust", "Python"]

// 排序
strList.Sort(func(a, b string) bool {
    return a < b
}) // ["Go", "Python", "Rust"]
```

#### 队列 (queue)

```go
// 基本队列
queue := queue.NewArrayQueue[int]()
queue.Enqueue(1, 2, 3)
value, _ := queue.Dequeue() // 1

// 优先队列
pq := queue.NewPriorityQueue[string]()
pq.EnqueueWithPriority("低优先级任务", 0)
pq.EnqueueWithPriority("高优先级任务", 10)
value, _ = pq.Dequeue() // "高优先级任务"

// 延迟队列
delayQ := queue.NewDelayQueue[string]()
delayQ.EnqueueWithDelay("延迟消息", 5*time.Second)
// 5秒后可以取出
```

### 切片工具 (sliceutils)

提供了丰富的切片操作工具，大大简化了切片处理的代码量。

```go
// 切片查找
items := []int{1, 2, 3, 4, 5}
found, index := sliceutils.Contains(items, 3) // true, 2

// 切片过滤
evens := sliceutils.Filter(items, func(x int) bool {
    return x%2 == 0
}) // [2, 4]

// 切片映射
doubled := sliceutils.Map(items, func(x int) int {
    return x * 2
}) // [2, 4, 6, 8, 10]

// 切片去重
duplicated := []int{1, 2, 2, 3, 3, 3, 4}
unique := sliceutils.Deduplicate(duplicated) // [1, 2, 3, 4]

// 切片分组
grouped := sliceutils.GroupBy(items, func(x int) string {
    if x%2 == 0 {
        return "even"
    }
    return "odd"
}) // map[even:[2, 4], odd:[1, 3, 5]]

// 线程安全切片
safeSlice := sliceutils.NewThreadSafeSlice(items)
safeSlice.Append(6)
allItems := safeSlice.AsSlice() // [1, 2, 3, 4, 5, 6]
```

### 树结构 (tree)

GGU提供了多种树结构实现，包括AVL树、B树等，适用于需要高效查找和有序数据存储的场景。

#### AVL树

```go
// 创建AVL树
tree, _ := tree.NewAVLTree[int, string](tree.IntComparator)

// 添加节点
tree.Put(10, "数据1")
tree.Put(5, "数据2")
tree.Put(15, "数据3")

// 查找节点
value, found := tree.Get(10) // "数据1", true

// 范围查询
keys, values, _ := tree.FindRange(5, 12)
// keys: [5, 10]
// values: ["数据2", "数据1"]

// 删除节点
tree.Remove(5)

// 遍历
tree.ForEach(func(key int, value string) bool {
    fmt.Printf("键: %d, 值: %s\n", key, value)
    return true
})
```

#### 电商树

针对电商场景优化的树结构，支持商品分类、价格区间查询等功能。

```go
// 创建电商分类树
ecomTree := tree.NewEComTree[string, string]()

// 添加分类
ecomTree.AddCategory("电子产品")
ecomTree.AddCategory("电子产品/手机")
ecomTree.AddCategory("电子产品/手机/苹果")
ecomTree.AddCategory("电子产品/手机/三星")

// 添加商品
ecomTree.AddProduct("电子产品/手机/苹果", "iPhone 13", "P001")
ecomTree.AddProduct("电子产品/手机/三星", "Galaxy S21", "P002")

// 查询分类下所有商品
products := ecomTree.GetProductsByCategory("电子产品/手机")
// ["iPhone 13", "Galaxy S21"]

// 获取路径
path := ecomTree.GetCategoryPath("电子产品/手机/苹果")
// ["电子产品", "手机", "苹果"]
```

### 同步工具 (syncx)

提供了对标准库`sync`包的增强，包括更易用的互斥锁、读写锁、单次执行工具等。

```go
// 增强的Once
once := syncx.NewOnce()
result := once.DoWithResult(func() (interface{}, error) {
    // 计算一次性结果
    return "计算结果", nil
})

// 可重置的Once
resetOnce := syncx.NewResettableOnce()
resetOnce.Do(func() {
    // 执行初始化
})
// 重置后可再次执行
resetOnce.Reset()
resetOnce.Do(func() {
    // 再次执行初始化
})

// 信号量
sem := syncx.NewSemaphore(10)
sem.Acquire()
defer sem.Release()
// 执行受限制的并发操作

// 读写锁工具
rwm := syncx.NewRWMutex()
value, _ := rwm.RLockFunc(func() (interface{}, error) {
    // 读取共享资源
    return sharedResource, nil
})
```

### 重试机制 (retry)

提供了灵活的重试策略，帮助处理可能失败的操作，特别适用于网络请求和外部服务调用。

```go
// 创建重试器
retrier := retry.NewRetrier(
    retry.WithMaxAttempts(3),
    retry.WithBackoff(retry.ExponentialBackoff(100*time.Millisecond)),
    retry.WithTimeout(5*time.Second),
)

// 执行可能失败的操作
result, err := retrier.Run(func() (interface{}, error) {
    // 发起HTTP请求或其他可能失败的操作
    return http.Get("https://api.example.com/data")
})

// 自定义重试条件
customRetrier := retry.NewRetrier(
    retry.WithRetryIf(func(err error) bool {
        // 只有特定错误才重试
        return errors.Is(err, io.ErrUnexpectedEOF)
    }),
)
```

### 对象池 (pool)

提供了泛型对象池实现，帮助减少GC压力和内存分配。

```go
// 创建对象池
bufferPool := pool.NewPool(
    // 创建函数
    func() *bytes.Buffer {
        return &bytes.Buffer{}
    },
    // 重置函数
    func(buf *bytes.Buffer) {
        buf.Reset()
    },
)

// 获取对象
buf := bufferPool.Get()
defer bufferPool.Put(buf)

// 使用对象
buf.WriteString("测试数据")
data := buf.Bytes()

// 带过期清理的对象池
timeoutPool := pool.NewTimeoutPool(
    func() *SomeResource { return &SomeResource{} },
    func(r *SomeResource) { r.Close() },
    30*time.Second, // 30秒未使用将被清理
)
```

### 网络工具 (net)

提供了网络通信相关的工具和增强功能。

```go
// 创建HTTP客户端
client := net.NewHTTPClient(
    net.WithTimeout(5*time.Second),
    net.WithRetry(3),
    net.WithBackoff(net.ExponentialBackoff),
)

// 发送GET请求
resp, err := client.Get("https://api.example.com/data")

// 发送带自定义头的POST请求
headers := map[string]string{
    "Content-Type": "application/json",
    "Authorization": "Bearer token",
}
resp, err = client.PostWithHeaders(
    "https://api.example.com/users",
    strings.NewReader(`{"name":"张三"}`),
    headers,
)

// 健康检查
checker := net.NewHealthChecker(
    net.WithEndpoint("https://api.example.com/health"),
    net.WithInterval(30*time.Second),
)
checker.Start()
defer checker.Stop()

isHealthy := checker.IsHealthy() // 检查服务是否健康
```

### Gin框架增强 (ginutil)

提供了对Gin框架的各种增强功能，使Web开发更加高效。

#### 统一响应渲染

```go
// 配置渲染器
render.Configure(render.Config{
    JSONPrettyPrint: true,
    HTMLTemplateDir: "templates/*",
})

// 成功响应
func GetUser(c *gin.Context) {
    user := fetchUser(c.Param("id"))
    
    if user == nil {
        render.NotFound(c, "用户不存在")
        return
    }
    
    render.Success(c, user)
}

// 分页响应
func ListUsers(c *gin.Context) {
    page, size := paginator.GetPagination(c)
    users, total := getUserList(page, size)
    
    render.Paginated(c, users, total, page, size)
}

// 错误响应
func UpdateUser(c *gin.Context) {
    if err := updateUser(c); err != nil {
        render.Error(c, "更新用户失败", err)
        return
    }
    
    render.Success(c, nil)
}
```

#### 路由注册

```go
// 控制器定义
type UserController struct{}

func (u *UserController) Routes(group *gin.RouterGroup) {
    users := group.Group("/users")
    
    users.GET("", u.List)
    users.GET("/:id", u.Get)
    users.POST("", u.Create)
    users.PUT("/:id", u.Update)
    users.DELETE("/:id", u.Delete)
}

func (u *UserController) List(c *gin.Context) {
    // 实现列表逻辑
}

// 自动注册路由
func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    api := r.Group("/api")
    register.RegisterRoutes(api,
        &UserController{},
        &ProductController{},
        &OrderController{},
    )
    
    return r
}
```

#### 中间件

```go
// 跟踪ID中间件
r.Use(middleware.TraceID())

// JWT认证中间件
protected := r.Group("/api")
protected.Use(middleware.JWT(middleware.JWTConfig{
    SecretKey: "your-secret-key",
    TokenLookup: "header:Authorization",
}))

// 限流中间件
r.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
    Rate: 100,
    Burst: 50,
}))

// CORS中间件
r.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: []string{"https://example.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))
```

## 最佳实践

### 性能优化

1. **预分配切片容量**：使用sliceutils时，尽可能预估切片大小
   
   ```go
   // 好的做法
   result := make([]int, 0, len(sourceSlice))
   
   // 或使用库提供的函数
   result := sliceutils.NewWithCapacity[int](len(sourceSlice))
   ```

2. **合理使用对象池**：对于频繁创建和销毁的对象，使用对象池
   
   ```go
   // 定义全局池
   var bufferPool = pool.NewPool(
       func() *bytes.Buffer { return &bytes.Buffer{} },
       func(b *bytes.Buffer) { b.Reset() },
   )
   
   func ProcessData(data []byte) string {
       buf := bufferPool.Get()
       defer bufferPool.Put(buf)
       
       // 使用buf处理数据
       return buf.String()
   }
   ```

3. **避免不必要的反射**：尽管泛型已经减少了反射需求，但仍应避免不必要的反射操作

### 并发安全

1. **使用线程安全的数据结构**：在并发环境中，使用库提供的线程安全实现
   
   ```go
   // 线程安全的切片
   safeSlice := sliceutils.NewThreadSafeSlice([]int{})
   
   // 线程安全的映射
   safeMap := maputils.NewSafeMap[string, int]()
   ```

2. **使用同步原语增强工具**：利用syncx包提供的工具简化同步逻辑
   
   ```go
   // 使用增强的读写锁
   rwm := syncx.NewRWMutex()
   rwm.LockFunc(func() {
       // 写入共享资源
   })
   ```

### Gin应用开发

1. **统一错误处理**：使用render包提供统一的错误响应格式
   
   ```go
   // 定义错误码
   var (
       ErrUserNotFound = ecode.New(1001, "用户不存在")
       ErrInvalidParam = ecode.New(1002, "无效的参数")
   )
   
   // 使用错误码
   if user == nil {
       render.WithError(c, ErrUserNotFound)
       return
   }
   ```

2. **结构化路由组织**：使用register包组织API路由
   
   ```go
   // 按领域组织控制器
   register.RegisterRoutes(api,
       // 用户领域
       &UserController{},
       &RoleController{},
       
       // 商品领域
       &ProductController{},
       &CategoryController{},
   )
   ```

## 常见问题

### 泛型约束相关问题

**问题**：为什么我的自定义类型无法用于某些泛型函数？

**解答**：GGU中的某些泛型函数对类型有特定约束。例如，树结构需要可比较的类型，或者提供一个比较器函数。确保您的类型满足相应的约束，或提供适当的比较器。

```go
// 为自定义类型提供比较器
type Product struct {
    ID    int
    Name  string
    Price float64
}

productComparator := func(a, b Product) int {
    if a.ID < b.ID {
        return -1
    }
    if a.ID > b.ID {
        return 1
    }
    return 0
}

// 创建使用自定义比较器的树
tree, _ := tree.NewAVLTree[Product, string](productComparator)
```

### 并发安全问题

**问题**：我的程序在并发环境下出现数据竞争，如何解决？

**解答**：确保在并发环境中使用线程安全的数据结构或适当的同步机制。GGU提供了许多线程安全的实现：

```go
// 使用线程安全的集合
safeSet := set.NewConcurrentSet[int]()

// 使用线程安全的队列
safeQueue := queue.NewConcurrentQueue[string]()

// 使用同步工具保护自定义逻辑
mutex := syncx.NewMutex()
mutex.LockFunc(func() {
    // 执行需要保护的操作
})
```

### 内存泄漏问题

**问题**：使用对象池后，内存使用量持续增长，可能存在内存泄漏？

**解答**：使用对象池时，务必确保正确归还对象，并考虑使用带超时清理的池。

```go
// 正确使用
obj := pool.Get()
defer pool.Put(obj) // 确保归还对象

// 使用带超时清理的池
timeoutPool := pool.NewTimeoutPool(
    createFunc,
    resetFunc,
    5*time.Minute, // 5分钟未使用的对象将被清理
)
```

## 性能考量

GGU库的各个组件都经过性能优化，但使用时仍需注意以下几点：

1. **选择合适的数据结构**：根据实际需求选择合适的数据结构
   - 频繁随机访问：使用ArrayList而非LinkedList
   - 频繁查找：使用MapSet而非普通切片
   - 有序数据：使用AVL树或B树而非普通Map

2. **避免不必要的类型转换**：利用泛型特性避免频繁的类型转换

3. **大数据集优化**：处理大数据集时，考虑分批处理和并行处理

```go
// 大数据集分批处理
batches := sliceutils.Chunk(largeSlice, 1000)
for _, batch := range batches {
    processBatch(batch)
}

// 并行处理
results := sliceutils.ParallelMap(largeSlice, func(item int) string {
    return processItem(item)
}, runtime.NumCPU())
```

## 调试技巧

1. **启用调试日志**：部分模块支持调试日志
   
   ```go
   // 启用网络模块的调试日志
   net.SetDebugMode(true)
   ```

2. **性能分析**：使用Go的性能分析工具分析性能瓶颈
   
   ```go
   import "github.com/Humphrey-He/go-generic-utils/pkg/profiler"
   
   // 启用CPU分析
   defer profiler.StartCPUProfile("cpu.prof").Stop()
   
   // 程序执行...
   
   // 分析：go tool pprof cpu.prof
   ```

3. **数据结构可视化**：某些数据结构支持导出可视化表示
   
   ```go
   // 导出树结构的DOT格式
   dot := tree.ExportDOT()
   ioutil.WriteFile("tree.dot", []byte(dot), 0644)
   // 使用Graphviz可视化：dot -Tpng tree.dot -o tree.png
   ```

## 贡献指南

我们欢迎您对GGU项目做出贡献。请遵循以下步骤：

1. Fork项目仓库
2. 创建您的特性分支：`git checkout -b feature/amazing-feature`
3. 提交您的更改：`git commit -m 'Add some amazing feature'`
4. 推送到分支：`git push origin feature/amazing-feature`
5. 提交Pull Request

在提交PR前，请确保：
- 所有测试通过：`make test`
- 代码符合格式规范：`make fmt`
- 添加了必要的文档和单元测试
- 更新了CHANGELOG.md文件（如适用）

## 联系我们

如有任何问题或建议，请通过以下方式联系我们：

- GitHub Issues: [https://github.com/Humphrey-He/go-generic-utils/issues](https://github.com/Humphrey-He/go-generic-utils/issues)
- 邮箱：steve1484121793@gmail.com 