package register_test

import (
	"fmt"
	"net/http"

	"github.com/Humphrey-He/go-generic-utils/ginutil/register"

	"github.com/gin-gonic/gin"
)

// UserController 是一个用户控制器示例。
type UserController struct {
	register.ControllerBase
	// 这里可以添加依赖，如数据库连接、配置等
}

// RegisterRoutes 实现 Routable 接口，注册路由。
func (c *UserController) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("", c.GetUsers)
	group.GET("/:id", c.GetUser)
	group.POST("", c.CreateUser)
	group.PUT("/:id", c.UpdateUser)
	group.DELETE("/:id", c.DeleteUser)
}

// GetUsers 获取用户列表。
func (c *UserController) GetUsers(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "获取用户列表"})
}

// GetUser 获取单个用户。
func (c *UserController) GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("获取用户 %s", id)})
}

// CreateUser 创建用户。
func (c *UserController) CreateUser(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, gin.H{"message": "创建用户"})
}

// UpdateUser 更新用户。
func (c *UserController) UpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("更新用户 %s", id)})
}

// DeleteUser 删除用户。
func (c *UserController) DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("删除用户 %s", id)})
}

// ProductController 是一个产品控制器示例，实现 RESTController 接口。
type ProductController struct {
	register.ControllerBase
}

// Index 获取产品列表。
func (c *ProductController) Index(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "获取产品列表"})
}

// Show 获取单个产品。
func (c *ProductController) Show(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("获取产品 %s", id)})
}

// Create 创建产品。
func (c *ProductController) Create(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, gin.H{"message": "创建产品"})
}

// Update 更新产品。
func (c *ProductController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("更新产品 %s", id)})
}

// Delete 删除产品。
func (c *ProductController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("删除产品 %s", id)})
}

// TaggedController 是一个使用结构体标签的控制器示例。
type TaggedController struct {
	// 路由配置字段
	RouteGetTags   register.RouteInfo `route:"GET /tags [获取标签列表] (获取所有标签)"`
	RouteGetTag    register.RouteInfo `route:"GET /tags/:id [获取单个标签] (获取指定ID的标签)"`
	RouteCreateTag register.RouteInfo `route:"POST /tags [创建标签] (创建新标签)"`
	RouteUpdateTag register.RouteInfo `route:"PUT /tags/:id [更新标签] (更新指定ID的标签)"`
	RouteDeleteTag register.RouteInfo `route:"DELETE /tags/:id [删除标签] (删除指定ID的标签)"`
}

// GetTags 获取标签列表。
func (c *TaggedController) GetTags(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "获取标签列表"})
}

// GetTag 获取单个标签。
func (c *TaggedController) GetTag(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("获取标签 %s", id)})
}

// CreateTag 创建标签。
func (c *TaggedController) CreateTag(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, gin.H{"message": "创建标签"})
}

// UpdateTag 更新标签。
func (c *TaggedController) UpdateTag(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("更新标签 %s", id)})
}

// DeleteTag 删除标签。
func (c *TaggedController) DeleteTag(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("删除标签 %s", id)})
}

// ExampleRegisterRoutes 展示如何使用 Routable 接口注册路由。
func ExampleRegisterRoutes() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建控制器
	userController := &UserController{}
	userController.SetName("users")

	// 注册控制器到全局注册器
	register.RegisterGlobal("users", userController)

	// 注册全局路由
	register.RegisterGlobalRoutes(router)

	// 输出：
	// GET    /users
	// GET    /users/:id
	// POST   /users
	// PUT    /users/:id
	// DELETE /users/:id
}

// ExampleResourceGroup 展示如何使用 ResourceGroup 注册 RESTful 资源。
func ExampleResourceGroup() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建 API 路由组
	api := register.NewAPIGroup(router, "/api", "v1")

	// 创建控制器
	productController := &ProductController{}

	// 注册资源路由
	api.Resource("products", productController)

	// 注册嵌套资源路由
	api.NestedResource("products", "productId", "reviews", &ProductReviewController{})

	// 输出：
	// GET    /api/v1/products
	// GET    /api/v1/products/:id
	// POST   /api/v1/products
	// PUT    /api/v1/products/:id
	// DELETE /api/v1/products/:id
	// GET    /api/v1/products/:productId/reviews
	// GET    /api/v1/products/:productId/reviews/:id
	// POST   /api/v1/products/:productId/reviews
	// PUT    /api/v1/products/:productId/reviews/:id
	// DELETE /api/v1/products/:productId/reviews/:id
}

// ProductReviewController 是一个产品评论控制器示例。
type ProductReviewController struct{}

// Index 获取评论列表。
func (c *ProductReviewController) Index(ctx *gin.Context) {
	productId := ctx.Param("productId")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("获取产品 %s 的评论列表", productId)})
}

// Show 获取单个评论。
func (c *ProductReviewController) Show(ctx *gin.Context) {
	productId := ctx.Param("productId")
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("获取产品 %s 的评论 %s", productId, id)})
}

// Create 创建评论。
func (c *ProductReviewController) Create(ctx *gin.Context) {
	productId := ctx.Param("productId")
	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("为产品 %s 创建评论", productId)})
}

// Update 更新评论。
func (c *ProductReviewController) Update(ctx *gin.Context) {
	productId := ctx.Param("productId")
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("更新产品 %s 的评论 %s", productId, id)})
}

// Delete 删除评论。
func (c *ProductReviewController) Delete(ctx *gin.Context) {
	productId := ctx.Param("productId")
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("删除产品 %s 的评论 %s", productId, id)})
}

// ExampleRegisterControllerWithTags 展示如何使用结构体标签注册路由。
func ExampleRegisterControllerWithTags() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建控制器
	taggedController := &TaggedController{}

	// 注册控制器路由
	err := register.RegisterGlobalControllerWithTags(router, taggedController)
	if err != nil {
		fmt.Println("注册控制器路由失败:", err)
		return
	}

	// 输出：
	// GET    /tags
	// GET    /tags/:id
	// POST   /tags
	// PUT    /tags/:id
	// DELETE /tags/:id
}

// ExampleSwaggerUI 展示如何注册 Swagger UI。
func ExampleSwaggerUI() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册 Swagger UI
	err := register.AddSwaggerEndpoints(router, "./docs/swagger.json")
	if err != nil {
		fmt.Println("注册 Swagger UI 失败:", err)
		return
	}

	// 输出：
	// GET    /swagger/doc.json
	// GET    /swagger
	// GET    /swagger/swagger-config.json
}

// ExampleFluentAPI 展示如何使用流式 API。
func ExampleFluentAPI() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建路由
	r := register.NewRouter(router)

	// 使用流式 API
	r.At("/hello").GET(func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello, World!")
	})

	r.Group("/api").
		At("/status").GET(func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 输出：
	// GET    /hello
	// GET    /api/status
}

// ExampleNamespace 展示如何使用命名空间组织 API。
func ExampleNamespace() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建 API 路由组
	api := register.NewAPIGroup(router, "/api", "v1")

	// 创建控制器
	userController := &UserRESTController{}
	productController := &ProductController{}

	// 使用命名空间组织 API
	api.Namespace("admin", func(ns *register.ResourceGroup) {
		ns.Resource("users", userController)
		ns.Resource("products", productController)
	})

	// 输出：
	// GET    /api/v1/admin/users
	// GET    /api/v1/admin/users/:id
	// POST   /api/v1/admin/users
	// PUT    /api/v1/admin/users/:id
	// DELETE /api/v1/admin/users/:id
	// GET    /api/v1/admin/products
	// GET    /api/v1/admin/products/:id
	// POST   /api/v1/admin/products
	// PUT    /api/v1/admin/products/:id
	// DELETE /api/v1/admin/products/:id
}

// UserRESTController 是一个实现了 RESTController 接口的用户控制器。
type UserRESTController struct {
	register.ControllerBase
}

// Index 获取用户列表。
func (c *UserRESTController) Index(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "获取用户列表"})
}

// Show 获取单个用户。
func (c *UserRESTController) Show(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("获取用户 %s", id)})
}

// Create 创建用户。
func (c *UserRESTController) Create(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, gin.H{"message": "创建用户"})
}

// Update 更新用户。
func (c *UserRESTController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("更新用户 %s", id)})
}

// Delete 删除用户。
func (c *UserRESTController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("删除用户 %s", id)})
}

// ExampleResourceWithActions 展示如何注册带自定义操作的资源。
func ExampleResourceWithActions() {
	// 创建 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建 API 路由组
	api := register.NewAPIGroup(router, "/api", "v1")

	// 创建控制器
	productController := &ProductController{}

	// 定义自定义操作
	actions := map[string]gin.HandlerFunc{
		// 资源实例操作，如 /products/:id/publish
		"PUT:publish": func(ctx *gin.Context) {
			id := ctx.Param("id")
			ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("发布产品 %s", id)})
		},
		// 集合操作，如 /products/export
		"collection:GET:export": func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"message": "导出所有产品"})
		},
	}

	// 注册带自定义操作的资源
	api.ResourceWithActions("products", productController, actions)

	// 输出：
	// GET    /api/v1/products
	// GET    /api/v1/products/:id
	// POST   /api/v1/products
	// PUT    /api/v1/products/:id
	// DELETE /api/v1/products/:id
	// PUT    /api/v1/products/:id/publish
	// GET    /api/v1/products/export
}
