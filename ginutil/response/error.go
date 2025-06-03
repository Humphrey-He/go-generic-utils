package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Humphrey-He/go-generic-utils/ginutil/ecode"

	"github.com/gin-gonic/gin"
)

// mapBusinessCodeToHTTPStatus 将业务错误码映射为HTTP状态码
func mapBusinessCodeToHTTPStatus(businessCode int) int {
	if businessCode == ecode.OK {
		return http.StatusOK
	}

	// 简化映射：4xxxx -> 4xx, 5xxxx -> 5xx, 7xxxx -> 502/503
	strCode := fmt.Sprintf("%d", businessCode)
	if strings.HasPrefix(strCode, "4") { // A类用户错误
		switch businessCode {
		case ecode.AccessUnauthorized:
			return http.StatusUnauthorized // 401
		case ecode.AccessPermissionDenied:
			return http.StatusForbidden // 403
		case ecode.ErrorCodeNotFound:
			return http.StatusNotFound // 404
		case ecode.ErrorCodeTooManyRequests:
			return http.StatusTooManyRequests // 429
		default:
			return http.StatusBadRequest // 400
		}
	} else if strings.HasPrefix(strCode, "5") { // B类系统错误
		return http.StatusInternalServerError // 500
	} else if strings.HasPrefix(strCode, "7") { // C类第三方错误
		return http.StatusBadGateway // 502
	}

	// 默认返回500
	return http.StatusInternalServerError
}

// Fail 发送一个标准的错误响应。它会将 businessCode 映射到合适的HTTP状态码。
func Fail(c *gin.Context, businessCode int, errorMessage string) {
	httpStatus := mapBusinessCodeToHTTPStatus(businessCode)
	resp := StandardResponse[any]{
		Code:       businessCode,
		Message:    errorMessage,
		Data:       nil,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendAbortJSON(c, httpStatus, resp)
}

// FailWithData 发送一个带有额外数据的标准错误响应。
func FailWithData[T any](c *gin.Context, businessCode int, errorMessage string, data T) {
	httpStatus := mapBusinessCodeToHTTPStatus(businessCode)
	resp := StandardResponse[T]{
		Code:       businessCode,
		Message:    errorMessage,
		Data:       data,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendAbortJSON(c, httpStatus, resp)
}

// BadRequest 发送一个400 Bad Request错误响应。
// 可选地提供错误详情。
func BadRequest(c *gin.Context, message string, errorDetails ...any) {
	var data any = nil
	if len(errorDetails) > 0 {
		data = errorDetails[0]
	}

	if message == "" {
		message = ecode.GetMessage(ecode.ErrorCodeBadRequest)
	}

	FailWithData(c, ecode.ErrorCodeBadRequest, message, data)
}

// Unauthorized 发送一个401 Unauthorized错误响应。
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.AccessUnauthorized)
	}
	Fail(c, ecode.AccessUnauthorized, message)
}

// Forbidden 发送一个403 Forbidden错误响应。
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.AccessPermissionDenied)
	}
	Fail(c, ecode.AccessPermissionDenied, message)
}

// NotFound 发送一个404 Not Found错误响应。
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.ErrorCodeNotFound)
	}
	Fail(c, ecode.ErrorCodeNotFound, message)
}

// InternalError 发送一个500 Internal Server Error错误响应。
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.ErrorCodeInternal)
	}
	Fail(c, ecode.ErrorCodeInternal, message)
}

// TooManyRequests 发送一个429 Too Many Requests错误响应。
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.ErrorCodeTooManyRequests)
	}
	Fail(c, ecode.ErrorCodeTooManyRequests, message)
}

// DatabaseError 发送一个数据库错误响应。
func DatabaseError(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.ErrorCodeDatabase)
	}
	Fail(c, ecode.ErrorCodeDatabase, message)
}

// ThirdPartyError 发送一个第三方服务错误响应。
func ThirdPartyError(c *gin.Context, message string) {
	if message == "" {
		message = ecode.GetMessage(ecode.ErrorCodeThirdParty)
	}
	Fail(c, ecode.ErrorCodeThirdParty, message)
}
