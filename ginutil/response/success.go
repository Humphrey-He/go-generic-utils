package response

import (
	"fmt"
	"net/http"

	"github.com/noobtrump/go-generic-utils/ginutil/ecode"

	"github.com/gin-gonic/gin"
)

// OK 发送一个标准的成功响应，HTTP状态码为200。
func OK[T any](c *gin.Context, data T) {
	resp := StandardResponse[T]{
		Code:       ecode.OK,
		Message:    ecode.SuccessMessage,
		Data:       data,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendJSON(c, http.StatusOK, resp)
}

// OKWithMessage 发送一个带有自定义消息的标准成功响应，HTTP状态码为200。
func OKWithMessage[T any](c *gin.Context, data T, message string) {
	resp := StandardResponse[T]{
		Code:       ecode.OK,
		Message:    message,
		Data:       data,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendJSON(c, http.StatusOK, resp)
}

// Created 发送一个资源创建成功的响应，HTTP状态码为201。
// 通常用于POST请求成功创建资源后的响应。
func Created[T any](c *gin.Context, data T) {
	resp := StandardResponse[T]{
		Code:       ecode.OK,
		Message:    "资源创建成功",
		Data:       data,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendJSON(c, http.StatusCreated, resp)
}

// Accepted 发送一个请求已接受但尚未处理完成的响应，HTTP状态码为202。
// 通常用于异步任务的接受确认。
func Accepted[T any](c *gin.Context, data T) {
	resp := StandardResponse[T]{
		Code:       ecode.OK,
		Message:    "请求已接受，正在处理中",
		Data:       data,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendJSON(c, http.StatusAccepted, resp)
}

// NoContent 发送一个无内容响应，HTTP状态码为204。
// 通常用于删除操作成功后的响应。
func NoContent(c *gin.Context) {
	// 虽然204状态码不应该有响应体，但我们仍然设置TraceID和ServerTime等信息到header中
	// 以便于日志记录和调试
	resp := StandardResponse[any]{
		Code:       ecode.OK,
		Message:    "操作成功",
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}

	// 手动设置需要的headers
	c.Header("X-Trace-ID", resp.TraceID)
	c.Header("X-Server-Time", fmt.Sprintf("%d", resp.ServerTime))
	c.Status(http.StatusNoContent)
}
