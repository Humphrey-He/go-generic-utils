// Package cors 提供了一个用于 Gin 框架的 CORS (跨域资源共享) 中间件。
//
// 该包允许开发者轻松配置 CORS 策略，包括允许的来源、方法、头部等，
// 并自动处理预检请求 (OPTIONS)。
package cors

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Config 定义了 CORS 中间件的配置选项。
type Config struct {
	// AllowedOrigins 是允许的来源列表。
	// 可以使用 "*" 表示允许所有来源。
	// 默认值是 ["*"]。
	AllowedOrigins []string

	// AllowOriginFunc 是一个函数，用于动态判断是否允许特定来源。
	// 如果设置了此函数，它将优先于 AllowedOrigins 使用。
	// 默认值是 nil。
	AllowOriginFunc func(origin string) bool

	// AllowedMethods 是允许的 HTTP 方法列表。
	// 默认值是 ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"]。
	AllowedMethods []string

	// AllowedHeaders 是允许的 HTTP 头部列表。
	// 可以使用 "*" 表示允许客户端请求的所有头部。
	// 默认值是 []。
	AllowedHeaders []string

	// ExposeHeaders 是允许客户端读取的响应头部列表。
	// 默认值是 []。
	ExposeHeaders []string

	// AllowCredentials 表示是否允许包含凭证的请求。
	// 注意：如果为 true，则 AllowedOrigins 不能包含 "*"，必须指定明确的来源。
	// 默认值是 false。
	AllowCredentials bool

	// MaxAge 表示预检请求结果的缓存时间。
	// 默认值是 0 (不缓存)。
	MaxAge time.Duration

	// OptionsPassthrough 表示是否将 OPTIONS 请求传递给下一个处理器。
	// 默认值是 false。
	OptionsPassthrough bool

	// Debug 表示是否启用调试模式，输出详细日志。
	// 默认值是 false。
	Debug bool
}

// Option 是配置 CORS 中间件的函数选项。
type Option func(*Config)

// DefaultConfig 返回默认的 CORS 配置。
func DefaultConfig() *Config {
	return &Config{
		AllowedOrigins:     []string{"*"},
		AllowOriginFunc:    nil,
		AllowedMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:     []string{},
		ExposeHeaders:      []string{},
		AllowCredentials:   false,
		MaxAge:             0,
		OptionsPassthrough: false,
		Debug:              false,
	}
}

// WithAllowedOrigins 设置允许的来源列表。
func WithAllowedOrigins(origins []string) Option {
	return func(c *Config) {
		c.AllowedOrigins = origins
	}
}

// WithAllowOriginFunc 设置动态判断来源是否允许的函数。
func WithAllowOriginFunc(fn func(origin string) bool) Option {
	return func(c *Config) {
		c.AllowOriginFunc = fn
	}
}

// WithAllowedMethods 设置允许的 HTTP 方法列表。
func WithAllowedMethods(methods []string) Option {
	return func(c *Config) {
		c.AllowedMethods = methods
	}
}

// WithAllowedHeaders 设置允许的 HTTP 头部列表。
func WithAllowedHeaders(headers []string) Option {
	return func(c *Config) {
		c.AllowedHeaders = headers
	}
}

// WithExposeHeaders 设置允许客户端读取的响应头部列表。
func WithExposeHeaders(headers []string) Option {
	return func(c *Config) {
		c.ExposeHeaders = headers
	}
}

// WithAllowCredentials 设置是否允许包含凭证的请求。
func WithAllowCredentials(allow bool) Option {
	return func(c *Config) {
		c.AllowCredentials = allow
	}
}

// WithMaxAge 设置预检请求结果的缓存时间。
func WithMaxAge(duration time.Duration) Option {
	return func(c *Config) {
		c.MaxAge = duration
	}
}

// WithOptionsPassthrough 设置是否将 OPTIONS 请求传递给下一个处理器。
func WithOptionsPassthrough(passthrough bool) Option {
	return func(c *Config) {
		c.OptionsPassthrough = passthrough
	}
}

// WithDebug 设置是否启用调试模式。
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// New 使用默认配置创建一个新的 CORS 中间件。
func New() gin.HandlerFunc {
	return NewWithConfig()
}

// NewWithConfig 使用自定义配置创建一个新的 CORS 中间件。
func NewWithConfig(options ...Option) gin.HandlerFunc {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return newCorsHandler(config)
}

// newCorsHandler 创建一个处理 CORS 的 Gin 中间件。
func newCorsHandler(config *Config) gin.HandlerFunc {
	// 确保配置有效
	validateConfig(config)

	// 将 AllowedMethods 转换为大写
	for i, method := range config.AllowedMethods {
		config.AllowedMethods[i] = strings.ToUpper(method)
	}

	return func(c *gin.Context) {
		// 获取请求来源
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// 不是 CORS 请求，继续处理
			c.Next()
			return
		}

		// 检查是否允许该来源
		allowedOrigin := ""
		if config.AllowOriginFunc != nil {
			if config.AllowOriginFunc(origin) {
				allowedOrigin = origin
			}
		} else {
			for _, o := range config.AllowedOrigins {
				if o == "*" {
					// 如果允许所有来源，但启用了凭证，则必须返回实际的来源而不是 "*"
					if config.AllowCredentials {
						allowedOrigin = origin
					} else {
						allowedOrigin = "*"
					}
					break
				} else if o == origin {
					allowedOrigin = origin
					break
				}
			}
		}

		// 如果来源不被允许，继续处理请求（可能会因为同源策略而失败）
		if allowedOrigin == "" {
			if config.Debug {
				log.Printf("CORS: 来源 '%s' 不被允许\n", origin)
			}
			c.Next()
			return
		}

		// 设置 CORS 头部
		c.Header("Access-Control-Allow-Origin", allowedOrigin)

		// 处理预检请求 (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			// 处理预检请求
			handlePreflight(c, config, allowedOrigin)
			if !config.OptionsPassthrough {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else {
			// 处理实际请求
			handleActual(c, config)
		}

		c.Next()
	}
}

// validateConfig 验证配置的有效性，并修复可能的冲突。
func validateConfig(config *Config) {
	// 如果允许凭证且允许所有来源，输出警告
	if config.AllowCredentials && contains(config.AllowedOrigins, "*") && config.AllowOriginFunc == nil {
		if config.Debug {
			log.Println("CORS 警告: 同时设置 AllowCredentials=true 和 AllowedOrigins=['*'] 可能导致安全问题")
		}
	}
}

// handlePreflight 处理 CORS 预检请求。
func handlePreflight(c *gin.Context, config *Config, allowedOrigin string) {
	// 设置允许的方法
	if len(config.AllowedMethods) > 0 {
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
	}

	// 设置允许的头部
	var allowedHeaders string
	if len(config.AllowedHeaders) > 0 {
		if contains(config.AllowedHeaders, "*") {
			// 如果允许所有头部，则回显请求的头部
			requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
			if requestHeaders != "" {
				allowedHeaders = requestHeaders
			} else {
				allowedHeaders = "*"
			}
		} else {
			allowedHeaders = strings.Join(config.AllowedHeaders, ", ")
		}
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
	}

	// 设置凭证
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// 设置缓存时间
	if config.MaxAge > 0 {
		maxAge := int64(config.MaxAge.Seconds())
		c.Header("Access-Control-Max-Age", strconv.FormatInt(maxAge, 10))
	}

	if config.Debug {
		log.Printf("CORS: 预检请求处理完成，来源: %s\n", allowedOrigin)
	}
}

// handleActual 处理 CORS 实际请求。
func handleActual(c *gin.Context, config *Config) {
	// 设置凭证
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// 设置可暴露的头部
	if len(config.ExposeHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
	}

	if config.Debug {
		log.Printf("CORS: 实际请求处理完成，方法: %s, 路径: %s\n", c.Request.Method, c.Request.URL.Path)
	}
}

// contains 检查切片是否包含特定字符串。
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// AllowAll 返回一个允许所有来源、方法和头部的 CORS 配置。
// 注意：这仅适用于开发环境，生产环境应使用更严格的配置。
func AllowAll() gin.HandlerFunc {
	return NewWithConfig(
		WithAllowedOrigins([]string{"*"}),
		WithAllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
		WithAllowedHeaders([]string{"*"}),
		WithExposeHeaders([]string{"Content-Length", "Content-Type"}),
	)
}

// DefaultHandler 返回一个使用默认配置的 CORS 中间件。
func DefaultHandler() gin.HandlerFunc {
	return New()
}
