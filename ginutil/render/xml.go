package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessXML 发送成功的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - data: 响应数据，可以是任意类型
//   - message: 可选的成功消息，不提供则使用默认成功消息
func SuccessXML[T any](c *gin.Context, data T, message ...string) {
	msg := DefaultSuccessMessage
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := buildStandardResponse(c, CodeSuccess, msg, data)
	c.XML(http.StatusOK, response)
}

// ErrorXML 发送错误的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - businessCode: 业务错误码
//   - errorMessage: 错误消息，如果为空则使用默认错误消息
//   - httpStatusCode: 可选的 HTTP 状态码，不提供则根据业务码自动映射
func ErrorXML(c *gin.Context, businessCode int, errorMessage string, httpStatusCode ...int) {
	if errorMessage == "" {
		errorMessage = GetDefaultMessage(businessCode)
	}

	response := buildStandardResponse(c, businessCode, errorMessage, gin.H{})

	status := MapBusinessCodeToHTTPStatus(businessCode)
	if len(httpStatusCode) > 0 && httpStatusCode[0] > 0 {
		status = httpStatusCode[0]
	}

	c.XML(status, response)
	c.Abort()
}

// ErrorXMLWithData 发送带数据的错误 XML 响应
// 参数:
//   - c: Gin 上下文
//   - businessCode: 业务错误码
//   - errorMessage: 错误消息，如果为空则使用默认错误消息
//   - data: 错误相关的数据
//   - httpStatusCode: 可选的 HTTP 状态码，不提供则根据业务码自动映射
func ErrorXMLWithData[D any](c *gin.Context, businessCode int, errorMessage string, data D, httpStatusCode ...int) {
	if errorMessage == "" {
		errorMessage = GetDefaultMessage(businessCode)
	}

	response := buildStandardResponse(c, businessCode, errorMessage, data)

	status := MapBusinessCodeToHTTPStatus(businessCode)
	if len(httpStatusCode) > 0 && httpStatusCode[0] > 0 {
		status = httpStatusCode[0]
	}

	c.XML(status, response)
	c.Abort()
}

// PaginatedXML 发送分页的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - list: 数据列表
//   - total: 总记录数
//   - pageNum: 当前页码
//   - pageSize: 每页记录数
//   - message: 可选的成功消息，不提供则使用默认成功消息
func PaginatedXML[T any](c *gin.Context, list []T, total int64, pageNum, pageSize int, message ...string) {
	msg := DefaultSuccessMessage
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	paginatedData := buildPaginatedResponse(list, total, pageNum, pageSize)
	response := buildStandardResponse(c, CodeSuccess, msg, paginatedData)

	c.XML(http.StatusOK, response)
}

// CustomXML 发送自定义的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - businessCode: 业务码
//   - message: 消息内容
//   - data: 响应数据
//   - httpStatusCode: HTTP 状态码
func CustomXML[T any](c *gin.Context, businessCode int, message string, data T, httpStatusCode int) {
	response := buildStandardResponse(c, businessCode, message, data)
	c.XML(httpStatusCode, response)
}

// UnauthorizedXML 发送未授权的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认未授权消息
func UnauthorizedXML(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeUnauthorized)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	ErrorXML(c, CodeUnauthorized, msg, http.StatusUnauthorized)
}

// ForbiddenXML 发送禁止访问的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认禁止访问消息
func ForbiddenXML(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeForbidden)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	ErrorXML(c, CodeForbidden, msg, http.StatusForbidden)
}

// NotFoundXML 发送资源不存在的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认资源不存在消息
func NotFoundXML(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeNotFound)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	ErrorXML(c, CodeNotFound, msg, http.StatusNotFound)
}

// BadRequestXML 发送参数错误的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认参数错误消息
func BadRequestXML(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeInvalidParams)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	ErrorXML(c, CodeInvalidParams, msg, http.StatusBadRequest)
}

// InternalErrorXML 发送内部错误的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认内部错误消息
func InternalErrorXML(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeInternalError)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	ErrorXML(c, CodeInternalError, msg, http.StatusInternalServerError)
}

// ValidationErrorXML 发送验证错误的 XML 响应
// 参数:
//   - c: Gin 上下文
//   - errors: 验证错误信息
func ValidationErrorXML(c *gin.Context, errors interface{}) {
	ErrorXMLWithData(c, CodeInvalidParams, GetDefaultMessage(CodeInvalidParams), errors, http.StatusBadRequest)
}
