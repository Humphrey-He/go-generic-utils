package auth

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/Humphrey-He/go-generic-utils/ginutil/ecode"
	"github.com/Humphrey-He/go-generic-utils/ginutil/response"

	"github.com/gin-gonic/gin"
)

// 常见错误
var (
	// ErrBasicAuthInvalidHeader 表示无效的Basic Auth头部
	ErrBasicAuthInvalidHeader = errors.New("无效的Basic Auth头部")

	// ErrBasicAuthInvalidCredentials 表示无效的Basic Auth凭证
	ErrBasicAuthInvalidCredentials = errors.New("无效的Basic Auth凭证")
)

// BasicAuthConfig 定义Basic Auth中间件的配置选项
type BasicAuthConfig struct {
	// Realm 是Basic Auth的域，显示在浏览器的认证对话框中
	Realm string

	// Validator 是用于验证用户名和密码的函数
	Validator interface{}

	// ErrorHandler 是自定义错误处理函数
	ErrorHandler func(c *gin.Context, err error)
}

// BasicAuthOption 是用于配置Basic Auth中间件的函数类型
type BasicAuthOption func(*BasicAuthConfig)

// WithRealm 设置Basic Auth的域
func WithRealm(realm string) BasicAuthOption {
	return func(config *BasicAuthConfig) {
		config.Realm = realm
	}
}

// WithBasicAuthErrorHandler 设置错误处理函数
func WithBasicAuthErrorHandler(handler func(c *gin.Context, err error)) BasicAuthOption {
	return func(config *BasicAuthConfig) {
		config.ErrorHandler = handler
	}
}

// 默认Basic Auth错误处理器
func defaultBasicAuthErrorHandler(c *gin.Context, err error) {
	// 设置WWW-Authenticate头部，触发浏览器的认证对话框
	realm := "Restricted"
	if config, exists := c.Get("basic_auth_config"); exists {
		if cfg, ok := config.(*BasicAuthConfig); ok && cfg.Realm != "" {
			realm = cfg.Realm
		}
	}

	c.Header("WWW-Authenticate", `Basic realm="`+realm+`"`)

	var message string
	switch {
	case errors.Is(err, ErrBasicAuthInvalidHeader):
		message = "需要认证"
	case errors.Is(err, ErrBasicAuthInvalidCredentials):
		message = "用户名或密码错误"
	default:
		message = "认证失败"
	}

	response.Fail(c, ecode.AccessUnauthorized, message)
}

// NewBasicAuthMiddleware 创建一个新的HTTP Basic Authentication中间件
func NewBasicAuthMiddleware[ID comparable, Role comparable](
	validator func(username, password string) (*UserIdentity[ID, Role], bool),
	options ...BasicAuthOption,
) gin.HandlerFunc {
	// 创建默认配置
	config := &BasicAuthConfig{
		Realm:        "Restricted",
		Validator:    validator,
		ErrorHandler: defaultBasicAuthErrorHandler,
	}

	// 应用选项
	for _, option := range options {
		option(config)
	}

	// 返回中间件处理函数
	return func(c *gin.Context) {
		// 将配置存储到上下文中，以便错误处理器可以访问
		c.Set("basic_auth_config", config)

		// 获取Authorization头部
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			config.ErrorHandler(c, ErrBasicAuthInvalidHeader)
			return
		}

		// 解析Basic认证头部
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "basic" {
			config.ErrorHandler(c, ErrBasicAuthInvalidHeader)
			return
		}

		// 解码Base64凭证
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			config.ErrorHandler(c, ErrBasicAuthInvalidHeader)
			return
		}

		// 解析用户名和密码
		credentials := string(decoded)
		colonIndex := strings.IndexByte(credentials, ':')
		if colonIndex < 0 {
			config.ErrorHandler(c, ErrBasicAuthInvalidCredentials)
			return
		}

		username := credentials[:colonIndex]
		password := credentials[colonIndex+1:]

		// 验证凭证
		if typedValidator, ok := config.Validator.(func(string, string) (*UserIdentity[ID, Role], bool)); ok {
			identity, valid := typedValidator(username, password)
			if !valid {
				config.ErrorHandler(c, ErrBasicAuthInvalidCredentials)
				return
			}

			// 将用户身份信息存储到上下文中
			if identity != nil {
				SetIdentityToContext(c, identity)
			}

			// 继续处理请求
			c.Next()
		} else {
			// 验证器类型不匹配
			panic("auth: Basic Auth验证器类型不匹配")
		}
	}
}

// RequireBasicAuth 创建一个HTTP Basic Authentication中间件的简化版本
func RequireBasicAuth[ID comparable, Role comparable](
	validator func(username, password string) (*UserIdentity[ID, Role], bool),
	realm string,
) gin.HandlerFunc {
	return NewBasicAuthMiddleware(validator, WithRealm(realm))
}

// SimpleBasicAuthValidator 创建一个简单的用户名/密码验证器，使用预定义的凭证映射
func SimpleBasicAuthValidator[ID comparable, Role comparable](
	credentials map[string]string,
	identityProvider func(username string) (*UserIdentity[ID, Role], bool),
) func(username, password string) (*UserIdentity[ID, Role], bool) {
	return func(username, password string) (*UserIdentity[ID, Role], bool) {
		// 验证用户名和密码
		expectedPassword, exists := credentials[username]
		if !exists || expectedPassword != password {
			return nil, false
		}

		// 获取用户身份信息
		if identityProvider != nil {
			return identityProvider(username)
		}

		// 如果没有提供身份提供者，则创建一个基本的身份信息
		return &UserIdentity[ID, Role]{
			Username: username,
		}, true
	}
}

// BasicAuthUserPass 是一个简单的用户名/密码对结构
type BasicAuthUserPass struct {
	Username string
	Password string
}

// SimpleBasicAuthValidatorFromList 创建一个简单的用户名/密码验证器，使用预定义的凭证列表
func SimpleBasicAuthValidatorFromList[ID comparable, Role comparable](
	userPassList []BasicAuthUserPass,
	identityProvider func(username string) (*UserIdentity[ID, Role], bool),
) func(username, password string) (*UserIdentity[ID, Role], bool) {
	// 将列表转换为映射
	credentials := make(map[string]string, len(userPassList))
	for _, up := range userPassList {
		credentials[up.Username] = up.Password
	}

	return SimpleBasicAuthValidator(credentials, identityProvider)
}
