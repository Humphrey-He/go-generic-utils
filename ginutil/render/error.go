package render

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationErrors 表示验证错误信息
type ValidationErrors map[string]string

// ErrWithCode 表示一个带有业务码的错误
type ErrWithCode struct {
	// Code 业务错误码
	Code int

	// Message 错误消息
	Message string

	// OriginalError 原始错误
	OriginalError error
}

// Error 实现 error 接口
func (e *ErrWithCode) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("错误码: %d, 消息: %s, 原因: %s", e.Code, e.Message, e.OriginalError.Error())
	}
	return fmt.Sprintf("错误码: %d, 消息: %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *ErrWithCode) Unwrap() error {
	return e.OriginalError
}

// NewError 创建一个带有业务码的错误
func NewError(code int, message string) error {
	return &ErrWithCode{
		Code:    code,
		Message: message,
	}
}

// WrapError 包装一个错误
func WrapError(code int, message string, err error) error {
	return &ErrWithCode{
		Code:          code,
		Message:       message,
		OriginalError: err,
	}
}

// HandleError 处理错误并发送响应
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var errWithCode *ErrWithCode
	if errors.As(err, &errWithCode) {
		// 如果是带有业务码的错误，使用其业务码和消息
		Error(c, errWithCode.Code, errWithCode.Message)
		return
	}

	// 处理验证错误
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		validationErrs := HandleValidationErrors(validationErrors)
		ValidationError(c, validationErrs)
		return
	}

	// 其他错误，默认为内部错误
	InternalError(c, err.Error())
}

// HandleValidationErrors 处理验证错误
func HandleValidationErrors(errs validator.ValidationErrors) ValidationErrors {
	result := make(ValidationErrors)

	for _, err := range errs {
		// 获取字段名和错误信息
		field := ToSnakeCase(err.Field())
		message := getValidationErrorMessage(err)

		result[field] = message
	}

	return result
}

// getValidationErrorMessage 获取验证错误消息
func getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "必填字段"
	case "email":
		return "无效的邮箱格式"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("长度不能小于 %s", err.Param())
		}
		return fmt.Sprintf("不能小于 %s", err.Param())
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("长度不能大于 %s", err.Param())
		}
		return fmt.Sprintf("不能大于 %s", err.Param())
	case "len":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("长度必须等于 %s", err.Param())
		}
		return fmt.Sprintf("必须等于 %s", err.Param())
	case "alphanum":
		return "只能包含字母和数字"
	case "oneof":
		return fmt.Sprintf("必须是 [%s] 其中之一", err.Param())
	}

	return fmt.Sprintf("验证失败: %s", err.Tag())
}

// ToSnakeCase 将驼峰命名转换为蛇形命名
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ErrorHandler 返回一个处理错误的中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 执行后续处理器
		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			HandleError(c, err)
			return
		}
	}
}

// PanicHandler 返回一个处理 panic 的中间件
func PanicHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var message string

				// 尝试转换为字符串
				switch v := err.(type) {
				case string:
					message = v
				case error:
					message = v.Error()
				default:
					message = "发生未知错误"
				}

				// 返回 500 响应
				Error(c, CodeInternalError, message, http.StatusInternalServerError)

				// 终止处理链
				c.Abort()
			}
		}()

		// 执行后续处理器
		c.Next()
	}
}

// BindAndValidate 绑定请求参数并进行验证
func BindAndValidate(c *gin.Context, obj interface{}) error {
	// 根据请求方法和内容类型选择合适的绑定方法
	var err error

	if c.Request.Method == http.MethodGet {
		err = c.ShouldBindQuery(obj)
	} else if c.ContentType() == "application/json" {
		err = c.ShouldBindJSON(obj)
	} else if c.ContentType() == "application/xml" {
		err = c.ShouldBindXML(obj)
	} else {
		err = c.ShouldBind(obj)
	}

	if err != nil {
		return err
	}

	return nil
}

// ValidateAndHandle 验证请求参数并处理错误
func ValidateAndHandle(c *gin.Context, obj interface{}) bool {
	if err := BindAndValidate(c, obj); err != nil {
		HandleError(c, err)
		return false
	}

	return true
}
