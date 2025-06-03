package binding

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 测试数据结构
// Test data structures
type TestUser struct {
	ID       int    `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,min=18,max=120"`
	IsActive bool   `json:"is_active"`
}

type TestQueryParams struct {
	Page   int    `form:"page" binding:"min=1"`
	Limit  int    `form:"limit" binding:"min=1,max=100"`
	Sort   string `form:"sort" binding:"oneof=asc desc"`
	Filter string `form:"filter"`
}

type TestURIParams struct {
	ID     int    `uri:"id" binding:"required,min=1"`
	Action string `uri:"action" binding:"required,oneof=view edit delete"`
}

type TestFormData struct {
	Name    string `form:"name" binding:"required"`
	Message string `form:"message" binding:"required"`
	Agree   bool   `form:"agree" binding:"required"`
}

func init() {
	// 设置Gin为测试模式
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// 创建一个新的Gin Context用于测试
// Create a new Gin Context for testing
func newTestContext(w *httptest.ResponseRecorder, req *http.Request) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

// 测试 BindJSON 函数
// Test BindJSON function
func TestBindJSON(t *testing.T) {
	tests := []struct {
		name         string
		jsonBody     string
		wantErr      bool
		errorCount   int
		errorFields  []string
		errorTags    []string
		errorMessage string
	}{
		{
			name:     "valid_data",
			jsonBody: `{"id": 1, "name": "John Doe", "email": "john@example.com", "age": 30, "is_active": true}`,
			wantErr:  false,
		},
		{
			name:        "missing_required_field",
			jsonBody:    `{"id": 1, "email": "john@example.com", "age": 30}`,
			wantErr:     true,
			errorCount:  1,
			errorFields: []string{"Name"},
			errorTags:   []string{"required"},
		},
		{
			name:        "invalid_email",
			jsonBody:    `{"id": 1, "name": "John Doe", "email": "invalid-email", "age": 30}`,
			wantErr:     true,
			errorCount:  1,
			errorFields: []string{"Email"},
			errorTags:   []string{"email"},
		},
		{
			name:        "multiple_validation_errors",
			jsonBody:    `{"id": 1, "name": "Jo", "email": "invalid-email", "age": 15}`,
			wantErr:     true,
			errorCount:  3,
			errorFields: []string{"Name", "Email", "Age"},
			errorTags:   []string{"min", "email", "min"},
		},
		{
			name:         "invalid_json",
			jsonBody:     `{"id": 1, "name": "John Doe", "email": "john@example.com", "age": 30, is_active: true}`,
			wantErr:      true,
			errorCount:   1,
			errorMessage: "invalid character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求
			// Prepare request
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c := newTestContext(w, req)

			// 执行绑定
			// Execute binding
			var user TestUser
			errs := BindJSON(c, &user)

			// 验证结果
			// Validate results
			if tt.wantErr {
				assert.True(t, len(errs) > 0, "Expected errors but got none")
				assert.Equal(t, tt.errorCount, len(errs), "Error count mismatch")

				if tt.errorFields != nil {
					fields := make([]string, 0, len(errs))
					tags := make([]string, 0, len(errs))
					for _, err := range errs {
						fields = append(fields, err.Field)
						tags = append(tags, err.Tag)
					}

					for _, field := range tt.errorFields {
						assert.Contains(t, fields, field, "Expected error field not found: "+field)
					}

					for _, tag := range tt.errorTags {
						assert.Contains(t, tags, tag, "Expected error tag not found: "+tag)
					}
				}

				if tt.errorMessage != "" {
					found := false
					for _, err := range errs {
						if strings.Contains(err.Message, tt.errorMessage) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message not found: "+tt.errorMessage)
				}
			} else {
				assert.Empty(t, errs, "Expected no errors but got: %v", errs)
				// 验证数据是否被正确绑定
				// Verify data was bound correctly
				assert.Equal(t, 1, user.ID)
				assert.Equal(t, "John Doe", user.Name)
				assert.Equal(t, "john@example.com", user.Email)
				assert.Equal(t, 30, user.Age)
				assert.True(t, user.IsActive)
			}
		})
	}
}

// 测试 BindQuery 函数
// Test BindQuery function
func TestBindQuery(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		wantErr     bool
		errorCount  int
		errorFields []string
		errorTags   []string
	}{
		{
			name:        "valid_data",
			queryString: "page=1&limit=20&sort=desc&filter=active",
			wantErr:     false,
		},
		{
			name:        "invalid_page",
			queryString: "page=0&limit=20&sort=desc",
			wantErr:     true,
			errorCount:  1,
			errorFields: []string{"Page"},
			errorTags:   []string{"min"},
		},
		{
			name:        "invalid_limit",
			queryString: "page=1&limit=200&sort=desc",
			wantErr:     true,
			errorCount:  1,
			errorFields: []string{"Limit"},
			errorTags:   []string{"max"},
		},
		{
			name:        "invalid_sort",
			queryString: "page=1&limit=20&sort=invalid",
			wantErr:     true,
			errorCount:  1,
			errorFields: []string{"Sort"},
			errorTags:   []string{"oneof"},
		},
		{
			name:        "multiple_errors",
			queryString: "page=0&limit=200&sort=invalid",
			wantErr:     true,
			errorCount:  3,
			errorFields: []string{"Page", "Limit", "Sort"},
			errorTags:   []string{"min", "max", "oneof"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求
			// Prepare request
			req := httptest.NewRequest("GET", "/test?"+tt.queryString, nil)
			w := httptest.NewRecorder()
			c := newTestContext(w, req)

			// 执行绑定
			// Execute binding
			var params TestQueryParams
			errs := BindQuery(c, &params)

			// 验证结果
			// Validate results
			if tt.wantErr {
				assert.True(t, len(errs) > 0, "Expected errors but got none")
				assert.Equal(t, tt.errorCount, len(errs), "Error count mismatch")

				if tt.errorFields != nil {
					fields := make([]string, 0, len(errs))
					tags := make([]string, 0, len(errs))
					for _, err := range errs {
						fields = append(fields, err.Field)
						tags = append(tags, err.Tag)
					}

					for _, field := range tt.errorFields {
						assert.Contains(t, fields, field, "Expected error field not found: "+field)
					}

					for _, tag := range tt.errorTags {
						assert.Contains(t, tags, tag, "Expected error tag not found: "+tag)
					}
				}
			} else {
				assert.Empty(t, errs, "Expected no errors but got: %v", errs)
				// 验证数据是否被正确绑定
				// Verify data was bound correctly
				assert.Equal(t, 1, params.Page)
				assert.Equal(t, 20, params.Limit)
				assert.Equal(t, "desc", params.Sort)
				assert.Equal(t, "active", params.Filter)
			}
		})
	}
}

// 测试 BindURI 函数
// Test BindURI function
func TestBindURI(t *testing.T) {
	// 创建路由
	// Create router
	router := gin.New()
	router.GET("/users/:id/:action", func(c *gin.Context) {
		var params TestURIParams
		errs := BindURI(c, &params)

		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
			return
		}

		c.JSON(http.StatusOK, params)
	})

	tests := []struct {
		name       string
		uri        string
		wantStatus int
	}{
		{
			name:       "valid_uri",
			uri:        "/users/1/edit",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid_id",
			uri:        "/users/0/edit",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid_action",
			uri:        "/users/1/invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.uri, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var params TestURIParams
				err := json.Unmarshal(w.Body.Bytes(), &params)
				assert.NoError(t, err)

				// 检查参数是否正确绑定
				// Check if parameters were bound correctly
				if tt.uri == "/users/1/edit" {
					assert.Equal(t, 1, params.ID)
					assert.Equal(t, "edit", params.Action)
				}
			} else {
				// 验证是否返回了错误信息
				// Verify error information was returned
				var response map[string][]FieldError
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["errors"])
			}
		})
	}
}

// 测试 BindForm 函数
// Test BindForm function
func TestBindForm(t *testing.T) {
	tests := []struct {
		name        string
		formValues  map[string]string
		wantErr     bool
		errorCount  int
		errorFields []string
	}{
		{
			name: "valid_form",
			formValues: map[string]string{
				"name":    "John Doe",
				"message": "Hello, World!",
				"agree":   "true",
			},
			wantErr: false,
		},
		{
			name: "missing_required_field",
			formValues: map[string]string{
				"name":  "John Doe",
				"agree": "true",
			},
			wantErr:     true,
			errorCount:  1,
			errorFields: []string{"Message"},
		},
		{
			name: "multiple_missing_fields",
			formValues: map[string]string{
				"name": "John Doe",
			},
			wantErr:     true,
			errorCount:  2,
			errorFields: []string{"Message", "Agree"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备表单数据
			// Prepare form data
			formData := strings.NewReader(makeFormData(tt.formValues))
			req := httptest.NewRequest("POST", "/test", formData)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			c := newTestContext(w, req)

			// 执行绑定
			// Execute binding
			var form TestFormData
			errs := BindForm(c, &form)

			// 验证结果
			// Validate results
			if tt.wantErr {
				assert.True(t, len(errs) > 0, "Expected errors but got none")
				assert.Equal(t, tt.errorCount, len(errs), "Error count mismatch")

				fields := make([]string, 0, len(errs))
				for _, err := range errs {
					fields = append(fields, err.Field)
				}

				for _, field := range tt.errorFields {
					assert.Contains(t, fields, field, "Expected error field not found: "+field)
				}
			} else {
				assert.Empty(t, errs, "Expected no errors but got: %v", errs)
				// 验证数据是否被正确绑定
				// Verify data was bound correctly
				assert.Equal(t, "John Doe", form.Name)
				assert.Equal(t, "Hello, World!", form.Message)
				assert.True(t, form.Agree)
			}
		})
	}
}

// 测试 Bind 函数 - 自动选择绑定方法
// Test Bind function - automatically selects binding method
func TestBind(t *testing.T) {
	// 测试 JSON 绑定
	// Test JSON binding
	t.Run("json_binding", func(t *testing.T) {
		jsonBody := `{"id": 1, "name": "John Doe", "email": "john@example.com", "age": 30, "is_active": true}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c := newTestContext(w, req)

		var user TestUser
		errs := Bind(c, &user)

		assert.Empty(t, errs, "Expected no errors but got: %v", errs)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
	})

	// 测试表单绑定
	// Test form binding
	t.Run("form_binding", func(t *testing.T) {
		formValues := map[string]string{
			"name":    "John Doe",
			"message": "Hello, World!",
			"agree":   "true",
		}
		formData := strings.NewReader(makeFormData(formValues))
		req := httptest.NewRequest("POST", "/test", formData)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		c := newTestContext(w, req)

		var form TestFormData
		errs := Bind(c, &form)

		assert.Empty(t, errs, "Expected no errors but got: %v", errs)
		assert.Equal(t, "John Doe", form.Name)
		assert.Equal(t, "Hello, World!", form.Message)
		assert.True(t, form.Agree)
	})

	// 测试查询参数绑定
	// Test query parameter binding
	t.Run("query_binding", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?page=1&limit=20&sort=desc&filter=active", nil)
		w := httptest.NewRecorder()
		c := newTestContext(w, req)

		var params TestQueryParams
		errs := Bind(c, &params)

		assert.Empty(t, errs, "Expected no errors but got: %v", errs)
		assert.Equal(t, 1, params.Page)
		assert.Equal(t, 20, params.Limit)
		assert.Equal(t, "desc", params.Sort)
		assert.Equal(t, "active", params.Filter)
	})
}

// 测试参数验证错误
// Test parameter validation errors
func TestFieldErrors(t *testing.T) {
	t.Run("error_string", func(t *testing.T) {
		errs := FieldErrors{
			{Field: "Username", Tag: "required", Message: "This field is required"},
			{Field: "Password", Tag: "min", Message: "Minimum length is 8"},
		}

		// 测试Error()方法
		// Test Error() method
		errStr := errs.Error()
		assert.Contains(t, errStr, "Username: This field is required")
		assert.Contains(t, errStr, "Password: Minimum length is 8")
	})

	t.Run("add_error", func(t *testing.T) {
		var errs FieldErrors

		// 测试Add方法
		// Test Add method
		errs.Add("Email", "email", "Invalid email format")
		assert.Len(t, errs, 1)
		assert.Equal(t, "Email", errs[0].Field)
		assert.Equal(t, "email", errs[0].Tag)
		assert.Equal(t, "Invalid email format", errs[0].Message)

		// 测试带值的Add方法
		// Test Add method with value
		errs.Add("Age", "min", "Must be at least 18", 15)
		assert.Len(t, errs, 2)
		assert.Equal(t, "Age", errs[1].Field)
		assert.Equal(t, 15, errs[1].Value)
	})

	t.Run("add_error_struct", func(t *testing.T) {
		var errs FieldErrors

		// 测试AddError方法
		// Test AddError method
		errs.AddError(FieldError{
			Field:   "Email",
			Tag:     "email",
			Message: "Invalid email format",
			Value:   "test",
		})

		assert.Len(t, errs, 1)
		assert.Equal(t, "Email", errs[0].Field)
		assert.Equal(t, "email", errs[0].Tag)
		assert.Equal(t, "Invalid email format", errs[0].Message)
		assert.Equal(t, "test", errs[0].Value)
	})

	t.Run("has_errors", func(t *testing.T) {
		var errs FieldErrors
		assert.False(t, errs.HasErrors(), "Empty FieldErrors should return false")

		errs.Add("Field", "tag", "message")
		assert.True(t, errs.HasErrors(), "Non-empty FieldErrors should return true")
	})
}

// 辅助函数：生成表单数据字符串
// Helper function: generate form data string
func makeFormData(values map[string]string) string {
	parts := make([]string, 0, len(values))
	for k, v := range values {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "&")
}
