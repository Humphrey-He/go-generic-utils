package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/Humphrey-He/go-generic-utils/ginutil/render"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 用户模型
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,gt=0"`
}

// 产品模型
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// 评论模型
type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// 模拟数据库
var (
	users = []User{
		{ID: 1, Username: "admin", Email: "admin@example.com", Age: 30},
		{ID: 2, Username: "user1", Email: "user1@example.com", Age: 25},
	}

	products = []Product{
		{ID: 1, Name: "笔记本电脑", Description: "高性能笔记本电脑", Price: 6999.00, Stock: 100},
		{ID: 2, Name: "智能手机", Description: "最新款智能手机", Price: 4999.00, Stock: 200},
		{ID: 3, Name: "无线耳机", Description: "蓝牙无线耳机", Price: 999.00, Stock: 500},
	}

	comments = []Comment{
		{ID: 1, UserID: 1, Content: "非常好用的产品", CreatedAt: time.Now().Add(-24 * time.Hour)},
		{ID: 2, UserID: 2, Content: "性价比很高", CreatedAt: time.Now().Add(-12 * time.Hour)},
	}
)

// TraceID 中间件，为每个请求生成追踪 ID
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 TraceID，如果没有则生成一个
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 将 TraceID 设置到上下文
		c.Set(render.ContextKeyTraceID, traceID)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// 简单的绑定和验证函数
func BindAndValidateJSON(c *gin.Context, obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

// 设置路由
func setupRouter() *gin.Engine {
	// 创建 Gin 引擎
	r := gin.Default()

	// 配置渲染器
	render.Configure(render.Config{
		JSONPrettyPrint:          true,
		HTMLTemplateDir:          "templates/*",
		DefaultHTMLErrorTemplate: "error.html",
	})

	// 注册中间件
	r.Use(TraceIDMiddleware())

	// 注册路由组
	api := r.Group("/api")
	{
		// 用户相关路由
		users := api.Group("/users")
		{
			users.GET("", listUsers)
			users.GET("/:id", getUserByID)
			users.POST("", createUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}

		// 产品相关路由
		products := api.Group("/products")
		{
			products.GET("", listProducts)
			products.GET("/:id", getProductByID)
		}

		// 评论相关路由
		comments := api.Group("/comments")
		{
			comments.GET("", listComments)
			comments.POST("", createComment)
		}

		// 错误处理示例
		errors := api.Group("/errors")
		{
			errors.GET("/not-found", notFoundError)
			errors.GET("/bad-request", badRequestError)
			errors.GET("/unauthorized", unauthorizedError)
			errors.GET("/forbidden", forbiddenError)
			errors.GET("/internal", internalError)
			errors.GET("/with-data", errorWithData)
			errors.GET("/custom", customError)
		}

		// XML 响应示例
		xml := api.Group("/xml")
		{
			xml.GET("/users", listUsersXML)
			xml.GET("/products", listProductsXML)
		}
	}

	// HTML 页面示例
	web := r.Group("/web")
	{
		web.GET("/users", listUsersHTML)
		web.GET("/users/:id", getUserHTMLByID)
		web.GET("/products", listProductsHTML)
		web.GET("/error", errorPageHTML)
	}

	return r
}

// 用户相关处理函数
func listUsers(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	// 模拟分页
	total := int64(len(users))
	start := (page - 1) * size
	end := start + size
	if start >= len(users) {
		start = 0
		end = 0
	}
	if end > len(users) {
		end = len(users)
	}

	// 发送分页响应
	render.Paginated(c, users[start:end], total, page, size)
}

func getUserByID(c *gin.Context) {
	// 获取用户 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.BadRequest(c, "无效的用户 ID")
		return
	}

	// 查找用户
	for _, user := range users {
		if user.ID == id {
			render.Success(c, user)
			return
		}
	}

	render.NotFound(c, fmt.Sprintf("用户 ID %d 不存在", id))
}

func createUser(c *gin.Context) {
	var user User
	if err := BindAndValidateJSON(c, &user); err != nil {
		render.ValidationError(c, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 模拟生成新用户 ID
	user.ID = len(users) + 1

	// 添加到用户列表
	users = append(users, user)

	render.Success(c, user, "用户创建成功")
}

func updateUser(c *gin.Context) {
	// 获取用户 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.BadRequest(c, "无效的用户 ID")
		return
	}

	var updatedUser User
	if err := BindAndValidateJSON(c, &updatedUser); err != nil {
		render.ValidationError(c, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 查找并更新用户
	for i, user := range users {
		if user.ID == id {
			updatedUser.ID = id
			users[i] = updatedUser
			render.Success(c, updatedUser, "用户更新成功")
			return
		}
	}

	render.NotFound(c, fmt.Sprintf("用户 ID %d 不存在", id))
}

func deleteUser(c *gin.Context) {
	// 获取用户 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.BadRequest(c, "无效的用户 ID")
		return
	}

	// 查找并删除用户
	for i, user := range users {
		if user.ID == id {
			// 删除用户（简单实现，实际应用中可能需要更复杂的逻辑）
			users = append(users[:i], users[i+1:]...)
			render.Success[interface{}](c, nil, "用户删除成功")
			return
		}
	}

	render.NotFound(c, fmt.Sprintf("用户 ID %d 不存在", id))
}

// 产品相关处理函数
func listProducts(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	// 模拟分页
	total := int64(len(products))
	start := (page - 1) * size
	end := start + size
	if start >= len(products) {
		start = 0
		end = 0
	}
	if end > len(products) {
		end = len(products)
	}

	// 发送分页响应
	render.Paginated(c, products[start:end], total, page, size)
}

func getProductByID(c *gin.Context) {
	// 获取产品 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.BadRequest(c, "无效的产品 ID")
		return
	}

	// 查找产品
	for _, product := range products {
		if product.ID == id {
			render.Success(c, product)
			return
		}
	}

	render.NotFound(c, fmt.Sprintf("产品 ID %d 不存在", id))
}

// 评论相关处理函数
func listComments(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	// 模拟分页
	total := int64(len(comments))
	start := (page - 1) * size
	end := start + size
	if start >= len(comments) {
		start = 0
		end = 0
	}
	if end > len(comments) {
		end = len(comments)
	}

	// 发送分页响应
	render.Paginated(c, comments[start:end], total, page, size)
}

func createComment(c *gin.Context) {
	var comment Comment
	if err := BindAndValidateJSON(c, &comment); err != nil {
		render.ValidationError(c, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 模拟生成新评论 ID
	comment.ID = len(comments) + 1
	comment.CreatedAt = time.Now()

	// 添加到评论列表
	comments = append(comments, comment)

	render.Success(c, comment, "评论创建成功")
}

// 错误处理示例
func notFoundError(c *gin.Context) {
	render.NotFound(c, "请求的资源不存在")
}

func badRequestError(c *gin.Context) {
	render.BadRequest(c, "请求参数错误")
}

func unauthorizedError(c *gin.Context) {
	render.Unauthorized(c, "请先登录")
}

func forbiddenError(c *gin.Context) {
	render.Forbidden(c, "没有权限访问此资源")
}

func internalError(c *gin.Context) {
	render.InternalError(c, "服务器内部错误")
}

func errorWithData(c *gin.Context) {
	errorData := map[string]interface{}{
		"error_code": "DB_CONNECTION_FAILED",
		"details":    "无法连接到数据库",
		"timestamp":  time.Now().Unix(),
	}
	render.ErrorWithData(c, render.CodeInternalError, "数据库连接错误", errorData)
}

func customError(c *gin.Context) {
	render.Custom(c, 2001, "自定义错误消息", gin.H{
		"custom_field": "自定义值",
	}, http.StatusTeapot)
}

// XML 响应示例
func listUsersXML(c *gin.Context) {
	render.SuccessXML(c, users)
}

func listProductsXML(c *gin.Context) {
	render.SuccessXML(c, products)
}

// HTML 页面示例
func listUsersHTML(c *gin.Context) {
	render.HTML(c, http.StatusOK, "users.html", gin.H{
		"title": "用户列表",
		"users": users,
	})
}

func getUserHTMLByID(c *gin.Context) {
	// 获取用户 ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		render.BadRequestPage(c, "无效的用户 ID")
		return
	}

	// 查找用户
	for _, user := range users {
		if user.ID == id {
			render.HTML(c, http.StatusOK, "user_detail.html", gin.H{
				"title": "用户详情",
				"user":  user,
			})
			return
		}
	}

	render.NotFoundPage(c, fmt.Sprintf("用户 ID %d 不存在", id))
}

func listProductsHTML(c *gin.Context) {
	render.HTML(c, http.StatusOK, "products.html", gin.H{
		"title":    "产品列表",
		"products": products,
	})
}

func errorPageHTML(c *gin.Context) {
	errorType := c.DefaultQuery("type", "not_found")

	switch errorType {
	case "not_found":
		render.NotFoundPage(c, "请求的页面不存在")
	case "bad_request":
		render.BadRequestPage(c, "请求参数错误")
	case "unauthorized":
		render.UnauthorizedPage(c, "请先登录")
	case "forbidden":
		render.ForbiddenPage(c, "没有权限访问此页面")
	default:
		render.InternalErrorPage(c, "服务器内部错误")
	}
}

// 主函数
func main() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 设置 Gin 模式
	gin.SetMode(gin.DebugMode)

	// 创建路由
	router := setupRouter()

	// 启动服务器
	log.Println("服务器启动在 http://localhost:8080")
	log.Println("基础用法示例:")
	log.Println("- JSON 响应: http://localhost:8080/api/users")
	log.Println("- 分页响应: http://localhost:8080/api/products?page=1&size=2")
	log.Println("- 错误响应: http://localhost:8080/api/errors/not-found")
	log.Println("- XML 响应: http://localhost:8080/api/xml/users")
	log.Println("- HTML 页面: http://localhost:8080/web/users")
	log.Println("- HTML 错误页面: http://localhost:8080/web/error?type=not_found")

	router.Run(":8080")
}
