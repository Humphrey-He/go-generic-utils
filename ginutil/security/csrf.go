// Package security 提供 Gin 框架的各种安全增强功能。
// 包括 CSRF 防护、XSS 防护和输入清理等功能。
package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// ErrTokenInvalid 表示 CSRF 令牌无效。
	ErrTokenInvalid = errors.New("CSRF 令牌无效")

	// ErrTokenNotFound 表示在请求中找不到 CSRF 令牌。
	ErrTokenNotFound = errors.New("CSRF 令牌未在请求中找到")

	// ErrCookieNotFound 表示在请求中找不到 CSRF Cookie。
	ErrCookieNotFound = errors.New("CSRF Cookie 未在请求中找到")

	// 安全的 HTTP 方法不需要 CSRF 保护
	safeMethods = map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"OPTIONS": true,
		"TRACE":   true,
	}
)

// CSRFConfig 是 CSRF 保护中间件的配置选项。
type CSRFConfig struct {
	// TokenLength 是 CSRF 令牌的长度（以字节为单位）。
	TokenLength int

	// TokenLookup 定义从请求中查找令牌的方式，格式为 "source:name"。
	// 例如：
	// - "header:X-CSRF-Token"
	// - "form:_csrf"
	// - "query:_csrf"
	// - "cookie:_csrf"
	// 多个源可以用逗号分隔：
	// - "header:X-CSRF-Token,form:_csrf,query:_csrf"
	TokenLookup string

	// CookieName 是存储 CSRF 令牌的 Cookie 名称。
	CookieName string

	// CookieDomain 设置 Cookie 的 Domain 属性。
	CookieDomain string

	// CookiePath 设置 Cookie 的 Path 属性。
	CookiePath string

	// CookieMaxAge 设置 Cookie 的 Max-Age 属性（以秒为单位）。
	CookieMaxAge int

	// CookieSecure 设置 Cookie 的 Secure 属性。
	CookieSecure bool

	// CookieHTTPOnly 设置 Cookie 的 HttpOnly 属性。
	// 注意：如果启用 HttpOnly，JavaScript 将无法读取 CSRF Cookie。
	// 如果使用 JS 框架将令牌作为请求头发送，应将此设置为 false。
	CookieHTTPOnly bool

	// CookieSameSite 设置 Cookie 的 SameSite 属性。
	CookieSameSite http.SameSite

	// ErrorFunc 是发生 CSRF 错误时调用的函数。
	// 如果为 nil，将使用默认错误处理函数。
	ErrorFunc CSRFErrorFunc

	// ExcludedPaths 是不需要 CSRF 保护的路径列表。
	ExcludedPaths []string

	// IgnoreFunc 用于自定义忽略某些请求的逻辑。
	// 如果返回 true，将不会对请求进行 CSRF 保护。
	IgnoreFunc func(c *gin.Context) bool

	// SuccessHandler 是 CSRF 验证成功后调用的函数。
	// 可用于添加自定义响应头或日志记录。
	SuccessHandler func(c *gin.Context)

	// TokenGetter 是从请求中获取令牌的自定义函数。
	// 如果设置了此选项，将忽略 TokenLookup。
	TokenGetter func(c *gin.Context) (string, error)

	// TokenContextKey 是存储 CSRF 令牌的上下文键名。
	TokenContextKey string

	// SessionKey 是用于签名和验证令牌的密钥。
	// 如果为空，将为每个请求生成一个新令牌。
	SessionKey []byte
}

// CSRFErrorFunc 是 CSRF 错误处理函数的类型定义。
type CSRFErrorFunc func(c *gin.Context, err error)

// DefaultCSRFConfig 返回默认的 CSRF 配置。
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:     32,
		TokenLookup:     "header:X-CSRF-Token,form:_csrf",
		CookieName:      "csrf_token",
		CookiePath:      "/",
		CookieMaxAge:    86400, // 1 天
		CookieSecure:    true,
		CookieHTTPOnly:  false, // 默认允许 JavaScript 读取
		CookieSameSite:  http.SameSiteLaxMode,
		TokenContextKey: "csrf_token",
	}
}

// 令牌源类型
const (
	tokenSourceHeader = "header"
	tokenSourceForm   = "form"
	tokenSourceQuery  = "query"
	tokenSourceCookie = "cookie"
)

type tokenExtractor struct {
	source string
	name   string
}

// CSRF 返回一个 Gin 中间件，用于 CSRF 保护。
func CSRF(config ...CSRFConfig) gin.HandlerFunc {
	// 使用默认配置
	cfg := DefaultCSRFConfig()

	// 应用自定义配置
	if len(config) > 0 {
		cfg = config[0]
	}

	// 设置默认值
	if cfg.TokenLength == 0 {
		cfg.TokenLength = 32
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "csrf_token"
	}
	if cfg.CookiePath == "" {
		cfg.CookiePath = "/"
	}
	if cfg.TokenContextKey == "" {
		cfg.TokenContextKey = "csrf_token"
	}
	if cfg.ErrorFunc == nil {
		cfg.ErrorFunc = defaultCSRFErrorHandler
	}

	// 解析令牌查找设置
	extractors := parseTokenLookup(cfg.TokenLookup)

	// 返回中间件处理函数
	return func(c *gin.Context) {
		// 检查是否跳过 CSRF 保护
		if shouldIgnoreCSRF(c, cfg) {
			c.Next()
			return
		}

		// 从请求中获取 CSRF 令牌
		var requestToken string
		var err error

		if cfg.TokenGetter != nil {
			// 使用自定义令牌获取函数
			requestToken, err = cfg.TokenGetter(c)
		} else {
			// 使用配置的提取器获取令牌
			requestToken, err = extractToken(c, extractors)
		}

		// 获取或生成 CSRF Cookie
		cookieToken, issued := getCSRFCookie(c, cfg)

		// 安全方法或新会话可以跳过验证
		if safeMethods[c.Request.Method] || issued {
			// 将令牌存储到上下文中，以便视图可以访问
			c.Set(cfg.TokenContextKey, cookieToken)
			c.Next()
			return
		}

		// 对于非安全方法，验证 CSRF 令牌
		if err != nil {
			cfg.ErrorFunc(c, err)
			return
		}

		if requestToken == "" {
			cfg.ErrorFunc(c, ErrTokenNotFound)
			return
		}

		if requestToken != cookieToken {
			cfg.ErrorFunc(c, ErrTokenInvalid)
			return
		}

		// CSRF 验证成功
		if cfg.SuccessHandler != nil {
			cfg.SuccessHandler(c)
		}

		// 将令牌存储到上下文中，以便视图可以访问
		c.Set(cfg.TokenContextKey, cookieToken)
		c.Next()
	}
}

// parseTokenLookup 解析 TokenLookup 字符串。
func parseTokenLookup(lookup string) []tokenExtractor {
	if lookup == "" {
		return []tokenExtractor{{source: tokenSourceHeader, name: "X-CSRF-Token"}}
	}

	sources := strings.Split(lookup, ",")
	extractors := make([]tokenExtractor, 0, len(sources))

	for _, source := range sources {
		parts := strings.Split(strings.TrimSpace(source), ":")
		if len(parts) != 2 {
			continue
		}

		extractors = append(extractors, tokenExtractor{
			source: strings.ToLower(parts[0]),
			name:   parts[1],
		})
	}

	return extractors
}

// extractToken 从请求中提取 CSRF 令牌。
func extractToken(c *gin.Context, extractors []tokenExtractor) (string, error) {
	for _, extractor := range extractors {
		switch extractor.source {
		case tokenSourceHeader:
			token := c.GetHeader(extractor.name)
			if token != "" {
				return token, nil
			}
		case tokenSourceForm:
			token := c.PostForm(extractor.name)
			if token != "" {
				return token, nil
			}
		case tokenSourceQuery:
			token := c.Query(extractor.name)
			if token != "" {
				return token, nil
			}
		case tokenSourceCookie:
			cookie, err := c.Cookie(extractor.name)
			if err == nil && cookie != "" {
				return cookie, nil
			}
		}
	}

	return "", ErrTokenNotFound
}

// getCSRFCookie 获取或生成 CSRF Cookie。
func getCSRFCookie(c *gin.Context, cfg CSRFConfig) (string, bool) {
	// 尝试从请求中获取现有的 CSRF Cookie
	cookie, err := c.Cookie(cfg.CookieName)
	if err == nil && cookie != "" {
		return cookie, false
	}

	// 生成新的 CSRF 令牌
	token := generateCSRFToken(cfg.TokenLength)

	// 设置 CSRF Cookie
	c.SetSameSite(cfg.CookieSameSite)
	c.SetCookie(
		cfg.CookieName,
		token,
		cfg.CookieMaxAge,
		cfg.CookiePath,
		cfg.CookieDomain,
		cfg.CookieSecure,
		cfg.CookieHTTPOnly,
	)

	return token, true
}

// generateCSRFToken 生成一个随机的 CSRF 令牌。
func generateCSRFToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机生成失败，使用时间戳作为备用
		// 这不是加密安全的，但总比没有好
		now := time.Now().UnixNano()
		for i := 0; i < length; i++ {
			b[i] = byte(now & 0xff)
			now >>= 8
		}
	}
	return base64.StdEncoding.EncodeToString(b)
}

// shouldIgnoreCSRF 检查是否应该忽略 CSRF 保护。
func shouldIgnoreCSRF(c *gin.Context, cfg CSRFConfig) bool {
	// 检查是否为安全方法
	if safeMethods[c.Request.Method] {
		return true
	}

	// 检查自定义忽略函数
	if cfg.IgnoreFunc != nil && cfg.IgnoreFunc(c) {
		return true
	}

	// 检查排除路径
	path := c.Request.URL.Path
	for _, excludedPath := range cfg.ExcludedPaths {
		if path == excludedPath || (strings.HasSuffix(excludedPath, "*") && strings.HasPrefix(path, excludedPath[:len(excludedPath)-1])) {
			return true
		}
	}

	return false
}

// defaultCSRFErrorHandler 是默认的 CSRF 错误处理函数。
func defaultCSRFErrorHandler(c *gin.Context, err error) {
	status := http.StatusForbidden
	errorMsg := "CSRF 验证失败"

	if errors.Is(err, ErrTokenNotFound) {
		errorMsg = "CSRF 令牌未在请求中找到"
	} else if errors.Is(err, ErrTokenInvalid) {
		errorMsg = "CSRF 令牌无效"
	} else if errors.Is(err, ErrCookieNotFound) {
		errorMsg = "CSRF Cookie 未在请求中找到"
	}

	c.AbortWithStatusJSON(status, gin.H{
		"error":   errorMsg,
		"message": "请刷新页面并重试",
		"code":    status,
	})
}

// GetCSRFToken 从 Gin 上下文中获取 CSRF 令牌。
func GetCSRFToken(c *gin.Context) string {
	v, exists := c.Get("csrf_token")
	if !exists {
		return ""
	}
	if token, ok := v.(string); ok {
		return token
	}
	return ""
}

// CSRFErrorHandlerFunc 创建一个自定义的 CSRF 错误处理函数。
func CSRFErrorHandlerFunc(handler func(c *gin.Context, err error)) CSRFErrorFunc {
	return func(c *gin.Context, err error) {
		handler(c, err)
	}
}

// WithCSRFConfig 创建一个带有自定义配置的 CSRF 中间件。
func WithCSRFConfig(cfg CSRFConfig) gin.HandlerFunc {
	return CSRF(cfg)
}

// CSRFWithIgnorePaths 创建一个 CSRF 中间件，指定要忽略的路径。
func CSRFWithIgnorePaths(paths ...string) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	cfg.ExcludedPaths = paths
	return CSRF(cfg)
}

// CSRFWithErrorHandler 创建一个 CSRF 中间件，使用自定义错误处理函数。
func CSRFWithErrorHandler(handler func(c *gin.Context, err error)) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	cfg.ErrorFunc = handler
	return CSRF(cfg)
}

// CSRFWithTokenLength 创建一个 CSRF 中间件，指定令牌长度。
func CSRFWithTokenLength(length int) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	cfg.TokenLength = length
	return CSRF(cfg)
}

// CSRFWithTokenSource 创建一个 CSRF 中间件，指定令牌源。
func CSRFWithTokenSource(source string) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	cfg.TokenLookup = source
	return CSRF(cfg)
}

// CSRFWithCookieOptions 创建一个 CSRF 中间件，指定 Cookie 选项。
func CSRFWithCookieOptions(name string, maxAge int, httpOnly, secure bool, sameSite http.SameSite) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	cfg.CookieName = name
	cfg.CookieMaxAge = maxAge
	cfg.CookieHTTPOnly = httpOnly
	cfg.CookieSecure = secure
	cfg.CookieSameSite = sameSite
	return CSRF(cfg)
}

// RenderCSRFField 生成一个包含 CSRF 令牌的隐藏表单字段。
func RenderCSRFField(c *gin.Context) string {
	token := GetCSRFToken(c)
	if token == "" {
		return ""
	}
	return fmt.Sprintf(`<input type="hidden" name="_csrf" value="%s">`, token)
}

// CSRFMetaTag 生成一个包含 CSRF 令牌的 meta 标签。
func CSRFMetaTag(c *gin.Context) string {
	token := GetCSRFToken(c)
	if token == "" {
		return ""
	}
	return fmt.Sprintf(`<meta name="csrf-token" content="%s">`, token)
}
