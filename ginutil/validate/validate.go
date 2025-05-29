// Package validate 提供基于 go-playground/validator 的增强校验功能。
// 它支持结构体校验、自定义校验规则和本地化错误消息。
package validate

import (
	"reflect"
	"strings"
	"sync"

	"github.com/noobtrump/go-generic-utils/ginutil/binding"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidatorV10 是对 go-playground/validator/v10 的封装，提供增强的校验功能。
type ValidatorV10 struct {
	validate *validator.Validate // 底层校验器实例
	once     sync.Once           // 确保只初始化一次

	// 用于获取字段名称的函数
	tagNameFunc func(field reflect.StructField) string

	// 消息存储，用于本地化错误消息
	messageStore *MessageStore
}

// 全局默认校验器实例
var defaultValidator *ValidatorV10

// 初始化默认校验器
func init() {
	defaultValidator = NewV10()

	// 设置默认的中文错误消息
	defaultValidator.messageStore = NewDefaultMessageStore()

	// 注册内置的自定义校验规则
	registerBuiltinRules(defaultValidator)
}

// NewV10 创建一个新的 ValidatorV10 实例。
// 如果提供了 v 参数，则使用该校验器实例；否则创建一个新的实例。
func NewV10(v ...*validator.Validate) *ValidatorV10 {
	validator := &ValidatorV10{
		messageStore: NewDefaultMessageStore(),
	}

	if len(v) > 0 && v[0] != nil {
		validator.validate = v[0]
	}

	return validator
}

// lazyInit 延迟初始化校验器。
func (v *ValidatorV10) lazyInit() {
	v.once.Do(func() {
		if v.validate == nil {
			v.validate = validator.New()
		}

		// 使用 JSON 标签名作为字段名
		v.tagNameFunc = func(field reflect.StructField) string {
			name := field.Tag.Get("json")
			if name == "" {
				return field.Name
			}

			// 处理 `json:"name,omitempty"` 格式
			parts := strings.SplitN(name, ",", 2)
			return parts[0]
		}

		// 注册获取字段名的函数
		v.validate.RegisterTagNameFunc(v.tagNameFunc)
	})
}

// Engine 返回底层的 validator.Validate 实例。
func (v *ValidatorV10) Engine() *validator.Validate {
	v.lazyInit()
	return v.validate
}

// RegisterValidation 注册一个自定义校验函数。
func (v *ValidatorV10) RegisterValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	v.lazyInit()
	return v.validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

// SetTagNameFunc 设置用于获取字段名的函数。
func (v *ValidatorV10) SetTagNameFunc(fn func(field reflect.StructField) string) {
	v.lazyInit()
	v.tagNameFunc = fn
	v.validate.RegisterTagNameFunc(fn)
}

// SetMessageStore 设置消息存储。
func (v *ValidatorV10) SetMessageStore(ms *MessageStore) {
	v.messageStore = ms
}

// Validate 校验结构体，返回 binding.FieldErrors。
func (v *ValidatorV10) Validate(obj interface{}) binding.FieldErrors {
	v.lazyInit()

	// 检查 obj 是否为 nil
	if obj == nil {
		return nil
	}

	// 执行校验
	err := v.validate.Struct(obj)
	if err == nil {
		return nil
	}

	// 转换校验错误为 FieldErrors
	fieldErrors := make(binding.FieldErrors, 0)

	// 类型断言为 validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// 如果不是 ValidationErrors 类型，返回一个通用错误
		fieldErrors.Add("", "", err.Error())
		return fieldErrors
	}

	// 处理每个校验错误
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		tag := fieldError.Tag()

		// 获取本地化的错误消息
		message := v.messageStore.GetMessage(fieldError, "zh")
		if message == "" {
			message = v.defaultMessage(fieldError)
		}

		fieldErrors.Add(field, tag, message, fieldError.Value())
	}

	return fieldErrors
}

// defaultMessage 生成默认的错误消息。
func (v *ValidatorV10) defaultMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "此字段是必填的"
	case "email":
		return "无效的邮箱格式"
	case "min":
		return "值必须大于或等于 " + fieldError.Param()
	case "max":
		return "值必须小于或等于 " + fieldError.Param()
	case "len":
		return "长度必须等于 " + fieldError.Param()
	default:
		return "校验失败: '" + fieldError.Tag() + "'"
	}
}

// GetDefaultValidator 返回默认的校验器实例。
func GetDefaultValidator() *ValidatorV10 {
	return defaultValidator
}

// SetDefaultValidator 设置默认的校验器实例。
func SetDefaultValidator(v *ValidatorV10) {
	if v != nil {
		defaultValidator = v
	}
}

// ValidateStruct 使用默认校验器校验结构体。
// 这是一个泛型函数，可以接受任意类型的结构体。
func ValidateStruct[T any](obj T) binding.FieldErrors {
	return defaultValidator.Validate(obj)
}

// MustBindAndValidate 绑定并校验请求数据。
// 如果绑定或校验失败，则使用 response 包发送标准错误响应并调用 c.Abort()，返回 false。
// 成功返回 true。
func MustBindAndValidate[T any](c *gin.Context, obj *T) bool {
	if c == nil {
		return false
	}

	// 绑定请求数据
	if err := c.ShouldBind(obj); err != nil {
		// 处理绑定错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "请求数据绑定失败",
			"error":   err.Error(),
		})
		c.Abort()
		return false
	}

	// 校验数据
	if errs := ValidateStruct(*obj); errs != nil && errs.HasErrors() {
		// 处理校验错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "数据校验失败",
			"errors":  errs,
		})
		c.Abort()
		return false
	}

	return true
}

// MustBindJSONAndValidate 绑定 JSON 数据并校验。
func MustBindJSONAndValidate[T any](c *gin.Context, obj *T) bool {
	if c == nil {
		return false
	}

	// 绑定 JSON 数据
	if err := c.ShouldBindJSON(obj); err != nil {
		// 处理绑定错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "JSON 数据绑定失败",
			"error":   err.Error(),
		})
		c.Abort()
		return false
	}

	// 校验数据
	if errs := ValidateStruct(*obj); errs != nil && errs.HasErrors() {
		// 处理校验错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "数据校验失败",
			"errors":  errs,
		})
		c.Abort()
		return false
	}

	return true
}

// MustBindQueryAndValidate 绑定查询参数并校验。
func MustBindQueryAndValidate[T any](c *gin.Context, obj *T) bool {
	if c == nil {
		return false
	}

	// 绑定查询参数
	if err := c.ShouldBindQuery(obj); err != nil {
		// 处理绑定错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "查询参数绑定失败",
			"error":   err.Error(),
		})
		c.Abort()
		return false
	}

	// 校验数据
	if errs := ValidateStruct(*obj); errs != nil && errs.HasErrors() {
		// 处理校验错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "数据校验失败",
			"errors":  errs,
		})
		c.Abort()
		return false
	}

	return true
}

// MustBindURIAndValidate 绑定 URI 参数并校验。
func MustBindURIAndValidate[T any](c *gin.Context, obj *T) bool {
	if c == nil {
		return false
	}

	// 绑定 URI 参数
	if err := c.ShouldBindUri(obj); err != nil {
		// 处理绑定错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "URI 参数绑定失败",
			"error":   err.Error(),
		})
		c.Abort()
		return false
	}

	// 校验数据
	if errs := ValidateStruct(*obj); errs != nil && errs.HasErrors() {
		// 处理校验错误
		c.JSON(400, gin.H{
			"code":    400,
			"message": "数据校验失败",
			"errors":  errs,
		})
		c.Abort()
		return false
	}

	return true
}
