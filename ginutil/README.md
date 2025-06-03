# GinUtil - Enhanced Gin Framework Utilities

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Framework-Gin-00ADD8?style=for-the-badge" alt="Gin Framework">
  <img src="https://img.shields.io/badge/Type-Library-green?style=for-the-badge" alt="Type">
</p>

## 项目描述

GinUtil 是一个功能丰富的 Gin 框架扩展工具库，为 Gin Web 开发提供一站式增强解决方案，简化常见开发任务并提高开发效率。

### 核心特性

- **标准化响应**: 提供统一的 JSON/XML/HTML 响应格式，支持成功、错误、分页等多种场景
- **路由管理**: 支持自动注册、RESTful API 构建、路由分组和命名空间管理
- **中间件集成**: 包含日志、恢复、超时、认证等开箱即用的中间件
- **上下文增强**: 扩展 Gin 上下文功能，提供请求跟踪、用户身份管理等能力
- **参数验证**: 增强请求参数绑定和验证功能
- **安全工具**: 提供 CSRF 防护、XSS 过滤、CORS 等安全相关组件
- **API 文档**: 内置 Swagger UI 集成，轻松生成和管理 API 文档
- **泛型支持**: 充分利用 Go 1.18+ 泛型特性，提供类型安全的 API

### 适用场景

- 构建 RESTful API 服务
- 微服务架构的 Web 层
- 后台管理系统
- 需要标准化响应格式的企业级应用
- 对代码组织和开发效率有较高要求的项目

## 快速开始

### 安装

```bash
go get github.com/Humphrey-He/go-generic-utils/ginutil
```

### 基本用法

```go
package main

import (
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/logger"
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/recovery"
	"github.com/Humphrey-He/go-generic-utils/ginutil/render"
	"github.com/Humphrey-He/go-generic-utils/ginutil/register"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建 Gin 引擎
	r := gin.New()

	// 使用中间件
	r.Use(logger.New())
	r.Use(recovery.New())

	// 创建 API 路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		users := api.Group("/users")
		users.GET("", ListUsers)
		users.GET("/:id", GetUserByID)
		users.POST("", CreateUser)
	}

	// 启动服务器
	r.Run(":8080")
}

// 处理函数示例
func GetUserByID(c *gin.Context) {
	// 获取用户
	user := fetchUser(c.Param("id"))
	
	if user == nil {
		// 用户不存在，返回 404
		render.Error(c, render.CodeNotFound, "用户不存在")
		return
	}
	
	// 返回用户数据
	render.Success(c, user)
}
```

## 包结构

GinUtil 由多个子包组成，每个子包提供特定功能：

- **render**: 统一的响应渲染工具
- **register**: 路由注册和管理工具
- **middleware**: 常用中间件集合
- **contextx**: 上下文增强工具
- **security**: 安全相关组件
- **validate**: 请求验证工具
- **binding**: 请求绑定增强
- **paginator**: 分页处理工具
- **response**: 响应格式化工具
- **ecode**: 错误码管理

## 核心功能模块

### 响应渲染 (render)

提供统一的响应格式和渲染方法，支持 JSON、XML 和 HTML 等多种响应格式。

```go
// 成功响应
render.Success(c, user)

// 带消息的成功响应
render.Success(c, user, "获取用户成功")

// 错误响应
render.Error(c, render.CodeNotFound, "用户不存在")

// 分页响应
render.Paginated(c, users, total, page, pageSize)

// 自定义响应
render.Custom(c, myCode, "自定义消息", data, http.StatusOK)

// 流式 API
render.Resp(c).
    Code(200).
    Message("操作成功").
    Data(user).
    Success()
```

### 路由注册 (register)

提供多种路由注册方式，支持控制器自动注册、RESTful API 路由、命名空间等。

```go
// 使用控制器接口注册
type UserController struct{}

func (c *UserController) RegisterRoutes(group *gin.RouterGroup) {
    group.GET("", c.List)
    group.GET("/:id", c.Get)
    group.POST("", c.Create)
    group.PUT("/:id", c.Update)
    group.DELETE("/:id", c.Delete)
}

// 注册控制器
register.RegisterGlobal("users", &UserController{})
register.RegisterGlobalRoutes(router)

// 使用 RESTful 控制器
type ProductController struct{}
func (c *ProductController) Index(ctx *gin.Context) { /* 列表 */ }
func (c *ProductController) Show(ctx *gin.Context) { /* 详情 */ }
func (c *ProductController) Create(ctx *gin.Context) { /* 创建 */ }
func (c *ProductController) Update(ctx *gin.Context) { /* 更新 */ }
func (c *ProductController) Delete(ctx *gin.Context) { /* 删除 */ }

// 注册 RESTful 路由
group := router.Group("/products")
register.RegisterRESTRoutes(group, &ProductController{})

// 使用流式 API 和命名空间
api := register.NewAPIGroup(router, "/api", "v1")
api.Namespace("admin", func(ns *register.ResourceGroup) {
    ns.Resource("users", &UserController{})
    ns.Resource("products", &ProductController{})
})
```

### 中间件 (middleware)

提供常用的中间件，包括日志、恢复、超时、认证等。

```go
// 使用日志中间件
r.Use(logger.New())

// 使用自定义日志中间件
r.Use(logger.NewWithConfig(
    logger.WithOutput(os.Stdout),
    logger.WithFormatter(logger.JSONFormatter),
    logger.WithSkipPaths([]string{"/health"}),
))

// 使用恢复中间件
r.Use(recovery.New())

// 使用超时中间件
r.Use(timeout.New(timeout.WithTimeout(5 * time.Second)))

// 使用认证中间件
api := r.Group("/api")
api.Use(auth.RequireAuth[uint, string]())
```

### 上下文增强 (contextx)

提供对 Gin 上下文的增强功能，支持请求跟踪、用户身份管理等。

```go
// 设置和获取用户身份
contextx.SetUserIdentity(c, &auth.UserIdentity[uint, string]{
    UserID:   1,
    Username: "admin",
    Role:     "admin",
})

identity, ok := contextx.GetUserIdentity[uint, string](c)
if ok {
    userID := identity.UserID
    username := identity.Username
}

// 设置和获取请求跟踪 ID
contextx.SetTraceID(c, "req-123456")
traceID := contextx.GetTraceID(c)

// 设置和获取请求开始时间
contextx.SetRequestStartTime(c, time.Now())
duration, ok := contextx.GetRequestDuration(c)
```

### API 文档 (Swagger)

提供 Swagger UI 集成，支持 API 文档的生成和展示。

```go
// 注册 Swagger UI
register.AddSwaggerEndpoints(router, "./docs/swagger.json")

// 使用自定义配置
config := register.DefaultSwaggerConfig()
config.Title = "我的 API 文档"
config.Version = "2.0.0"
register.RegisterSwaggerWithConfig(router, config)
```

## 使用示例

### RESTful API 服务

```go
package main

import (
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/logger"
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/recovery"
	"github.com/Humphrey-He/go-generic-utils/ginutil/render"
	"github.com/Humphrey-He/go-generic-utils/ginutil/register"
	"github.com/gin-gonic/gin"
)

// 用户结构体
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserController 实现 RESTful 控制器
type UserController struct{}

// Index 列出所有用户
func (c *UserController) Index(ctx *gin.Context) {
	users := []User{
		{ID: 1, Username: "user1", Email: "user1@example.com"},
		{ID: 2, Username: "user2", Email: "user2@example.com"},
	}
	render.Success(ctx, users)
}

// Show 获取单个用户
func (c *UserController) Show(ctx *gin.Context) {
	id := ctx.Param("id")
	user := User{ID: 1, Username: "user1", Email: "user1@example.com"}
	render.Success(ctx, user)
}

// Create 创建用户
func (c *UserController) Create(ctx *gin.Context) {
	var user User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		render.Error(ctx, render.CodeInvalidParams, "无效的请求参数")
		return
	}
	
	// 模拟创建用户
	user.ID = 3
	
	render.Success(ctx, user, "用户创建成功")
}

// Update 更新用户
func (c *UserController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var user User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		render.Error(ctx, render.CodeInvalidParams, "无效的请求参数")
		return
	}
	
	render.Success(ctx, user, "用户更新成功")
}

// Delete 删除用户
func (c *UserController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	render.Success(ctx, nil, "用户删除成功")
}

func main() {
	r := gin.New()
	
	// 使用中间件
	r.Use(logger.New())
	r.Use(recovery.New())
	
	// 创建 API 路由组
	api := register.NewAPIGroup(r, "/api", "v1")
	
	// 注册用户控制器
	api.Resource("users", &UserController{})
	
	// 添加 Swagger UI
	register.AddSwaggerEndpoints(r, "./docs/swagger.json")
	
	// 启动服务器
	r.Run(":8080")
}
```

### 分页查询

```go
package main

import (
	"github.com/Humphrey-He/go-generic-utils/ginutil/paginator"
	"github.com/Humphrey-He/go-generic-utils/ginutil/render"
	"github.com/gin-gonic/gin"
)

// 产品结构体
type Product struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// ListProducts 分页获取产品列表
func ListProducts(c *gin.Context) {
	// 获取分页参数
	page, size := paginator.GetPagination(c)
	
	// 模拟产品列表
	products := []Product{
		{ID: 1, Name: "产品1", Price: 100},
		{ID: 2, Name: "产品2", Price: 200},
		{ID: 3, Name: "产品3", Price: 300},
	}
	
	// 总记录数
	total := int64(len(products))
	
	// 返回分页响应
	render.Paginated(c, products, total, page, size)
}

func main() {
	r := gin.Default()
	
	r.GET("/products", ListProducts)
	
	r.Run(":8080")
}
```

### 认证和授权

```go
package main

import (
	"github.com/Humphrey-He/go-generic-utils/ginutil/contextx"
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/auth"
	"github.com/Humphrey-He/go-generic-utils/ginutil/render"
	"github.com/gin-gonic/gin"
)

// UserProfile 获取用户资料
func UserProfile(c *gin.Context) {
	// 获取当前用户 ID
	userID, ok := contextx.GetUserID[uint](c)
	if !ok {
		render.Unauthorized(c, "未登录")
		return
	}
	
	// 返回用户资料
	profile := map[string]interface{}{
		"id":       userID,
		"username": "测试用户",
		"email":    "test@example.com",
	}
	
	render.Success(c, profile)
}

// AdminOnly 仅管理员可访问
func AdminOnly(c *gin.Context) {
	// 获取用户身份
	identity, ok := contextx.GetUserIdentity[uint, string](c)
	if !ok || identity == nil {
		render.Unauthorized(c, "未登录")
		return
	}
	
	// 检查角色
	if identity.Role != "admin" {
		render.Forbidden(c, "需要管理员权限")
		return
	}
	
	render.Success(c, gin.H{"message": "管理员专属内容"})
}

func main() {
	r := gin.Default()
	
	// 模拟认证中间件
	mockAuth := func(c *gin.Context) {
		// 模拟设置用户身份
		identity := &auth.UserIdentity[uint, string]{
			UserID:   1,
			Username: "test",
			Role:     "user",
		}
		contextx.SetUserIdentity(c, identity)
		c.Next()
	}
	
	// 用户路由
	user := r.Group("/user")
	user.Use(mockAuth)
	user.GET("/profile", UserProfile)
	
	// 管理员路由
	admin := r.Group("/admin")
	admin.Use(mockAuth)
	admin.GET("/dashboard", AdminOnly)
	
	r.Run(":8080")
}
```

## 常见问题 (FAQ)

### Q: 如何自定义响应格式？

可以通过 `render.Configure` 方法自定义响应格式：

```go
render.Configure(render.Config{
    JSONPrettyPrint: true,
    HTMLTemplateDir: "templates/*",
})
```

### Q: 如何处理文件上传？

可以使用 Gin 的文件处理功能，配合 GinUtil 的响应工具：

```go
func UploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        render.Error(c, render.CodeInvalidParams, "文件上传失败")
        return
    }
    
    // 保存文件
    dst := path.Join("uploads", file.Filename)
    if err := c.SaveUploadedFile(file, dst); err != nil {
        render.Error(c, render.CodeInternalError, "文件保存失败")
        return
    }
    
    render.Success(c, gin.H{
        "file_name": file.Filename,
        "file_size": file.Size,
        "file_path": dst,
    }, "文件上传成功")
}
```

### Q: 如何集成数据库？

GinUtil 专注于 Web 层，可以与任何数据库操作库配合使用：

```go
func GetUsers(c *gin.Context) {
    // 使用数据库获取用户列表
    users, err := db.QueryUsers()
    if err != nil {
        render.Error(c, render.CodeInternalError, "获取用户失败")
        return
    }
    
    render.Success(c, users)
}
```

### Q: 如何处理大型项目的路由组织？

对于大型项目，建议使用模块化的路由组织方式：

```go
// 用户模块路由
func RegisterUserRoutes(r *gin.Engine) {
    users := r.Group("/users")
    users.GET("", ListUsers)
    users.GET("/:id", GetUser)
    // ...
}

// 产品模块路由
func RegisterProductRoutes(r *gin.Engine) {
    products := r.Group("/products")
    products.GET("", ListProducts)
    products.GET("/:id", GetProduct)
    // ...
}

// 在主函数中注册所有模块路由
func main() {
    r := gin.Default()
    
    RegisterUserRoutes(r)
    RegisterProductRoutes(r)
    
    r.Run(":8080")
}
```

## 贡献与支持

欢迎提交 Issues 和 Pull Requests 来帮助改进 GinUtil。

## 许可证

本项目基于 Apache License 2.0 许可证开源。 