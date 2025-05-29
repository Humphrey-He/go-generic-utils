// Package binding provides enhanced request binding utilities for Gin applications.
//
// This package extends Gin's binding capabilities with generic type support,
// making request binding type-safe and reducing boilerplate code. It offers
// a set of generic binding functions for different data sources (JSON, Query, URI, Form, XML)
// along with enhanced validation and error reporting.
//
// Key features:
//   - Generic binding functions for type-safe request binding
//   - Enhanced validation with detailed field-level error reporting
//   - Custom validation rules and messages
//   - Extensible decoder system for different content types
//
// Example usage:
//
//	type UserRequest struct {
//	    Name  string `json:"name" binding:"required"`
//	    Email string `json:"email" binding:"required,email"`
//	}
//
//	func CreateUser(c *gin.Context) {
//	    var req UserRequest
//	    if errs := binding.BindJSON(&req, c); len(errs) > 0 {
//	        c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
//	        return
//	    }
//	    // Process the validated request
//	}
//
// Relationship with Gin's native binding:
// This package builds upon Gin's native binding functionality but enhances it with:
//   - Generic type support for compile-time type safety
//   - More detailed error reporting through FieldErrors
//   - Consistent API across different binding sources
//   - Extensibility through custom validators and decoders
package binding
