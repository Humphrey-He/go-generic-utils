// Package logger 提供了一个用于 Gin 框架的可配置日志中间件。
//
// 该中间件可以记录请求的详细信息，包括时间戳、状态码、延迟时间、
// 客户端 IP、HTTP 方法、请求路径、User-Agent 等，并支持自定义格式化和输出。
package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LogFormatter 是日志格式化器接口，用于自定义日志输出格式。
type LogFormatter interface {
	// Format 将日志条目格式化为字节数组。
	Format(*LogEntry) []byte
}

// LogEntry 包含一个 HTTP 请求的日志信息。
type LogEntry struct {
	// 时间戳
	Timestamp time.Time `json:"timestamp"`
	// 状态码
	StatusCode int `json:"status_code"`
	// 延迟时间（毫秒）
	Latency float64 `json:"latency_ms"`
	// 客户端 IP
	ClientIP string `json:"client_ip"`
	// HTTP 方法
	Method string `json:"method"`
	// 请求路径
	Path string `json:"path"`
	// 原始查询参数
	RawQuery string `json:"raw_query,omitempty"`
	// User-Agent
	UserAgent string `json:"user_agent"`
	// 错误信息
	ErrorMessage string `json:"error_message,omitempty"`
	// 请求体大小
	RequestSize int64 `json:"request_size"`
	// 响应体大小
	ResponseSize int `json:"response_size"`
	// 请求 ID
	RequestID string `json:"request_id,omitempty"`
	// 额外信息
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// Config 定义了日志中间件的配置选项。
type Config struct {
	// Output 是日志输出的目标，默认为 os.Stdout。
	Output io.Writer
	// Formatter 是日志格式化器，默认为 TextFormatter。
	Formatter LogFormatter
	// SkipPaths 是不需要记录日志的路径列表。
	SkipPaths []string
	// SkipPathRegexps 是不需要记录日志的路径正则表达式列表。
	SkipPathRegexps []*regexp.Regexp
	// TimeFormat 是时间戳的格式，默认为 RFC3339。
	TimeFormat string
	// UTC 表示是否使用 UTC 时间，默认为 false。
	UTC bool
	// ContextKeys 是需要从 Gin Context 中提取并记录的键列表。
	ContextKeys []string
	// RequestHeader 是需要记录的请求头列表。
	RequestHeader []string
	// ResponseHeader 是需要记录的响应头列表。
	ResponseHeader []string
	// DisableRequestLog 表示是否禁用请求日志，默认为 false。
	DisableRequestLog bool
	// DisableResponseLog 表示是否禁用响应日志，默认为 false。
	DisableResponseLog bool
	// LogFunc 是自定义日志函数，如果设置了该函数，将使用它来记录日志而不是 Output。
	LogFunc func(*LogEntry)
}

// Option 是配置日志中间件的函数选项。
type Option func(*Config)

// WithOutput 设置日志输出的目标。
func WithOutput(output io.Writer) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// WithFormatter 设置日志格式化器。
func WithFormatter(formatter LogFormatter) Option {
	return func(c *Config) {
		c.Formatter = formatter
	}
}

// WithSkipPaths 设置不需要记录日志的路径列表。
func WithSkipPaths(paths ...string) Option {
	return func(c *Config) {
		c.SkipPaths = append(c.SkipPaths, paths...)
	}
}

// WithSkipPathRegexps 设置不需要记录日志的路径正则表达式列表。
func WithSkipPathRegexps(regexps ...*regexp.Regexp) Option {
	return func(c *Config) {
		c.SkipPathRegexps = append(c.SkipPathRegexps, regexps...)
	}
}

// WithTimeFormat 设置时间戳的格式。
func WithTimeFormat(format string) Option {
	return func(c *Config) {
		c.TimeFormat = format
	}
}

// WithUTC 设置是否使用 UTC 时间。
func WithUTC(utc bool) Option {
	return func(c *Config) {
		c.UTC = utc
	}
}

// WithContextKeys 设置需要从 Gin Context 中提取并记录的键列表。
func WithContextKeys(keys ...string) Option {
	return func(c *Config) {
		c.ContextKeys = append(c.ContextKeys, keys...)
	}
}

// WithRequestHeader 设置需要记录的请求头列表。
func WithRequestHeader(headers ...string) Option {
	return func(c *Config) {
		c.RequestHeader = append(c.RequestHeader, headers...)
	}
}

// WithResponseHeader 设置需要记录的响应头列表。
func WithResponseHeader(headers ...string) Option {
	return func(c *Config) {
		c.ResponseHeader = append(c.ResponseHeader, headers...)
	}
}

// WithDisableRequestLog 设置是否禁用请求日志。
func WithDisableRequestLog(disable bool) Option {
	return func(c *Config) {
		c.DisableRequestLog = disable
	}
}

// WithDisableResponseLog 设置是否禁用响应日志。
func WithDisableResponseLog(disable bool) Option {
	return func(c *Config) {
		c.DisableResponseLog = disable
	}
}

// WithLogFunc 设置自定义日志函数。
func WithLogFunc(logFunc func(*LogEntry)) Option {
	return func(c *Config) {
		c.LogFunc = logFunc
	}
}

// DefaultConfig 返回默认的日志中间件配置。
func DefaultConfig() *Config {
	return &Config{
		Output:             os.Stdout,
		Formatter:          &TextFormatter{},
		SkipPaths:          []string{},
		SkipPathRegexps:    []*regexp.Regexp{},
		TimeFormat:         time.RFC3339,
		UTC:                false,
		ContextKeys:        []string{},
		RequestHeader:      []string{},
		ResponseHeader:     []string{},
		DisableRequestLog:  false,
		DisableResponseLog: false,
		LogFunc:            nil,
	}
}

// JSONFormatter 实现了 LogFormatter 接口，将日志格式化为 JSON。
type JSONFormatter struct {
	// PrettyPrint 表示是否美化输出的 JSON，默认为 false。
	PrettyPrint bool
}

// Format 将日志条目格式化为 JSON 字节数组。
func (f *JSONFormatter) Format(entry *LogEntry) []byte {
	var output []byte
	var err error

	if f.PrettyPrint {
		output, err = json.MarshalIndent(entry, "", "  ")
	} else {
		output, err = json.Marshal(entry)
	}

	if err != nil {
		return []byte(fmt.Sprintf("格式化日志失败: %v", err))
	}

	return append(output, '\n')
}

// TextFormatter 实现了 LogFormatter 接口，将日志格式化为文本。
type TextFormatter struct {
	// TimeFormat 是时间戳的格式，默认为 RFC3339。
	TimeFormat string
	// DisableColors 表示是否禁用颜色输出，默认为 false。
	DisableColors bool
}

// Format 将日志条目格式化为文本字节数组。
func (f *TextFormatter) Format(entry *LogEntry) []byte {
	var output bytes.Buffer

	// 设置默认时间格式
	timeFormat := f.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	// 格式化时间戳
	timestamp := entry.Timestamp.Format(timeFormat)

	// 格式化状态码（带颜色）
	statusCodeColor := ""
	statusCodeEndColor := ""
	if !f.DisableColors {
		if entry.StatusCode >= 200 && entry.StatusCode < 300 {
			statusCodeColor = "\033[97;42m" // 绿色
		} else if entry.StatusCode >= 300 && entry.StatusCode < 400 {
			statusCodeColor = "\033[90;47m" // 灰色
		} else if entry.StatusCode >= 400 && entry.StatusCode < 500 {
			statusCodeColor = "\033[97;43m" // 黄色
		} else {
			statusCodeColor = "\033[97;41m" // 红色
		}
		statusCodeEndColor = "\033[0m"
	}

	// 格式化日志内容
	fmt.Fprintf(&output, "[%s] %s%d%s %s %s %s %.2fms %d bytes",
		timestamp,
		statusCodeColor, entry.StatusCode, statusCodeEndColor,
		entry.Method,
		entry.Path,
		entry.ClientIP,
		entry.Latency,
		entry.ResponseSize,
	)

	// 添加错误信息
	if entry.ErrorMessage != "" {
		fmt.Fprintf(&output, " | 错误: %s", entry.ErrorMessage)
	}

	// 添加请求 ID
	if entry.RequestID != "" {
		fmt.Fprintf(&output, " | 请求ID: %s", entry.RequestID)
	}

	// 添加额外信息
	if len(entry.Extra) > 0 {
		extraBytes, _ := json.Marshal(entry.Extra)
		fmt.Fprintf(&output, " | 额外信息: %s", string(extraBytes))
	}

	output.WriteByte('\n')
	return output.Bytes()
}

// 日志写入的互斥锁，用于保证并发安全
var logMutex sync.Mutex

// New 使用默认配置创建一个新的日志中间件。
func New() gin.HandlerFunc {
	return NewWithConfig()
}

// NewWithConfig 使用自定义配置创建一个新的日志中间件。
func NewWithConfig(options ...Option) gin.HandlerFunc {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return newLoggerHandler(config)
}

// newLoggerHandler 创建一个处理日志的 Gin 中间件。
func newLoggerHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过该路径
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 检查路径是否在跳过列表中
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 检查路径是否匹配正则表达式
		for _, re := range config.SkipPathRegexps {
			if re.MatchString(path) {
				c.Next()
				return
			}
		}

		// 开始时间
		start := time.Now()

		// 记录请求体大小
		requestSize := c.Request.ContentLength

		// 如果不禁用请求日志，则记录请求信息
		if !config.DisableRequestLog {
			// 这里可以添加请求日志的记录逻辑
		}

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()

		// 计算延迟时间（毫秒）
		latency := float64(end.Sub(start).Nanoseconds()) / 1e6

		// 获取状态码
		statusCode := c.Writer.Status()

		// 获取客户端 IP
		clientIP := c.ClientIP()

		// 获取 HTTP 方法
		method := c.Request.Method

		// 获取 User-Agent
		userAgent := c.Request.UserAgent()

		// 获取错误信息
		var errorMessage string
		if len(c.Errors) > 0 {
			errorMessage = c.Errors.String()
		}

		// 获取响应体大小
		responseSize := c.Writer.Size()

		// 获取请求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			if id, exists := c.Get("RequestID"); exists {
				if idStr, ok := id.(string); ok {
					requestID = idStr
				}
			}
		}

		// 创建日志条目
		entry := &LogEntry{
			Timestamp:    end,
			StatusCode:   statusCode,
			Latency:      latency,
			ClientIP:     clientIP,
			Method:       method,
			Path:         path,
			RawQuery:     raw,
			UserAgent:    userAgent,
			ErrorMessage: errorMessage,
			RequestSize:  requestSize,
			ResponseSize: responseSize,
			RequestID:    requestID,
			Extra:        make(map[string]interface{}),
		}

		// 如果使用 UTC 时间
		if config.UTC {
			entry.Timestamp = entry.Timestamp.UTC()
		}

		// 从 Context 中提取额外信息
		for _, key := range config.ContextKeys {
			if value, exists := c.Get(key); exists {
				entry.Extra[key] = value
			}
		}

		// 提取请求头
		if len(config.RequestHeader) > 0 {
			headers := make(map[string]string)
			for _, header := range config.RequestHeader {
				if value := c.GetHeader(header); value != "" {
					// 对敏感信息进行脱敏
					if header == "Authorization" || header == "Cookie" {
						headers[header] = "[REDACTED]"
					} else {
						headers[header] = value
					}
				}
			}
			if len(headers) > 0 {
				entry.Extra["request_headers"] = headers
			}
		}

		// 提取响应头
		if len(config.ResponseHeader) > 0 {
			headers := make(map[string]string)
			for _, header := range config.ResponseHeader {
				if values := c.Writer.Header()[header]; len(values) > 0 {
					headers[header] = values[0]
				}
			}
			if len(headers) > 0 {
				entry.Extra["response_headers"] = headers
			}
		}

		// 使用自定义日志函数或格式化并输出日志
		if config.LogFunc != nil {
			config.LogFunc(entry)
		} else {
			// 格式化日志
			output := config.Formatter.Format(entry)

			// 并发安全地写入日志
			logMutex.Lock()
			config.Output.Write(output)
			logMutex.Unlock()
		}
	}
}
