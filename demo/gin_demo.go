package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Humphrey-He/go-generic-utils/ginutil/ecode"
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/auth"
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/logger"
	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/recovery"
	"github.com/Humphrey-He/go-generic-utils/ginutil/paginator"
	"github.com/Humphrey-He/go-generic-utils/ginutil/register"
)

// GinDemoConfig 是gin演示的配置
type GinDemoConfig struct {
	Port        int    `json:"port"`
	JWTSecret   string `json:"jwt_secret"`
	Environment string `json:"environment"`
}

// 默认配置
var defaultConfig = GinDemoConfig{
	Port:        8080,
	JWTSecret:   "your-jwt-secret-key",
	Environment: "development",
}

// 全局配置变量
var config = defaultConfig

// 定义用户相关结构
type UserID int64
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

// UserModel 用户模型
type UserModel struct {
	ID       UserID     `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Roles    []UserRole `json:"roles"`
	Created  time.Time  `json:"created"`
	Updated  time.Time  `json:"updated"`
}

// UserIdentity 用户身份信息
type UserIdentity struct {
	ID    UserID     `json:"id"`
	Roles []UserRole `json:"roles"`
}

// ProductModel 产品模型
type ProductModel struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CategoryID  int       `json:"category_id"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// 请求和响应结构体
// 创建用户请求
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Roles    []string `json:"roles" binding:"dive,oneof=admin user guest"`
}

// 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expires_at"`
	User      UserModel `json:"user"`
}

// 创建产品请求
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=3,max=100"`
	Description string  `json:"description" binding:"max=1000"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	CategoryID  int     `json:"category_id" binding:"required,gt=0"`
}

// 产品过滤请求
type ProductFilterRequest struct {
	CategoryID  *int     `form:"category_id"`
	MinPrice    *float64 `form:"min_price"`
	MaxPrice    *float64 `form:"max_price"`
	SearchQuery *string  `form:"q"`
	SortBy      *string  `form:"sort_by" binding:"omitempty,oneof=price created name"`
	SortOrder   *string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page        int      `form:"page" binding:"omitempty,gt=0"`
	PageSize    int      `form:"page_size" binding:"omitempty,gt=0,max=100"`
}

// JWT Claims结构
type UserClaims struct {
	UserIdentity
	jwt.RegisteredClaims
}

// 控制器 - 用户控制器
type UserController struct {
	register.ControllerBase
}

// 用户控制器方法

// CreateUser 创建用户
// @route:"POST /create [创建用户] (创建新用户账号)"
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest

	// 使用绑定和校验
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// 使用标准错误响应
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    ecode.ErrorCodeUserInputInvalid,
			"message": err.Error(),
		})
		return
	}

	// 模拟创建用户
	user := UserModel{
		ID:       1001,
		Username: req.Username,
		Email:    req.Email,
		Roles:    []UserRole{},
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	// 转换角色
	for _, r := range req.Roles {
		user.Roles = append(user.Roles, UserRole(r))
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, gin.H{
		"code":    ecode.OK,
		"message": ecode.SuccessMessage,
		"data":    user,
	})
}

// Login 用户登录
// @route:"POST /login [用户登录] (验证用户凭据并返回JWT令牌)"
func (c *UserController) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    ecode.ErrorCodeUserInputInvalid,
			"message": err.Error(),
		})
		return
	}

	// 模拟验证用户
	if req.Username != "admin" || req.Password != "password123" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    ecode.AccessUnauthorized,
			"message": "用户名或密码错误",
		})
		return
	}

	// 创建用户模型
	user := UserModel{
		ID:       1001,
		Username: "admin",
		Email:    "admin@example.com",
		Roles:    []UserRole{RoleAdmin},
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	// 创建JWT令牌
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	claims := UserClaims{
		UserIdentity: UserIdentity{
			ID:    user.ID,
			Roles: user.Roles,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    ecode.ErrorCodeInternal,
			"message": "生成令牌失败",
		})
		return
	}

	// 返回登录响应
	loginResp := LoginResponse{
		Token:     signedToken,
		ExpiresAt: expiresAt.Unix(),
		User:      user,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    ecode.OK,
		"message": ecode.SuccessMessage,
		"data":    loginResp,
	})
}

// GetProfile 获取用户资料
// @route:"GET /profile [获取用户资料] (获取当前登录用户的详细资料)"
func (c *UserController) GetProfile(ctx *gin.Context) {
	// 从上下文获取用户身份
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    ecode.AccessUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	// 模拟获取用户资料
	user := UserModel{
		ID:       userID.(UserID),
		Username: "admin",
		Email:    "admin@example.com",
		Roles:    []UserRole{RoleAdmin},
		Created:  time.Now().Add(-24 * time.Hour),
		Updated:  time.Now(),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    ecode.OK,
		"message": ecode.SuccessMessage,
		"data":    user,
	})
}

// RegisterRoutes 注册路由
func (c *UserController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/create", c.CreateUser)
	group.POST("/login", c.Login)

	// 需要认证的路由
	authGroup := group.Group("/")
	jwtMiddleware := auth.RequireJWT[UserID, UserRole](
		auth.HMACKeyFunc([]byte(config.JWTSecret)),
		auth.WithClaimsFactory(func() jwt.Claims {
			return &UserClaims{}
		}),
		auth.WithSuccessHandler(auth.UserIdentitySuccessHandler[UserID, UserRole]()),
	)

	// 使用中间件
	authGroup.Use(jwtMiddleware)

	// 需要认证的路由
	authGroup.GET("/profile", c.GetProfile)
}

// 控制器 - 产品控制器
type ProductController struct {
	register.ControllerBase
}

// GetProducts 获取产品列表
// @route:"GET / [获取产品列表] (获取产品列表，支持分页和过滤)"
func (c *ProductController) GetProducts(ctx *gin.Context) {
	var req ProductFilterRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    ecode.ErrorCodeUserInputInvalid,
			"message": err.Error(),
		})
		return
	}

	// 使用分页器绑定分页参数
	page := req.Page
	if page <= 0 {
		page = paginator.DefaultPageNum
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = paginator.DefaultPageSize
	}

	// 模拟获取产品
	products := []ProductModel{
		{
			ID:          1,
			Name:        "产品1",
			Description: "这是产品1的描述",
			Price:       99.99,
			CategoryID:  1,
			Created:     time.Now().Add(-48 * time.Hour),
			Updated:     time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          2,
			Name:        "产品2",
			Description: "这是产品2的描述",
			Price:       199.99,
			CategoryID:  2,
			Created:     time.Now().Add(-24 * time.Hour),
			Updated:     time.Now(),
		},
	}

	// 模拟总数
	total := int64(2)

	// 计算总页数
	totalPages := int(0)
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	// 创建分页信息
	pageInfo := &paginator.PageInfo{
		PageNum:  page,
		PageSize: pageSize,
		Pages:    totalPages,
	}

	// 使用分页响应
	ctx.JSON(http.StatusOK, gin.H{
		"code":    ecode.OK,
		"message": ecode.SuccessMessage,
		"data": gin.H{
			"items":    products,
			"total":    total,
			"pageInfo": pageInfo,
		},
	})
}

// GetProduct 获取单个产品
// @route:"GET /:id [获取产品详情] (获取单个产品的详细信息)"
func (c *ProductController) GetProduct(ctx *gin.Context) {
	id := ctx.Param("id")

	// 简单验证
	if id != "1" && id != "2" {
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    ecode.ErrorCodeNotFound,
			"message": "产品不存在",
		})
		return
	}

	// 模拟获取产品
	product := ProductModel{
		ID:          1,
		Name:        "产品1",
		Description: "这是产品1的详细描述，包含更多信息",
		Price:       99.99,
		CategoryID:  1,
		Created:     time.Now().Add(-48 * time.Hour),
		Updated:     time.Now().Add(-24 * time.Hour),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    ecode.OK,
		"message": ecode.SuccessMessage,
		"data":    product,
	})
}

// CreateProduct 创建产品
// @route:"POST / [创建产品] (创建新产品)"
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	var req CreateProductRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    ecode.ErrorCodeUserInputInvalid,
			"message": err.Error(),
		})
		return
	}

	// 模拟创建产品
	product := ProductModel{
		ID:          3,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    ecode.OK,
		"message": ecode.SuccessMessage,
		"data":    product,
	})
}

// RegisterRoutes 注册产品控制器路由
func (c *ProductController) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/", c.GetProducts)
	group.GET("/:id", c.GetProduct)

	// 需要认证和授权的路由
	authGroup := group.Group("/")
	jwtMiddleware := auth.RequireJWT[UserID, UserRole](
		auth.HMACKeyFunc([]byte(config.JWTSecret)),
		auth.WithClaimsFactory(func() jwt.Claims {
			return &UserClaims{}
		}),
		auth.WithSuccessHandler(auth.UserIdentitySuccessHandler[UserID, UserRole]()),
	)

	// 需要管理员角色
	adminRoleMiddleware := auth.RequireRoles[UserID, UserRole](RoleAdmin)

	authGroup.Use(jwtMiddleware, adminRoleMiddleware)
	authGroup.POST("/", c.CreateProduct)
}

// RunGinDemo 运行Gin演示
func RunGinDemo() {
	// 设置Gin模式
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 配置全局中间件
	// 日志中间件
	r.Use(logger.New())

	// 恢复中间件
	r.Use(recovery.New())

	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 限流中间件 - 全局限流
	r.Use(func(c *gin.Context) {
		// 模拟限流中间件，实际生产中需要实现真正的限流逻辑
		c.Next()
	})

	// API版本前缀
	api := r.Group("/api/v1")

	// 注册控制器
	userController := &UserController{}
	userController.SetName("User")

	productController := &ProductController{}
	productController.SetName("Product")

	// 将控制器注册到全局注册表
	registry := register.NewRegistry()
	registry.RegisterController("user", userController)
	registry.RegisterController("product", productController)

	// 自动注册所有控制器的路由
	registry.RegisterRoutesWithGroup(api)

	// 注册健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 启动服务器
	log.Printf("Gin 演示服务器启动在端口 %d\n", config.Port)
	log.Printf("API地址: http://localhost:%d/api/v1\n", config.Port)
	log.Printf("示例API:\n")
	log.Printf("- 用户注册: POST http://localhost:%d/api/v1/user/create\n", config.Port)
	log.Printf("- 用户登录: POST http://localhost:%d/api/v1/user/login\n", config.Port)
	log.Printf("- 用户资料: GET http://localhost:%d/api/v1/user/profile (需要认证)\n", config.Port)
	log.Printf("- 产品列表: GET http://localhost:%d/api/v1/product/\n", config.Port)
	log.Printf("- 产品详情: GET http://localhost:%d/api/v1/product/1\n", config.Port)
	log.Printf("- 创建产品: POST http://localhost:%d/api/v1/product/ (需要认证和管理员角色)\n", config.Port)

	// 由于这是一个演示，我们不实际启动服务器
	// 在实际应用中会使用以下代码启动服务器
	// r.Run(fmt.Sprintf(":%d", config.Port))
}

// 如果想单独运行这个演示，可以使用以下main函数
/*
func main() {
	RunGinDemo()
}
*/
