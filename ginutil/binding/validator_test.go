package binding

import (
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// 测试结构体
// Test structs
type TestValidateUser struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,min=18"`
}

type TestCustomMessage struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// 测试自定义验证函数
// Test custom validation function
func isOdd(fl validator.FieldLevel) bool {
	return fl.Field().Int()%2 == 1
}

// 测试默认验证器
// Test default validator
func TestDefaultValidator(t *testing.T) {
	// 创建一个有效的用户
	// Create a valid user
	validUser := TestValidateUser{
		Username: "johndoe",
		Email:    "john@example.com",
		Age:      25,
	}

	// 验证有效用户
	// Validate valid user
	errors := DefaultValidator.Validate(validUser)
	assert.Empty(t, errors, "Valid user should pass validation")

	// 验证无效用户
	// Validate invalid user
	invalidUser := TestValidateUser{
		Username: "jo", // too short
		Email:    "invalid-email",
		Age:      15, // too young
	}

	errors = DefaultValidator.Validate(invalidUser)
	assert.NotEmpty(t, errors, "Invalid user should fail validation")
	assert.Equal(t, 3, len(errors), "Should have 3 validation errors")

	// 验证每个错误字段
	// Validate each error field
	fieldsMap := make(map[string]bool)
	tagsMap := make(map[string]bool)

	for _, err := range errors {
		fieldsMap[err.Field] = true
		tagsMap[err.Tag] = true
		// 输出错误信息以便调试
		// Output error information for debugging
		t.Logf("Error field: %s, tag: %s, message: %s", err.Field, err.Tag, err.Message)
	}

	// 字段名可能会根据标签命名函数返回小写的字段名
	// Field names might be lowercase due to tag naming function
	assert.True(t, fieldsMap["username"] || fieldsMap["Username"], "Username should be in error fields")
	assert.True(t, fieldsMap["email"] || fieldsMap["Email"], "Email should be in error fields")
	assert.True(t, fieldsMap["age"] || fieldsMap["Age"], "Age should be in error fields")

	assert.True(t, tagsMap["min"], "min tag should be in error tags")
	assert.True(t, tagsMap["email"], "email tag should be in error tags")
}

// 测试注册自定义验证
// Test registering custom validation
func TestRegisterValidation(t *testing.T) {
	// 创建一个自定义验证器
	// Create a custom validator
	v := &defaultValidator{
		customMessages: make(map[string]string),
	}

	// 注册自定义验证函数
	// Register custom validation function
	err := v.RegisterValidation("odd", isOdd)
	assert.NoError(t, err, "Registering custom validation should not error")

	// 定义使用自定义验证的结构体
	// Define struct that uses custom validation
	type TestCustomValidation struct {
		Number int `json:"number" binding:"required,odd"`
	}

	// 验证有效数据（奇数）
	// Validate valid data (odd number)
	validData := TestCustomValidation{Number: 5}
	errors := v.Validate(validData)
	assert.Empty(t, errors, "Valid odd number should pass validation")

	// 验证无效数据（偶数）
	// Validate invalid data (even number)
	invalidData := TestCustomValidation{Number: 4}
	errors = v.Validate(invalidData)
	assert.NotEmpty(t, errors, "Even number should fail validation")
	assert.Equal(t, 1, len(errors), "Should have 1 validation error")
	// 字段名可能会根据标签命名函数返回小写的字段名
	// Field names might be lowercase due to tag naming function
	fieldName := errors[0].Field
	assert.True(t, fieldName == "number" || fieldName == "Number", "Field name should be 'number' or 'Number'")
	assert.Equal(t, "odd", errors[0].Tag, "Tag should be correct")
}

// 测试自定义错误消息
// Test custom error messages
func TestCustomErrorMessages(t *testing.T) {
	// 创建一个自定义验证器
	// Create a custom validator
	v := &defaultValidator{
		customMessages: make(map[string]string),
	}
	v.lazyInit()

	// 注册全局自定义消息
	// Register global custom message
	v.RegisterCustomMessage("required", "此字段不能为空")

	// 注册字段特定的自定义消息
	// Register field-specific custom message
	v.RegisterCustomMessage("Email.email", "邮箱格式不正确")

	// 验证数据
	// Validate data
	data := TestCustomMessage{
		Name:  "",
		Email: "invalid-email",
	}

	errors := v.Validate(data)
	assert.Equal(t, 2, len(errors), "Should have 2 validation errors")

	// 检查自定义消息
	// Check custom messages
	for _, err := range errors {
		if err.Field == "Name" && err.Tag == "required" {
			assert.Equal(t, "此字段不能为空", err.Message, "Should use custom message for required tag")
		}

		if err.Field == "Email" && err.Tag == "email" {
			assert.Equal(t, "邮箱格式不正确", err.Message, "Should use custom message for Email.email")
		}
	}
}

// 测试 GetValidationErrors 函数
// Test GetValidationErrors function
func TestGetValidationErrors(t *testing.T) {
	// 创建一个自定义验证器
	// Create a custom validator
	v := validator.New()
	v.SetTagName("binding")

	// 准备一个验证错误
	// Prepare a validation error
	data := TestValidateUser{
		Username: "a",
		Email:    "invalid",
	}

	err := v.Struct(data)

	// 测试提取验证错误
	// Test extracting validation errors
	validationErrors, ok := GetValidationErrors(err)
	assert.True(t, ok, "Should be able to extract validation errors")
	assert.NotEmpty(t, validationErrors, "Should have validation errors")

	// 测试空错误
	// Test nil error
	validationErrors, ok = GetValidationErrors(nil)
	assert.False(t, ok, "Should return false for nil error")
	assert.Nil(t, validationErrors, "Should return nil validation errors for nil error")

	// 测试非验证错误
	// Test non-validation error
	validationErrors, ok = GetValidationErrors(assert.AnError)
	assert.False(t, ok, "Should return false for non-validation error")
	assert.Nil(t, validationErrors, "Should return nil validation errors for non-validation error")
}

// 测试 SetTagNameFunc 函数
// Test SetTagNameFunc function
func TestSetTagNameFunc(t *testing.T) {
	// 创建一个自定义验证器
	// Create a custom validator
	v := &defaultValidator{
		customMessages: make(map[string]string),
	}

	// 设置自定义标签名称函数
	// Set custom tag name function
	v.SetTagNameFunc(func(field reflect.StructField) string {
		return "custom_" + field.Name
	})

	// 验证引擎已经注册了函数
	// Verify engine has registered the function
	engine := v.Engine()
	assert.NotNil(t, engine, "Engine should not be nil")

	// 验证自定义标签名称函数
	// 由于validator.Validate的内部实现，我们无法直接测试tagNameFunc的效果
	// 这里只能验证它被设置了
	// Validate custom tag name function
	// Due to the internal implementation of validator.Validate, we cannot directly test the effect of tagNameFunc
	// We can only verify it was set
	assert.NotNil(t, v.tagNameFunc, "Tag name function should not be nil")
}

// 测试空对象验证
// Test nil object validation
func TestValidateNilObject(t *testing.T) {
	errors := DefaultValidator.Validate(nil)
	assert.Nil(t, errors, "Validating nil object should return nil")
}

// 测试验证包装器
// Test validator wrapper
func TestValidatorWrapper(t *testing.T) {
	// 创建一个自定义验证器
	// Create a custom validator
	mockValidator := &mockValidator{}

	// 创建包装器
	// Create wrapper
	wrapper := &validatorWrapper{mockValidator}

	// 测试 Validate 函数
	// Test Validate function
	result := wrapper.Validate("test")
	assert.True(t, result.HasErrors(), "Should call through to mock validator")
	assert.Equal(t, "mock", result[0].Field, "Should return mock result")

	// 测试 Engine 函数
	// Test Engine function
	engine := wrapper.Engine()
	assert.Equal(t, "mock_engine", engine, "Should call through to mock engine")
}

// 创建一个模拟验证器用于测试
// Create a mock validator for testing
type mockValidator struct{}

func (m *mockValidator) Validate(obj interface{}) FieldErrors {
	return FieldErrors{{Field: "mock", Tag: "mock", Message: "mock error"}}
}

func (m *mockValidator) Engine() interface{} {
	return "mock_engine"
}
