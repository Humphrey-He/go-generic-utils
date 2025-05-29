# Gin 响应工具库 — `response`

`response` 包为 Gin 应用提供了一套标准化的 HTTP 响应工具，使 API 响应格式统一、简洁且易于扩展。

## 主要特性

* **统一的响应结构**：使用泛型 `StandardResponse[T]` 支持各种业务数据类型
* **成功响应辅助函数**：`OK`、`Created`、`NoContent` 等
* **错误响应辅助函数**：`Fail`、`BadRequest`、`NotFound` 等
* **分页数据支持**：内置 `PaginatedData` 结构和响应函数
* **业务错误码系统**：与 `ecode` 包集成，实现细粒度的错误分类
* **链路追踪支持**：自动包含 `trace_id` 字段

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 成功响应

```go
import (
    "ggu/ginutil/response"
    "github.com/gin-gonic/gin"
)

// 返回单个资源
func GetUser(c *gin.Context) {
    user := User{ID: 1, Name: "张三"}
    response.OK(c, user)
}

// 自定义成功消息
func UpdateUser(c *gin.Context) {
    user := User{ID: 1, Name: "张三"}
    response.OKWithMessage(c, user, "用户更新成功")
}

// 资源创建成功
func CreateUser(c *gin.Context) {
    user := User{ID: 1, Name: "张三"}
    response.Created(c, user)
}

// 无内容响应（如删除操作）
func DeleteUser(c *gin.Context) {
    // 删除用户...
    response.NoContent(c)
}
```

### 错误响应

```go
// 请求参数错误
func CreateUser(c *gin.Context) {
    var req UserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "请求参数错误", err.Error())
        return
    }
    // ...
}

// 资源不存在
func GetUser(c *gin.Context) {
    user, found := findUser(id)
    if !found {
        response.NotFound(c, "用户不存在")
        return
    }
    // ...
}

// 未授权
func SecureEndpoint(c *gin.Context) {
    if !isAuthenticated(c) {
        response.Unauthorized(c, "请先登录")
        return
    }
    // ...
}

// 自定义业务错误
func ProcessOrder(c *gin.Context) {
    // ...
    if insufficientStock {
        response.Fail(c, 40050, "商品库存不足")
        return
    }
    // ...
}
```

### 分页数据响应

```go
func ListUsers(c *gin.Context) {
    // 从请求中获取分页参数
    pageInfo := response.GetPageInfo(c)
    
    // 查询数据库
    users, totalCount := getUsersFromDB(pageInfo.GetOffset(), pageInfo.GetLimit())
    
    // 返回分页响应
    response.RespondPaginated(c, users, totalCount, pageInfo.PageNum, pageInfo.PageSize)
}
```

## 响应结构

### 成功响应示例

```json
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "id": 1,
        "name": "张三",
        "email": "zhangsan@example.com"
    },
    "trace_id": "1234567890abcdef",
    "server_time": 1620000000000
}
```

### 错误响应示例

```json
{
    "code": 40001,
    "message": "请求参数错误",
    "data": {
        "fields": [
            {
                "field": "email",
                "message": "邮箱格式不正确"
            }
        ]
    },
    "trace_id": "1234567890abcdef",
    "server_time": 1620000000000
}
```

### 分页响应示例

```json
{
    "code": 0,
    "message": "操作成功",
    "data": {
        "list": [
            {"id": 1, "name": "张三"},
            {"id": 2, "name": "李四"}
        ],
        "total": 100,
        "pageNum": 1,
        "pageSize": 10,
        "totalPages": 10
    },
    "trace_id": "1234567890abcdef",
    "server_time": 1620000000000
}
```

## 与错误码系统集成

本包与 `ginutil/ecode` 包紧密集成，使用预定义的业务错误码。您可以通过 `ecode.RegisterMessage` 注册自定义错误码消息。

```go
// 注册自定义错误码消息
func init() {
    ecode.RegisterMessage(40050, "商品库存不足")
}
```

## 链路追踪

响应中的 `trace_id` 字段默认从 Gin 上下文中的 `X-Trace-ID` 键获取，通常由中间件设置。您可以实现自己的中间件来设置这个值：

```go
func TraceMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := generateTraceID() // 生成或从请求头获取
        c.Set(response.GinTraceIDKey, traceID)
        c.Next()
    }
}
```

## 扩展

您可以通过创建自定义的响应函数来扩展这个包：

```go
// 自定义业务响应
func RespondWithAuditLog[T any](c *gin.Context, data T, logInfo string) {
    // 记录审计日志
    auditLogger.Log(c, logInfo)
    
    // 发送标准响应
    response.OK(c, data)
}
``` 