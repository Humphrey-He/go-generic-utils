# Gin 认证与授权中间件 — `auth`

`auth` 包提供了一套用于 Gin 应用的身份认证和授权中间件，支持多种认证方式，并提供统一的用户身份表示和授权机制。

## 主要特性

* **多种认证方式**：
  * JWT 认证（JSON Web Token）
  * Basic 认证（HTTP Basic Authentication）
  * OAuth 2.0 认证
* **统一的用户身份表示**：所有认证方式使用相同的 `UserIdentity` 结构
* **基于角色的访问控制**：支持角色和权限检查
* **泛型支持**：用户 ID 和角色类型可以是任何可比较类型
* **可扩展性**：通过选项模式提供灵活的配置能力
* **与 response 包集成**：使用统一的响应格式

## 安装

```bash
# 假设您已经在项目中引入了该包
```

## 基本用法

### 用户身份

```go
import (
    "ggu/ginutil/middleware/auth"
    "github.com/gin-gonic/gin"
)

// 使用字符串类型的用户ID和角色
type StringID = string
type StringRole = string

// 从上下文中获取用户身份
func GetCurrentUser(c *gin.Context) (*auth.UserIdentity[StringID, StringRole], bool) {
    return auth.GetIdentityFromContext[StringID, StringRole](c)
}

// 检查用户是否具有特定角色
func HasAdminRole(c *gin.Context) bool {
    identity, exists := GetCurrentUser(c)
    if !exists {
        return false
    }
    return identity.HasRole("admin")
}
```

### JWT 认证

```go
import (
    "ggu/ginutil/middleware/auth"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // 创建JWT认证中间件
    jwtMiddleware := auth.RequireJWT[StringID, StringRole](
        auth.HMACKeyFunc([]byte("your-secret-key")),
    )
    
    // 应用到路由组
    authorized := r.Group("/api")
    authorized.Use(jwtMiddleware)
    
    // 受保护的路由
    authorized.GET("/profile", getProfile)
    
    return r
}

// 创建JWT令牌
func createToken(userID string, username string, roles []string) (string, error) {
    claims := &auth.UserIdentityClaims[StringID, StringRole]{
        UserIdentity: auth.UserIdentity[StringID, StringRole]{
            UserID:   userID,
            Username: username,
            Roles:    roles,
        },
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "your-app",
            Subject:   userID,
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("your-secret-key"))
}
```

### Basic 认证

```go
import (
    "ggu/ginutil/middleware/auth"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // 创建Basic认证中间件
    basicAuthMiddleware := auth.RequireBasicAuth[StringID, StringRole](
        validateCredentials,
        "My API",
    )
    
    // 应用到路由组
    authorized := r.Group("/api")
    authorized.Use(basicAuthMiddleware)
    
    // 受保护的路由
    authorized.GET("/profile", getProfile)
    
    return r
}

// 验证用户名和密码
func validateCredentials(username, password string) (*auth.UserIdentity[StringID, StringRole], bool) {
    // 在实际应用中，应该从数据库中查询用户信息
    if username == "admin" && password == "secret" {
        return &auth.UserIdentity[StringID, StringRole]{
            UserID:   "1",
            Username: "admin",
            Roles:    []StringRole{"admin"},
        }, true
    }
    return nil, false
}

// 使用预定义的凭证
func setupSimpleBasicAuth() gin.HandlerFunc {
    credentials := map[string]string{
        "admin": "secret",
        "user":  "password",
    }
    
    validator := auth.SimpleBasicAuthValidator[StringID, StringRole](
        credentials,
        func(username string) (*auth.UserIdentity[StringID, StringRole], bool) {
            // 根据用户名获取用户信息
            var roles []StringRole
            if username == "admin" {
                roles = []StringRole{"admin"}
            } else {
                roles = []StringRole{"user"}
            }
            
            return &auth.UserIdentity[StringID, StringRole]{
                UserID:   username,
                Username: username,
                Roles:    roles,
            }, true
        },
    )
    
    return auth.RequireBasicAuth(validator, "My API")
}
```

### OAuth 2.0 认证

```go
import (
    "ggu/ginutil/middleware/auth"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // 创建OAuth认证中间件
    oauthMiddleware := auth.RequireOAuth[StringID, StringRole](
        "https://oauth-provider.com/introspect",
        "client-id",
        "client-secret",
        extractUserInfo,
    )
    
    // 应用到路由组
    authorized := r.Group("/api")
    authorized.Use(oauthMiddleware)
    
    // 受保护的路由
    authorized.GET("/profile", getProfile)
    
    return r
}

// 从OAuth令牌内省响应中提取用户信息
func extractUserInfo(response *auth.TokenIntrospectionResponse) (*auth.UserIdentity[StringID, StringRole], error) {
    // 提取用户ID
    userID := response.Sub
    
    // 提取用户名
    username := response.Username
    
    // 提取角色（假设在自定义字段中）
    var roles []StringRole
    if rolesAny, ok := response.Extra["roles"]; ok {
        if rolesArr, ok := rolesAny.([]interface{}); ok {
            for _, role := range rolesArr {
                if roleStr, ok := role.(string); ok {
                    roles = append(roles, StringRole(roleStr))
                }
            }
        }
    }
    
    return &auth.UserIdentity[StringID, StringRole]{
        UserID:   userID,
        Username: username,
        Roles:    roles,
    }, nil
}

// 使用简化的提取器
func setupSimpleOAuth() gin.HandlerFunc {
    extractor := auth.SimpleOAuthUserInfoExtractor[StringID, StringRole](
        "sub",           // 用户ID字段
        "username",      // 用户名字段
        "roles",         // 角色字段
        "tenant_id",     // 租户ID字段
        func(val interface{}) []StringRole {
            // 将角色值转换为[]StringRole
            var roles []StringRole
            if rolesArr, ok := val.([]interface{}); ok {
                for _, role := range rolesArr {
                    if roleStr, ok := role.(string); ok {
                        roles = append(roles, StringRole(roleStr))
                    }
                }
            }
            return roles
        },
        func(val interface{}) StringID {
            // 将用户ID值转换为StringID
            if idStr, ok := val.(string); ok {
                return StringID(idStr)
            }
            return ""
        },
    )
    
    return auth.RequireOAuth(
        "https://oauth-provider.com/introspect",
        "client-id",
        "client-secret",
        extractor,
    )
}
```

### 授权中间件

```go
import (
    "ggu/ginutil/middleware/auth"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // 认证中间件
    jwtMiddleware := auth.RequireJWT[StringID, StringRole](
        auth.HMACKeyFunc([]byte("your-secret-key")),
    )
    
    // 应用到路由组
    authorized := r.Group("/api")
    authorized.Use(jwtMiddleware)
    
    // 要求特定角色
    adminOnly := authorized.Group("/admin")
    adminOnly.Use(auth.RequireRoles[StringID, StringRole]("admin"))
    
    // 要求任一角色
    moderatorOrAdmin := authorized.Group("/moderate")
    moderatorOrAdmin.Use(auth.RequireAnyRole[StringID, StringRole]("admin", "moderator"))
    
    // 要求特定权限
    authorized.GET("/users", auth.RequirePermission[StringID, StringRole]("users:read"), listUsers)
    authorized.POST("/users", auth.RequirePermission[StringID, StringRole]("users:create"), createUser)
    
    // 自定义授权逻辑
    authorized.GET("/resource/:id", auth.Authorize[StringID, StringRole](canAccessResource), getResource)
    
    return r
}

// 自定义授权函数
func canAccessResource(identity *auth.UserIdentity[StringID, StringRole], c *gin.Context) bool {
    resourceID := c.Param("id")
    
    // 管理员可以访问所有资源
    if identity.HasRole("admin") {
        return true
    }
    
    // 检查资源是否属于当前用户
    ownedResources, exists := identity.GetExtraData("owned_resources")
    if !exists {
        return false
    }
    
    if resources, ok := ownedResources.([]string); ok {
        for _, id := range resources {
            if id == resourceID {
                return true
            }
        }
    }
    
    return false
}
```

## 高级配置

### JWT 配置选项

```go
jwtMiddleware := auth.NewJWTMiddleware(
    // 必需的选项
    auth.WithValidationKeyGetter(auth.HMACKeyFunc([]byte("your-secret-key"))),
    auth.WithClaimsFactory(auth.UserIdentityClaimsFactory[StringID, StringRole]()),
    
    // 可选的选项
    auth.WithSuccessHandler(auth.UserIdentitySuccessHandler[StringID, StringRole]()),
    auth.WithTokenExtractor(auth.CombineExtractors(
        auth.ExtractJWTFromHeader,
        auth.ExtractJWTFromQuery("token"),
        auth.ExtractJWTFromCookie("jwt"),
    )),
    auth.WithSigningMethods("HS256", "RS256"),
    auth.WithErrorHandler(func(c *gin.Context, err error) {
        // 自定义错误处理
        c.JSON(401, gin.H{"error": err.Error()})
        c.Abort()
    }),
)
```

### Basic Auth 配置选项

```go
basicAuthMiddleware := auth.NewBasicAuthMiddleware[StringID, StringRole](
    validateCredentials,
    auth.WithRealm("Secure API"),
    auth.WithBasicAuthErrorHandler(func(c *gin.Context, err error) {
        // 自定义错误处理
        c.JSON(401, gin.H{"error": err.Error()})
        c.Abort()
    }),
)
```

### OAuth 配置选项

```go
oauthMiddleware := auth.NewOAuth2Middleware[StringID, StringRole](
    extractUserInfo,
    auth.WithTokenIntrospectionEndpoint("https://oauth-provider.com/introspect"),
    auth.WithClientCredentials("client-id", "client-secret"),
    auth.WithOAuthTokenExtractor(auth.ExtractJWTFromHeader),
    auth.WithOAuthErrorHandler(func(c *gin.Context, err error) {
        // 自定义错误处理
        c.JSON(401, gin.H{"error": err.Error()})
        c.Abort()
    }),
    auth.WithHTTPClient(&http.Client{
        Timeout: 5 * time.Second,
    }),
)
```

## 安全性注意事项

1. **密钥管理**：JWT 签名密钥应妥善保管，建议使用环境变量或密钥管理服务。
2. **HTTPS**：Basic Authentication 只有在 HTTPS 下才是安全的，明文传输密码极其危险。
3. **令牌过期**：为 JWT 和 OAuth 令牌设置合理的过期时间，并实现令牌刷新机制。
4. **最小权限原则**：为用户分配最小必要的角色和权限。
5. **防止暴力破解**：实现请求限流和账户锁定机制。 