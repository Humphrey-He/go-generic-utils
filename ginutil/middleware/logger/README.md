# Gin 日志中间件 — `logger`

`logger` 包提供了一个用于 Gin 框架的可配置日志中间件，用于记录 HTTP 请求的详细信息，包括时间戳、状态码、延迟时间、客户端 IP、HTTP 方法、请求路径、User-Agent 等。

## 主要特性

* **可配置输出**：支持将日志输出到标准输出、文件或任何实现了 `io.Writer` 接口的对象。
* **可配置格式化器**：提供 JSON 和文本两种格式化器，也可以实现自定义格式化器。
* **路径过滤**：支持通过路径列表和正则表达式跳过特定路径的日志记录。
* **上下文信息**：可以记录 Gin Context 中的自定义键值对。
* **请求/响应头**：可以选择性地记录请求头和响应头。
* **自定义日志函数**：支持完全自定义日志处理逻辑。
* **并发安全**：使用互斥锁确保日志写入的并发安全。

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 默认配置

最简单的使用方式是使用默认配置：

```go
import (
    "ggu/ginutil/middleware/logger"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用默认日志配置
    r.Use(logger.New())
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义配置

使用函数选项模式自定义日志配置：

```go
import (
    "ggu/ginutil/middleware/logger"
    "github.com/gin-gonic/gin"
    "os"
    "regexp"
    "time"
)

func main() {
    r := gin.New()
    
    // 创建日志文件
    logFile, _ := os.Create("gin.log")
    
    // 使用自定义日志配置
    r.Use(logger.NewWithConfig(
        logger.WithOutput(logFile),
        logger.WithFormatter(&logger.JSONFormatter{}),
        logger.WithSkipPaths("/health", "/metrics"),
        logger.WithSkipPathRegexps(regexp.MustCompile(`^/static/.*`)),
        logger.WithTimeFormat(time.RFC3339),
        logger.WithUTC(true),
        logger.WithContextKeys("user_id", "trace_id"),
        logger.WithRequestHeader("Content-Type", "Authorization"),
        logger.WithResponseHeader("Content-Type", "X-Request-ID"),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义日志函数

使用自定义日志函数将日志发送到第三方服务：

```go
import (
    "ggu/ginutil/middleware/logger"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用自定义日志函数
    r.Use(logger.NewWithConfig(
        logger.WithLogFunc(func(entry *logger.LogEntry) {
            // 将日志发送到 Elasticsearch、Logstash、Fluentd 等
            // 例如：
            // client.Index().
            //     Index("gin-logs").
            //     Type("log").
            //     BodyJson(entry).
            //     Do(context.Background())
        }),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

## 配置选项

### WithOutput

设置日志输出的目标。

```go
logger.WithOutput(os.Stdout)
logger.WithOutput(logFile)
```

### WithFormatter

设置日志格式化器。

```go
// JSON 格式化器
logger.WithFormatter(&logger.JSONFormatter{
    PrettyPrint: true, // 美化输出
})

// 文本格式化器
logger.WithFormatter(&logger.TextFormatter{
    DisableColors: false, // 启用颜色
    TimeFormat: "2006-01-02 15:04:05", // 自定义时间格式
})
```

### WithSkipPaths

设置不需要记录日志的路径列表。

```go
logger.WithSkipPaths("/health", "/metrics", "/favicon.ico")
```

### WithSkipPathRegexps

设置不需要记录日志的路径正则表达式列表。

```go
logger.WithSkipPathRegexps(
    regexp.MustCompile(`^/static/.*`),
    regexp.MustCompile(`^/assets/.*`),
)
```

### WithTimeFormat

设置时间戳的格式。

```go
logger.WithTimeFormat(time.RFC3339)
logger.WithTimeFormat("2006-01-02 15:04:05")
```

### WithUTC

设置是否使用 UTC 时间。

```go
logger.WithUTC(true)
```

### WithContextKeys

设置需要从 Gin Context 中提取并记录的键列表。

```go
logger.WithContextKeys("user_id", "trace_id", "session_id")
```

### WithRequestHeader

设置需要记录的请求头列表。

```go
logger.WithRequestHeader("Content-Type", "Accept", "User-Agent")
```

### WithResponseHeader

设置需要记录的响应头列表。

```go
logger.WithResponseHeader("Content-Type", "X-Request-ID")
```

### WithDisableRequestLog

设置是否禁用请求日志。

```go
logger.WithDisableRequestLog(true)
```

### WithDisableResponseLog

设置是否禁用响应日志。

```go
logger.WithDisableResponseLog(true)
```

### WithLogFunc

设置自定义日志函数。

```go
logger.WithLogFunc(func(entry *logger.LogEntry) {
    // 自定义日志处理逻辑
})
```

## 自定义格式化器

实现 `LogFormatter` 接口来创建自定义格式化器：

```go
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logger.LogEntry) []byte {
    // 自定义格式化逻辑
    var output bytes.Buffer
    fmt.Fprintf(&output, "%s - %s %s %d %s %.2fms\n",
        entry.Timestamp.Format(time.RFC3339),
        entry.Method,
        entry.Path,
        entry.StatusCode,
        entry.ClientIP,
        entry.Latency,
    )
    return output.Bytes()
}

// 使用自定义格式化器
r.Use(logger.NewWithConfig(
    logger.WithFormatter(&CustomFormatter{}),
))
```

## 日志条目字段

`LogEntry` 结构体包含以下字段：

| 字段 | 类型 | 描述 |
|------|------|------|
| Timestamp | time.Time | 请求处理完成的时间戳 |
| StatusCode | int | HTTP 状态码 |
| Latency | float64 | 请求处理耗时（毫秒） |
| ClientIP | string | 客户端 IP 地址 |
| Method | string | HTTP 方法 |
| Path | string | 请求路径 |
| RawQuery | string | 原始查询参数 |
| UserAgent | string | 用户代理 |
| ErrorMessage | string | 错误信息（如果有） |
| RequestSize | int64 | 请求体大小 |
| ResponseSize | int | 响应体大小 |
| RequestID | string | 请求 ID |
| Extra | map[string]interface{} | 额外信息 |

## 最佳实践

1. **在生产环境中使用 JSON 格式**：JSON 格式更容易被日志聚合系统（如 ELK）处理。
2. **跳过健康检查和静态资源路径**：这些路径通常不需要记录日志，可以减少日志量。
3. **记录请求 ID**：为每个请求生成一个唯一的请求 ID，并在日志中记录，便于跟踪请求。
4. **脱敏敏感信息**：确保不记录敏感信息，如密码、令牌等。
5. **使用自定义日志函数发送到集中式日志系统**：在大型应用中，将日志发送到集中式日志系统更有利于管理和分析。 