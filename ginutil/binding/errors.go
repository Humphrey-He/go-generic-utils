package binding

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrUnsupportedContentType is returned when the content type is not supported
	// ErrUnsupportedContentType 在内容类型不受支持时返回
	ErrUnsupportedContentType = errors.New("unsupported content type")

	// ErrBindingFailed is returned when binding fails for a generic reason
	// ErrBindingFailed 在由于通用原因导致绑定失败时返回
	ErrBindingFailed = errors.New("binding failed")

	// ErrInvalidTarget is returned when the binding target is not a pointer
	// ErrInvalidTarget 在绑定目标不是指针时返回
	ErrInvalidTarget = errors.New("binding target must be a pointer")

	// ErrNilContext is returned when the gin context is nil
	// ErrNilContext 在 gin 上下文为 nil 时返回
	ErrNilContext = errors.New("gin context is nil")
)

// BindingError wraps a binding error with additional context
// BindingError 使用附加上下文包装绑定错误
type BindingError struct {
	Err     error  // Original error 原始错误
	Message string // User-friendly message 用户友好的消息
}

// Error implements the error interface
// Error 实现 error 接口
func (e *BindingError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "binding error"
}

// Unwrap returns the wrapped error
// Unwrap 返回被包装的错误
func (e *BindingError) Unwrap() error {
	return e.Err
}

// FieldError describes a validation error for a specific field
// FieldError 描述特定字段的验证错误
type FieldError struct {
	Field   string `json:"field"`           // Field name (JSON name or struct field name) 字段名称（JSON 名称或结构体字段名称）
	Tag     string `json:"tag"`             // Validation tag that failed (e.g., required, email, min) 失败的验证标签（例如，required、email、min）
	Message string `json:"message"`         // User-readable error message 用户可读的错误消息
	Value   any    `json:"value,omitempty"` // Value that caused the error (optional) 导致错误的值（可选）
}

// FieldErrors is a slice of FieldError that implements the error interface
// FieldErrors 是 FieldError 的切片，实现了 error 接口
type FieldErrors []FieldError

// Error implements the error interface
// Error 实现 error 接口
func (fe FieldErrors) Error() string {
	if len(fe) == 0 {
		return ""
	}

	if len(fe) == 1 {
		return fmt.Sprintf("%s: %s", fe[0].Field, fe[0].Message)
	}

	var sb strings.Builder
	sb.WriteString("validation failed: ")

	for i, err := range fe {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", err.Field, err.Message))
		if i >= 2 && len(fe) > 3 {
			sb.WriteString(fmt.Sprintf("; and %d more errors", len(fe)-i-1))
			break
		}
	}

	return sb.String()
}

// HasErrors returns true if there are any field errors
// HasErrors 如果存在任何字段错误，则返回 true
func (fe FieldErrors) HasErrors() bool {
	return len(fe) > 0
}

// Add adds a field error to the collection
// Add 向集合中添加字段错误
func (fe *FieldErrors) Add(field, tag, message string, value ...any) {
	var val any
	if len(value) > 0 {
		val = value[0]
	}

	*fe = append(*fe, FieldError{
		Field:   field,
		Tag:     tag,
		Message: message,
		Value:   val,
	})
}

// AddError adds a FieldError to the collection
// AddError 向集合中添加 FieldError
func (fe *FieldErrors) AddError(err FieldError) {
	*fe = append(*fe, err)
}

// NewBindingError creates a new BindingError
// NewBindingError 创建一个新的 BindingError
func NewBindingError(err error, message string) error {
	return &BindingError{
		Err:     err,
		Message: message,
	}
}
