# Gin 请求绑定增强库 — `binding`

`binding` 包为 Gin 应用提供了增强的请求绑定工具。

本包通过泛型支持扩展了 Gin 的原生绑定能力，使得请求绑定过程类型安全，并显著减少了样板代码。它针对不同的数据源（如 JSON、查询参数、URI、表单、XML）提供了一套泛型绑定函数，并辅以增强的校验功能和错误报告机制。

## 主要特性

* **泛型绑定函数**：实现类型安全的请求数据绑定。
* **增强的校验机制**：提供详细的、字段级别的错误报告。
* **自定义校验规则与消息**：允许开发者定义自己的校验逻辑和错误提示。
* **可扩展的解码器系统**：能够方便地针对不同的内容类型扩展解码能力。

## 使用示例

```go
package main

import (
	"net/http"

	"[github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)"

)

// UserRequest 定义了用户请求的数据结构。
// 使用 `json` 标签指定 JSON 字段名，
// 使用 `binding` 标签指定校验规则。
type UserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func CreateUser(c *gin.Context) {
	var req UserRequest
    // 假设 binding.BindJSON 是您库中定义的函数
    // errs 是一个包含详细错误信息的结构体切片或自定义错误类型
	// if errs := binding.BindJSON(&req, c); len(errs) > 0 { 
	//     c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
	//     return
	// }

    // 模拟绑定和校验过程 - 实际应替换为 binding.BindJSON(&req, c)
    // 为了示例能独立运行，这里我们先手动绑定并假设一个错误场景
    if err := c.ShouldBindJSON(&req); err != nil { // 使用 Gin 原生绑定作为示例的替代
        // 实际中，binding.BindJSON 会返回更详细的错误信息 errs
        c.JSON(http.StatusBadRequest, gin.H{"error_gin": err.Error()})
        return
    }
    
    // 假设这是 binding.BindJSON 返回的错误格式
    // var fieldErrors []binding.FieldError // 假设的错误类型
    // if req.Name == "" {
    //     fieldErrors = append(fieldErrors, binding.FieldError{Field: "Name", Message: "名称不能为空"})
    // }
    // if !strings.Contains(req.Email, "@") { // 简化的 email 校验
    //     fieldErrors = append(fieldErrors, binding.FieldError{Field: "Email", Message: "邮箱格式不正确"})
    // }
    // if len(fieldErrors) > 0 {
    //     c.JSON(http.StatusBadRequest, gin.H{"errors": fieldErrors})
    //     return
    // }

	// 处理已校验通过的请求数据
	// 例如：userService.Create(req.Name, req.Email)
	c.JSON(http.StatusOK, gin.H{
		"message": "用户创建成功",
		"name":    req.Name,
		"email":   req.Email,
	})
}

/*
// 假设的 binding.FieldError 结构，用于提供详细错误信息
package binding 

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Tag     string `json:"tag,omitempty"`     // 触发错误的校验标签
    Value   any    `json:"value,omitempty"`   // 导致错误的值
}
*/