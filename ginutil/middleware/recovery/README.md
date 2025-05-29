# Gin 错误恢复中间件 — `recovery`

`recovery` 包提供了一个用于 Gin 框架的错误恢复中间件，用于捕获处理 HTTP 请求过程中发生的 panic，记录详细的堆栈信息，并返回友好的错误响应。

## 主要特性

* **错误恢复**：捕获请求处理过程中的 panic，防止服务器崩溃。
* **堆栈跟踪**：记录详细的堆栈信息，便于调试。
* **自定义错误处理**：支持自定义错误处理函数，可以返回不同格式的错误响应。
* **优雅处理连接断开**：特殊处理 "broken pipe" 和 "connection reset by peer" 错误，避免不必要的日志。
* **敏感信息过滤**：对请求头中的敏感信息（如 Authorization、Cookie、Token）进行脱敏处理。
* **可配置输出**：支持将错误日志输出到标准错误、文件或任何实现了 `io.Writer` 接口的对象。

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 默认配置

最简单的使用方式是使用默认配置：

```go
import (
    "ggu/ginutil/middleware/recovery"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用默认恢复配置
    r.Use(recovery.New())
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义错误处理函数

使用自定义错误处理函数：

```go
import (
    "ggu/ginutil/middleware/recovery"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用自定义错误处理函数
    r.Use(recovery.NewWithConfig(
        recovery.WithErrorHandler(func(c *gin.Context, err interface{}) {
            // 自定义错误响应
            c.JSON(500, gin.H{
                "error": err,
                "message": "服务器内部错误",
                "request_id": c.GetHeader("X-Request-ID"),
            })
        }),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义堆栈跟踪配置

自定义堆栈跟踪配置：

```go
import (
    "ggu/ginutil/middleware/recovery"
    "github.com/gin-gonic/gin"
    "os"
)

func main() {
    r := gin.New()
    
    // 创建日志文件
    logFile, _ := os.Create("panic.log")
    
    // 自定义堆栈跟踪配置
    r.Use(recovery.NewWithConfig(
        recovery.WithOutput(logFile),
        recovery.WithStackAll(true),
        recovery.WithStackSize(8 << 10), // 8 KB
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 使用预定义的错误处理函数

使用预定义的 JSON 错误处理函数：

```go
import (
    "ggu/ginutil/middleware/recovery"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用 JSON 错误处理函数
    r.Use(recovery.NewWithConfig(
        recovery.WithErrorHandler(recovery.JSONErrorHandler(500, "服务器内部错误")),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 集成 Sentry 错误跟踪

集成 Sentry 错误跟踪（示例）：

```go
import (
    "ggu/ginutil/middleware/recovery"
    "github.com/gin-gonic/gin"
    "github.com/getsentry/sentry-go"
)

func main() {
    // 初始化 Sentry
    sentry.Init(sentry.ClientOptions{
        Dsn: "your-sentry-dsn",
    })
    
    r := gin.New()
    
    // 使用 Sentry 错误处理函数
    r.Use(recovery.NewWithConfig(
        recovery.WithErrorHandler(func(c *gin.Context, err interface{}) {
            // 发送错误到 Sentry
            sentry.CaptureException(fmt.Errorf("%v", err))
            
            // 返回错误响应
            c.JSON(500, gin.H{
                "message": "服务器内部错误",
            })
        }),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

## 配置选项

### WithErrorHandler

设置自定义错误处理函数。

```go
recovery.WithErrorHandler(func(c *gin.Context, err interface{}) {
    // 自定义错误处理逻辑
})
```

### WithStackAll

设置是否记录完整的堆栈信息。

```go
recovery.WithStackAll(true)
```

### WithStackSize

设置堆栈缓冲区的大小。

```go
recovery.WithStackSize(8 << 10) // 8 KB
```

### WithOutput

设置日志输出的目标。

```go
recovery.WithOutput(os.Stderr)
recovery.WithOutput(logFile)
```

### WithDisableStackAll

设置是否禁用完整的堆栈跟踪。

```go
recovery.WithDisableStackAll(true)
```

### WithDisablePrintStack

设置是否禁用打印堆栈信息。

```go
recovery.WithDisablePrintStack(true)
```

### WithDisableRecovery

设置是否禁用恢复功能。

```go
recovery.WithDisableRecovery(true)
```

## 预定义的错误处理函数

### JSONErrorHandler

创建一个返回 JSON 格式错误的处理函数。

```go
recovery.JSONErrorHandler(500, "服务器内部错误")
```

### ErrorHandlerWithLogger

创建一个将错误记录到日志的处理函数。

```go
logger := log.New(os.Stderr, "", log.LstdFlags)
recovery.ErrorHandlerWithLogger(defaultHandler, logger)
```

### ErrorHandlerWithSentry

创建一个将错误发送到 Sentry 的处理函数（示例）。

```go
recovery.ErrorHandlerWithSentry(defaultHandler)
```

## 辅助函数

### IsBrokenPipeError

检查错误是否是连接断开错误。

```go
if recovery.IsBrokenPipeError(err) {
    // 处理连接断开错误
}
```

### RecoveryWithWriter

使用指定的输出创建一个恢复中间件。

```go
r.Use(recovery.RecoveryWithWriter(os.Stderr))
```

### RecoveryWithCustomErrorHandler

使用自定义错误处理函数创建一个恢复中间件。

```go
r.Use(recovery.RecoveryWithCustomErrorHandler(myErrorHandler))
```

## 最佳实践

1. **总是使用恢复中间件**：在生产环境中，恢复中间件是必不可少的，可以防止服务器因为未处理的 panic 而崩溃。
2. **记录详细的错误信息**：在开发环境中，启用完整的堆栈跟踪，便于调试。
3. **自定义错误响应**：根据应用的需求，自定义错误响应格式，提供友好的错误信息。
4. **集成错误跟踪系统**：在生产环境中，将错误发送到 Sentry、Bugsnag 等错误跟踪系统，便于监控和分析。
5. **脱敏敏感信息**：确保不记录敏感信息，如密码、令牌等。 