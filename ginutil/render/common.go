// Package render 提供用于 Gin 框架的标准化响应渲染功能。
// 包括 JSON、XML 和 HTML 模板渲染，支持统一的响应格式、
// 错误处理、分页响应和追踪 ID 等功能。
package render

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 常量定义
const (
	// 默认消息
	DefaultSuccessMessage = "操作成功"
	DefaultErrorMessage   = "操作失败"

	// 业务码
	CodeSuccess            = 0    // 成功
	CodeInvalidParams      = 1001 // 参数错误
	CodeUnauthorized       = 1002 // 未授权
	CodeForbidden          = 1003 // 禁止访问
	CodeNotFound           = 1004 // 资源不存在
	CodeInternalError      = 1005 // 内部错误
	CodeServiceUnavailable = 1006 // 服务不可用
	CodeTimeout            = 1007 // 请求超时
	CodeTooManyRequests    = 1008 // 请求过多
	CodeBadGateway         = 1009 // 网关错误

	// 上下文键
	ContextKeyTraceID   = "X-Trace-ID" // 用于从上下文获取追踪 ID
	ContextKeyUserID    = "UserID"     // 用于从上下文获取用户 ID
	ContextKeyRequestID = "RequestID"  // 用于从上下文获取请求 ID
)

// StandardResponse 是标准的响应结构
type StandardResponse[T any] struct {
	Code       int       `json:"code" xml:"code"`               // 业务码
	Message    string    `json:"message" xml:"message"`         // 提示消息
	Data       T         `json:"data" xml:"data"`               // 业务数据
	TraceID    string    `json:"trace_id" xml:"trace_id"`       // 追踪 ID
	ServerTime time.Time `json:"server_time" xml:"server_time"` // 服务器时间
}

// PaginatedResponse 是分页响应结构
type PaginatedResponse[T any] struct {
	List     []T   `json:"list" xml:"list"`           // 数据列表
	Total    int64 `json:"total" xml:"total"`         // 总记录数
	PageNum  int   `json:"page_num" xml:"page_num"`   // 当前页码
	PageSize int   `json:"page_size" xml:"page_size"` // 每页记录数
	Pages    int   `json:"pages" xml:"pages"`         // 总页数
	HasNext  bool  `json:"has_next" xml:"has_next"`   // 是否有下一页
	HasPrev  bool  `json:"has_prev" xml:"has_prev"`   // 是否有上一页
}

// 配置选项
var (
	// JSONPrettyPrint 是否美化 JSON 输出
	JSONPrettyPrint = false

	// HTMLTemplateDir HTML 模板目录
	HTMLTemplateDir = "templates/*"

	// DefaultHTMLErrorTemplate 默认的 HTML 错误模板
	DefaultHTMLErrorTemplate = "error.html"
)

// Config 配置结构
type Config struct {
	// JSONPrettyPrint 是否美化 JSON 输出
	JSONPrettyPrint bool

	// HTMLTemplateDir HTML 模板目录
	HTMLTemplateDir string

	// DefaultHTMLErrorTemplate 默认的 HTML 错误模板
	DefaultHTMLErrorTemplate string
}

// Configure 配置渲染器
func Configure(config Config) {
	if config.JSONPrettyPrint {
		JSONPrettyPrint = config.JSONPrettyPrint
	}

	if config.HTMLTemplateDir != "" {
		HTMLTemplateDir = config.HTMLTemplateDir
	}

	if config.DefaultHTMLErrorTemplate != "" {
		DefaultHTMLErrorTemplate = config.DefaultHTMLErrorTemplate
	}
}

// getTraceID 从上下文获取追踪 ID
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(ContextKeyTraceID); exists {
		if tid, ok := traceID.(string); ok && tid != "" {
			return tid
		}
	}

	// 尝试从请求头获取
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		return traceID
	}

	return "" // 如果没有找到追踪 ID，则返回空字符串
}

// buildStandardResponse 构建标准响应
func buildStandardResponse[T any](c *gin.Context, code int, message string, data T) StandardResponse[T] {
	return StandardResponse[T]{
		Code:       code,
		Message:    message,
		Data:       data,
		TraceID:    getTraceID(c),
		ServerTime: time.Now(),
	}
}

// calculatePages 计算总页数
func calculatePages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}

	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}

	return pages
}

// buildPaginatedResponse 构建分页响应
func buildPaginatedResponse[T any](list []T, total int64, pageNum, pageSize int) PaginatedResponse[T] {
	if pageSize <= 0 {
		pageSize = 10 // 默认每页 10 条
	}

	if pageNum <= 0 {
		pageNum = 1 // 默认第一页
	}

	pages := calculatePages(total, pageSize)

	return PaginatedResponse[T]{
		List:     list,
		Total:    total,
		PageNum:  pageNum,
		PageSize: pageSize,
		Pages:    pages,
		HasNext:  pageNum < pages,
		HasPrev:  pageNum > 1,
	}
}

// MapBusinessCodeToHTTPStatus 将业务码映射到 HTTP 状态码
func MapBusinessCodeToHTTPStatus(businessCode int) int {
	switch businessCode {
	case CodeSuccess:
		return http.StatusOK
	case CodeInvalidParams:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeInternalError:
		return http.StatusInternalServerError
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeBadGateway:
		return http.StatusBadGateway
	default:
		// 对于未定义的业务码，如果是以 4 开头，返回 400，以 5 开头返回 500
		if businessCode >= 4000 && businessCode < 5000 {
			return http.StatusBadRequest
		} else if businessCode >= 5000 {
			return http.StatusInternalServerError
		}
		return http.StatusOK // 默认返回 200
	}
}

// GetDefaultMessage 根据业务码获取默认消息
func GetDefaultMessage(businessCode int) string {
	switch businessCode {
	case CodeSuccess:
		return DefaultSuccessMessage
	case CodeInvalidParams:
		return "参数错误"
	case CodeUnauthorized:
		return "未授权"
	case CodeForbidden:
		return "禁止访问"
	case CodeNotFound:
		return "资源不存在"
	case CodeInternalError:
		return "服务器内部错误"
	case CodeServiceUnavailable:
		return "服务不可用"
	case CodeTimeout:
		return "请求超时"
	case CodeTooManyRequests:
		return "请求过多，请稍后再试"
	case CodeBadGateway:
		return "网关错误"
	default:
		return DefaultErrorMessage
	}
}

// HTML 模板辅助函数类型
type HTMLHelperFunc func(c *gin.Context, data gin.H) gin.H

// 全局 HTML 辅助函数列表
var htmlHelperFuncs []HTMLHelperFunc

// RegisterHTMLHelper 注册 HTML 辅助函数
func RegisterHTMLHelper(fn HTMLHelperFunc) {
	htmlHelperFuncs = append(htmlHelperFuncs, fn)
}

// applyHTMLHelpers 应用所有 HTML 辅助函数
func applyHTMLHelpers(c *gin.Context, data gin.H) gin.H {
	result := make(gin.H)

	// 复制原始数据
	for k, v := range data {
		result[k] = v
	}

	// 应用辅助函数
	for _, fn := range htmlHelperFuncs {
		helper := fn(c, result)
		for k, v := range helper {
			result[k] = v
		}
	}

	return result
}
