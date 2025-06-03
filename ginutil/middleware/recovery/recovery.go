// Package recovery 提供了一个用于 Gin 框架的错误恢复中间件。
//
// 该中间件可以捕获处理 HTTP 请求过程中发生的 panic，
// 记录详细的堆栈信息，并返回友好的错误响应。
package recovery

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Humphrey-He/go-generic-utils/ginutil/response"

	"github.com/gin-gonic/gin"
)

// 错误码常量
const (
	// InternalServerError 表示服务器内部错误
	InternalServerError = 500
)

// ErrorHandler 是一个处理 panic 错误的函数类型。
type ErrorHandler func(c *gin.Context, err interface{})

// Config 定义了恢复中间件的配置选项。
type Config struct {
	// ErrorHandler 是自定义错误处理函数。
	// 如果不提供，将使用默认的错误处理函数。
	ErrorHandler ErrorHandler

	// StackAll 表示是否记录完整的堆栈信息，默认为 false。
	StackAll bool

	// StackSize 是堆栈缓冲区的大小，默认为 4KB。
	StackSize int

	// Output 是日志输出的目标，默认为 os.Stderr。
	Output io.Writer

	// DisableStackAll 表示是否禁用完整的堆栈跟踪，默认为 false。
	DisableStackAll bool

	// DisablePrintStack 表示是否禁用打印堆栈信息，默认为 false。
	DisablePrintStack bool

	// DisableRecovery 表示是否禁用恢复功能，默认为 false。
	// 如果设置为 true，将不会捕获 panic，而是让其传播到上层。
	DisableRecovery bool
}

// Option 是配置恢复中间件的函数选项。
type Option func(*Config)

// WithErrorHandler 设置自定义错误处理函数。
func WithErrorHandler(handler ErrorHandler) Option {
	return func(c *Config) {
		c.ErrorHandler = handler
	}
}

// WithStackAll 设置是否记录完整的堆栈信息。
func WithStackAll(stackAll bool) Option {
	return func(c *Config) {
		c.StackAll = stackAll
	}
}

// WithStackSize 设置堆栈缓冲区的大小。
func WithStackSize(stackSize int) Option {
	return func(c *Config) {
		c.StackSize = stackSize
	}
}

// WithOutput 设置日志输出的目标。
func WithOutput(output io.Writer) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// WithDisableStackAll 设置是否禁用完整的堆栈跟踪。
func WithDisableStackAll(disable bool) Option {
	return func(c *Config) {
		c.DisableStackAll = disable
	}
}

// WithDisablePrintStack 设置是否禁用打印堆栈信息。
func WithDisablePrintStack(disable bool) Option {
	return func(c *Config) {
		c.DisablePrintStack = disable
	}
}

// WithDisableRecovery 设置是否禁用恢复功能。
func WithDisableRecovery(disable bool) Option {
	return func(c *Config) {
		c.DisableRecovery = disable
	}
}

// DefaultConfig 返回默认的恢复中间件配置。
func DefaultConfig() *Config {
	return &Config{
		ErrorHandler:      defaultErrorHandler,
		StackAll:          false,
		StackSize:         4 << 10, // 4 KB
		Output:            os.Stderr,
		DisableStackAll:   false,
		DisablePrintStack: false,
		DisableRecovery:   false,
	}
}

// 默认的错误处理函数
func defaultErrorHandler(c *gin.Context, err interface{}) {
	// 检查是否是已知错误
	var errMsg string
	switch e := err.(type) {
	case string:
		errMsg = e
	case error:
		errMsg = e.Error()
	default:
		errMsg = "服务器内部错误"
	}

	// 使用 response 包返回统一的错误响应
	response.Fail(c, InternalServerError, errMsg)
}

// New 使用默认配置创建一个新的恢复中间件。
func New() gin.HandlerFunc {
	return NewWithConfig()
}

// NewWithConfig 使用自定义配置创建一个新的恢复中间件。
func NewWithConfig(options ...Option) gin.HandlerFunc {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return newRecoveryHandler(config)
}

// newRecoveryHandler 创建一个处理恢复的 Gin 中间件。
func newRecoveryHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果禁用了恢复功能，直接执行后续处理器
		if config.DisableRecovery {
			c.Next()
			return
		}

		// 创建恢复函数
		defer func() {
			if err := recover(); err != nil {
				// 检查是否是连接断开错误
				if ne, ok := checkBrokenPipe(err); ok {
					// 如果是连接断开错误，只记录日志，不返回错误响应
					log.Printf("连接断开: %v", ne.Error())
					c.Abort()
					return
				}

				// 获取堆栈信息
				stack := stack(config.StackAll, config.StackSize)

				// 打印请求信息
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := formatRequestHeaders(c.Request.Header)

				// 如果不禁用打印堆栈信息，则输出到日志
				if !config.DisablePrintStack {
					logger := log.New(config.Output, "\n\n", log.LstdFlags)
					logger.Printf("[Recovery] panic 恢复，路径: %s\n", c.Request.URL.Path)
					logger.Printf("时间: %v\n", time.Now())
					logger.Printf("请求: %s\n%s", string(httpRequest), headers)
					logger.Printf("错误: %s\n", err)
					logger.Printf("堆栈跟踪:\n%s", stack)
				}

				// 调用错误处理函数
				config.ErrorHandler(c, err)
			}
		}()

		// 执行后续处理器
		c.Next()
	}
}

// stack 返回格式化的堆栈跟踪。
func stack(all bool, size int) []byte {
	buf := make([]byte, size)
	n := runtime.Stack(buf, all)
	return buf[:n]
}

// checkBrokenPipe 检查错误是否是连接断开错误。
func checkBrokenPipe(err interface{}) (netErr net.Error, ok bool) {
	// 首先尝试将 err 转换为 net.Error
	if netErr, ok = err.(net.Error); ok {
		return netErr, ok
	}

	// 然后尝试将 err 转换为 error
	if e, ok := err.(error); ok {
		// 检查错误是否包含 "broken pipe" 或 "connection reset by peer"
		errMsg := e.Error()
		if strings.Contains(errMsg, "broken pipe") ||
			strings.Contains(errMsg, "connection reset by peer") {
			return nil, true
		}
	}

	return nil, false
}

// formatRequestHeaders 格式化请求头，脱敏敏感信息。
func formatRequestHeaders(header http.Header) string {
	var buf bytes.Buffer
	buf.WriteString("请求头:\n")
	for k, v := range header {
		// 脱敏敏感头部
		if k == "Authorization" || k == "Cookie" || strings.Contains(strings.ToLower(k), "token") {
			buf.WriteString(fmt.Sprintf("%s: [REDACTED]\n", k))
		} else {
			buf.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}
	}
	return buf.String()
}

// CustomErrorHandler 创建一个自定义错误处理函数。
func CustomErrorHandler(handler func(c *gin.Context, err interface{})) ErrorHandler {
	return func(c *gin.Context, err interface{}) {
		handler(c, err)
	}
}

// JSONErrorHandler 创建一个返回 JSON 格式错误的处理函数。
func JSONErrorHandler(code int, defaultMessage string) ErrorHandler {
	return func(c *gin.Context, err interface{}) {
		var errMsg string
		switch e := err.(type) {
		case string:
			errMsg = e
		case error:
			errMsg = e.Error()
		default:
			errMsg = defaultMessage
		}

		c.AbortWithStatusJSON(code, gin.H{
			"code":    code,
			"message": errMsg,
		})
	}
}

// ErrorHandlerWithSentry 创建一个将错误发送到 Sentry 的处理函数。
// 注意：这只是一个示例，实际使用时需要导入 Sentry 相关包。
func ErrorHandlerWithSentry(handler ErrorHandler) ErrorHandler {
	return func(c *gin.Context, err interface{}) {
		// 这里可以添加将错误发送到 Sentry 的代码
		// 例如：
		// sentry.CaptureException(fmt.Errorf("%v", err))

		// 调用原始错误处理函数
		handler(c, err)
	}
}

// ErrorHandlerWithLogger 创建一个将错误记录到日志的处理函数。
func ErrorHandlerWithLogger(handler ErrorHandler, logger *log.Logger) ErrorHandler {
	return func(c *gin.Context, err interface{}) {
		// 记录错误到日志
		logger.Printf("处理请求时发生错误: %v, 路径: %s, 方法: %s",
			err, c.Request.URL.Path, c.Request.Method)

		// 调用原始错误处理函数
		handler(c, err)
	}
}

// IsBrokenPipeError 检查错误是否是连接断开错误。
func IsBrokenPipeError(err error) bool {
	if err == nil {
		return false
	}

	// 检查错误是否包含 "broken pipe" 或 "connection reset by peer"
	errMsg := err.Error()
	return strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "connection reset by peer")
}

// RecoveryWithWriter 使用指定的输出创建一个恢复中间件。
func RecoveryWithWriter(out io.Writer) gin.HandlerFunc {
	return NewWithConfig(WithOutput(out))
}

// RecoveryWithCustomErrorHandler 使用自定义错误处理函数创建一个恢复中间件。
func RecoveryWithCustomErrorHandler(handler ErrorHandler) gin.HandlerFunc {
	return NewWithConfig(WithErrorHandler(handler))
}
