package logger_test

import (
	"bytes"
	"fmt"
	"ggu/ginutil/middleware/logger"
	"net/http/httptest"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// 本示例展示如何使用默认日志配置
func Example_default() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用默认日志配置
	r.Use(logger.New())

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("User-Agent", "ExampleBot/1.0")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应内容类型:", w.Header().Get("Content-Type"))

	// Output:
	// 状态码: 200
	// 响应内容类型: application/json; charset=utf-8
}

// 本示例展示如何使用自定义日志配置
func Example_customConfig() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个缓冲区来捕获日志输出
	var buf bytes.Buffer

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义日志配置
	r.Use(logger.NewWithConfig(
		logger.WithOutput(&buf),
		logger.WithFormatter(&logger.JSONFormatter{}),
		logger.WithSkipPaths("/health", "/metrics"),
		logger.WithSkipPathRegexps(regexp.MustCompile(`^/static/.*`)),
		logger.WithTimeFormat(time.RFC3339),
		logger.WithUTC(true),
		logger.WithContextKeys("user_id", "trace_id"),
		logger.WithRequestHeader("Content-Type", "Accept"),
		logger.WithResponseHeader("Content-Type", "X-Request-ID"),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/user/:id", func(c *gin.Context) {
		// 设置上下文值
		c.Set("user_id", c.Param("id"))
		c.Set("trace_id", "trace-123456")

		// 设置响应头
		c.Header("X-Request-ID", "req-123456")

		// 返回响应
		c.JSON(200, gin.H{"id": c.Param("id"), "name": "张三"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/user/123", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ExampleBot/1.0")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("响应内容类型:", w.Header().Get("Content-Type"))
	fmt.Println("日志格式:", "JSON")
	fmt.Println("日志包含 user_id:", buf.String() != "" && bytes.Contains(buf.Bytes(), []byte(`"user_id":"123"`)))

	// Output:
	// 状态码: 200
	// 响应内容类型: application/json; charset=utf-8
	// 日志格式: JSON
	// 日志包含 user_id: true
}

// 本示例展示如何跳过特定路径的日志记录
func Example_skipPaths() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个缓冲区来捕获日志输出
	var buf bytes.Buffer

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义日志配置，跳过健康检查路径
	r.Use(logger.NewWithConfig(
		logger.WithOutput(&buf),
		logger.WithSkipPaths("/health"),
	))

	// 添加 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// 创建并处理 API 请求
	reqAPI := httptest.NewRequest("GET", "/api/data", nil)
	wAPI := httptest.NewRecorder()
	r.ServeHTTP(wAPI, reqAPI)

	// 创建并处理健康检查请求
	reqHealth := httptest.NewRequest("GET", "/health", nil)
	wHealth := httptest.NewRecorder()
	r.ServeHTTP(wHealth, reqHealth)

	// 检查两个请求的状态码
	fmt.Println("API 状态码:", wAPI.Code)
	fmt.Println("健康检查状态码:", wHealth.Code)

	// 检查日志缓冲区是否包含 API 路径但不包含健康检查路径
	logContainsAPI := bytes.Contains(buf.Bytes(), []byte("/api/data"))
	logContainsHealth := bytes.Contains(buf.Bytes(), []byte("/health"))

	fmt.Println("日志包含 API 路径:", logContainsAPI)
	fmt.Println("日志包含健康检查路径:", logContainsHealth)

	// Output:
	// API 状态码: 200
	// 健康检查状态码: 200
	// 日志包含 API 路径: true
	// 日志包含健康检查路径: false
}

// 本示例展示如何使用自定义日志函数
func Example_customLogFunc() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 自定义日志记录变量
	var (
		recordedPath     string
		recordedMethod   string
		recordedStatus   int
		recordedLatency  float64
		recordedClientIP string
	)

	// 使用自定义日志函数
	r.Use(logger.NewWithConfig(
		logger.WithLogFunc(func(entry *logger.LogEntry) {
			recordedPath = entry.Path
			recordedMethod = entry.Method
			recordedStatus = entry.StatusCode
			recordedLatency = entry.Latency
			recordedClientIP = entry.ClientIP

			// 在实际应用中，可以将日志发送到任何地方
			// 例如：将日志发送到 Elasticsearch、Logstash、Fluentd 等
		}),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("User-Agent", "ExampleBot/1.0")
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 输出记录的日志信息
	fmt.Println("记录的路径:", recordedPath)
	fmt.Println("记录的方法:", recordedMethod)
	fmt.Println("记录的状态码:", recordedStatus)
	fmt.Println("记录的客户端IP:", recordedClientIP)
	fmt.Println("是否记录了延迟时间:", recordedLatency > 0)

	// Output:
	// 记录的路径: /api/data
	// 记录的方法: GET
	// 记录的状态码: 200
	// 记录的客户端IP: 192.0.2.1:1234
	// 是否记录了延迟时间: true
}

// 本示例展示如何使用文本格式化器
func Example_textFormatter() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个缓冲区来捕获日志输出
	var buf bytes.Buffer

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义文本格式化器
	r.Use(logger.NewWithConfig(
		logger.WithOutput(&buf),
		logger.WithFormatter(&logger.TextFormatter{
			DisableColors: true, // 禁用颜色以便于测试输出比较
			TimeFormat:    "2006-01-02",
		}),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, world!"})
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 检查日志格式
	logContainsDate := bytes.Contains(buf.Bytes(), []byte(time.Now().Format("2006-01-02")))
	logContainsPath := bytes.Contains(buf.Bytes(), []byte("/api/data"))
	logContainsStatus := bytes.Contains(buf.Bytes(), []byte("200"))

	fmt.Println("日志包含日期:", logContainsDate)
	fmt.Println("日志包含路径:", logContainsPath)
	fmt.Println("日志包含状态码:", logContainsStatus)

	// Output:
	// 日志包含日期: true
	// 日志包含路径: true
	// 日志包含状态码: true
}
