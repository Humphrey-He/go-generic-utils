package binding

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// BindJSON binds JSON data from the request body to the provided object
// and performs validation.
//
// BindJSON 将请求体中的 JSON 数据绑定到指定对象并执行验证。
func BindJSON[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	// 使用 Gin 的绑定功能
	if err := c.ShouldBindJSON(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	// 验证对象
	return DefaultValidator.Validate(obj)
}

// BindQuery binds URL query parameters to the provided object
// and performs validation.
//
// BindQuery 将 URL 查询参数绑定到指定对象并执行验证。
func BindQuery[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	// 使用 Gin 的绑定功能
	if err := c.ShouldBindQuery(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	// 验证对象
	return DefaultValidator.Validate(obj)
}

// BindURI binds URI parameters to the provided object
// and performs validation.
//
// BindURI 将 URI 参数绑定到指定对象并执行验证。
func BindURI[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	// 使用 Gin 的绑定功能
	if err := c.ShouldBindUri(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	// 验证对象
	return DefaultValidator.Validate(obj)
}

// BindForm binds form data to the provided object
// and performs validation.
//
// BindForm 将表单数据绑定到指定对象并执行验证。
func BindForm[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	// 使用 Gin 的绑定功能
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	// 验证对象
	return DefaultValidator.Validate(obj)
}

// BindXML binds XML data from the request body to the provided object
// and performs validation.
//
// BindXML 将请求体中的 XML 数据绑定到指定对象并执行验证。
func BindXML[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	// 使用 Gin 的绑定功能
	if err := c.ShouldBindXML(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	// 验证对象
	return DefaultValidator.Validate(obj)
}

// Bind automatically selects the appropriate binding method based on Content-Type
// and binds the request data to the provided object.
//
// Bind 根据 Content-Type 自动选择适当的绑定方法，并将请求数据绑定到指定对象。
func Bind[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	// 使用 Gin 的绑定功能
	if err := c.ShouldBind(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	// 验证对象
	return DefaultValidator.Validate(obj)
}

// convertBindingError converts a standard error to FieldErrors
// convertBindingError 将标准错误转换为 FieldErrors
func convertBindingError(err error) FieldErrors {
	if err == nil {
		return nil
	}

	// Check if it's already a FieldErrors
	// 检查是否已经是 FieldErrors 类型
	if fieldErrors, ok := err.(FieldErrors); ok {
		return fieldErrors
	}

	// Try to convert validator.ValidationErrors
	// 尝试转换 validator.ValidationErrors
	if validationErrors, ok := GetValidationErrors(err); ok {
		fieldErrors := make(FieldErrors, 0, len(validationErrors))

		// Process each validation error
		// 处理每个验证错误
		for _, fieldError := range validationErrors {
			field := fieldError.Field()
			tag := fieldError.Tag()

			// Get message based on tag
			// 根据标签获取消息
			var message string
			switch tag {
			case "required":
				message = "This field is required"
			case "email":
				message = "Invalid email format"
			case "min":
				message = "Value must be greater than or equal to " + fieldError.Param()
			case "max":
				message = "Value must be less than or equal to " + fieldError.Param()
			case "len":
				message = "Length must be " + fieldError.Param()
			default:
				message = "Validation failed on the '" + tag + "' tag"
			}

			fieldErrors.Add(field, tag, message, fieldError.Value())
		}

		return fieldErrors
	}

	// Create a generic binding error
	// 创建一个通用的绑定错误
	return FieldErrors{{
		Field:   "",
		Tag:     "binding_failed",
		Message: err.Error(),
	}}
}
