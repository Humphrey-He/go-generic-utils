package contextx_test

import (
	"ggu/ginutil/contextx"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 测试正常情况
	contextx.Set(c, "test-key", "test-value")
	value, exists := c.Get("test-key")
	assert.True(t, exists, "键应该存在")
	assert.Equal(t, "test-value", value, "值应该正确")

	// 测试结构体值
	type TestStruct struct {
		Name string
		Age  int
	}
	expected := TestStruct{Name: "测试", Age: 30}
	contextx.Set(c, "test-struct", expected)
	value, exists = c.Get("test-struct")
	assert.True(t, exists, "键应该存在")
	assert.Equal(t, expected, value, "值应该正确")

	// 测试 nil 上下文
	assert.NotPanics(t, func() {
		contextx.Set[string](nil, "nil-test", "value")
	}, "传入 nil 上下文不应该 panic")
}

func TestGet(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 设置不同类型的值
	c.Set("string-key", "string-value")
	c.Set("int-key", 42)
	c.Set("bool-key", true)

	// 测试正常情况
	strVal, exists := contextx.Get[string](c, "string-key")
	assert.True(t, exists, "字符串键应该存在")
	assert.Equal(t, "string-value", strVal, "字符串值应该正确")

	intVal, exists := contextx.Get[int](c, "int-key")
	assert.True(t, exists, "整数键应该存在")
	assert.Equal(t, 42, intVal, "整数值应该正确")

	boolVal, exists := contextx.Get[bool](c, "bool-key")
	assert.True(t, exists, "布尔键应该存在")
	assert.Equal(t, true, boolVal, "布尔值应该正确")

	// 测试键不存在的情况
	strVal, exists = contextx.Get[string](c, "non-existent-key")
	assert.False(t, exists, "不存在的键应该返回 false")
	assert.Equal(t, "", strVal, "不存在的键应该返回零值")

	// 测试类型不匹配的情况
	intVal, exists = contextx.Get[int](c, "string-key")
	assert.False(t, exists, "类型不匹配应该返回 false")
	assert.Equal(t, 0, intVal, "类型不匹配应该返回零值")

	// 测试 nil 上下文
	strVal, exists = contextx.Get[string](nil, "key")
	assert.False(t, exists, "nil 上下文应该返回 false")
	assert.Equal(t, "", strVal, "nil 上下文应该返回零值")
}

func TestMustGet(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 设置一个值
	c.Set("must-key", "must-value")

	// 测试正常情况
	assert.NotPanics(t, func() {
		val := contextx.MustGet[string](c, "must-key")
		assert.Equal(t, "must-value", val, "值应该正确")
	})

	// 测试键不存在的情况
	assert.Panics(t, func() {
		contextx.MustGet[string](c, "non-existent-key")
	}, "不存在的键应该 panic")

	// 测试类型不匹配的情况
	c.Set("int-key", 42)
	assert.Panics(t, func() {
		contextx.MustGet[string](c, "int-key")
	}, "类型不匹配应该 panic")

	// 测试 nil 上下文
	assert.Panics(t, func() {
		contextx.MustGet[string](nil, "key")
	}, "nil 上下文应该 panic")
}
