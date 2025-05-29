package contextx_test

import (
	"testing"

	"github.com/noobtrump/go-generic-utils/ginutil/contextx"
	"github.com/noobtrump/go-generic-utils/ginutil/middleware/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUserIdentity(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		UserID:   123,
		Username: "testuser",
		Roles:    []string{"admin", "user"},
		TenantID: "tenant-1",
	}

	// 测试设置和获取用户身份信息
	contextx.SetUserIdentity(c, identity)
	retrieved, exists := contextx.GetUserIdentity[int, string](c)

	assert.True(t, exists, "用户身份信息应该存在")
	assert.Equal(t, identity, retrieved, "用户身份信息应该正确")

	// 测试 MustGetUserIdentity
	assert.NotPanics(t, func() {
		retrieved := contextx.MustGetUserIdentity[int, string](c)
		assert.Equal(t, identity, retrieved, "用户身份信息应该正确")
	})

	// 测试 nil 上下文
	_, exists = contextx.GetUserIdentity[int, string](nil)
	assert.False(t, exists, "nil 上下文应该返回 false")

	assert.Panics(t, func() {
		contextx.MustGetUserIdentity[int, string](nil)
	}, "nil 上下文应该 panic")
}

func TestGetUserID(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		UserID: 123,
	}

	// 设置用户身份信息
	contextx.SetUserIdentity(c, identity)

	// 测试 GetUserID
	userID, exists := contextx.GetUserID[int](c)
	assert.True(t, exists, "用户 ID 应该存在")
	assert.Equal(t, 123, userID, "用户 ID 应该正确")

	// 测试 MustGetUserID
	assert.NotPanics(t, func() {
		userID := contextx.MustGetUserID[int](c)
		assert.Equal(t, 123, userID, "用户 ID 应该正确")
	})

	// 测试不存在的用户身份信息
	c2, _ := gin.CreateTestContext(nil)
	_, exists = contextx.GetUserID[int](c2)
	assert.False(t, exists, "不存在的用户身份信息应该返回 false")

	assert.Panics(t, func() {
		contextx.MustGetUserID[int](c2)
	}, "不存在的用户身份信息应该 panic")
}

func TestGetUsername(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		Username: "testuser",
	}

	// 设置用户身份信息
	contextx.SetUserIdentity(c, identity)

	// 测试 GetUsername
	username, exists := contextx.GetUsername[int](c)
	assert.True(t, exists, "用户名应该存在")
	assert.Equal(t, "testuser", username, "用户名应该正确")
}

func TestGetUserRoles(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		Roles: []string{"admin", "user"},
	}

	// 设置用户身份信息
	contextx.SetUserIdentity(c, identity)

	// 测试 GetUserRoles
	roles, exists := contextx.GetUserRoles[int, string](c)
	assert.True(t, exists, "用户角色应该存在")
	assert.Equal(t, []string{"admin", "user"}, roles, "用户角色应该正确")
}

func TestHasRole(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		Roles: []string{"admin", "user"},
	}

	// 设置用户身份信息
	contextx.SetUserIdentity(c, identity)

	// 测试 HasRole
	assert.True(t, contextx.HasRole[int, string](c, "admin"), "用户应该具有 admin 角色")
	assert.True(t, contextx.HasRole[int, string](c, "user"), "用户应该具有 user 角色")
	assert.False(t, contextx.HasRole[int, string](c, "guest"), "用户不应该具有 guest 角色")

	// 测试 HasAnyRole
	assert.True(t, contextx.HasAnyRole[int, string](c, "admin", "guest"), "用户应该具有 admin 角色")
	assert.False(t, contextx.HasAnyRole[int, string](c, "guest", "visitor"), "用户不应该具有 guest 或 visitor 角色")

	// 测试 HasAllRoles
	assert.True(t, contextx.HasAllRoles[int, string](c, "admin", "user"), "用户应该同时具有 admin 和 user 角色")
	assert.False(t, contextx.HasAllRoles[int, string](c, "admin", "guest"), "用户不应该同时具有 admin 和 guest 角色")
}

func TestGetTenantID(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		TenantID: "tenant-1",
	}

	// 设置用户身份信息
	contextx.SetUserIdentity(c, identity)

	// 测试 GetTenantID
	tenantID, exists := contextx.GetTenantID[int](c)
	assert.True(t, exists, "租户 ID 应该存在")
	assert.Equal(t, "tenant-1", tenantID, "租户 ID 应该正确")
}

func TestGetUserExtraData(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 创建用户身份信息
	identity := &auth.UserIdentity[int, string]{
		ExtraData: map[string]any{
			"age":     30,
			"email":   "test@example.com",
			"premium": true,
		},
	}

	// 设置用户身份信息
	contextx.SetUserIdentity(c, identity)

	// 测试 GetUserExtraData
	age, exists := contextx.GetUserExtraData[int, int](c, "age")
	assert.True(t, exists, "年龄应该存在")
	assert.Equal(t, 30, age, "年龄应该正确")

	email, exists := contextx.GetUserExtraData[int, string](c, "email")
	assert.True(t, exists, "邮箱应该存在")
	assert.Equal(t, "test@example.com", email, "邮箱应该正确")

	premium, exists := contextx.GetUserExtraData[int, bool](c, "premium")
	assert.True(t, exists, "会员状态应该存在")
	assert.True(t, premium, "会员状态应该正确")

	// 测试不存在的数据
	_, exists = contextx.GetUserExtraData[int, string](c, "not-exist")
	assert.False(t, exists, "不存在的数据应该返回 false")

	// 测试类型不匹配
	_, exists = contextx.GetUserExtraData[int, int](c, "email")
	assert.False(t, exists, "类型不匹配应该返回 false")
}
