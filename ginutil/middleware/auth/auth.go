// Package auth 提供了一套用于Gin应用的身份认证和授权中间件。
//
// 本包包含了多种认证方式的实现：
//   - JWT认证：基于JSON Web Token的认证
//   - Basic认证：基于HTTP Basic Authentication的认证
//   - OAuth 2.0认证：基于OAuth 2.0协议的认证
//
// 所有认证方式都使用统一的UserIdentity结构来表示认证后的用户身份信息，
// 便于在不同的认证方式之间进行切换，以及在后续的授权中间件中使用。
package auth

import (
	"errors"

	"github.com/noobtrump/go-generic-utils/ginutil/response"

	"github.com/gin-gonic/gin"
)

// 常见错误
var (
	// ErrAuthRequired 表示需要认证
	ErrAuthRequired = errors.New("需要认证")

	// ErrPermissionDenied 表示权限不足
	ErrPermissionDenied = errors.New("权限不足")
)

// RequireAuth 是一个通用的认证中间件，要求请求中包含有效的用户身份信息。
// 如果没有找到用户身份信息，则返回401 Unauthorized错误。
func RequireAuth[ID comparable, Role comparable]() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文中获取用户身份信息
		identity, exists := GetIdentityFromContext[ID, Role](c)
		if !exists || identity == nil {
			response.Unauthorized(c, "需要登录")
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// RequireTenant 创建一个中间件，要求用户属于指定的租户。
func RequireTenant[ID comparable, Role comparable](tenantID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文中获取用户身份信息
		identity, exists := GetIdentityFromContext[ID, Role](c)
		if !exists || identity == nil {
			response.Unauthorized(c, "需要登录")
			return
		}

		// 验证租户ID
		if identity.TenantID != tenantID {
			response.Forbidden(c, "无权访问该租户的资源")
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// AuthorizeFunc 是一个授权函数类型，用于自定义授权逻辑。
type AuthorizeFunc[ID comparable, Role comparable] func(identity *UserIdentity[ID, Role], c *gin.Context) bool

// Authorize 创建一个中间件，使用自定义的授权函数进行授权。
func Authorize[ID comparable, Role comparable](authorizeFunc AuthorizeFunc[ID, Role]) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文中获取用户身份信息
		identity, exists := GetIdentityFromContext[ID, Role](c)
		if !exists || identity == nil {
			response.Unauthorized(c, "需要登录")
			return
		}

		// 调用授权函数
		if !authorizeFunc(identity, c) {
			response.Forbidden(c, "权限不足")
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// RequirePermission 创建一个中间件，要求用户具有指定的权限。
// 权限检查是通过ExtraData中的permissions字段进行的。
func RequirePermission[ID comparable, Role comparable](permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文中获取用户身份信息
		identity, exists := GetIdentityFromContext[ID, Role](c)
		if !exists || identity == nil {
			response.Unauthorized(c, "需要登录")
			return
		}

		// 从ExtraData中获取权限
		permissionsAny, exists := identity.GetExtraData("permissions")
		if !exists {
			response.Forbidden(c, "权限不足")
			return
		}

		// 尝试将权限转换为字符串切片
		var permissions []string
		switch p := permissionsAny.(type) {
		case []string:
			permissions = p
		case []interface{}:
			permissions = make([]string, 0, len(p))
			for _, item := range p {
				if str, ok := item.(string); ok {
					permissions = append(permissions, str)
				}
			}
		default:
			response.Forbidden(c, "权限格式无效")
			return
		}

		// 检查是否具有指定的权限
		for _, p := range permissions {
			if p == permission {
				// 继续处理请求
				c.Next()
				return
			}
		}

		// 权限不足
		response.Forbidden(c, "权限不足")
	}
}

// RequireAnyPermission 创建一个中间件，要求用户具有指定的任一权限。
func RequireAnyPermission[ID comparable, Role comparable](permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文中获取用户身份信息
		identity, exists := GetIdentityFromContext[ID, Role](c)
		if !exists || identity == nil {
			response.Unauthorized(c, "需要登录")
			return
		}

		// 从ExtraData中获取权限
		permissionsAny, exists := identity.GetExtraData("permissions")
		if !exists {
			response.Forbidden(c, "权限不足")
			return
		}

		// 尝试将权限转换为字符串切片
		var userPermissions []string
		switch p := permissionsAny.(type) {
		case []string:
			userPermissions = p
		case []interface{}:
			userPermissions = make([]string, 0, len(p))
			for _, item := range p {
				if str, ok := item.(string); ok {
					userPermissions = append(userPermissions, str)
				}
			}
		default:
			response.Forbidden(c, "权限格式无效")
			return
		}

		// 检查是否具有指定的任一权限
		for _, required := range permissions {
			for _, userPerm := range userPermissions {
				if userPerm == required {
					// 继续处理请求
					c.Next()
					return
				}
			}
		}

		// 权限不足
		response.Forbidden(c, "权限不足")
	}
}

// RequireAllPermissions 创建一个中间件，要求用户具有指定的所有权限。
func RequireAllPermissions[ID comparable, Role comparable](permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从上下文中获取用户身份信息
		identity, exists := GetIdentityFromContext[ID, Role](c)
		if !exists || identity == nil {
			response.Unauthorized(c, "需要登录")
			return
		}

		// 从ExtraData中获取权限
		permissionsAny, exists := identity.GetExtraData("permissions")
		if !exists {
			response.Forbidden(c, "权限不足")
			return
		}

		// 尝试将权限转换为字符串切片
		var userPermissions []string
		switch p := permissionsAny.(type) {
		case []string:
			userPermissions = p
		case []interface{}:
			userPermissions = make([]string, 0, len(p))
			for _, item := range p {
				if str, ok := item.(string); ok {
					userPermissions = append(userPermissions, str)
				}
			}
		default:
			response.Forbidden(c, "权限格式无效")
			return
		}

		// 将用户权限转换为映射，便于快速查找
		userPermMap := make(map[string]struct{}, len(userPermissions))
		for _, p := range userPermissions {
			userPermMap[p] = struct{}{}
		}

		// 检查是否具有指定的所有权限
		for _, required := range permissions {
			if _, ok := userPermMap[required]; !ok {
				response.Forbidden(c, "权限不足")
				return
			}
		}

		// 继续处理请求
		c.Next()
	}
}
