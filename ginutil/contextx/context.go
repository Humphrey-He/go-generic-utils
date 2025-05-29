// Package contextx 提供了对 gin.Context 操作的工具函数，
// 主要用于类型安全地存取上下文数据、获取请求信息和处理常见的请求场景。
package contextx

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Set 将值（类型为 T）以指定的键存储到 gin.Context 的 Keys 映射中。
// 如果 Context 为 nil，此函数将不执行任何操作。
func Set[T any](c *gin.Context, key string, value T) {
	if c == nil {
		return
	}

	c.Set(key, value)
}

// Get 从 gin.Context 的 Keys 映射中获取指定键的值，并尝试将其类型断言为 T。
// 返回值：
//   - 第一个返回值是获取到的值（如果获取失败则为类型 T 的零值）
//   - 第二个返回值表示是否成功获取并正确断言类型
//
// 在以下情况下将返回 (T的零值, false)：
//   - Context 为 nil
//   - 指定的键不存在
//   - 存储的值无法断言为类型 T
func Get[T any](c *gin.Context, key string) (value T, exists bool) {
	if c == nil {
		return value, false
	}

	v, exists := c.Get(key)
	if !exists {
		return value, false
	}

	// 尝试类型断言
	typedValue, ok := v.(T)
	if !ok {
		return value, false
	}

	return typedValue, true
}

// MustGet 从 gin.Context 的 Keys 映射中获取指定键的值，
// 并尝试将其类型断言为 T。如果获取失败或类型断言失败，将会 panic。
//
// 在以下情况下将会 panic：
//   - Context 为 nil
//   - 指定的键不存在
//   - 存储的值无法断言为类型 T
//
// 仅当确定键存在且类型正确时才使用此函数，否则应使用 Get。
func MustGet[T any](c *gin.Context, key string) T {
	value, exists := Get[T](c, key)
	if !exists {
		panic(fmt.Sprintf("contextx: key %q not found in context or type assertion failed", key))
	}
	return value
}
