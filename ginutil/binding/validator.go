package binding

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

// Validator defines the interface for data validation
type Validator interface {
	// Validate performs validation on the provided object
	// Returns FieldErrors if validation fails, nil or empty FieldErrors otherwise
	Validate(obj interface{}) FieldErrors

	// Engine returns the underlying validation engine
	Engine() interface{}
}

// defaultValidator is the default implementation of Validator interface
type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate

	// tagNameFunc is used to determine the field name from struct tags
	tagNameFunc func(field reflect.StructField) string

	// customMessages maps validation tags to custom error messages
	customMessages map[string]string
}

// DefaultValidator is the default validator instance
var DefaultValidator Validator = &defaultValidator{
	customMessages: make(map[string]string),
}

// lazyInit initializes the validator instance once
func (v *defaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")

		// Use JSON tag name as the field name by default
		v.tagNameFunc = func(field reflect.StructField) string {
			name := field.Tag.Get("json")
			if name == "" {
				return field.Name
			}

			// Handle cases like `json:"name,omitempty"`
			parts := strings.SplitN(name, ",", 2)
			return parts[0]
		}

		// Register the function to get the field name
		v.validate.RegisterTagNameFunc(v.tagNameFunc)
	})
}

// Engine returns the underlying validator engine
func (v *defaultValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

// Validate performs validation on the provided object
func (v *defaultValidator) Validate(obj interface{}) FieldErrors {
	v.lazyInit()

	// Check if obj is nil
	if obj == nil {
		return nil
	}

	// Perform validation
	err := v.validate.Struct(obj)
	if err == nil {
		return nil
	}

	// Convert validation errors to FieldErrors
	fieldErrors := make(FieldErrors, 0)

	// Type assertion for validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// If not a ValidationErrors type, return a generic error
		fieldErrors.Add("", "", err.Error())
		return fieldErrors
	}

	// Process each validation error
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		tag := fieldError.Tag()

		// Get custom message if available
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
func (v *defaultValidator) getCustomMessage(tag, field string) string {
	// Try field-specific message first
	if msg, ok := v.customMessages[field+"."+tag]; ok {
		return msg
	}

	// Fall back to tag-only message
	return v.customMessages[tag]
}

// SetTagNameFunc sets the function used to determine field names
func (v *defaultValidator) SetTagNameFunc(fn func(field reflect.StructField) string) {
	v.tagNameFunc = fn
	if v.validate != nil {
		v.validate.RegisterTagNameFunc(fn)
	}
}

// RegisterCustomMessage registers a custom error message for a validation tag
// The key can be either just the tag name (e.g., "required") for a global message
// or a field-specific message (e.g., "Email.required")
func (v *defaultValidator) RegisterCustomMessage(key, message string) {
	v.customMessages[key] = message
}

// RegisterValidation registers a custom validation function
func (v *defaultValidator) RegisterValidation(tag string, fn validator.Func) error {
	v.lazyInit()
	return v.validate.RegisterValidation(tag, fn)
}

// SetValidator sets the global validator instance
func SetValidator(val Validator) {
	if v, ok := val.(*defaultValidator); ok {
		DefaultValidator = v
	} else {
		// If it's not a defaultValidator, create a wrapper that delegates to the provided validator
		DefaultValidator = &validatorWrapper{val}
	}
}

// validatorWrapper wraps a custom validator implementation
type validatorWrapper struct {
	v Validator
}

func (w *validatorWrapper) Validate(obj interface{}) FieldErrors {
	return w.v.Validate(obj)
}

func (w *validatorWrapper) Engine() interface{} {
	return w.v.Engine()
}
