package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success 发送成功的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - data: 响应数据，可以是任意类型
//   - message: 可选的成功消息，不提供则使用默认成功消息
func Success[T any](c *gin.Context, data T, message ...string) {
	msg := DefaultSuccessMessage
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := buildStandardResponse(c, CodeSuccess, msg, data)

	if JSONPrettyPrint {
		c.IndentedJSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// Error 发送错误的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - businessCode: 业务错误码
//   - errorMessage: 错误消息，如果为空则使用默认错误消息
//   - httpStatusCode: 可选的 HTTP 状态码，不提供则根据业务码自动映射
func Error(c *gin.Context, businessCode int, errorMessage string, httpStatusCode ...int) {
	if errorMessage == "" {
		errorMessage = GetDefaultMessage(businessCode)
	}

	response := buildStandardResponse(c, businessCode, errorMessage, gin.H{})

	status := MapBusinessCodeToHTTPStatus(businessCode)
	if len(httpStatusCode) > 0 && httpStatusCode[0] > 0 {
		status = httpStatusCode[0]
	}

	if JSONPrettyPrint {
		c.AbortWithStatusJSON(status, response)
	} else {
		c.AbortWithStatusJSON(status, response)
	}
}

// ErrorWithData 发送带数据的错误 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - businessCode: 业务错误码
//   - errorMessage: 错误消息，如果为空则使用默认错误消息
//   - data: 错误相关的数据
//   - httpStatusCode: 可选的 HTTP 状态码，不提供则根据业务码自动映射
func ErrorWithData[D any](c *gin.Context, businessCode int, errorMessage string, data D, httpStatusCode ...int) {
	if errorMessage == "" {
		errorMessage = GetDefaultMessage(businessCode)
	}

	response := buildStandardResponse(c, businessCode, errorMessage, data)

	status := MapBusinessCodeToHTTPStatus(businessCode)
	if len(httpStatusCode) > 0 && httpStatusCode[0] > 0 {
		status = httpStatusCode[0]
	}

	if JSONPrettyPrint {
		c.AbortWithStatusJSON(status, response)
	} else {
		c.AbortWithStatusJSON(status, response)
	}
}

// Paginated 发送分页的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - list: 数据列表
//   - total: 总记录数
//   - pageNum: 当前页码
//   - pageSize: 每页记录数
//   - message: 可选的成功消息，不提供则使用默认成功消息
func Paginated[T any](c *gin.Context, list []T, total int64, pageNum, pageSize int, message ...string) {
	msg := DefaultSuccessMessage
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	paginatedData := buildPaginatedResponse(list, total, pageNum, pageSize)
	response := buildStandardResponse(c, CodeSuccess, msg, paginatedData)

	if JSONPrettyPrint {
		c.IndentedJSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// Custom 发送自定义的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - businessCode: 业务码
//   - message: 消息内容
//   - data: 响应数据
//   - httpStatusCode: HTTP 状态码
func Custom[T any](c *gin.Context, businessCode int, message string, data T, httpStatusCode int) {
	response := buildStandardResponse(c, businessCode, message, data)

	if JSONPrettyPrint {
		c.IndentedJSON(httpStatusCode, response)
	} else {
		c.JSON(httpStatusCode, response)
	}
}

// NoContent 发送无内容的响应
// 参数:
//   - c: Gin 上下文
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Unauthorized 发送未授权的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认未授权消息
func Unauthorized(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeUnauthorized)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	Error(c, CodeUnauthorized, msg, http.StatusUnauthorized)
}

// Forbidden 发送禁止访问的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认禁止访问消息
func Forbidden(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeForbidden)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	Error(c, CodeForbidden, msg, http.StatusForbidden)
}

// NotFound 发送资源不存在的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认资源不存在消息
func NotFound(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeNotFound)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	Error(c, CodeNotFound, msg, http.StatusNotFound)
}

// BadRequest 发送参数错误的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认参数错误消息
func BadRequest(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeInvalidParams)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	Error(c, CodeInvalidParams, msg, http.StatusBadRequest)
}

// InternalError 发送内部错误的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - message: 可选的错误消息，不提供则使用默认内部错误消息
func InternalError(c *gin.Context, message ...string) {
	msg := GetDefaultMessage(CodeInternalError)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	Error(c, CodeInternalError, msg, http.StatusInternalServerError)
}

// ValidationError 发送验证错误的 JSON 响应
// 参数:
//   - c: Gin 上下文
//   - errors: 验证错误信息
func ValidationError(c *gin.Context, errors interface{}) {
	ErrorWithData(c, CodeInvalidParams, GetDefaultMessage(CodeInvalidParams), errors, http.StatusBadRequest)
}
