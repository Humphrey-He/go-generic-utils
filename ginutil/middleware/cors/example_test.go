package cors_test

import (
	"fmt"
	"ggu/ginutil/middleware/cors"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 本示例展示如何使用默认 CORS 配置
func Example_default() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用默认 CORS 配置
	r.Use(cors.New())

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Access-Control-Allow-Origin:", w.Header().Get("Access-Control-Allow-Origin"))

	// Output:
	// Status Code: 200
	// Access-Control-Allow-Origin: *
}

// 本示例展示如何处理预检请求 (OPTIONS)
func Example_preflight() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义 CORS 配置
	r.Use(cors.NewWithConfig(
		cors.WithAllowedOrigins([]string{"https://example.com"}),
		cors.WithAllowedMethods([]string{"GET", "POST", "PUT"}),
		cors.WithAllowedHeaders([]string{"Content-Type", "Authorization"}),
		cors.WithMaxAge(12*time.Hour),
	))

	// 添加一个 API 端点
	r.POST("/api/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "created"})
	})

	// 创建一个模拟预检请求
	req := httptest.NewRequest("OPTIONS", "/api/users", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Access-Control-Allow-Origin:", w.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("Access-Control-Allow-Methods:", w.Header().Get("Access-Control-Allow-Methods"))
	fmt.Println("Access-Control-Allow-Headers:", w.Header().Get("Access-Control-Allow-Headers"))
	fmt.Println("Access-Control-Max-Age:", w.Header().Get("Access-Control-Max-Age"))

	// Output:
	// Status Code: 204
	// Access-Control-Allow-Origin: https://example.com
	// Access-Control-Allow-Methods: GET, POST, PUT
	// Access-Control-Allow-Headers: Content-Type, Authorization
	// Access-Control-Max-Age: 43200
}

// 本示例展示如何使用动态来源验证
func Example_dynamicOrigin() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用动态来源验证
	r.Use(cors.NewWithConfig(
		cors.WithAllowOriginFunc(func(origin string) bool {
			// 允许所有以 example.com 结尾的来源
			return strings.HasSuffix(origin, "example.com")
		}),
		cors.WithAllowedMethods([]string{"GET"}),
	))

	// 添加一个 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个允许的来源请求
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	req1.Header.Set("Origin", "https://subdomain.example.com")
	w1 := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w1, req1)

	// 创建一个不允许的来源请求
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	req2.Header.Set("Origin", "https://malicious-site.com")
	w2 := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w2, req2)

	// 验证响应
	fmt.Println("允许的来源:")
	fmt.Println("Status Code:", w1.Code)
	fmt.Println("Access-Control-Allow-Origin:", w1.Header().Get("Access-Control-Allow-Origin"))

	fmt.Println("\n不允许的来源:")
	fmt.Println("Status Code:", w2.Code)
	fmt.Println("Access-Control-Allow-Origin:", w2.Header().Get("Access-Control-Allow-Origin"))

	// Output:
	// 允许的来源:
	// Status Code: 200
	// Access-Control-Allow-Origin: https://subdomain.example.com
	//
	// 不允许的来源:
	// Status Code: 200
	// Access-Control-Allow-Origin:
}

// 本示例展示如何配置凭证支持
func Example_credentials() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用支持凭证的 CORS 配置
	r.Use(cors.NewWithConfig(
		cors.WithAllowedOrigins([]string{"https://example.com"}),
		cors.WithAllowCredentials(true),
		cors.WithAllowedMethods([]string{"GET", "POST"}),
		cors.WithAllowedHeaders([]string{"Content-Type", "Authorization"}),
	))

	// 添加一个 API 端点
	r.GET("/api/user", func(c *gin.Context) {
		c.JSON(200, gin.H{"user": "张三"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/user", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Access-Control-Allow-Origin:", w.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("Access-Control-Allow-Credentials:", w.Header().Get("Access-Control-Allow-Credentials"))

	// Output:
	// Status Code: 200
	// Access-Control-Allow-Origin: https://example.com
	// Access-Control-Allow-Credentials: true
}

// 本示例展示如何配置暴露的头部
func Example_exposeHeaders() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义 CORS 配置，暴露特定头部
	r.Use(cors.NewWithConfig(
		cors.WithAllowedOrigins([]string{"*"}),
		cors.WithExposeHeaders([]string{"Content-Length", "X-Request-ID"}),
	))

	// 添加一个 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.Header("X-Request-ID", "12345")
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Access-Control-Allow-Origin:", w.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("Access-Control-Expose-Headers:", w.Header().Get("Access-Control-Expose-Headers"))
	fmt.Println("X-Request-ID:", w.Header().Get("X-Request-ID"))

	// Output:
	// Status Code: 200
	// Access-Control-Allow-Origin: *
	// Access-Control-Expose-Headers: Content-Length, X-Request-ID
	// X-Request-ID: 12345
}

// 本示例展示如何使用 AllowAll 配置（开发环境）
func Example_allowAll() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用 AllowAll 配置
	r.Use(cors.AllowAll())

	// 添加一个 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("X-Custom-Header", "value")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Access-Control-Allow-Origin:", w.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("Access-Control-Expose-Headers:", w.Header().Get("Access-Control-Expose-Headers"))

	// 创建一个模拟预检请求
	reqOptions := httptest.NewRequest("OPTIONS", "/api/data", nil)
	reqOptions.Header.Set("Origin", "https://example.com")
	reqOptions.Header.Set("Access-Control-Request-Method", "POST")
	reqOptions.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	wOptions := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(wOptions, reqOptions)

	// 验证预检响应
	fmt.Println("\n预检请求:")
	fmt.Println("Status Code:", wOptions.Code)
	fmt.Println("Access-Control-Allow-Origin:", wOptions.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("Access-Control-Allow-Headers:", wOptions.Header().Get("Access-Control-Allow-Headers"))

	// Output:
	// Status Code: 200
	// Access-Control-Allow-Origin: *
	// Access-Control-Expose-Headers: Content-Length, Content-Type
	//
	// 预检请求:
	// Status Code: 204
	// Access-Control-Allow-Origin: *
	// Access-Control-Allow-Headers: *
}
