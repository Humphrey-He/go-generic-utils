package validate_test

import (
	"fmt"
	"ggu/ginutil/validate"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// 示例用户结构体
type User struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,no-special-chars"`
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	Country  string `json:"country" validate:"required,is-valid-country-code"`
	Phone    string `json:"phone" validate:"required,is-valid-phone-for-region=CN"`
}

// ExampleValidateStruct 展示如何使用 ValidateStruct 函数校验结构体
func ExampleValidateStruct() {
	// 创建一个有效的用户
	validUser := User{
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Password: "password123",
		Age:      30,
		Country:  "CN",
		Phone:    "13812345678",
	}

	// 校验有效用户
	errs := validate.ValidateStruct(validUser)
	if errs == nil || !errs.HasErrors() {
		fmt.Println("Valid user passed validation")
	} else {
		fmt.Println("Valid user failed validation:", errs)
	}

	// 创建一个无效的用户
	invalidUser := User{
		Name:     "李",             // 名字太短
		Email:    "invalid-email", // 无效的邮箱
		Password: "pwd",           // 密码太短
		Age:      17,              // 年龄太小
		Country:  "CHN",           // 无效的国家代码
		Phone:    "12345678",      // 无效的手机号
	}

	// 校验无效用户
	errs = validate.ValidateStruct(invalidUser)
	if errs != nil && errs.HasErrors() {
		fmt.Println("Invalid user failed validation as expected")
		fmt.Printf("Number of validation errors: %d\n", len(errs))
	} else {
		fmt.Println("Invalid user unexpectedly passed validation")
	}

	// Output:
	// Valid user passed validation
	// Invalid user failed validation as expected
	// Number of validation errors: 6
}

// ExampleRegisterRule 展示如何注册自定义校验规则
func ExampleRegisterRule() {
	// 注册一个自定义校验规则，检查字符串是否是回文
	validate.RegisterRule("palindrome", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		if str == "" {
			return true
		}

		// 忽略大小写和空格
		str = strings.ToLower(strings.ReplaceAll(str, " ", ""))

		// 检查是否是回文
		for i := 0; i < len(str)/2; i++ {
			if str[i] != str[len(str)-i-1] {
				return false
			}
		}

		return true
	})

	// 定义使用自定义规则的结构体
	type PalindromeTest struct {
		Text string `json:"text" validate:"required,palindrome"`
	}

	// 测试有效的回文
	valid := PalindromeTest{Text: "A man a plan a canal Panama"}
	errs := validate.ValidateStruct(valid)
	if errs == nil || !errs.HasErrors() {
		fmt.Println("Valid palindrome passed validation")
	} else {
		fmt.Println("Valid palindrome failed validation:", errs)
	}

	// 测试无效的回文
	invalid := PalindromeTest{Text: "This is not a palindrome"}
	errs = validate.ValidateStruct(invalid)
	if errs != nil && errs.HasErrors() {
		fmt.Println("Invalid palindrome failed validation as expected")
	} else {
		fmt.Println("Invalid palindrome unexpectedly passed validation")
	}

	// Output:
	// Valid palindrome passed validation
	// Invalid palindrome failed validation as expected
}

// ExampleMustBindJSONAndValidate 展示如何在 Gin 处理函数中使用 MustBindJSONAndValidate
func ExampleMustBindJSONAndValidate() {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 定义处理函数
	r.POST("/users", func(c *gin.Context) {
		var user User
		if !validate.MustBindJSONAndValidate(c, &user) {
			// 绑定或校验失败，MustBindJSONAndValidate 已经发送了错误响应
			return
		}

		// 处理业务逻辑...
		c.JSON(http.StatusCreated, gin.H{
			"message": "User created successfully",
			"user":    user,
		})
	})

	// 创建有效的请求
	validReq := httptest.NewRequest("POST", "/users", strings.NewReader(`{
		"name": "张三",
		"email": "zhangsan@example.com",
		"password": "password123",
		"age": 30,
		"country": "CN",
		"phone": "13812345678"
	}`))
	validReq.Header.Set("Content-Type", "application/json")

	// 处理有效请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, validReq)
	fmt.Printf("Valid request status code: %d\n", w.Code)

	// 创建无效的请求
	invalidReq := httptest.NewRequest("POST", "/users", strings.NewReader(`{
		"name": "李",
		"email": "invalid-email",
		"password": "pwd",
		"age": 17,
		"country": "CHN",
		"phone": "12345678"
	}`))
	invalidReq.Header.Set("Content-Type", "application/json")

	// 处理无效请求
	w = httptest.NewRecorder()
	r.ServeHTTP(w, invalidReq)
	fmt.Printf("Invalid request status code: %d\n", w.Code)

	// Output:
	// Valid request status code: 201
	// Invalid request status code: 400
}

// ExampleCustomMessageStore 展示如何自定义错误消息
func ExampleCustomMessageStore() {
	// 创建自定义消息存储
	ms := validate.NewMessageStore()

	// 注册自定义错误消息
	ms.RegisterMessage("zh", "required", "此字段不能为空")
	ms.RegisterMessage("zh", "email", "请提供有效的电子邮件地址")
	ms.RegisterMessage("zh", "min", "此字段的长度不能小于 %s")

	// 为特定字段注册自定义错误消息
	ms.RegisterMessage("zh", "required", "用户名不能为空", "Name")
	ms.RegisterMessage("zh", "required", "密码不能为空", "Password")

	// 创建新的校验器
	v := validate.NewV10()
	v.SetMessageStore(ms)

	// 定义测试结构体
	type TestForm struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	// 创建一个空的测试表单
	form := TestForm{}

	// 使用自定义校验器校验
	errs := v.Validate(form)
	if errs != nil && errs.HasErrors() {
		fmt.Println("Validation failed with custom messages:")
		for _, err := range errs {
			fmt.Printf("  %s: %s\n", err.Field, err.Message)
		}
	}

	// 输出将包含自定义错误消息
	// Output:
	// Validation failed with custom messages:
	//   Name: 用户名不能为空
	//   Email: 此字段不能为空
	//   Password: 密码不能为空
}
