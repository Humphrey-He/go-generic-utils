package binding

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// Validator defines the interface for data validation
// Validator 定义了数据验证的接口
type Validator interface {
	// Validate performs validation on the provided object
	// Returns FieldErrors if validation fails, nil or empty FieldErrors otherwise
	//
	// Validate 对提供的对象执行验证
	// 如果验证失败，返回 FieldErrors；否则返回 nil 或空的 FieldErrors
	Validate(obj interface{}) FieldErrors

	// Engine returns the underlying validation engine
	// Engine 返回底层验证引擎
	Engine() interface{}
}

// defaultValidator is the default implementation of Validator interface
// defaultValidator 是 Validator 接口的默认实现
type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate

	// tagNameFunc is used to determine the field name from struct tags
	// tagNameFunc 用于从结构体标签中确定字段名称
	tagNameFunc func(field reflect.StructField) string

	// customMessages maps validation tags to custom error messages
	// customMessages 将验证标签映射到自定义错误消息
	customMessages map[string]string
}

// DefaultValidator is the default validator instance
// DefaultValidator 是默认的验证器实例
var DefaultValidator Validator = &defaultValidator{
	customMessages: make(map[string]string),
}

// lazyInit initializes the validator instance once
// lazyInit 一次性初始化验证器实例
func (v *defaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// Use JSON tag name as the field name by default
		// 默认使用 JSON 标签名称作为字段名
		v.tagNameFunc = func(field reflect.StructField) string {
			name := field.Tag.Get("json")
			if name == "" {
				return field.Name
			}

			// Handle cases like `json:"name,omitempty"`
			// 处理类似 `json:"name,omitempty"` 的情况
			parts := strings.SplitN(name, ",", 2)
			return parts[0]
		}

		// Register the function to get the field name
		// 注册获取字段名称的函数
		v.validate.RegisterTagNameFunc(v.tagNameFunc)
	})
}

// Engine returns the underlying validator engine
// Engine 返回底层验证引擎
func (v *defaultValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

// Validate performs validation on the provided object
// Validate 对提供的对象执行验证
func (v *defaultValidator) Validate(obj interface{}) FieldErrors {
	v.lazyInit()

	// Check if obj is nil
	// 检查对象是否为 nil
	if obj == nil {
		return nil
	}

	// Perform validation
	// 执行验证
	err := v.validate.Struct(obj)
	if err == nil {
		return nil
	}

	// Convert validation errors to FieldErrors
	// 将验证错误转换为 FieldErrors
	fieldErrors := make(FieldErrors, 0)

	// Type assertion for validator.ValidationErrors
	// 对 validator.ValidationErrors 进行类型断言
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// If not a ValidationErrors type, return a generic error
		// 如果不是 ValidationErrors 类型，返回一个通用错误
		fieldErrors.Add("", "", err.Error())
		return fieldErrors
	}

	// Process each validation error
	// 处理每个验证错误
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		tag := fieldError.Tag()

		// Get custom message if available
		// 获取自定义消息（如果可用）
		message := v.getCustomMessage(tag, field)
		if message == "" {
			message = v.defaultMessage(fieldError)
		}

		fieldErrors.Add(field, tag, message, fieldError.Value())
	}

	return fieldErrors
}

// GetValidationErrors 从错误中提取验证错误。
// 如果错误是 validator.ValidationErrors 类型，返回 true 和验证错误列表。
// 否则返回 false 和 nil。
//
// GetValidationErrors extracts validation errors from an error.
// If the error is of type validator.ValidationErrors, returns true and the list of validation errors.
// Otherwise returns false and nil.
func GetValidationErrors(err error) ([]validator.FieldError, bool) {
	if err == nil {
		return nil, false
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, false
	}

	return validationErrors, true
}

// defaultMessage generates a default error message for a validation error
// defaultMessage 为验证错误生成默认错误消息
func (v *defaultValidator) defaultMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value must be greater than or equal to " + fieldError.Param()
	case "max":
		return "Value must be less than or equal to " + fieldError.Param()
	case "len":
		return "Length must be " + fieldError.Param()
	default:
		return "Validation failed on the '" + fieldError.Tag() + "' tag"
	}
}

// getCustomMessage returns a custom error message for a specific tag and field
// getCustomMessage 返回特定标签和字段的自定义错误消息
func (v *defaultValidator) getCustomMessage(tag, field string) string {
	// Try field-specific message first
	// 首先尝试字段特定的消息
	if msg, ok := v.customMessages[field+"."+tag]; ok {
		return msg
	}

	// Fall back to tag-only message
	// 回退到仅标签的消息
	return v.customMessages[tag]
}

// SetTagNameFunc sets the function used to determine field names
// SetTagNameFunc 设置用于确定字段名称的函数
func (v *defaultValidator) SetTagNameFunc(fn func(field reflect.StructField) string) {
	v.tagNameFunc = fn
	if v.validate != nil {
		v.validate.RegisterTagNameFunc(fn)
	}
}

// RegisterCustomMessage registers a custom error message for a validation tag
// The key can be either just the tag name (e.g., "required") for a global message
// or a field-specific message (e.g., "Email.required")
//
// RegisterCustomMessage 为验证标签注册自定义错误消息
// 键可以是标签名（例如 "required"）用于全局消息，
// 也可以是字段特定的消息（例如 "Email.required"）
func (v *defaultValidator) RegisterCustomMessage(key, message string) {
	v.customMessages[key] = message
}

// RegisterValidation registers a custom validation function
// RegisterValidation 注册自定义验证函数
func (v *defaultValidator) RegisterValidation(tag string, fn validator.Func) error {
	v.lazyInit()
	return v.validate.RegisterValidation(tag, fn)
}

// SetValidator sets the global validator instance
// SetValidator 设置全局验证器实例
func SetValidator(val Validator) {
	if v, ok := val.(*defaultValidator); ok {
		DefaultValidator = v
	} else {
		// If it's not a defaultValidator, create a wrapper that delegates to the provided validator
		// 如果不是 defaultValidator，创建一个委托给提供的验证器的包装器
		DefaultValidator = &validatorWrapper{val}
	}
}

// validatorWrapper wraps a custom validator implementation
// validatorWrapper 包装自定义验证器实现
type validatorWrapper struct {
	v Validator
}

func (w *validatorWrapper) Validate(obj interface{}) FieldErrors {
	return w.v.Validate(obj)
}

func (w *validatorWrapper) Engine() interface{} {
	return w.v.Engine()
}
