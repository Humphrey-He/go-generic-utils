package timeout_test

import (
	"fmt"
	"ggu/ginutil/middleware/timeout"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
)

// 本示例展示如何使用默认超时配置
func Example_default() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用默认超时配置（30 秒）
	r.Use(timeout.New())

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		// 模拟处理时间为 100 毫秒
		time.Sleep(100 * time.Millisecond)
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应内容类型:", w.Header().Get("Content-Type"))

	// Output:
	// 状态码: 200
	// 响应内容类型: application/json; charset=utf-8
}

// 本示例展示如何使用自定义超时配置
func Example_customConfig() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义超时配置
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutMessage("自定义超时消息"),
		timeout.WithTimeoutCode(http.StatusServiceUnavailable),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/api/slow", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.JSON(200, gin.H{"message": "这个响应不应该被发送"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/slow", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应包含自定义消息:", w.Body.String() != "" && w.Body.String() != "{}")

	// Output:
	// 状态码: 503
	// 响应包含自定义消息: true
}

// 本示例展示如何使用自定义超时处理函数
func Example_customHandler() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义超时处理函数
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutHandler(func(c *gin.Context) {
			c.String(http.StatusServiceUnavailable, "自定义超时处理函数")
		}),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/api/slow", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.JSON(200, gin.H{"message": "这个响应不应该被发送"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/slow", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应内容:", w.Body.String())

	// Output:
	// 状态码: 503
	// 响应内容: 自定义超时处理函数
}

// 本示例展示如何使用自定义超时 JSON 响应
func Example_customJSON() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义超时 JSON 响应
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutJSON(gin.H{
			"error":   "请求超时",
			"code":    504,
			"timeout": true,
		}),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/api/slow", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.JSON(200, gin.H{"message": "这个响应不应该被发送"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/slow", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应包含错误信息:", w.Body.String() != "" && w.Body.String() != "{}")
	fmt.Println("响应内容类型:", w.Header().Get("Content-Type"))

	// Output:
	// 状态码: 504
	// 响应包含错误信息: true
	// 响应内容类型: application/json; charset=utf-8
}

// 本示例展示如何禁用超时功能
func Example_disableTimeout() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用超时中间件，但禁用超时功能
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithDisableTimeout(true),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/api/slow", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 由于禁用了超时功能，这个响应应该被发送
		c.String(200, "正常响应")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/slow", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应内容:", w.Body.String())

	// Output:
	// 状态码: 200
	// 响应内容: 正常响应
}

// 本示例展示如何使用辅助函数
func Example_helperFunctions() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用 TimeoutWithMessage 辅助函数
	r.GET("/message", timeout.TimeoutWithMessage(100*time.Millisecond, "消息超时"), func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		c.String(200, "这个响应不应该被发送")
	})

	// 使用 TimeoutWithJSON 辅助函数
	r.GET("/json", timeout.TimeoutWithJSON(100*time.Millisecond, gin.H{"error": "JSON 超时"}), func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		c.String(200, "这个响应不应该被发送")
	})

	// 测试 TimeoutWithMessage
	reqMessage := httptest.NewRequest("GET", "/message", nil)
	wMessage := httptest.NewRecorder()
	r.ServeHTTP(wMessage, reqMessage)

	// 测试 TimeoutWithJSON
	reqJSON := httptest.NewRequest("GET", "/json", nil)
	wJSON := httptest.NewRecorder()
	r.ServeHTTP(wJSON, reqJSON)

	fmt.Println("消息超时状态码:", wMessage.Code)
	fmt.Println("JSON 超时状态码:", wJSON.Code)
	fmt.Println("JSON 响应包含错误信息:", wJSON.Body.String() != "" && wJSON.Body.String() != "{}")

	// Output:
	// 消息超时状态码: 504
	// JSON 超时状态码: 504
	// JSON 响应包含错误信息: true
}
