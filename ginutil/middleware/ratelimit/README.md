# Gin 限流中间件 — `ratelimit`

`ratelimit` 包提供了一个用于 Gin 框架的请求限流中间件，用于限制客户端的请求频率，防止恶意请求或过度使用 API。

## 主要特性

* **灵活的存储后端**：支持内存存储和 Redis 存储，可以轻松扩展到其他存储后端。
* **多种限流策略**：支持基于 IP、路径、用户等多种限流键。
* **自定义限流规则**：可以自定义每秒请求数、突发请求数和每个请求消耗的令牌数。
* **白名单机制**：支持设置 IP 白名单，跳过限流检查。
* **自定义错误处理**：支持自定义限流错误的响应格式和内容。
* **限流信息头部**：自动添加限流相关的响应头部，方便客户端了解限流状态。

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 默认配置

最简单的使用方式是使用默认配置：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.New()
    
    // 使用默认限流配置（每秒 10 个请求，突发 20 个）
    r.Use(ratelimit.New())
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义限流参数

设置自定义限流参数：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

func main() {
    r := gin.New()
    
    // 使用自定义限流配置
    r.Use(ratelimit.NewWithConfig(
        ratelimit.WithLimit(100),           // 每秒允许 100 个请求
        ratelimit.WithBurst(50),            // 突发请求数为 50
        ratelimit.WithTokensPerRequest(1),  // 每个请求消耗 1 个令牌
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 基于 IP 的限流

限制每个 IP 的请求频率：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

func main() {
    r := gin.New()
    
    // 使用基于 IP 的限流
    r.Use(ratelimit.RateLimitPerIP(100, 50))  // 每个 IP 每秒最多 100 个请求，突发 50 个
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 基于路径的限流

限制每个路径的请求频率：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

func main() {
    r := gin.New()
    
    // 使用基于路径的限流
    r.Use(ratelimit.RateLimitPerPath(100, 50))  // 每个路径每秒最多 100 个请求，突发 50 个
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义限流键

使用自定义函数提取限流键：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

func main() {
    r := gin.New()
    
    // 使用组合键函数
    r.Use(ratelimit.NewWithConfig(
        ratelimit.WithLimit(100),
        ratelimit.WithBurst(50),
        ratelimit.WithKeyFunc(ratelimit.CombinedKeyFunc(
            ratelimit.IPKeyFunc(),
            ratelimit.PathKeyFunc(),
        )),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 使用白名单

设置 IP 白名单，跳过限流检查：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
)

func main() {
    r := gin.New()
    
    // 使用白名单
    r.Use(ratelimit.NewWithConfig(
        ratelimit.WithLimit(100),
        ratelimit.WithBurst(50),
        ratelimit.WithSkipper(ratelimit.WhitelistSkipper(
            "127.0.0.1",
            "192.168.1.100",
        )),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 自定义错误处理

自定义限流错误的响应：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "golang.org/x/time/rate"
    "net/http"
    "time"
)

func main() {
    r := gin.New()
    
    // 使用自定义错误处理函数
    r.Use(ratelimit.NewWithConfig(
        ratelimit.WithLimit(100),
        ratelimit.WithBurst(50),
        ratelimit.WithErrorHandler(func(c *gin.Context, retryAfter time.Duration) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "请求频率超限",
                "retry_after": retryAfter.Seconds(),
                "code": 429,
            })
        }),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

### 使用 Redis 存储

使用 Redis 作为限流状态的存储后端：

```go
import (
    "ggu/ginutil/middleware/ratelimit"
    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
    "golang.org/x/time/rate"
)

func main() {
    r := gin.New()
    
    // 创建 Redis 客户端
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建 Redis 存储
    store := ratelimit.NewRedisStore(redisClient,
        ratelimit.WithRedisKeyPrefix("myapp:ratelimit:"),
        ratelimit.WithRedisDefaultTTL(time.Hour),
    )
    
    // 使用 Redis 存储
    r.Use(ratelimit.NewWithConfig(
        ratelimit.WithLimit(100),
        ratelimit.WithBurst(50),
        ratelimit.WithStore(store),
    ))
    
    // 添加路由...
    
    r.Run(":8080")
}
```

## 配置选项

### WithStore

设置限流状态的存储。

```go
ratelimit.WithStore(store)
```

### WithKeyFunc

设置从 gin.Context 中提取限流键的函数。

```go
ratelimit.WithKeyFunc(keyFunc)
```

### WithErrorHandler

设置处理限流错误的函数。

```go
ratelimit.WithErrorHandler(func(c *gin.Context, retryAfter time.Duration) {
    // 自定义错误处理逻辑
})
```

### WithSkipper

设置判断是否跳过限流的函数。

```go
ratelimit.WithSkipper(skipperFunc)
```

### WithLimit

设置每秒允许的请求数。

```go
ratelimit.WithLimit(100)
```

### WithBurst

设置允许的突发请求数。

```go
ratelimit.WithBurst(50)
```

### WithTokensPerRequest

设置每个请求消耗的令牌数。

```go
ratelimit.WithTokensPerRequest(2)
```

### WithDisableHeaders

设置是否禁用限流相关的响应头。

```go
ratelimit.WithDisableHeaders(true)
```

## 限流键函数

### IPKeyFunc

使用客户端 IP 作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.IPKeyFunc())
```

### PathKeyFunc

使用请求路径作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.PathKeyFunc())
```

### MethodKeyFunc

使用请求方法作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.MethodKeyFunc())
```

### CombinedKeyFunc

组合多个键函数。

```go
ratelimit.WithKeyFunc(ratelimit.CombinedKeyFunc(
    ratelimit.IPKeyFunc(),
    ratelimit.PathKeyFunc(),
))
```

### HeaderKeyFunc

使用请求头作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.HeaderKeyFunc("X-API-Key"))
```

### QueryKeyFunc

使用查询参数作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.QueryKeyFunc("user_id"))
```

### ParamKeyFunc

使用路径参数作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.ParamKeyFunc("id"))
```

### UserKeyFunc

使用用户标识作为限流键。

```go
ratelimit.WithKeyFunc(ratelimit.UserKeyFunc())
```

## 跳过函数

### WhitelistSkipper

跳过白名单 IP。

```go
ratelimit.WithSkipper(ratelimit.WhitelistSkipper("127.0.0.1", "192.168.1.100"))
```

### MethodSkipper

跳过指定 HTTP 方法。

```go
ratelimit.WithSkipper(ratelimit.MethodSkipper("OPTIONS", "HEAD"))
```

### PathSkipper

跳过指定路径。

```go
ratelimit.WithSkipper(ratelimit.PathSkipper("/health", "/metrics"))
```

### CombinedSkipper

组合多个跳过函数。

```go
ratelimit.WithSkipper(ratelimit.CombinedSkipper(
    ratelimit.WhitelistSkipper("127.0.0.1"),
    ratelimit.PathSkipper("/health"),
))
```

## 辅助函数

### RateLimit

使用指定的限流参数创建一个限流中间件。

```go
r.GET("/api/data", ratelimit.RateLimit(100, 50), handler)
```

### RateLimitPerIP

使用指定的限流参数创建一个基于 IP 的限流中间件。

```go
r.GET("/api/data", ratelimit.RateLimitPerIP(100, 50), handler)
```

### RateLimitPerPath

使用指定的限流参数创建一个基于路径的限流中间件。

```go
r.GET("/api/data", ratelimit.RateLimitPerPath(100, 50), handler)
```

### RateLimitPerUser

使用指定的限流参数创建一个基于用户的限流中间件。

```go
r.GET("/api/data", ratelimit.RateLimitPerUser(100, 50), handler)
```

## 最佳实践

1. **根据资源敏感度设置限流**：对于敏感或资源消耗大的 API，设置更严格的限流。
2. **使用 Redis 存储实现分布式限流**：在多实例部署时，使用 Redis 存储确保限流在所有实例间共享。
3. **为不同类型的客户端设置不同的限流规则**：例如，为登录用户和匿名用户设置不同的限流规则。
4. **监控限流情况**：记录被限流的请求，以便分析和优化限流规则。
5. **在响应中提供限流信息**：通过响应头部告知客户端当前的限流状态和剩余配额。 