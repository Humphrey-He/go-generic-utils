# Gin 超时中间件 — `timeout`

`timeout` 包提供了一个用于 Gin 框架的请求超时中间件，用于为 HTTP 请求设置最大处理时间，超时后会中断请求处理并返回超时响应。

## 主要特性

* **请求超时控制**：为 HTTP 请求设置最大处理时间，防止请求处理时间过长。
* **自定义超时响应**：支持自定义超时响应的状态码、消息和 JSON 体。
* **自定义超时处理函数**：支持完全自定义超时处理逻辑。
* **Panic 传递**：确保在处理请求的 goroutine 中发生的 panic 能够被正确传递和捕获。
* **响应写入原子性**：通过检查响应是否已经写入，避免重复写入响应。

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 默认配置

最简单的使用方式是使用默认配置：

```go
import (
    "ggu/ginutil/middleware/timeout"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用默认超时配置（30 秒）
    r.Use(timeout.New())
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义超时时间

设置自定义超时时间：

```go
import (
    "ggu/ginutil/middleware/timeout"
    "github.com/gin-gonic/gin"
    "time"
)

func main() {
    r := gin.New()
    
    // 设置超时时间为 5 秒
    r.Use(timeout.NewWithConfig(
        timeout.WithTimeout(5 * time.Second),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义超时处理函数

使用自定义超时处理函数：

```go
import (
    "ggu/ginutil/middleware/timeout"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

func main() {
    r := gin.New()
    
    // 使用自定义超时处理函数
    r.Use(timeout.NewWithConfig(
        timeout.WithTimeout(5 * time.Second),
        timeout.WithTimeoutHandler(func(c *gin.Context) {
            c.JSON(http.StatusServiceUnavailable, gin.H{
                "error": "服务暂时不可用，请稍后重试",
                "code": 503,
            })
        }),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义超时消息

设置自定义超时消息：

```go
import (
    "ggu/ginutil/middleware/timeout"
    "github.com/gin-gonic/gin"
    "time"
)

func main() {
    r := gin.New()
    
    // 设置自定义超时消息
    r.Use(timeout.NewWithConfig(
        timeout.WithTimeout(5 * time.Second),
        timeout.WithTimeoutMessage("请求处理时间过长，请稍后重试"),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义超时 JSON 响应

设置自定义超时 JSON 响应：

```go
import (
    "ggu/ginutil/middleware/timeout"
    "github.com/gin-gonic/gin"
    "time"
)

func main() {
    r := gin.New()
    
    // 设置自定义超时 JSON 响应
    r.Use(timeout.NewWithConfig(
        timeout.WithTimeout(5 * time.Second),
        timeout.WithTimeoutJSON(gin.H{
            "error": "请求超时",
            "code": 504,
            "details": "服务器处理请求时间过长",
        }),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 禁用超时功能

在特定情况下禁用超时功能：

```go
import (
    "ggu/ginutil/middleware/timeout"
    "github.com/gin-gonic/gin"
    "time"
)

func main() {
    r := gin.New()
    
    // 禁用超时功能
    r.Use(timeout.NewWithConfig(
        timeout.WithTimeout(5 * time.Second),
        timeout.WithDisableTimeout(true),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

## 配置选项

### WithTimeout

设置请求处理的最大时间。

```go
timeout.WithTimeout(5 * time.Second)
```

### WithTimeoutHandler

设置自定义超时处理函数。

```go
timeout.WithTimeoutHandler(func(c *gin.Context) {
    // 自定义超时处理逻辑
})
```

### WithTimeoutMessage

设置超时响应的错误消息。

```go
timeout.WithTimeoutMessage("请求处理超时")
```

### WithTimeoutCode

设置超时响应的 HTTP 状态码。

```go
timeout.WithTimeoutCode(http.StatusServiceUnavailable)
```

### WithTimeoutJSON

设置超时响应的 JSON 体。

```go
timeout.WithTimeoutJSON(gin.H{
    "error": "请求超时",
    "code": 504,
})
```

### WithDisableTimeout

设置是否禁用超时功能。

```go
timeout.WithDisableTimeout(true)
```

## 辅助函数

### TimeoutWithHandler

使用指定的超时时间和处理函数创建一个超时中间件。

```go
r.Use(timeout.TimeoutWithHandler(5 * time.Second, func(c *gin.Context) {
    c.String(http.StatusServiceUnavailable, "请求处理超时")
}))
```

### TimeoutWithMessage

使用指定的超时时间和错误消息创建一个超时中间件。

```go
r.Use(timeout.TimeoutWithMessage(5 * time.Second, "请求处理超时"))
```

### TimeoutWithCode

使用指定的超时时间和 HTTP 状态码创建一个超时中间件。

```go
r.Use(timeout.TimeoutWithCode(5 * time.Second, http.StatusServiceUnavailable))
```

### TimeoutWithJSON

使用指定的超时时间和 JSON 体创建一个超时中间件。

```go
r.Use(timeout.TimeoutWithJSON(5 * time.Second, gin.H{
    "error": "请求超时",
    "code": 504,
}))
```

## 最佳实践

1. **设置合理的超时时间**：根据请求的复杂度和预期处理时间设置合理的超时时间。
2. **为不同的路由设置不同的超时时间**：对于不同的路由，可以使用不同的超时时间。
3. **提供友好的超时响应**：为用户提供清晰的超时错误信息，帮助他们理解发生了什么。
4. **监控超时情况**：记录超时事件，以便分析和优化系统性能。
5. **处理长时间运行的任务**：对于预期会长时间运行的任务，考虑使用异步处理方式，而不是依赖超时中间件。 