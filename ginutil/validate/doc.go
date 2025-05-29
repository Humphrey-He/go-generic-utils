/*
Package validate 提供基于 go-playground/validator 的增强校验功能。

# 概述

validate 包是对 go-playground/validator/v10 的封装，提供了更易用的 API、自定义校验规则和本地化错误消息。
它主要用于结构体校验，特别适合与 Gin 框架集成，用于校验请求参数。

# 主要功能

1. 类型安全的校验函数：使用泛型提供类型安全的校验函数。
2. 自定义校验规则：提供常用的自定义校验规则，如 no-special-chars、is-safe-html 等。
3. 本地化错误消息：支持自定义错误消息，默认提供中文错误消息。
4. 与 Gin 框架集成：提供便捷的绑定和校验函数，如 MustBindAndValidate、MustBindJSONAndValidate 等。

# 使用方法

## 基本用法

	import "ggu/ginutil/validate"

	type User struct {
		Name     string `json:"name" validate:"required,min=2,max=50"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8,no-special-chars"`
	}

	func ValidateUser(user User) error {
		return validate.ValidateStruct(user)
	}

## 与 Gin 框架集成

	func CreateUser(c *gin.Context) {
		var user User
		if !validate.MustBindJSONAndValidate(c, &user) {
			// 绑定或校验失败，MustBindJSONAndValidate 已经发送了错误响应
			return
		}

		// 处理业务逻辑...
	}

## 注册自定义校验规则

	// 注册全局校验规则
	validate.RegisterRule("custom-rule", func(fl validator.FieldLevel) bool {
		// 实现校验逻辑
		return true
	})

	// 注册特定校验器的校验规则
	v := validate.NewV10()
	v.RegisterValidation("custom-rule", func(fl validator.FieldLevel) bool {
		// 实现校验逻辑
		return true
	})

## 自定义错误消息

	// 创建自定义消息存储
	ms := validate.NewMessageStore()
	ms.RegisterMessage("zh", "required", "此字段不能为空")
	ms.RegisterMessage("zh", "email", "请输入有效的电子邮箱地址")

	// 设置到校验器
	v := validate.NewV10()
	v.SetMessageStore(ms)

# 内置校验规则

validate 包提供了以下内置的自定义校验规则：

- no-special-chars: 不允许特殊字符
- is-safe-html: 简单的 HTML 安全校验
- is-valid-country-code: 校验国家码是否有效
- is-valid-phone-for-region: 校验手机号是否符合特定地区的格式

此外，还提供了以下工厂函数来创建自定义校验规则：

- IsUniqueInDB: 检查字段值在数据库中是否唯一
- IsUniqueInSlice: 检查切片中的元素是否唯一

# 错误处理

validate 包返回的错误类型是 binding.FieldErrors，它提供了丰富的错误信息，包括字段名、校验标签、错误消息和字段值。
这使得 API 层可以统一处理错误响应，提供更友好的错误信息给客户端。

	if errs := validate.ValidateStruct(user); errs != nil && errs.HasErrors() {
		// 处理错误
		for _, err := range errs {
			fmt.Printf("字段: %s, 错误: %s\n", err.Field, err.Message)
		}
	}
*/
package validate
