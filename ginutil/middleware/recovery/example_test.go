package recovery_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/noobtrump/go-generic-utils/ginutil/middleware/recovery"

	"github.com/gin-gonic/gin"
)

// 本示例展示如何使用默认恢复配置
func Example_default() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用默认恢复配置
	r.Use(recovery.New())

	// 添加一个会 panic 的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("测试 panic")
	})

	// 添加一个正常的路由
	r.GET("/normal", func(c *gin.Context) {
		c.String(200, "正常响应")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("状态码:", w.Code)
	fmt.Println("响应类型:", w.Header().Get("Content-Type"))
	fmt.Println("响应包含错误信息:", strings.Contains(w.Body.String(), "测试 panic"))

	// 测试正常路由
	reqNormal := httptest.NewRequest("GET", "/normal", nil)
	wNormal := httptest.NewRecorder()
	r.ServeHTTP(wNormal, reqNormal)

	fmt.Println("正常路由状态码:", wNormal.Code)
	fmt.Println("正常路由响应:", wNormal.Body.String())

	// Output:
	// 状态码: 500
	// 响应类型: application/json; charset=utf-8
	// 响应包含错误信息: true
	// 正常路由状态码: 200
	// 正常路由响应: 正常响应
}

// 本示例展示如何使用自定义错误处理函数
func Example_customErrorHandler() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义错误处理函数
	r.Use(recovery.NewWithConfig(
		recovery.WithErrorHandler(func(c *gin.Context, err interface{}) {
			c.String(500, fmt.Sprintf("发生错误: %v", err))
		}),
	))

	// 添加一个会 panic 的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("自定义错误处理")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("状态码:", w.Code)
	fmt.Println("响应类型:", w.Header().Get("Content-Type"))
	fmt.Println("响应内容:", w.Body.String())

	// Output:
	// 状态码: 500
	// 响应类型: text/plain; charset=utf-8
	// 响应内容: 发生错误: 自定义错误处理
}

// 本示例展示如何使用 JSON 错误处理函数
func Example_jsonErrorHandler() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用 JSON 错误处理函数
	r.Use(recovery.NewWithConfig(
		recovery.WithErrorHandler(recovery.JSONErrorHandler(500, "服务器内部错误")),
	))

	// 添加一个会 panic 的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("JSON 错误处理")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	fmt.Println("状态码:", w.Code)
	fmt.Println("响应类型:", w.Header().Get("Content-Type"))
	fmt.Println("响应包含 code:", strings.Contains(w.Body.String(), `"code":500`))
	fmt.Println("响应包含错误信息:", strings.Contains(w.Body.String(), "JSON 错误处理"))

	// Output:
	// 状态码: 500
	// 响应类型: application/json; charset=utf-8
	// 响应包含 code: true
	// 响应包含错误信息: true
}

// 本示例展示如何自定义堆栈跟踪配置
func Example_customStackConfig() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个缓冲区来捕获日志输出
	var buf bytes.Buffer

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义堆栈跟踪配置
	r.Use(recovery.NewWithConfig(
		recovery.WithOutput(&buf),
		recovery.WithStackAll(true),
		recovery.WithStackSize(8<<10), // 8 KB
	))

	// 添加一个会 panic 的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("堆栈跟踪测试")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证日志输出
	logOutput := buf.String()
	fmt.Println("日志包含 panic 信息:", strings.Contains(logOutput, "堆栈跟踪测试"))
	fmt.Println("日志包含堆栈跟踪:", strings.Contains(logOutput, "goroutine"))
	fmt.Println("日志包含请求路径:", strings.Contains(logOutput, "/panic"))

	// Output:
	// 日志包含 panic 信息: true
	// 日志包含堆栈跟踪: true
	// 日志包含请求路径: true
}

// 本示例展示如何禁用打印堆栈信息
func Example_disablePrintStack() {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个缓冲区来捕获日志输出
	var buf bytes.Buffer

	// 创建 Gin 路由
	r := gin.New()

	// 禁用打印堆栈信息
	r.Use(recovery.NewWithConfig(
		recovery.WithOutput(&buf),
		recovery.WithDisablePrintStack(true),
	))

	// 添加一个会 panic 的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("禁用堆栈打印")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证日志输出
	logOutput := buf.String()
	fmt.Println("日志是否为空:", len(logOutput) == 0)

	// Output:
	// 日志是否为空: true
}
