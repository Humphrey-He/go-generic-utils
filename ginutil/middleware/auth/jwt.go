package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Humphrey-He/go-generic-utils/ginutil/ecode"
	"github.com/Humphrey-He/go-generic-utils/ginutil/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 常见错误
var (
	// ErrJWTMissingToken 表示请求中缺少JWT令牌
	ErrJWTMissingToken = errors.New("缺少认证令牌")

	// ErrJWTInvalidToken 表示JWT令牌无效
	ErrJWTInvalidToken = errors.New("无效的认证令牌")

	// ErrJWTExpiredToken 表示JWT令牌已过期
	ErrJWTExpiredToken = errors.New("认证令牌已过期")

	// ErrJWTInvalidSigningMethod 表示JWT令牌使用了不支持的签名方法
	ErrJWTInvalidSigningMethod = errors.New("不支持的令牌签名方法")
)

// JWTConfig 定义JWT中间件的配置选项
type JWTConfig struct {
	// ValidationKeyGetter 用于获取验证签名的密钥（对称密钥或公钥）
	ValidationKeyGetter func(token *jwt.Token) (interface{}, error)

	// ClaimsFactory 用于指定解析Payload时使用的Claims结构体实例
	ClaimsFactory func() jwt.Claims

	// ErrorHandler 自定义错误处理函数
	ErrorHandler func(c *gin.Context, err error)

	// SuccessHandler 认证成功后的回调
	SuccessHandler func(c *gin.Context, token *jwt.Token, claims jwt.Claims)

	// TokenExtractor 自定义从请求中提取Token的方式
	TokenExtractor func(c *gin.Context) (string, error)

	// SigningMethods 指定期望的签名算法列表
	SigningMethods []string
}

// JWTOption 是用于配置JWT中间件的函数类型
type JWTOption func(*JWTConfig)

// WithValidationKeyGetter 设置验证密钥获取函数
func WithValidationKeyGetter(getter func(token *jwt.Token) (interface{}, error)) JWTOption {
	return func(config *JWTConfig) {
		config.ValidationKeyGetter = getter
	}
}

// WithClaimsFactory 设置Claims工厂函数
func WithClaimsFactory(factory func() jwt.Claims) JWTOption {
	return func(config *JWTConfig) {
		config.ClaimsFactory = factory
	}
}

// WithErrorHandler 设置错误处理函数
func WithErrorHandler(handler func(c *gin.Context, err error)) JWTOption {
	return func(config *JWTConfig) {
		config.ErrorHandler = handler
	}
}

// WithSuccessHandler 设置成功处理函数
func WithSuccessHandler(handler func(c *gin.Context, token *jwt.Token, claims jwt.Claims)) JWTOption {
	return func(config *JWTConfig) {
		config.SuccessHandler = handler
	}
}

// WithTokenExtractor 设置令牌提取函数
func WithTokenExtractor(extractor func(c *gin.Context) (string, error)) JWTOption {
	return func(config *JWTConfig) {
		config.TokenExtractor = extractor
	}
}

// WithSigningMethods 设置允许的签名方法
func WithSigningMethods(methods ...string) JWTOption {
	return func(config *JWTConfig) {
		config.SigningMethods = methods
	}
}

// 默认令牌提取器，从Authorization头部提取Bearer令牌
func defaultTokenExtractor(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", ErrJWTMissingToken
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrJWTMissingToken
	}

	return parts[1], nil
}

// 默认错误处理器，使用response包返回统一错误格式
func defaultErrorHandler(c *gin.Context, err error) {
	var message string
	var code int

	switch {
	case errors.Is(err, ErrJWTMissingToken):
		message = "缺少认证令牌"
		code = ecode.AccessUnauthorized
	case errors.Is(err, ErrJWTInvalidToken):
		message = "无效的认证令牌"
		code = ecode.AccessUnauthorized
	case errors.Is(err, ErrJWTExpiredToken):
		message = "认证令牌已过期"
		code = ecode.AccessUnauthorized
	case errors.Is(err, ErrJWTInvalidSigningMethod):
		message = "不支持的令牌签名方法"
		code = ecode.AccessUnauthorized
	default:
		message = "认证失败"
		code = ecode.AccessUnauthorized
	}

	response.Fail(c, code, message)
}

// 默认成功处理器，将解析出的Claims存入上下文
func defaultSuccessHandler(c *gin.Context, token *jwt.Token, claims jwt.Claims) {
	// 注意：这里不做任何操作，因为不同的Claims类型需要不同的处理方式
	// 用户应该提供自己的SuccessHandler来处理特定类型的Claims
}

// NewJWTMiddleware 创建一个新的JWT认证中间件
func NewJWTMiddleware(options ...JWTOption) gin.HandlerFunc {
	// 创建默认配置
	config := &JWTConfig{
		TokenExtractor: defaultTokenExtractor,
		ErrorHandler:   defaultErrorHandler,
		SuccessHandler: defaultSuccessHandler,
		SigningMethods: []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512", "ES256", "ES384", "ES512"},
	}

	// 应用选项
	for _, option := range options {
		option(config)
	}

	// 验证必需的配置
	if config.ValidationKeyGetter == nil {
		panic("auth: JWT中间件需要ValidationKeyGetter选项")
	}

	if config.ClaimsFactory == nil {
		panic("auth: JWT中间件需要ClaimsFactory选项")
	}

	// 返回中间件处理函数
	return func(c *gin.Context) {
		// 提取令牌
		tokenString, err := config.TokenExtractor(c)
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		// 创建Claims实例
		claims := config.ClaimsFactory()

		// 解析并验证令牌
		token, err := jwt.ParseWithClaims(tokenString, claims, config.ValidationKeyGetter)

		// 处理解析错误
		if err != nil {
			// 根据错误类型设置更具体的错误
			if errors.Is(err, jwt.ErrTokenExpired) {
				err = ErrJWTExpiredToken
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				err = ErrJWTInvalidToken
			} else {
				err = ErrJWTInvalidToken
			}

			config.ErrorHandler(c, err)
			return
		}

		// 验证签名方法
		if len(config.SigningMethods) > 0 {
			var validMethod bool
			for _, method := range config.SigningMethods {
				if token.Method.Alg() == method {
					validMethod = true
					break
				}
			}

			if !validMethod {
				err := fmt.Errorf("%w: %s", ErrJWTInvalidSigningMethod, token.Method.Alg())
				config.ErrorHandler(c, err)
				return
			}
		}

		// 验证令牌有效性
		if !token.Valid {
			config.ErrorHandler(c, ErrJWTInvalidToken)
			return
		}

		// 调用成功处理器
		config.SuccessHandler(c, token, claims)

		// 继续处理请求
		c.Next()
	}
}

// ExtractJWTFromHeader 从Authorization头部提取JWT令牌
func ExtractJWTFromHeader(c *gin.Context) (string, error) {
	return defaultTokenExtractor(c)
}

// ExtractJWTFromQuery 从URL查询参数提取JWT令牌
func ExtractJWTFromQuery(paramName string) func(c *gin.Context) (string, error) {
	return func(c *gin.Context) (string, error) {
		token := c.Query(paramName)
		if token == "" {
			return "", ErrJWTMissingToken
		}
		return token, nil
	}
}

// ExtractJWTFromCookie 从Cookie提取JWT令牌
func ExtractJWTFromCookie(cookieName string) func(c *gin.Context) (string, error) {
	return func(c *gin.Context) (string, error) {
		cookie, err := c.Cookie(cookieName)
		if err != nil || cookie == "" {
			return "", ErrJWTMissingToken
		}
		return cookie, nil
	}
}

// CombineExtractors 组合多个令牌提取器，按顺序尝试提取，直到成功
func CombineExtractors(extractors ...func(c *gin.Context) (string, error)) func(c *gin.Context) (string, error) {
	return func(c *gin.Context) (string, error) {
		for _, extractor := range extractors {
			token, err := extractor(c)
			if err == nil && token != "" {
				return token, nil
			}
		}
		return "", ErrJWTMissingToken
	}
}

// UserIdentityClaimsFactory 创建一个Claims工厂函数，返回嵌入了jwt.RegisteredClaims的UserIdentity
func UserIdentityClaimsFactory[ID comparable, Role comparable]() func() jwt.Claims {
	return func() jwt.Claims {
		return &UserIdentityClaims[ID, Role]{}
	}
}

// UserIdentityClaims 是嵌入了jwt.RegisteredClaims的UserIdentity
type UserIdentityClaims[ID comparable, Role comparable] struct {
	UserIdentity[ID, Role]
	jwt.RegisteredClaims
}

// UserIdentitySuccessHandler 创建一个成功处理器，将UserIdentityClaims中的UserIdentity提取到上下文
func UserIdentitySuccessHandler[ID comparable, Role comparable]() func(c *gin.Context, token *jwt.Token, claims jwt.Claims) {
	return func(c *gin.Context, token *jwt.Token, claims jwt.Claims) {
		if userClaims, ok := claims.(*UserIdentityClaims[ID, Role]); ok {
			SetIdentityToContext(c, &userClaims.UserIdentity)
		}
	}
}

// RequireJWT 创建一个JWT认证中间件，使用UserIdentity作为Claims
func RequireJWT[ID comparable, Role comparable](keyFunc func(token *jwt.Token) (interface{}, error), options ...JWTOption) gin.HandlerFunc {
	// 创建基本选项
	baseOptions := []JWTOption{
		WithValidationKeyGetter(keyFunc),
		WithClaimsFactory(UserIdentityClaimsFactory[ID, Role]()),
		WithSuccessHandler(UserIdentitySuccessHandler[ID, Role]()),
	}

	// 合并用户提供的选项
	allOptions := append(baseOptions, options...)

	return NewJWTMiddleware(allOptions...)
}

// 创建一个简单的HMAC密钥验证函数
func HMACKeyFunc(secret []byte) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否为HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrJWTInvalidSigningMethod, token.Header["alg"])
		}
		return secret, nil
	}
}

// 创建一个简单的RSA公钥验证函数
func RSAKeyFunc(publicKey interface{}) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否为RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: %v", ErrJWTInvalidSigningMethod, token.Header["alg"])
		}
		return publicKey, nil
	}
}

// RequireRoles 创建一个中间件，要求用户具有指定的所有角色
func RequireRoles[ID comparable, Role comparable](roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := GetIdentityFromContext[ID, Role](c)
		if !ok {
			response.Unauthorized(c, "需要登录")
			return
		}

		if !identity.HasAllRoles(roles...) {
			response.Forbidden(c, "权限不足")
			return
		}

		c.Next()
	}
}

// RequireAnyRole 创建一个中间件，要求用户具有指定的任一角色
func RequireAnyRole[ID comparable, Role comparable](roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := GetIdentityFromContext[ID, Role](c)
		if !ok {
			response.Unauthorized(c, "需要登录")
			return
		}

		if !identity.HasAnyRole(roles...) {
			response.Forbidden(c, "权限不足")
			return
		}

		c.Next()
	}
}
