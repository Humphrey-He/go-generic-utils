package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/noobtrump/go-generic-utils/ginutil/render"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 用户设置模型
type UserSettings struct {
	UserID      int                    `json:"user_id" binding:"required"`
	Theme       string                 `json:"theme" binding:"required"`
	Language    string                 `json:"language" binding:"required"`
	Preferences map[string]interface{} `json:"preferences"`
}

// 评论模型
type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
}

// 模拟数据库
var (
	// 用户设置存储
	userSettings = map[int]UserSettings{
		1: {UserID: 1, Theme: "dark", Language: "zh-CN", Preferences: map[string]interface{}{
			"notifications": true,
			"display_mode":  "compact",
		}},
	}

	// 评论存储
	comments = []Comment{
		{ID: 1, UserID: 1, Content: "这是一条评论", CreatedAt: time.Now().Add(-24 * time.Hour)},
	}

	// 模拟令牌存储
	tokenStore = sync.Map{}
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

// 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		latency := endTime.Sub(startTime)

		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()
		traceID := c.GetString(render.ContextKeyTraceID)

		// 记录日志
		log.Printf("[GIN] %s | %3d | %13v | %s | %s | %s",
			endTime.Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			method,
			path,
			traceID,
		)
	}
}

// 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误
				log.Printf("Panic recovered: %v", err)

				// 发送错误响应
				render.InternalError(c, "服务器内部错误")
				c.Abort()
			}
		}()

		c.Next()
	}
}

// XSS 防护中间件
func XSSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全相关的 HTTP 头部
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// CSRF 中间件
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于非 GET 请求，验证 CSRF 令牌
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions {
			// 从请求头获取 CSRF 令牌
			token := c.GetHeader("X-CSRF-Token")
			cookie, err := c.Cookie("csrf_token")

			// 验证令牌
			if err != nil || token == "" || token != cookie {
				render.Forbidden(c, "CSRF 令牌无效")
				c.Abort()
				return
			}
		}

		// 生成新的 CSRF 令牌
		token := uuid.New().String()

		// 将令牌存储在 Cookie 中
		c.SetCookie("csrf_token", token, 3600, "/", "", false, true)

		// 将令牌存储在上下文中，以便在模板中使用
		c.Set("csrf_token", token)

		c.Next()
	}
}

// 限流中间件
func RateLimitMiddleware(limit int, duration time.Duration) gin.HandlerFunc {
	// 创建令牌桶
	type bucket struct {
		tokens     int
		lastRefill time.Time
	}

	// 存储 IP 地址对应的令牌桶
	buckets := sync.Map{}

	return func(c *gin.Context) {
		// 获取客户端 IP
		ip := c.ClientIP()

		// 获取或创建令牌桶
		b, _ := buckets.LoadOrStore(ip, &bucket{
			tokens:     limit,
			lastRefill: time.Now(),
		})
		bkt := b.(*bucket)

		// 加锁以确保并发安全
		mu := &sync.Mutex{}
		mu.Lock()
		defer mu.Unlock()

		// 计算自上次填充以来经过的时间
		elapsed := time.Since(bkt.lastRefill)

		// 填充令牌
		tokensToAdd := int(elapsed.Seconds() * float64(limit) / duration.Seconds())
		if tokensToAdd > 0 {
			bkt.tokens = min(bkt.tokens+tokensToAdd, limit)
			bkt.lastRefill = time.Now()
		}

		// 检查是否有足够的令牌
		if bkt.tokens <= 0 {
			render.Error(c, render.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		// 消耗一个令牌
		bkt.tokens--

		c.Next()
	}
}

// 用户认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取认证令牌
		token := c.GetHeader("Authorization")
		if token == "" {
			render.Unauthorized(c, "未提供认证令牌")
			c.Abort()
			return
		}

		// 验证令牌（简化实现）
		userID, ok := tokenStore.Load(token)
		if !ok {
			render.Unauthorized(c, "认证令牌无效")
			c.Abort()
			return
		}

		// 将用户 ID 存储在上下文中
		c.Set(render.ContextKeyUserID, userID)

		c.Next()
	}
}

// 链式响应示例
func chainResponseDemo(c *gin.Context) {
	// 使用链式 API 构建响应
	render.Resp(c).
		Code(0).
		Message("操作成功").
		Data(gin.H{
			"user_id": 1,
			"balance": 100.5,
			"status":  "active",
		}).
		Status(http.StatusOK).
		JSON()
}

// 错误处理示例
func errorHandlingDemo(c *gin.Context) {
	// 创建一个自定义错误
	err := render.NewError(render.CodeNotFound, "找不到指定的资源")

	// 使用 From 方法从错误创建响应
	render.From(c, err).JSON()
}

// 包装错误示例
func errorWrappingDemo(c *gin.Context) {
	// 模拟数据库错误
	dbErr := errors.New("数据库连接失败")

	// 包装错误
	err := render.WrapError(render.CodeInternalError, "无法获取用户数据", dbErr)

	// 使用 HandleError 处理错误
	render.HandleError(c, err)
}

// 文件下载示例
func fileDownloadDemo(c *gin.Context) {
	// 设置文件路径和名称
	filePath := "./example/data/sample.pdf"
	fileName := "用户手册.pdf"

	// 使用 Download 函数发送文件
	render.Download(c, filePath, fileName)
}

// 自定义 HTML 辅助函数示例
func setupHTMLHelpers() {
	// 添加用户数据辅助函数
	render.AddUserDataHelper()

	// 添加 CSRF 令牌辅助函数
	render.AddCSRFHelper()

	// 添加 Flash 消息辅助函数
	render.AddFlashMessageHelper()

	// 添加自定义辅助函数
	render.RegisterHTMLHelper(func(c *gin.Context, data gin.H) gin.H {
		return gin.H{
			"app_name":     "高级示例应用",
			"app_version":  "1.0.0",
			"current_year": time.Now().Year(),
		}
	})
}

// 获取用户设置
func getUserSettings(c *gin.Context) {
	// 获取用户 ID
	userIDStr := c.Param("id")
	userID := 0
	fmt.Sscanf(userIDStr, "%d", &userID)

	// 查找用户设置
	settings, ok := userSettings[userID]
	if !ok {
		render.NotFound(c, "未找到用户设置")
		return
	}

	render.Success(c, settings)
}

// 更新用户设置
func updateUserSettings(c *gin.Context) {
	// 获取用户 ID
	userIDStr := c.Param("id")
	userID := 0
	fmt.Sscanf(userIDStr, "%d", &userID)

	// 绑定请求数据
	var settings UserSettings
	if err := render.BindAndValidate(c, &settings); err != nil {
		render.HandleError(c, err)
		return
	}

	// 确保 UserID 匹配
	if settings.UserID != userID {
		render.BadRequest(c, "用户 ID 不匹配")
		return
	}

	// 更新设置
	userSettings[userID] = settings

	render.Success(c, settings, "设置已更新")
}

// 创建评论
func createComment(c *gin.Context) {
	// 绑定请求数据
	var comment Comment
	if !render.ValidateAndHandle(c, &comment) {
		return
	}

	// 设置评论 ID 和创建时间
	comment.ID = len(comments) + 1
	comment.CreatedAt = time.Now()

	// 添加到评论列表
	comments = append(comments, comment)

	render.Success(c, comment, "评论已创建")
}

// 获取 CSRF 令牌
func getCSRFToken(c *gin.Context) {
	token, exists := c.Get("csrf_token")
	if !exists {
		render.InternalError(c, "无法生成 CSRF 令牌")
		return
	}

	render.Success(c, gin.H{
		"csrf_token": token,
	})
}

// 模拟 panic
func simulatePanic(c *gin.Context) {
	// 故意引发 panic
	panic("这是一个模拟的 panic")
}

// 设置高级路由
func setupAdvancedRouter() *gin.Engine {
	// 创建 Gin 引擎
	r := gin.New()

	// 配置渲染器
	render.Configure(render.Config{
		JSONPrettyPrint: true,
	})

	// 注册中间件
	r.Use(TraceIDMiddleware())
	r.Use(LoggerMiddleware())
	r.Use(RecoveryMiddleware())
	r.Use(XSSMiddleware())

	// 设置 HTML 辅助函数
	setupHTMLHelpers()

	// 高级 API 路由
	api := r.Group("/api/advanced")
	{
		// 应用限流中间件
		api.Use(RateLimitMiddleware(10, time.Minute))

		// 链式响应示例
		api.GET("/chain", chainResponseDemo)

		// 错误处理示例
		api.GET("/error", errorHandlingDemo)
		api.GET("/error-wrap", errorWrappingDemo)

		// 文件下载示例
		api.GET("/download", fileDownloadDemo)

		// 用户设置路由
		settings := api.Group("/settings")
		{
			settings.GET("/:id", getUserSettings)

			// 应用 CSRF 中间件
			settings.Use(CSRFMiddleware())
			settings.POST("/:id", updateUserSettings)
		}

		// 评论路由
		api.POST("/comments", createComment)

		// CSRF 令牌路由
		api.GET("/csrf-token", CSRFMiddleware(), getCSRFToken)

		// panic 示例
		api.GET("/panic", simulatePanic)
	}

	// 需要认证的路由
	auth := r.Group("/api/auth")
	{
		auth.Use(AuthMiddleware())

		auth.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get(render.ContextKeyUserID)
			render.Success(c, gin.H{
				"user_id": userID,
				"message": "已认证的用户资料",
			})
		})
	}

	return r
}

// 辅助函数：返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 主函数
func main() {
	// 设置 Gin 模式
	gin.SetMode(gin.DebugMode)

	// 创建路由
	router := setupAdvancedRouter()

	// 模拟添加认证令牌
	tokenStore.Store("Bearer token123", 1)

	// 启动服务器
	log.Println("服务器启动在 http://localhost:8081")
	log.Println("高级用法示例:")
	log.Println("- 链式响应: http://localhost:8081/api/advanced/chain")
	log.Println("- 错误处理: http://localhost:8081/api/advanced/error")
	log.Println("- 错误包装: http://localhost:8081/api/advanced/error-wrap")
	log.Println("- 文件下载: http://localhost:8081/api/advanced/download")
	log.Println("- 用户设置: http://localhost:8081/api/advanced/settings/1")
	log.Println("- CSRF 令牌: http://localhost:8081/api/advanced/csrf-token")
	log.Println("- 认证 API: http://localhost:8081/api/auth/profile (需要 Authorization: Bearer token123 头)")
	log.Println("- 模拟 panic: http://localhost:8081/api/advanced/panic")

	router.Run(":8081")
}
