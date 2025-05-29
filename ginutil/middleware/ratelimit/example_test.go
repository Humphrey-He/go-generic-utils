package ratelimit_test

import (
	"fmt"
	"ggu/ginutil/middleware/ratelimit"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
)

// 本示例展示如何使用默认限流配置
func Example_default() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用默认限流配置（每秒 10 个请求，突发 20 个）
	r.Use(ratelimit.New())

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应内容类型:", w.Header().Get("Content-Type"))
	fmt.Println("是否包含限流头部:", w.Header().Get("X-RateLimit-Limit") != "")

	// Output:
	// 状态码: 200
	// 响应内容类型: application/json; charset=utf-8
	// 是否包含限流头部: true
}

// 本示例展示如何使用自定义限流配置
func Example_customConfig() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义限流配置
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(2),
		ratelimit.WithBurst(2),
		ratelimit.WithTokensPerRequest(1),
		ratelimit.WithDisableHeaders(false),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建并处理第一个请求
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// 创建并处理第二个请求
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// 创建并处理第三个请求（超过限制）
	req3 := httptest.NewRequest("GET", "/api/data", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	fmt.Println("第一个请求状态码:", w1.Code)
	fmt.Println("第二个请求状态码:", w2.Code)
	fmt.Println("第三个请求状态码:", w3.Code)
	fmt.Println("第三个请求是否包含错误消息:", w3.Body.String() != "")

	// Output:
	// 第一个请求状态码: 200
	// 第二个请求状态码: 200
	// 第三个请求状态码: 429
	// 第三个请求是否包含错误消息: true
}

// 本示例展示如何使用基于 IP 的限流
func Example_ipBasedRateLimit() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用基于 IP 的限流
	r.Use(ratelimit.RateLimitPerIP(1, 1))

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建并处理来自 IP1 的请求
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// 创建并处理来自 IP1 的第二个请求（超过限制）
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// 创建并处理来自 IP2 的请求
	req3 := httptest.NewRequest("GET", "/api/data", nil)
	req3.RemoteAddr = "192.168.1.2:1234"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	fmt.Println("IP1 第一个请求状态码:", w1.Code)
	fmt.Println("IP1 第二个请求状态码:", w2.Code)
	fmt.Println("IP2 请求状态码:", w3.Code)

	// Output:
	// IP1 第一个请求状态码: 200
	// IP1 第二个请求状态码: 429
	// IP2 请求状态码: 200
}

// 本示例展示如何使用自定义错误处理函数
func Example_customErrorHandler() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义错误处理函数
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(1),
		ratelimit.WithBurst(1),
		ratelimit.WithErrorHandler(func(c *gin.Context, retryAfter time.Duration) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":      "请求频率超限",
				"retryAfter": retryAfter.String(),
				"code":       503,
			})
		}),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建并处理第一个请求
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// 创建并处理第二个请求（超过限制）
	req2 := httptest.NewRequest("GET", "/api/data", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	fmt.Println("第一个请求状态码:", w1.Code)
	fmt.Println("第二个请求状态码:", w2.Code)
	fmt.Println("第二个请求是否包含自定义错误:", w2.Body.String() != "" && w2.Body.String() != "{}")

	// Output:
	// 第一个请求状态码: 200
	// 第二个请求状态码: 503
	// 第二个请求是否包含自定义错误: true
}

// 本示例展示如何使用白名单跳过限流
func Example_whitelistSkipper() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用白名单跳过限流
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(1),
		ratelimit.WithBurst(1),
		ratelimit.WithSkipper(ratelimit.WhitelistSkipper("192.168.1.100")),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建并处理来自白名单 IP 的多个请求
	req1 := httptest.NewRequest("GET", "/api/data", nil)
	req1.RemoteAddr = "192.168.1.100:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/api/data", nil)
	req2.RemoteAddr = "192.168.1.100:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// 创建并处理来自非白名单 IP 的多个请求
	req3 := httptest.NewRequest("GET", "/api/data", nil)
	req3.RemoteAddr = "192.168.1.200:1234"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	req4 := httptest.NewRequest("GET", "/api/data", nil)
	req4.RemoteAddr = "192.168.1.200:1234"
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)

	fmt.Println("白名单 IP 第一个请求状态码:", w1.Code)
	fmt.Println("白名单 IP 第二个请求状态码:", w2.Code)
	fmt.Println("非白名单 IP 第一个请求状态码:", w3.Code)
	fmt.Println("非白名单 IP 第二个请求状态码:", w4.Code)

	// Output:
	// 白名单 IP 第一个请求状态码: 200
	// 白名单 IP 第二个请求状态码: 200
	// 非白名单 IP 第一个请求状态码: 200
	// 非白名单 IP 第二个请求状态码: 429
}

// 本示例展示如何使用辅助函数
func Example_helperFunctions() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 为不同的路由使用不同的限流策略
	r.GET("/api/low", ratelimit.RateLimit(1, 1), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "低频率 API"})
	})

	r.GET("/api/high", ratelimit.RateLimit(10, 10), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "高频率 API"})
	})

	// 测试低频率 API
	req1 := httptest.NewRequest("GET", "/api/low", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/api/low", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// 测试高频率 API
	req3 := httptest.NewRequest("GET", "/api/high", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	req4 := httptest.NewRequest("GET", "/api/high", nil)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)

	fmt.Println("低频率 API 第一个请求状态码:", w1.Code)
	fmt.Println("低频率 API 第二个请求状态码:", w2.Code)
	fmt.Println("高频率 API 第一个请求状态码:", w3.Code)
	fmt.Println("高频率 API 第二个请求状态码:", w4.Code)

	// Output:
	// 低频率 API 第一个请求状态码: 200
	// 低频率 API 第二个请求状态码: 429
	// 高频率 API 第一个请求状态码: 200
	// 高频率 API 第二个请求状态码: 200
}
