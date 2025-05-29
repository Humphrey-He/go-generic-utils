package render

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// HTML 渲染 HTML 模板
// 参数:
//   - c: Gin 上下文
//   - status: HTTP 状态码
//   - templateName: 模板名称
//   - data: 传递给模板的数据
func HTML(c *gin.Context, status int, templateName string, data gin.H) {
	// 应用 HTML 辅助函数，注入通用数据
	enhancedData := applyHTMLHelpers(c, data)

	// 渲染模板
	c.HTML(status, templateName, enhancedData)
}

// HTMLSuccessPage 渲染成功页面
// 参数:
//   - c: Gin 上下文
//   - templateName: 模板名称
//   - data: 传递给模板的数据
func HTMLSuccessPage(c *gin.Context, templateName string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}

	// 添加成功状态信息
	data["success"] = true
	data["code"] = CodeSuccess
	data["message"] = DefaultSuccessMessage

	// 应用 HTML 辅助函数，注入通用数据
	enhancedData := applyHTMLHelpers(c, data)

	// 渲染模板
	c.HTML(http.StatusOK, templateName, enhancedData)
}

// HTMLErrorPage 渲染错误页面
// 参数:
//   - c: Gin 上下文
//   - status: HTTP 状态码
//   - businessCode: 业务错误码
//   - message: 错误消息
//   - data: 可选的额外数据
func HTMLErrorPage(c *gin.Context, status int, businessCode int, message string, data ...gin.H) {
	var templateData gin.H

	if len(data) > 0 && data[0] != nil {
		templateData = data[0]
	} else {
		templateData = gin.H{}
	}

	if message == "" {
		message = GetDefaultMessage(businessCode)
	}

	// 添加错误状态信息
	templateData["success"] = false
	templateData["code"] = businessCode
	templateData["message"] = message
	templateData["error"] = message
	templateData["status"] = status

	// 应用 HTML 辅助函数，注入通用数据
	enhancedData := applyHTMLHelpers(c, templateData)

	// 渲染错误模板
	c.HTML(status, DefaultHTMLErrorTemplate, enhancedData)

	// 终止处理链
	c.Abort()
}

// BadRequestPage 渲染参数错误页面
// 参数:
//   - c: Gin 上下文
//   - message: 错误消息
//   - data: 可选的额外数据
func BadRequestPage(c *gin.Context, message string, data ...gin.H) {
	if message == "" {
		message = GetDefaultMessage(CodeInvalidParams)
	}

	HTMLErrorPage(c, http.StatusBadRequest, CodeInvalidParams, message, data...)
}

// NotFoundPage 渲染资源不存在页面
// 参数:
//   - c: Gin 上下文
//   - message: 错误消息
//   - data: 可选的额外数据
func NotFoundPage(c *gin.Context, message string, data ...gin.H) {
	if message == "" {
		message = GetDefaultMessage(CodeNotFound)
	}

	HTMLErrorPage(c, http.StatusNotFound, CodeNotFound, message, data...)
}

// InternalErrorPage 渲染内部错误页面
// 参数:
//   - c: Gin 上下文
//   - message: 错误消息
//   - data: 可选的额外数据
func InternalErrorPage(c *gin.Context, message string, data ...gin.H) {
	if message == "" {
		message = GetDefaultMessage(CodeInternalError)
	}

	HTMLErrorPage(c, http.StatusInternalServerError, CodeInternalError, message, data...)
}

// UnauthorizedPage 渲染未授权页面
// 参数:
//   - c: Gin 上下文
//   - message: 错误消息
//   - data: 可选的额外数据
func UnauthorizedPage(c *gin.Context, message string, data ...gin.H) {
	if message == "" {
		message = GetDefaultMessage(CodeUnauthorized)
	}

	HTMLErrorPage(c, http.StatusUnauthorized, CodeUnauthorized, message, data...)
}

// ForbiddenPage 渲染禁止访问页面
// 参数:
//   - c: Gin 上下文
//   - message: 错误消息
//   - data: 可选的额外数据
func ForbiddenPage(c *gin.Context, message string, data ...gin.H) {
	if message == "" {
		message = GetDefaultMessage(CodeForbidden)
	}

	HTMLErrorPage(c, http.StatusForbidden, CodeForbidden, message, data...)
}

// Redirect 执行重定向
// 参数:
//   - c: Gin 上下文
//   - status: HTTP 状态码
//   - location: 重定向目标 URL
func Redirect(c *gin.Context, status int, location string) {
	c.Redirect(status, location)
}

// RedirectPermanent 执行永久重定向 (301)
// 参数:
//   - c: Gin 上下文
//   - location: 重定向目标 URL
func RedirectPermanent(c *gin.Context, location string) {
	c.Redirect(http.StatusMovedPermanently, location)
}

// RedirectTemporary 执行临时重定向 (302)
// 参数:
//   - c: Gin 上下文
//   - location: 重定向目标 URL
func RedirectTemporary(c *gin.Context, location string) {
	c.Redirect(http.StatusFound, location)
}

// RedirectWithFlash 执行重定向并设置 Flash 消息
// 参数:
//   - c: Gin 上下文
//   - status: HTTP 状态码
//   - location: 重定向目标 URL
//   - flashKey: Flash 消息的键
//   - flashValue: Flash 消息的值
func RedirectWithFlash(c *gin.Context, status int, location, flashKey string, flashValue interface{}) {
	// 设置 Flash 消息
	c.Set(flashKey, flashValue)

	// 重定向
	c.Redirect(status, location)
}

// RegisterTemplates 注册 HTML 模板
// 参数:
//   - engine: Gin 引擎
//   - pattern: 模板文件路径模式，例如 "templates/*"
func RegisterTemplates(engine *gin.Engine, pattern ...string) {
	templatePattern := HTMLTemplateDir
	if len(pattern) > 0 && pattern[0] != "" {
		templatePattern = pattern[0]
	}

	// 注册模板
	engine.LoadHTMLGlob(templatePattern)
}

// RegisterTemplatesFromFiles 从指定文件注册 HTML 模板
// 参数:
//   - engine: Gin 引擎
//   - files: 模板文件路径列表
func RegisterTemplatesFromFiles(engine *gin.Engine, files ...string) {
	// 注册模板
	engine.LoadHTMLFiles(files...)
}

// SetTemplateDelimiters 设置模板定界符
// 参数:
//   - engine: Gin 引擎
//   - left: 左定界符
//   - right: 右定界符
func SetTemplateDelimiters(engine *gin.Engine, left, right string) {
	engine.Delims(left, right)
}

// RegisterGlobTemplates 递归注册目录中的所有模板
// 参数:
//   - engine: Gin 引擎
//   - dir: 模板目录
func RegisterGlobTemplates(engine *gin.Engine, dir string) error {
	// 递归查找所有模板文件
	pattern := filepath.Join(dir, "**", "*.html")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// 注册模板
	engine.LoadHTMLFiles(files...)
	return nil
}

// AddUserDataHelper 添加用户数据注入辅助函数
func AddUserDataHelper() {
	RegisterHTMLHelper(func(c *gin.Context, data gin.H) gin.H {
		result := gin.H{}

		// 检查上下文中是否存在用户信息
		if userID, exists := c.Get(ContextKeyUserID); exists {
			result["user_id"] = userID
		}

		return result
	})
}

// AddCSRFHelper 添加 CSRF 令牌注入辅助函数
func AddCSRFHelper() {
	RegisterHTMLHelper(func(c *gin.Context, data gin.H) gin.H {
		result := gin.H{}

		// 检查上下文中是否存在 CSRF 令牌
		if csrfToken, exists := c.Get("csrf_token"); exists {
			result["csrf_token"] = csrfToken
		}

		return result
	})
}

// AddFlashMessageHelper 添加 Flash 消息注入辅助函数
func AddFlashMessageHelper() {
	RegisterHTMLHelper(func(c *gin.Context, data gin.H) gin.H {
		result := gin.H{}

		// 检查上下文中是否存在 Flash 消息
		if flashMsg, exists := c.Get("flash_message"); exists {
			result["flash_message"] = flashMsg
		}

		return result
	})
}
