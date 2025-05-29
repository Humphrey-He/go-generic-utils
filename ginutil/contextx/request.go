// Package contextx 提供了对 gin.Context 操作的工具函数。
package contextx

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceIDKey 是用于在上下文中存储请求跟踪 ID 的键。
	TraceIDKey = "X-Trace-ID"

	// RequestStartTimeKey 是用于在上下文中存储请求开始时间的键。
	RequestStartTimeKey = "request.start_time"

	// ClientIPKey 是用于在上下文中存储客户端 IP 的键。
	ClientIPKey = "request.client_ip"
)

// SetTraceID 在上下文中设置请求跟踪 ID。
// 如果提供的 traceID 为空，将生成一个新的 UUID 作为跟踪 ID。
func SetTraceID(c *gin.Context, traceID string) {
	if c == nil {
		return
	}

	if traceID == "" {
		traceID = uuid.New().String()
	}

	Set(c, TraceIDKey, traceID)
	// 同时设置到响应头中
	c.Header(TraceIDKey, traceID)
}

// GetTraceID 从上下文中获取请求跟踪 ID。
// 如果不存在，将生成一个新的 UUID，设置到上下文中并返回。
func GetTraceID(c *gin.Context) string {
	if c == nil {
		return ""
	}

	// 尝试从上下文中获取
	traceID, exists := Get[string](c, TraceIDKey)
	if exists && traceID != "" {
		return traceID
	}

	// 尝试从请求头中获取
	traceID = c.GetHeader(TraceIDKey)
	if traceID != "" {
		// 设置到上下文中
		Set(c, TraceIDKey, traceID)
		return traceID
	}

	// 生成新的跟踪 ID
	traceID = uuid.New().String()
	SetTraceID(c, traceID)
	return traceID
}

// SetRequestStartTime 在上下文中设置请求开始时间。
func SetRequestStartTime(c *gin.Context, startTime time.Time) {
	if c == nil {
		return
	}
	Set(c, RequestStartTimeKey, startTime)
}

// GetRequestStartTime 从上下文中获取请求开始时间。
// 如果不存在，将返回当前时间和 false。
func GetRequestStartTime(c *gin.Context) (time.Time, bool) {
	if c == nil {
		return time.Now(), false
	}
	return Get[time.Time](c, RequestStartTimeKey)
}

// GetRequestDuration 计算从请求开始到现在的持续时间。
// 如果请求开始时间不存在，将返回 0 和 false。
func GetRequestDuration(c *gin.Context) (time.Duration, bool) {
	startTime, exists := GetRequestStartTime(c)
	if !exists {
		return 0, false
	}
	return time.Since(startTime), true
}

// SetClientIP 在上下文中设置客户端 IP。
func SetClientIP(c *gin.Context, ip string) {
	if c == nil {
		return
	}
	Set(c, ClientIPKey, ip)
}

// GetClientIP 获取客户端 IP。
// 首先尝试从上下文中获取，如果不存在，将使用 c.ClientIP() 方法。
func GetClientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}

	// 尝试从上下文中获取
	ip, exists := Get[string](c, ClientIPKey)
	if exists && ip != "" {
		return ip
	}

	// 使用 Gin 的 ClientIP 方法
	ip = c.ClientIP()
	if ip != "" {
		// 设置到上下文中以便后续使用
		SetClientIP(c, ip)
	}

	return ip
}

// GetRealIP 尝试获取客户端的真实 IP 地址。
// 按照以下顺序检查：
// 1. X-Real-IP 头
// 2. X-Forwarded-For 头的第一个值
// 3. 使用 c.ClientIP() 方法
func GetRealIP(c *gin.Context) string {
	if c == nil {
		return ""
	}

	// 检查 X-Real-IP 头
	ip := c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 检查 X-Forwarded-For 头
	ip = c.GetHeader("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能包含多个 IP，我们取第一个
		for i := 0; i < len(ip) && i < 50; i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}

	// 使用 Gin 的 ClientIP 方法
	return c.ClientIP()
}

// GetUserAgent 获取用户代理字符串。
func GetUserAgent(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetHeader("User-Agent")
}

// GetReferer 获取请求的 Referer。
func GetReferer(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetHeader("Referer")
}

// GetRequestMethod 获取请求的 HTTP 方法。
func GetRequestMethod(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.Request.Method
}

// GetRequestPath 获取请求的路径。
func GetRequestPath(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.Request.URL.Path
}

// GetRequestQuery 获取请求的查询字符串。
func GetRequestQuery(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.Request.URL.RawQuery
}

// GetRequestHost 获取请求的主机名。
func GetRequestHost(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.Request.Host
}

// GetRequestProtocol 获取请求的协议。
func GetRequestProtocol(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}
