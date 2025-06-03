package auth_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 示例类型定义
type StringID = string
type StringRole = string

// ExampleUserIdentity 展示如何使用 UserIdentity 结构
func Example_userIdentity() {
	// 创建用户身份
	identity := &auth.UserIdentity[StringID, StringRole]{
		UserID:   "123",
		Username: "张三",
		Roles:    []StringRole{"admin", "user"},
		TenantID: "tenant1",
		ExtraData: map[string]any{
			"department":  "技术部",
			"permissions": []string{"users:read", "users:write"},
		},
	}

	// 检查角色
	hasAdmin := identity.HasRole("admin")
	hasAnyRole := identity.HasAnyRole("editor", "admin")
	hasAllRoles := identity.HasAllRoles("admin", "user")

	// 获取额外数据
	department, exists := identity.GetExtraData("department")

	// 输出结果
	fmt.Printf("用户ID: %s\n", identity.UserID)
	fmt.Printf("用户名: %s\n", identity.Username)
	fmt.Printf("是否为管理员: %v\n", hasAdmin)
	fmt.Printf("是否有任一角色: %v\n", hasAnyRole)
	fmt.Printf("是否有所有角色: %v\n", hasAllRoles)
	fmt.Printf("部门: %s (存在: %v)\n", department, exists)

	// Output:
	// 用户ID: 123
	// 用户名: 张三
	// 是否为管理员: true
	// 是否有任一角色: true
	// 是否有所有角色: true
	// 部门: 技术部 (存在: true)
}

// ExampleJWT 展示如何使用 JWT 认证中间件
func Example_jwt() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 创建 JWT 中间件
	jwtMiddleware := auth.RequireJWT[StringID, StringRole](
		auth.HMACKeyFunc([]byte("test-secret")),
	)

	// 应用中间件
	r.GET("/api/profile", jwtMiddleware, func(c *gin.Context) {
		// 获取用户身份
		identity, exists := auth.GetIdentityFromContext[StringID, StringRole](c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}

		// 返回用户信息
		c.JSON(http.StatusOK, gin.H{
			"user_id":   identity.UserID,
			"username":  identity.Username,
			"roles":     identity.Roles,
			"tenant_id": identity.TenantID,
		})
	})

	// 创建有效的 JWT 令牌
	claims := &auth.UserIdentityClaims[StringID, StringRole]{
		UserIdentity: auth.UserIdentity[StringID, StringRole]{
			UserID:   "123",
			Username: "张三",
			Roles:    []StringRole{"admin"},
			TenantID: "tenant1",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	// 创建请求
	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出结果
	fmt.Printf("状态码: %d\n", w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	fmt.Printf("用户ID: %s\n", resp["user_id"])
	fmt.Printf("用户名: %s\n", resp["username"])

	// Output:
	// 状态码: 200
	// 用户ID: 123
	// 用户名: 张三
}

// ExampleBasicAuth 展示如何使用 Basic Auth 认证中间件
func Example_basicAuth() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 创建验证函数
	validator := func(username, password string) (*auth.UserIdentity[StringID, StringRole], bool) {
		if username == "admin" && password == "secret" {
			return &auth.UserIdentity[StringID, StringRole]{
				UserID:   "1",
				Username: "admin",
				Roles:    []StringRole{"admin"},
			}, true
		}
		return nil, false
	}

	// 创建 Basic Auth 中间件
	basicAuthMiddleware := auth.RequireBasicAuth[StringID, StringRole](validator, "Test API")

	// 应用中间件
	r.GET("/api/profile", basicAuthMiddleware, func(c *gin.Context) {
		// 获取用户身份
		identity, exists := auth.GetIdentityFromContext[StringID, StringRole](c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}

		// 返回用户信息
		c.JSON(http.StatusOK, gin.H{
			"user_id":  identity.UserID,
			"username": identity.Username,
			"roles":    identity.Roles,
		})
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.SetBasicAuth("admin", "secret")
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出结果
	fmt.Printf("状态码: %d\n", w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	fmt.Printf("用户ID: %s\n", resp["user_id"])
	fmt.Printf("用户名: %s\n", resp["username"])

	// Output:
	// 状态码: 200
	// 用户ID: 1
	// 用户名: admin
}

// ExampleRoleBasedAuthorization 展示如何使用基于角色的授权中间件
func Example_roleBasedAuthorization() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 模拟认证中间件，直接设置用户身份
	authMiddleware := func(c *gin.Context) {
		identity := &auth.UserIdentity[StringID, StringRole]{
			UserID:   "123",
			Username: "张三",
			Roles:    []StringRole{"editor"},
			ExtraData: map[string]any{
				"permissions": []string{"articles:read", "articles:write"},
			},
		}
		auth.SetIdentityToContext(c, identity)
		c.Next()
	}

	// 创建需要管理员角色的路由
	r.GET("/admin/dashboard", authMiddleware, auth.RequireRoles[StringID, StringRole]("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "管理员面板")
	})

	// 创建需要编辑者或管理员角色的路由
	r.GET("/articles", authMiddleware, auth.RequireAnyRole[StringID, StringRole]("editor", "admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "文章列表")
	})

	// 创建需要特定权限的路由
	r.POST("/articles", authMiddleware, auth.RequirePermission[StringID, StringRole]("articles:write"), func(c *gin.Context) {
		c.String(http.StatusOK, "创建文章成功")
	})

	// 测试管理员面板（应该失败，因为用户是编辑者而非管理员）
	req1 := httptest.NewRequest("GET", "/admin/dashboard", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	fmt.Printf("管理员面板访问状态码: %d\n", w1.Code)

	// 测试文章列表（应该成功，因为用户是编辑者）
	req2 := httptest.NewRequest("GET", "/articles", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	fmt.Printf("文章列表访问状态码: %d\n", w2.Code)

	// 测试创建文章（应该成功，因为用户有articles:write权限）
	req3 := httptest.NewRequest("POST", "/articles", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	fmt.Printf("创建文章访问状态码: %d\n", w3.Code)

	// Output:
	// 管理员面板访问状态码: 403
	// 文章列表访问状态码: 200
	// 创建文章访问状态码: 200
}
