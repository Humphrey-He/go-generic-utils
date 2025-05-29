# Gin 上下文工具包 — `contextx`

`contextx` 包提供了一系列对 `gin.Context` 进行操作的工具函数，使得在 Gin 框架中处理上下文数据更加类型安全和便捷。

## 主要特性

* **类型安全的上下文值存取**：使用 Go 1.18+ 泛型特性，提供类型安全的方式存取 `gin.Context` 中的数据。
* **用户身份信息管理**：封装对用户身份信息的存取和权限判断操作。
* **请求元数据处理**：简化对请求跟踪ID、客户端IP、请求时间等信息的获取。
* **分页参数处理**：提供标准化的分页参数解析和验证功能。

## 安装

该包是 `ggu` 项目的一部分，无需单独安装。

## 核心功能

### 类型安全的上下文值存取 (context.go)

使用泛型提供类型安全的方式存取 `gin.Context` 中的数据，避免手动类型断言。

```go
// 设置值
contextx.Set(c, "user.id", 12345)

// 获取值（带类型安全）
userID, exists := contextx.Get[int](c, "user.id")
if exists {
    fmt.Println("用户ID:", userID) // userID 的类型为 int
}

// 必须获取值（如果键不存在或类型不匹配将 panic）
userID := contextx.MustGet[int](c, "user.id")
```

### 用户身份信息管理 (user.go)

封装用户身份信息的存取和权限判断逻辑，与 `auth` 包紧密集成。

```go
// 设置用户身份信息
identity := &auth.UserIdentity[int, string]{
    UserID:   12345,
    Username: "test_user",
    Roles:    []string{"admin", "editor"},
    TenantID: "tenant1",
}
contextx.SetUserIdentity(c, identity)

// 获取用户身份信息
userIdentity, exists := contextx.GetUserIdentity[int, string](c)
if exists {
    fmt.Println("用户名:", userIdentity.Username)
}

// 获取用户ID
userID, exists := contextx.GetUserID[int](c)
if exists {
    fmt.Println("用户ID:", userID)
}

// 获取用户角色
roles, exists := contextx.GetUserRoles[int, string](c)
if exists {
    fmt.Println("用户角色:", strings.Join(roles, ", "))
}

// 检查用户角色
if contextx.HasRole[int, string](c, "admin") {
    // 用户拥有 admin 角色的处理逻辑
}

// 检查用户是否拥有任一角色
if contextx.HasAnyRole[int, string](c, "admin", "editor") {
    // 用户拥有 admin 或 editor 角色的处理逻辑
}

// 检查用户是否拥有所有角色
if contextx.HasAllRoles[int, string](c, "admin", "editor") {
    // 用户同时拥有 admin 和 editor 角色的处理逻辑
}
```

### 请求元数据处理 (request.go)

提供获取和设置请求相关元数据的工具函数。

```go
// 获取或生成跟踪ID
traceID := contextx.GetTraceID(c)
fmt.Println("跟踪ID:", traceID)

// 设置请求开始时间
contextx.SetRequestStartTime(c, time.Now())

// 获取请求持续时间
duration, exists := contextx.GetRequestDuration(c)
if exists {
    fmt.Printf("请求已处理 %.2f 毫秒\n", float64(duration)/float64(time.Millisecond))
}

// 获取客户端IP
clientIP := contextx.GetClientIP(c)
fmt.Println("客户端IP:", clientIP)

// 获取真实IP（考虑代理情况）
realIP := contextx.GetRealIP(c)
fmt.Println("真实IP:", realIP)

// 获取请求其他信息
fmt.Println("User-Agent:", contextx.GetUserAgent(c))
fmt.Println("Referer:", contextx.GetReferer(c))
fmt.Println("请求方法:", contextx.GetRequestMethod(c))
fmt.Println("请求路径:", contextx.GetRequestPath(c))
fmt.Println("请求查询:", contextx.GetRequestQuery(c))
fmt.Println("请求主机:", contextx.GetRequestHost(c))
fmt.Println("请求协议:", contextx.GetRequestProtocol(c))
```

### 分页参数处理 (pagination.go)

提供标准化的分页参数解析和验证功能。

```go
// 从请求中获取分页参数
pageInfo, errs := contextx.GetPageInfo(c)
if errs != nil && errs.HasErrors() {
    // 处理参数错误
    response.FailWithValidation(c, http.StatusBadRequest, "分页参数无效", errs)
    return
}

// 使用分页参数
fmt.Printf("页码: %d, 每页条数: %d\n", pageInfo.PageNum, pageInfo.PageSize)
fmt.Printf("SQL LIMIT %d OFFSET %d\n", pageInfo.Limit(), pageInfo.Offset())

// 或者直接获取 limit 和 offset
limit, offset, errs := contextx.GetLimitOffset(c)
if errs != nil && errs.HasErrors() {
    // 处理参数错误
    response.FailWithValidation(c, http.StatusBadRequest, "分页参数无效", errs)
    return
}

// 使用 limit 和 offset
fmt.Printf("LIMIT %d OFFSET %d\n", limit, offset)
```

## 完整使用示例

### HTTP 处理函数中使用

```go
func GetUserList(c *gin.Context) {
    // 1. 获取并验证分页参数
    limit, offset, errs := contextx.GetLimitOffset(c)
    if errs != nil && errs.HasErrors() {
        response.FailWithValidation(c, http.StatusBadRequest, "分页参数无效", errs)
        return
    }
    
    // 2. 获取请求的跟踪ID
    traceID := contextx.GetTraceID(c)
    
    // 3. 获取当前用户信息
    userID, exists := contextx.GetUserID[int](c)
    if !exists {
        response.Fail(c, http.StatusUnauthorized, "未认证")
        return
    }
    
    // 4. 检查权限
    if !contextx.HasAnyRole[int, string](c, "admin", "viewer") {
        response.Fail(c, http.StatusForbidden, "权限不足")
        return
    }
    
    // 5. 执行业务逻辑（获取用户列表）
    users, total, err := service.GetUserList(c.Request.Context(), limit, offset)
    if err != nil {
        // 记录错误日志（包含跟踪ID）
        log.Printf("获取用户列表失败, traceID=%s: %v", traceID, err)
        response.Fail(c, http.StatusInternalServerError, "获取用户列表失败")
        return
    }
    
    // 6. 返回成功响应
    response.SuccessWithPagination(c, users, total, limit, offset)
}
```

### 中间件中使用

```go
// 请求日志中间件
func RequestLoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 设置请求开始时间
        startTime := time.Now()
        contextx.SetRequestStartTime(c, startTime)
        
        // 设置或获取跟踪ID
        traceID := contextx.GetTraceID(c)
        
        // 获取请求信息
        path := contextx.GetRequestPath(c)
        method := contextx.GetRequestMethod(c)
        clientIP := contextx.GetClientIP(c)
        
        // 记录请求开始日志
        log.Printf("[%s] 请求开始: %s %s, IP=%s", traceID, method, path, clientIP)
        
        // 继续处理请求
        c.Next()
        
        // 获取响应状态码
        statusCode := c.Writer.Status()
        
        // 计算请求处理时间
        duration, _ := contextx.GetRequestDuration(c)
        
        // 记录请求结束日志
        log.Printf("[%s] 请求结束: %s %s, 状态=%d, 耗时=%.2fms", 
            traceID, method, path, statusCode, float64(duration)/float64(time.Millisecond))
    }
}
```

## 最佳实践

1. **使用泛型函数获取上下文值**：始终使用 `contextx.Get[T]` 代替 `c.Get` 和手动类型断言，避免运行时类型错误。

2. **合理使用 MustGet 函数**：仅在确信键存在且类型正确时使用 `MustGet`，否则使用 `Get` 并检查第二个返回值。

3. **处理用户身份信息**：使用 `contextx` 包中的用户身份函数，而不是直接操作 `auth.UserIdentity`，这样可以在将来更改身份存储方式时减少修改。

4. **标准化分页处理**：始终使用 `GetPageInfo` 或 `GetLimitOffset` 处理分页参数，确保一致的用户体验和错误处理。

5. **请求跟踪**：在所有请求日志中包含跟踪ID，便于问题排查。

6. **类型参数推导**：尽可能让编译器推导泛型类型参数，例如：
   ```go
   // 推荐
   value, exists := contextx.Get[string](c, "key")
   
   // 不推荐（多余的类型参数）
   value, exists := contextx.Get[string](c, "key")
   ```

7. **结合 response 包使用**：与 `response` 包一起使用，处理分页参数错误：
   ```go
   pageInfo, errs := contextx.GetPageInfo(c)
   if errs != nil && errs.HasErrors() {
       response.FailWithValidation(c, http.StatusBadRequest, "分页参数无效", errs)
       return
   }
   ```

8. **错误处理一致性**：当处理类似的函数时，保持一致的错误处理模式，例如：
   ```go
   if !contextx.HasRole[int, string](c, "admin") {
       response.Fail(c, http.StatusForbidden, "需要管理员权限")
       return
   }
   ``` 