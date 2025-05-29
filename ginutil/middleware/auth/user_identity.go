package auth

import "github.com/gin-gonic/gin"

// UserIdentityKey 是用于在Gin上下文中存储 UserIdentity 的键。
const UserIdentityKey = "gokits.auth.userIdentity"

// UserIdentity 保存已认证用户的信息。
// 如果不同服务有截然不同的身份结构，可以考虑泛型，
// 但为了系统内一致性，通用结构通常更好。
// 目前设为具体结构，但用 `ExtraData any` 增加灵活性。
type UserIdentity[ID comparable, Role comparable] struct {
	UserID    ID             `json:"user_id"`              // 用户唯一标识
	Username  string         `json:"username,omitempty"`   // 用户名 (可选)
	Roles     []Role         `json:"roles,omitempty"`      // 用户角色列表 (可选)
	TenantID  string         `json:"tenant_id,omitempty"`  // 租户ID (可选, B2B场景)
	ExtraData map[string]any `json:"extra_data,omitempty"` // 额外的自定义数据
}

// SetIdentityToContext 将 UserIdentity 设置到 Gin 上下文中。
func SetIdentityToContext[ID comparable, Role comparable](c *gin.Context, identity *UserIdentity[ID, Role]) {
	if c != nil && identity != nil {
		c.Set(UserIdentityKey, identity)
	}
}

// GetIdentityFromContext 从 Gin 上下文中检索 UserIdentity。
// 如果找到则返回身份信息和 true，否则返回 nil 和 false。
func GetIdentityFromContext[ID comparable, Role comparable](c *gin.Context) (*UserIdentity[ID, Role], bool) {
	if c == nil {
		return nil, false
	}

	identity, exists := c.Get(UserIdentityKey)
	if !exists {
		return nil, false
	}

	castedIdentity, ok := identity.(*UserIdentity[ID, Role])
	if !ok {
		return nil, false // 类型断言失败
	}

	return castedIdentity, true
}

// MustGetIdentityFromContext 检索 UserIdentity，如果未找到或类型错误则 panic。
func MustGetIdentityFromContext[ID comparable, Role comparable](c *gin.Context) *UserIdentity[ID, Role] {
	identity, ok := GetIdentityFromContext[ID, Role](c)
	if !ok {
		panic("auth: UserIdentity not found in context or type mismatch")
	}
	return identity
}

// HasRole 检查用户是否具有指定角色。
func (u *UserIdentity[ID, Role]) HasRole(role Role) bool {
	if u == nil || len(u.Roles) == 0 {
		return false
	}

	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// HasAnyRole 检查用户是否具有任一指定角色。
func (u *UserIdentity[ID, Role]) HasAnyRole(roles ...Role) bool {
	if u == nil || len(u.Roles) == 0 || len(roles) == 0 {
		return false
	}

	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}

	return false
}

// HasAllRoles 检查用户是否具有所有指定角色。
func (u *UserIdentity[ID, Role]) HasAllRoles(roles ...Role) bool {
	if u == nil || len(u.Roles) == 0 || len(roles) == 0 {
		return false
	}

	roleMap := make(map[Role]struct{}, len(u.Roles))
	for _, r := range u.Roles {
		roleMap[r] = struct{}{}
	}

	for _, role := range roles {
		if _, exists := roleMap[role]; !exists {
			return false
		}
	}

	return true
}

// GetExtraData 获取额外数据中的特定字段值。
func (u *UserIdentity[ID, Role]) GetExtraData(key string) (any, bool) {
	if u == nil || u.ExtraData == nil {
		return nil, false
	}

	value, exists := u.ExtraData[key]
	return value, exists
}

// SetExtraData 设置额外数据中的特定字段值。
func (u *UserIdentity[ID, Role]) SetExtraData(key string, value any) {
	if u == nil {
		return
	}

	if u.ExtraData == nil {
		u.ExtraData = make(map[string]any)
	}

	u.ExtraData[key] = value
}
