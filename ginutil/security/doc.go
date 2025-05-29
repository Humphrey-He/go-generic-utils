/*
Package security 提供 Gin 框架的各种安全增强功能。

# 概述

security 包为 Gin 应用程序提供了全面的安全增强功能，包括 CSRF 防护、XSS 防护和输入清理。
这些功能旨在帮助开发者构建更安全的 Web 应用程序，防御常见的安全威胁。

# 主要功能

## CSRF 防护

CSRF（跨站请求伪造）防护使用双重提交 Cookie 模式，适合 API 和现代 Web 应用程序。

主要功能：
- 为每个会话生成高熵的 CSRF 令牌
- 将令牌存储在安全的 Cookie 中
- 在非安全方法的请求中验证令牌
- 支持从请求头、表单字段、查询参数或 Cookie 中获取令牌
- 可配置的错误处理
- 可排除特定路径或请求
- 生成包含 CSRF 令牌的表单字段或 meta 标签

使用示例：

	// 添加 CSRF 保护中间件
	r := gin.New()
	r.Use(security.CSRF())

	// 获取并在表单中使用 CSRF 令牌
	r.GET("/form", func(c *gin.Context) {
		token := security.GetCSRFToken(c)
		html := fmt.Sprintf(`<form method="POST">
			<input type="hidden" name="_csrf" value="%s">
			...
		</form>`, token)
		c.HTML(200, "form.html", gin.H{"csrf_field": security.RenderCSRFField(c)})
	})

## XSS 防护

XSS（跨站脚本）防护提供了内容安全策略（CSP）和其他安全响应头部的设置，以及输出编码辅助函数。

主要功能：
- 设置内容安全策略（CSP）头部
- 设置其他安全响应头部（X-Frame-Options, X-Content-Type-Options, X-XSS-Protection 等）
- 提供 HTML 和 JavaScript 的输出编码辅助函数
- 支持 CSP nonce 生成和管理
- 便捷的 CSP 构建器 API

使用示例：

	// 添加安全头部中间件
	r := gin.New()
	r.Use(security.SecurityHeaders())

	// 使用 CSP 构建器
	builder := security.NewCSPBuilder()
	builder.Set(security.CSPDefaultSrc, security.CSPSelf)
	builder.Set(security.CSPScriptSrc, security.CSPSelf, "https://cdn.example.com")
	r.Use(builder.BuildMiddleware())

	// 使用 CSP nonce
	r.Use(security.WithCSPNonce())
	r.GET("/", func(c *gin.Context) {
		script := security.RenderScriptTag(c, "console.log('Hello');")
		c.HTML(200, "index.html", gin.H{"script": script})
	})

## 输入清理

输入清理功能提供了一组工具，用于清理和验证用户输入，防止 XSS 和注入攻击。

主要功能：
- 根据不同策略清理 HTML 内容（严格、基本、宽松）
- 清理和规范化纯文本字符串
- 移除控制字符和多余空白
- 清理结构体中的字符串字段
- 清理 JSON 字符串、文件名、查询字符串和搜索词
- 验证电子邮件地址、URL 和电话号码

使用示例：

	// 清理 HTML 内容
	safeHTML := security.SanitizeHTML(unsafeHTML, security.PolicyStrict)

	// 清理结构体
	comment := Comment{...}
	security.SanitizeStruct(&comment, security.PolicyBasic)

	// 验证并清理输入
	safeInput, errMsg := security.ValidateAndSanitize(userInput, 1000, security.PolicyStrict)
	if errMsg != "" {
		// 处理错误
	}

# 组合使用

这些安全功能可以组合使用，提供全面的安全保护：

	r := gin.New()

	// 添加各种安全中间件
	r.Use(security.CSRF())
	r.Use(security.SecurityHeaders())
	r.Use(security.WithCSPNonce())

	r.POST("/comment", func(c *gin.Context) {
		var comment Comment
		if err := c.ShouldBind(&comment); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 清理用户输入
		security.SanitizeStruct(&comment, security.PolicyBasic)

		// 处理评论...

		c.JSON(200, gin.H{"message": "评论已发布"})
	})

# 安全最佳实践

除了使用本包提供的功能外，还建议遵循以下安全最佳实践：

1. 使用 HTTPS 传输所有流量
2. 实施正确的身份验证和授权机制
3. 遵循最小权限原则
4. 定期更新依赖项
5. 对敏感操作实施速率限制
6. 记录安全相关事件
7. 定期进行安全审计和渗透测试
*/
package security
