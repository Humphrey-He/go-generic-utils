package binding_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/noobtrump/go-generic-utils/ginutil/binding"

	"github.com/gin-gonic/gin"
)

// This example demonstrates how to use the binding package to bind and validate JSON data
func Example_bindJSON() {
	// Define a struct with validation tags
	type LoginRequest struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=8"`
	}

	// Create a Gin router
	r := gin.New()

	// Define a handler that uses BindJSON
	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest

		// Bind and validate the request
		if errs := binding.BindJSON(c, &req); len(errs) > 0 {
			// Handle validation errors
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		// Process the valid request
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Welcome, %s!", req.Username),
		})
	})

	// Example of a valid request
	validReq := httptest.NewRequest("POST", "/login", strings.NewReader(`{
		"username": "johndoe",
		"password": "password123"
	}`))
	validReq.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, validReq)

	fmt.Println("Valid request status:", w1.Code)
	fmt.Println("Valid request body:", w1.Body.String())

	// Example of an invalid request
	invalidReq := httptest.NewRequest("POST", "/login", strings.NewReader(`{
		"username": "jo",
		"password": "pass"
	}`))
	invalidReq.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, invalidReq)

	fmt.Println("Invalid request status:", w2.Code)
	fmt.Println("Invalid request body:", w2.Body.String())

	// Output:
	// Valid request status: 200
	// Valid request body: {"message":"Welcome, johndoe!"}
	// Invalid request status: 400
	// Invalid request body: {"errors":[{"field":"Username","tag":"min","message":"Value must be greater than or equal to 3"},{"field":"Password","tag":"min","message":"Value must be greater than or equal to 8"}]}
}

// This example demonstrates how to use the binding package to bind and validate query parameters
func Example_bindQuery() {
	// Define a struct with validation tags
	type SearchRequest struct {
		Query  string `form:"q" binding:"required"`
		Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
		Offset int    `form:"offset" binding:"omitempty,min=0"`
	}

	// Create a Gin router
	r := gin.New()

	// Define a handler that uses BindQuery
	r.GET("/search", func(c *gin.Context) {
		var req SearchRequest

		// Bind and validate the request
		if errs := binding.BindQuery(c, &req); len(errs) > 0 {
			// Handle validation errors
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
			return
		}

		// Process the valid request
		limit := req.Limit
		if limit == 0 {
			limit = 10 // Default limit
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   req.Query,
			"limit":   limit,
			"offset":  req.Offset,
			"results": []string{"result1", "result2", "result3"},
		})
	})

	// Example of a valid request
	validReq := httptest.NewRequest("GET", "/search?q=golang&limit=20&offset=40", nil)

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, validReq)

	fmt.Println("Valid request status:", w1.Code)
	fmt.Println("Valid request body:", w1.Body.String())

	// Example of an invalid request
	invalidReq := httptest.NewRequest("GET", "/search?limit=200", nil)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, invalidReq)

	fmt.Println("Invalid request status:", w2.Code)
	fmt.Println("Invalid request body:", w2.Body.String())

	// Output:
	// Valid request status: 200
	// Valid request body: {"limit":20,"offset":40,"query":"golang","results":["result1","result2","result3"]}
	// Invalid request status: 400
	// Invalid request body: {"errors":[{"field":"Query","tag":"required","message":"This field is required"},{"field":"Limit","tag":"max","message":"Value must be less than or equal to 100"}]}
}
