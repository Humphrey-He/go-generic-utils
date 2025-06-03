package security_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Humphrey-He/go-generic-utils/ginutil/security"

	"github.com/gin-gonic/gin"
)

// ExampleCSRF 展示如何使用 CSRF 保护中间件。
func ExampleCSRF() {
	// 创建 Gin 路由
	r := gin.New()

	// 添加 CSRF 保护中间件
	r.Use(security.CSRF())

	// 定义路由处理函数
	r.GET("/form", func(c *gin.Context) {
		// 获取 CSRF 令牌
		token := security.GetCSRFToken(c)

		// 渲染包含 CSRF 令牌的表单
		html := fmt.Sprintf(`
			<form action="/submit" method="POST">
				<input type="hidden" name="_csrf" value="%s">
				<input type="text" name="username">
				<button type="submit">Submit</button>
			</form>
		`, token)

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	})

	r.POST("/submit", func(c *gin.Context) {
		// CSRF 保护中间件已经验证了令牌，可以安全地处理请求
		username := c.PostForm("username")
		c.String(http.StatusOK, "表单提交成功，用户名: "+username)
	})

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/form", nil)
	r.ServeHTTP(w, req)

	fmt.Println("状态码:", w.Code)
	fmt.Println("包含 CSRF 令牌:", w.Body.String() != "" && w.Body.String() != "<nil>")

	// Output:
	// 状态码: 200
	// 包含 CSRF 令牌: true
}

// ExampleSecurityHeaders 展示如何使用安全头部中间件。
func ExampleSecurityHeaders() {
	// 创建 Gin 路由
	r := gin.New()

	// 添加安全头部中间件
	r.Use(security.SecurityHeaders())

	// 定义路由处理函数
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// 输出安全头部
	fmt.Println("Content-Security-Policy 头部存在:", w.Header().Get("Content-Security-Policy") != "")
	fmt.Println("X-Frame-Options 头部:", w.Header().Get("X-Frame-Options"))
	fmt.Println("X-Content-Type-Options 头部:", w.Header().Get("X-Content-Type-Options"))
	fmt.Println("X-XSS-Protection 头部:", w.Header().Get("X-XSS-Protection"))

	// Output:
	// Content-Security-Policy 头部存在: true
	// X-Frame-Options 头部: SAMEORIGIN
	// X-Content-Type-Options 头部: nosniff
	// X-XSS-Protection 头部: 1; mode=block
}

// ExampleCSPBuilder 展示如何使用 CSP 构建器。
func ExampleCSPBuilder() {
	// 创建 CSP 构建器
	builder := security.NewCSPBuilder()

	// 配置 CSP 策略
	builder.Set(security.CSPDefaultSrc, security.CSPSelf)
	builder.Set(security.CSPScriptSrc, security.CSPSelf, security.CSPUnsafeInline, "https://cdn.example.com")
	builder.Set(security.CSPStyleSrc, security.CSPSelf, security.CSPUnsafeInline, "https://cdn.example.com")
	builder.Set(security.CSPImgSrc, security.CSPSelf, security.CSPData, "https://img.example.com")
	builder.Set(security.CSPConnectSrc, security.CSPSelf, "https://api.example.com")
	builder.Set(security.CSPObjectSrc, security.CSPNone)
	builder.Set(security.CSPFrameSrc, security.CSPNone)

	// 构建 CSP 选项
	options := builder.Build()

	// 生成 CSP 头部值
	cspHeader := security.GenerateCSPHeader(options)

	fmt.Println("CSP 头部包含 default-src:", strings.Contains(cspHeader, "default-src 'self'"))
	fmt.Println("CSP 头部包含 script-src:", strings.Contains(cspHeader, "script-src 'self' 'unsafe-inline' https://cdn.example.com"))

	// 创建 Gin 路由
	r := gin.New()

	// 添加 CSP 中间件
	r.Use(builder.BuildMiddleware())

	// 定义路由处理函数
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	fmt.Println("CSP 头部设置成功:", w.Header().Get("Content-Security-Policy") != "")

	// Output:
	// CSP 头部包含 default-src: true
	// CSP 头部包含 script-src: true
	// CSP 头部设置成功: true
}

// ExampleSanitizeHTML 展示如何使用 HTML 清理功能。
func ExampleSanitizeHTML() {
	// 待清理的 HTML 字符串
	dirtyHTML := `<p>这是<strong>合法的</strong> HTML，但是 <script>alert('XSS攻击');</script> 是危险的。</p>
	<img src="x" onerror="alert('XSS');" />
	<a href="javascript:alert('XSS')">危险链接</a>`

	// 使用不同的清理策略
	strictHTML := security.SanitizeHTML(dirtyHTML, security.PolicyStrict)
	basicHTML := security.SanitizeHTML(dirtyHTML, security.PolicyBasic)
	relaxedHTML := security.SanitizeHTML(dirtyHTML, security.PolicyRelaxed)

	fmt.Println("严格模式是否移除了所有标签:", !strings.Contains(strictHTML, "<"))
	fmt.Println("基本模式是否保留了 <strong> 标签:", strings.Contains(basicHTML, "<strong>"))
	fmt.Println("基本模式是否移除了 <script> 标签:", !strings.Contains(basicHTML, "<script>"))
	fmt.Println("宽松模式是否移除了 JavaScript 协议:", !strings.Contains(relaxedHTML, "javascript:"))

	// Output:
	// 严格模式是否移除了所有标签: true
	// 基本模式是否保留了 <strong> 标签: true
	// 基本模式是否移除了 <script> 标签: true
	// 宽松模式是否移除了 JavaScript 协议: true
}

// ExampleSanitizeStruct 展示如何清理结构体。
func ExampleSanitizeStruct() {
	// 定义包含 HTML 的结构体
	type Comment struct {
		Author  string
		Content string
		Email   string
	}

	// 创建带有不安全内容的结构体实例
	comment := Comment{
		Author:  "<script>alert('XSS')</script>John",
		Content: "<p>This is a comment with <script>evil code</script></p>",
		Email:   "john@example.com<script>alert('XSS')</script>",
	}

	// 清理结构体
	security.SanitizeStruct(&comment, security.PolicyStrict)

	fmt.Println("作者字段是否被清理:", !strings.Contains(comment.Author, "<script>"))
	fmt.Println("内容字段是否被清理:", !strings.Contains(comment.Content, "<script>"))
	fmt.Println("邮箱字段是否被清理:", !strings.Contains(comment.Email, "<script>"))

	// Output:
	// 作者字段是否被清理: true
	// 内容字段是否被清理: true
	// 邮箱字段是否被清理: true
}

// ExampleWithCSPNonce 展示如何使用 CSP nonce。
func ExampleWithCSPNonce() {
	// 创建 Gin 路由
	r := gin.New()

	// 添加 CSP nonce 中间件
	r.Use(security.WithCSPNonce())

	// 定义路由处理函数
	r.GET("/", func(c *gin.Context) {
		// 生成带有 nonce 的脚本标签
		scriptTag := security.RenderScriptTag(c, "console.log('Hello, World!');")

		// 渲染带有 nonce 的 HTML
		html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>CSP Nonce Example</title>
			</head>
			<body>
				<h1>CSP Nonce Example</h1>
				%s
			</body>
			</html>
		`, scriptTag)

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	})

	// 创建请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	// 输出结果
	fmt.Println("CSP 头部包含 nonce:", strings.Contains(w.Header().Get("Content-Security-Policy"), "nonce-"))
	fmt.Println("响应体包含 nonce 脚本标签:", strings.Contains(w.Body.String(), "nonce="))

	// Output:
	// CSP 头部包含 nonce: true
	// 响应体包含 nonce 脚本标签: true
}

// ExampleSanitizeFilename 展示如何清理文件名。
func ExampleSanitizeFilename() {
	// 不安全的文件名
	unsafeFilename := "../../../etc/passwd"

	// 清理文件名
	safeFilename := security.SanitizeFilename(unsafeFilename)

	fmt.Println("清理后的文件名:", safeFilename)
	fmt.Println("是否移除了路径分隔符:", !strings.Contains(safeFilename, "/"))

	// Output:
	// 清理后的文件名: ______etc_passwd
	// 是否移除了路径分隔符: true
}

// ExampleValidateAndSanitize 展示如何验证并清理输入。
func ExampleValidateAndSanitize() {
	// 包含 XSS 的输入
	xssInput := "<script>alert('XSS')</script>Hello, World!"

	// 验证并清理输入
	sanitized, errMsg := security.ValidateAndSanitize(xssInput, 100, security.PolicyStrict)

	fmt.Println("是否成功清理:", errMsg == "")
	fmt.Println("清理后的输入:", sanitized)
	fmt.Println("是否移除了脚本标签:", !strings.Contains(sanitized, "<script>"))

	// 过长的输入
	longInput := strings.Repeat("A", 101)

	// 验证并清理输入
	sanitized, errMsg = security.ValidateAndSanitize(longInput, 100, security.PolicyStrict)

	fmt.Println("过长输入是否返回错误:", errMsg != "")

	// Output:
	// 是否成功清理: true
	// 清理后的输入: Hello, World!
	// 是否移除了脚本标签: true
	// 过长输入是否返回错误: true
}
