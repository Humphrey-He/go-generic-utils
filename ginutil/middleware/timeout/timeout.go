// Package timeout 提供了一个用于 Gin 框架的请求超时中间件。
//
// 该中间件可以为 HTTP 请求设置最大处理时间，超时后会中断请求处理并返回超时响应。
// 它使用 context.WithTimeout 和 goroutine 来实现超时控制，并确保 panic 能够被正确传递。
package timeout

import (
	"context"
	"ggu/ginutil/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 错误码常量
const (
	// GatewayTimeout 表示请求处理超时
	GatewayTimeout = http.StatusGatewayTimeout
)

// TimeoutHandler 是一个处理超时的函数类型。
type TimeoutHandler func(c *gin.Context)

// Config 定义了超时中间件的配置选项。
type Config struct {
	// Timeout 是请求处理的最大时间，默认为 30 秒。
	Timeout time.Duration

	// TimeoutHandler 是自定义超时处理函数。
	// 如果不提供，将使用默认的超时处理函数。
	TimeoutHandler TimeoutHandler

	// TimeoutMessage 是超时响应的错误消息，默认为 "请求处理超时"。
	TimeoutMessage string

	// TimeoutCode 是超时响应的 HTTP 状态码，默认为 504 Gateway Timeout。
	TimeoutCode int

	// TimeoutJSON 是超时响应的 JSON 体，如果设置了，将覆盖 TimeoutMessage。
	TimeoutJSON interface{}

	// DisableTimeout 表示是否禁用超时功能，默认为 false。
	DisableTimeout bool
}

// Option 是配置超时中间件的函数选项。
type Option func(*Config)

// WithTimeout 设置请求处理的最大时间。
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithTimeoutHandler 设置自定义超时处理函数。
func WithTimeoutHandler(handler TimeoutHandler) Option {
	return func(c *Config) {
		c.TimeoutHandler = handler
	}
}

// WithTimeoutMessage 设置超时响应的错误消息。
func WithTimeoutMessage(message string) Option {
	return func(c *Config) {
		c.TimeoutMessage = message
	}
}

// WithTimeoutCode 设置超时响应的 HTTP 状态码。
func WithTimeoutCode(code int) Option {
	return func(c *Config) {
		c.TimeoutCode = code
	}
}

// WithTimeoutJSON 设置超时响应的 JSON 体。
func WithTimeoutJSON(json interface{}) Option {
	return func(c *Config) {
		c.TimeoutJSON = json
	}
}

// WithDisableTimeout 设置是否禁用超时功能。
func WithDisableTimeout(disable bool) Option {
	return func(c *Config) {
		c.DisableTimeout = disable
	}
}

// DefaultConfig 返回默认的超时中间件配置。
func DefaultConfig() *Config {
	return &Config{
		Timeout:        30 * time.Second,
		TimeoutHandler: defaultTimeoutHandler,
		TimeoutMessage: "请求处理超时",
		TimeoutCode:    GatewayTimeout,
		TimeoutJSON:    nil,
		DisableTimeout: false,
	}
}

// 默认的超时处理函数
func defaultTimeoutHandler(c *gin.Context) {
	// 使用 response 包返回统一的超时响应
	response.Fail(c, GatewayTimeout, "请求处理超时")
}

// New 使用默认配置创建一个新的超时中间件。
func New() gin.HandlerFunc {
	return NewWithConfig()
}

// NewWithConfig 使用自定义配置创建一个新的超时中间件。
func NewWithConfig(options ...Option) gin.HandlerFunc {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return newTimeoutHandler(config)
}

// newTimeoutHandler 创建一个处理超时的 Gin 中间件。
func newTimeoutHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果禁用了超时功能，直接执行后续处理器
		if config.DisableTimeout {
			c.Next()
			return
		}

		// 创建一个带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 创建用于传递 panic 的通道
		panicChan := make(chan interface{}, 1)
		// 创建用于通知处理完成的通道
		doneChan := make(chan struct{}, 1)

		// 在单独的 goroutine 中执行请求处理
		go func() {
			defer func() {
				// 捕获 panic 并通过通道传递
				if r := recover(); r != nil {
					panicChan <- r
				}
			}()

			// 执行后续处理器
			c.Next()

			// 通知处理完成
			doneChan <- struct{}{}
		}()

		// 等待请求处理完成或超时
		select {
		case <-doneChan:
			// 请求处理正常完成
			return
		case r := <-panicChan:
			// 请求处理发生 panic，重新抛出 panic
			panic(r)
		case <-ctx.Done():
			// 请求处理超时
			if ctx.Err() == context.DeadlineExceeded {
				// 检查是否已经写入响应
				if !c.Writer.Written() {
					// 调用超时处理函数
					if config.TimeoutJSON != nil {
						// 使用自定义 JSON 体
						c.JSON(config.TimeoutCode, config.TimeoutJSON)
					} else if config.TimeoutHandler != nil {
						// 使用自定义超时处理函数
						config.TimeoutHandler(c)
					} else {
						// 使用默认超时处理函数
						response.Fail(c, config.TimeoutCode, config.TimeoutMessage)
					}
				}

				// 中止后续处理器
				c.Abort()
			}
		}
	}
}

// TimeoutWithHandler 使用指定的超时时间和处理函数创建一个超时中间件。
func TimeoutWithHandler(timeout time.Duration, handler TimeoutHandler) gin.HandlerFunc {
	return NewWithConfig(WithTimeout(timeout), WithTimeoutHandler(handler))
}

// TimeoutWithMessage 使用指定的超时时间和错误消息创建一个超时中间件。
func TimeoutWithMessage(timeout time.Duration, message string) gin.HandlerFunc {
	return NewWithConfig(WithTimeout(timeout), WithTimeoutMessage(message))
}

// TimeoutWithCode 使用指定的超时时间和 HTTP 状态码创建一个超时中间件。
func TimeoutWithCode(timeout time.Duration, code int) gin.HandlerFunc {
	return NewWithConfig(WithTimeout(timeout), WithTimeoutCode(code))
}

// TimeoutWithJSON 使用指定的超时时间和 JSON 体创建一个超时中间件。
func TimeoutWithJSON(timeout time.Duration, json interface{}) gin.HandlerFunc {
	return NewWithConfig(WithTimeout(timeout), WithTimeoutJSON(json))
}
