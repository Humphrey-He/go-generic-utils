// Package contextx 提供了对 gin.Context 操作的工具函数。
package contextx

import (
	"ggu/ginutil/middleware/auth"

	"github.com/gin-gonic/gin"
)

// UserIdentityKey 是存储用户身份信息的上下文键。
// 为保持一致性，我们使用与 auth 包相同的键。
const UserIdentityKey = auth.UserIdentityKey

// SetUserIdentity 将用户身份信息设置到 gin.Context 中。
func SetUserIdentity[ID comparable, Role comparable](c *gin.Context, identity *auth.UserIdentity[ID, Role]) {
	if c == nil || identity == nil {
		return
	}
	Set(c, UserIdentityKey, identity)
}

// GetUserIdentity 从 gin.Context 中获取用户身份信息。
// 返回值：
//   - 第一个返回值是获取到的用户身份信息（如果获取失败则为 nil）
//   - 第二个返回值表示是否成功获取
func GetUserIdentity[ID comparable, Role comparable](c *gin.Context) (*auth.UserIdentity[ID, Role], bool) {
	return Get[*auth.UserIdentity[ID, Role]](c, UserIdentityKey)
}

// MustGetUserIdentity 从 gin.Context 中获取用户身份信息。
// 如果获取失败，将会 panic。
func MustGetUserIdentity[ID comparable, Role comparable](c *gin.Context) *auth.UserIdentity[ID, Role] {
	return MustGet[*auth.UserIdentity[ID, Role]](c, UserIdentityKey)
}

// GetUserID 从上下文中获取用户ID。
// 返回值：
//   - 第一个返回值是用户ID（如果获取失败则为零值）
//   - 第二个返回值表示是否成功获取
func GetUserID[ID comparable](c *gin.Context) (ID, bool) {
	var zeroID ID
	identity, ok := GetUserIdentity[ID, any](c)
	if !ok || identity == nil {
		return zeroID, false
	}
	return identity.UserID, true
}

// MustGetUserID 从上下文中获取用户ID。
// 如果获取失败，将会 panic。
func MustGetUserID[ID comparable](c *gin.Context) ID {
	identity := MustGetUserIdentity[ID, any](c)
	return identity.UserID
}

// GetUsername 从上下文中获取用户名。
// 返回值：
//   - 第一个返回值是用户名（如果获取失败则为空字符串）
//   - 第二个返回值表示是否成功获取
func GetUsername[ID comparable](c *gin.Context) (string, bool) {
	identity, ok := GetUserIdentity[ID, any](c)
	if !ok || identity == nil {
		return "", false
	}
	return identity.Username, true
}

// GetUserRoles 从上下文中获取用户角色列表。
// 返回值：
//   - 第一个返回值是用户角色列表（如果获取失败则为 nil）
//   - 第二个返回值表示是否成功获取
func GetUserRoles[ID comparable, Role comparable](c *gin.Context) ([]Role, bool) {
	identity, ok := GetUserIdentity[ID, Role](c)
	if !ok || identity == nil {
		return nil, false
	}
	return identity.Roles, true
}

// HasRole 检查上下文中的用户是否具有指定角色。
func HasRole[ID comparable, Role comparable](c *gin.Context, role Role) bool {
	identity, ok := GetUserIdentity[ID, Role](c)
	if !ok || identity == nil {
		return false
	}
	return identity.HasRole(role)
}

// HasAnyRole 检查上下文中的用户是否具有任一指定角色。
func HasAnyRole[ID comparable, Role comparable](c *gin.Context, roles ...Role) bool {
	identity, ok := GetUserIdentity[ID, Role](c)
	if !ok || identity == nil {
		return false
	}
	return identity.HasAnyRole(roles...)
}

// HasAllRoles 检查上下文中的用户是否具有所有指定角色。
func HasAllRoles[ID comparable, Role comparable](c *gin.Context, roles ...Role) bool {
	identity, ok := GetUserIdentity[ID, Role](c)
	if !ok || identity == nil {
		return false
	}
	return identity.HasAllRoles(roles...)
}

// GetTenantID 从上下文中获取租户ID。
// 返回值：
//   - 第一个返回值是租户ID（如果获取失败则为空字符串）
//   - 第二个返回值表示是否成功获取
func GetTenantID[ID comparable](c *gin.Context) (string, bool) {
	identity, ok := GetUserIdentity[ID, any](c)
	if !ok || identity == nil {
		return "", false
	}
	return identity.TenantID, true
}

// GetUserExtraData 从上下文中的用户身份信息获取额外数据。
// 返回值：
//   - 第一个返回值是额外数据值（如果获取失败则为 nil）
//   - 第二个返回值表示是否成功获取
func GetUserExtraData[ID comparable, T any](c *gin.Context, key string) (T, bool) {
	var zeroT T
	identity, ok := GetUserIdentity[ID, any](c)
	if !ok || identity == nil {
		return zeroT, false
	}

	value, exists := identity.GetExtraData(key)
	if !exists {
		return zeroT, false
	}

	typedValue, ok := value.(T)
	if !ok {
		return zeroT, false
	}

	return typedValue, true
}
