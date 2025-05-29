package binding

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrUnsupportedContentType is returned when the content type is not supported
	ErrUnsupportedContentType = errors.New("unsupported content type")

	// ErrBindingFailed is returned when binding fails for a generic reason
	ErrBindingFailed = errors.New("binding failed")

	// ErrInvalidTarget is returned when the binding target is not a pointer
	ErrInvalidTarget = errors.New("binding target must be a pointer")

	// ErrNilContext is returned when the gin context is nil
	ErrNilContext = errors.New("gin context is nil")
)

// BindingError wraps a binding error with additional context
type BindingError struct {
	Err     error  // Original error
	Message string // User-friendly message
}

// Error implements the error interface
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
func (e *BindingError) Unwrap() error {
	return e.Err
}

// FieldError describes a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`           // Field name (JSON name or struct field name)
	Tag     string `json:"tag"`             // Validation tag that failed (e.g., required, email, min)
	Message string `json:"message"`         // User-readable error message
	Value   any    `json:"value,omitempty"` // Value that caused the error (optional)
}

// FieldErrors is a slice of FieldError that implements the error interface
type FieldErrors []FieldError

// Error implements the error interface
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
func (fe FieldErrors) HasErrors() bool {
	return len(fe) > 0
}

// Add adds a field error to the collection
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
func (fe *FieldErrors) AddError(err FieldError) {
	*fe = append(*fe, err)
}

// NewBindingError creates a new BindingError
func NewBindingError(err error, message string) error {
	return &BindingError{
		Err:     err,
		Message: message,
	}
}
