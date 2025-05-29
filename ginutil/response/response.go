package response

import (
	// 引入错误码包

	"time"

	"github.com/gin-gonic/gin"
)

// StandardResponse 是标准的API响应结构体。
// 使用泛型 T 来支持不同类型的业务数据。
type StandardResponse[T any] struct {
	Code       int    `json:"code"`               // 业务状态码 (例如：ecode.OK 表示成功)
	Message    string `json:"message"`            // 用户可读的消息
	Data       T      `json:"data,omitempty"`     // 实际的业务数据荷载
	TraceID    string `json:"trace_id,omitempty"` // 链路追踪ID (可选)
	ServerTime int64  `json:"server_time"`        // 服务器时间戳 (Unix毫秒)
}

// GinTraceIDKey 是从Gin Context中获取TraceID时使用的键名。
const GinTraceIDKey = "X-Trace-ID" // 假设由中间件设置

// 用于测试的固定时间戳
var fixedServerTime int64 = 0

// SetFixedServerTimeForTest 设置一个固定的服务器时间戳，用于测试。
// 设置为0将恢复使用实时时间戳。
func SetFixedServerTimeForTest(timestamp int64) {
	fixedServerTime = timestamp
}

// getServerTime 获取服务器时间戳
func getServerTime() int64 {
	if fixedServerTime > 0 {
		return fixedServerTime
	}
	return time.Now().UnixMilli()
}

// sendJSON 是一个内部辅助函数，用于发送JSON响应。
// 它会自动填充TraceID和ServerTime。
func sendJSON[T any](c *gin.Context, httpStatus int, resp StandardResponse[T]) {
	// 如果TraceID未设置，尝试从Context中获取
	if resp.TraceID == "" {
		resp.TraceID = c.GetString(GinTraceIDKey)
	}

	// 设置服务器时间
	if resp.ServerTime == 0 {
		resp.ServerTime = getServerTime()
	}

	// 发送JSON响应
	c.JSON(httpStatus, resp)
}

// sendAbortJSON 是一个内部辅助函数，用于发送中止请求的JSON响应。
// 它与sendJSON类似，但会调用c.Abort()中止后续处理。
func sendAbortJSON[T any](c *gin.Context, httpStatus int, resp StandardResponse[T]) {
	// 如果TraceID未设置，尝试从Context中获取
	if resp.TraceID == "" {
		resp.TraceID = c.GetString(GinTraceIDKey)
	}

	// 设置服务器时间
	if resp.ServerTime == 0 {
		resp.ServerTime = getServerTime()
	}

	// 发送JSON响应并中止后续处理
	c.AbortWithStatusJSON(httpStatus, resp)
}
