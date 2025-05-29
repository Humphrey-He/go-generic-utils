package binding

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// BindJSON binds JSON data from the request body to the provided object
// and performs validation.
func BindJSON[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	if err := c.ShouldBindJSON(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	return DefaultValidator.Validate(obj)
}

// BindQuery binds URL query parameters to the provided object
// and performs validation.
func BindQuery[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	if err := c.ShouldBindQuery(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	return DefaultValidator.Validate(obj)
}

// BindURI binds URI parameters to the provided object
// and performs validation.
func BindURI[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	if err := c.ShouldBindUri(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	return DefaultValidator.Validate(obj)
}

// BindForm binds form data to the provided object
// and performs validation.
func BindForm[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	return DefaultValidator.Validate(obj)
}

// BindXML binds XML data from the request body to the provided object
// and performs validation.
func BindXML[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	if err := c.ShouldBindXML(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	return DefaultValidator.Validate(obj)
}

// Bind automatically selects the appropriate binding method based on Content-Type
// and binds the request data to the provided object.
func Bind[T any](c *gin.Context, obj *T) FieldErrors {
	if c == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrNilContext.Error()}}
	}

	if obj == nil {
		return FieldErrors{{Field: "", Tag: "", Message: ErrInvalidTarget.Error()}}
	}

	// Use Gin's binding
	if err := c.ShouldBind(obj); err != nil {
		return convertBindingError(err)
	}

	// Validate the object
	return DefaultValidator.Validate(obj)
}

// convertBindingError converts a standard error to FieldErrors
func convertBindingError(err error) FieldErrors {
	if err == nil {
		return nil
	}

	// Check if it's already a FieldErrors
	if fieldErrors, ok := err.(FieldErrors); ok {
		return fieldErrors
	}

	// Try to convert validator.ValidationErrors
	if DefaultValidator != nil {
		if engine := DefaultValidator.Engine(); engine != nil {
			// This would be implementation-specific based on the validator
			// For now, we'll just create a generic error
		}
	}

	// Create a generic binding error
	return FieldErrors{{
		Field:   "",
		Tag:     "binding_failed",
		Message: err.Error(),
	}}
}
