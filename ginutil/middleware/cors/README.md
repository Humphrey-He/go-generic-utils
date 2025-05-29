# Gin CORS 中间件 — `cors`

`cors` 包提供了一个用于 Gin 框架的 CORS (跨域资源共享) 中间件，允许开发者轻松配置 CORS 策略，包括允许的来源、方法、头部等，并自动处理预检请求 (OPTIONS)。

## 主要特性

* **灵活的配置选项**：支持所有标准 CORS 配置项
* **函数选项模式**：使用 `WithXxx()` 函数轻松配置中间件
* **支持通配符**：允许使用 `*` 匹配所有来源或头部
* **动态来源验证**：支持通过函数动态判断是否允许特定来源
* **预检请求处理**：自动处理 OPTIONS 请求
* **安全性检查**：自动检测并警告潜在的安全问题
* **调试模式**：可选的详细日志输出

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 默认配置

最简单的使用方式是使用默认配置，它允许所有来源的 GET、POST、PUT、PATCH、DELETE、HEAD 和 OPTIONS 请求：

```go
import (
    "ggu/ginutil/middleware/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 使用默认 CORS 配置
    r.Use(cors.New())
    
    r.GET("/api/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello, world!"})
    })
    
    r.Run(":8080")
}
```

### 自定义配置

使用函数选项模式自定义 CORS 配置：

```go
import (
    "ggu/ginutil/middleware/cors"
    "github.com/gin-gonic/gin"
    "time"
)

func main() {
    r := gin.Default()
    
    // 使用自定义 CORS 配置
    r.Use(cors.NewWithConfig(
        cors.WithAllowedOrigins([]string{"https://example.com", "https://api.example.com"}),
        cors.WithAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
        cors.WithAllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"}),
        cors.WithExposeHeaders([]string{"Content-Length", "X-Request-ID"}),
        cors.WithAllowCredentials(true),
        cors.WithMaxAge(12 * time.Hour),
    ))
    
    // 路由定义...
    
    r.Run(":8080")
}
```

### 动态来源验证

使用 `AllowOriginFunc` 动态判断是否允许特定来源：

```go
import (
    "ggu/ginutil/middleware/cors"
    "github.com/gin-gonic/gin"
    "strings"
)

func main() {
    r := gin.Default()
    
    // 使用动态来源验证
    r.Use(cors.NewWithConfig(
        cors.WithAllowOriginFunc(func(origin string) bool {
            // 允许所有以 example.com 结尾的来源
            return strings.HasSuffix(origin, "example.com")
        }),
        cors.WithAllowedMethods([]string{"GET", "POST"}),
        cors.WithAllowCredentials(true),
    ))
    
    // 路由定义...
    
    r.Run(":8080")
}
```

### 开发环境配置

在开发环境中，可以使用 `AllowAll()` 函数允许所有跨域请求：

```go
import (
    "ggu/ginutil/middleware/cors"
    "github.com/gin-gonic/gin"
    "os"
)

func main() {
    r := gin.Default()
    
    // 根据环境选择 CORS 配置
    if os.Getenv("ENV") == "development" {
        // 开发环境：允许所有跨域请求
        r.Use(cors.AllowAll())
    } else {
        // 生产环境：使用严格的 CORS 配置
        r.Use(cors.NewWithConfig(
            cors.WithAllowedOrigins([]string{"https://example.com"}),
            cors.WithAllowedMethods([]string{"GET", "POST"}),
            cors.WithAllowedHeaders([]string{"Content-Type", "Authorization"}),
        ))
    }
    
    // 路由定义...
    
    r.Run(":8080")
}
```

## 配置选项

### AllowedOrigins

设置允许的来源列表。可以使用 `*` 表示允许所有来源。

```go
cors.WithAllowedOrigins([]string{"https://example.com", "https://api.example.com"})
```

### AllowOriginFunc

设置动态判断来源是否允许的函数。如果设置了此函数，它将优先于 `AllowedOrigins` 使用。

```go
cors.WithAllowOriginFunc(func(origin string) bool {
    return strings.HasSuffix(origin, ".example.com")
})
```

### AllowedMethods

设置允许的 HTTP 方法列表。

```go
cors.WithAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
```

### AllowedHeaders

设置允许的 HTTP 头部列表。可以使用 `*` 表示允许客户端请求的所有头部。

```go
cors.WithAllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-With"})
```

### ExposeHeaders

设置允许客户端读取的响应头部列表。

```go
cors.WithExposeHeaders([]string{"Content-Length", "X-Request-ID"})
```

### AllowCredentials

设置是否允许包含凭证的请求。注意：如果为 `true`，则 `AllowedOrigins` 不能包含 `*`，必须指定明确的来源。

```go
cors.WithAllowCredentials(true)
```

### MaxAge

设置预检请求结果的缓存时间。

```go
cors.WithMaxAge(12 * time.Hour)
```

### OptionsPassthrough

设置是否将 OPTIONS 请求传递给下一个处理器。

```go
cors.WithOptionsPassthrough(true)
```

### Debug

设置是否启用调试模式，输出详细日志。

```go
cors.WithDebug(true)
```

## 安全性注意事项

1. **不要在生产环境中使用 `AllowAll()`**：这会允许任何网站向您的 API 发送请求。
2. **谨慎使用 `AllowCredentials`**：当设置 `AllowCredentials(true)` 时，不能将 `AllowedOrigins` 设置为 `["*"]`，必须指定明确的来源。
3. **限制允许的方法和头部**：只允许应用程序实际需要的 HTTP 方法和头部。
4. **使用 HTTPS**：在生产环境中，应该使用 HTTPS 保护您的 API。

## 常见问题

### 预检请求失败

如果预检请求 (OPTIONS) 失败，请检查以下几点：

1. `AllowedOrigins` 是否包含请求的来源
2. `AllowedMethods` 是否包含请求使用的 HTTP 方法
3. `AllowedHeaders` 是否包含请求使用的所有头部

### 凭证问题

如果使用 `withCredentials: true` 发送请求但收到错误，请确保：

1. 服务器配置了 `AllowCredentials(true)`
2. `AllowedOrigins` 不包含 `*`，而是指定了明确的来源

### 无法访问响应头部

如果客户端无法访问某些响应头部，请确保将这些头部添加到 `ExposeHeaders` 列表中：

```go
cors.WithExposeHeaders([]string{"Content-Length", "X-Request-ID"})
``` 