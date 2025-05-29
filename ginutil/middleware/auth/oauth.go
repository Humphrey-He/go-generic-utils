package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/noobtrump/go-generic-utils/ginutil/ecode"
	"github.com/noobtrump/go-generic-utils/ginutil/response"

	"github.com/gin-gonic/gin"
)

// 常见错误
var (
	// ErrOAuthMissingToken 表示请求中缺少OAuth令牌
	ErrOAuthMissingToken = errors.New("缺少OAuth令牌")

	// ErrOAuthInvalidToken 表示OAuth令牌无效
	ErrOAuthInvalidToken = errors.New("无效的OAuth令牌")

	// ErrOAuthExpiredToken 表示OAuth令牌已过期
	ErrOAuthExpiredToken = errors.New("OAuth令牌已过期")

	// ErrOAuthServerError 表示OAuth服务器错误
	ErrOAuthServerError = errors.New("OAuth服务器错误")

	// ErrOAuthIntrospectionFailed 表示令牌内省失败
	ErrOAuthIntrospectionFailed = errors.New("令牌内省失败")
)

// OAuthConfig 定义OAuth中间件的配置选项
type OAuthConfig struct {
	// TokenIntrospectionEndpoint 是令牌内省端点URL
	TokenIntrospectionEndpoint string

	// ClientID 是用于访问内省端点的客户端ID
	ClientID string

	// ClientSecret 是用于访问内省端点的客户端密钥
	ClientSecret string

	// UserInfoExtractor 是从内省响应中提取用户信息的函数
	UserInfoExtractor interface{}

	// TokenExtractor 是从请求中提取令牌的函数
	TokenExtractor func(c *gin.Context) (string, error)

	// ErrorHandler 是处理错误的函数
	ErrorHandler func(c *gin.Context, err error)

	// HTTPClient 是用于发送HTTP请求的客户端
	HTTPClient *http.Client
}

// OAuthOption 是用于配置OAuth中间件的函数类型
type OAuthOption func(*OAuthConfig)

// WithTokenIntrospectionEndpoint 设置令牌内省端点URL
func WithTokenIntrospectionEndpoint(url string) OAuthOption {
	return func(config *OAuthConfig) {
		config.TokenIntrospectionEndpoint = url
	}
}

// WithClientCredentials 设置客户端凭证
func WithClientCredentials(clientID, clientSecret string) OAuthOption {
	return func(config *OAuthConfig) {
		config.ClientID = clientID
		config.ClientSecret = clientSecret
	}
}

// WithOAuthTokenExtractor 设置令牌提取函数
func WithOAuthTokenExtractor(extractor func(c *gin.Context) (string, error)) OAuthOption {
	return func(config *OAuthConfig) {
		config.TokenExtractor = extractor
	}
}

// WithOAuthErrorHandler 设置错误处理函数
func WithOAuthErrorHandler(handler func(c *gin.Context, err error)) OAuthOption {
	return func(config *OAuthConfig) {
		config.ErrorHandler = handler
	}
}

// WithHTTPClient 设置HTTP客户端
func WithHTTPClient(client *http.Client) OAuthOption {
	return func(config *OAuthConfig) {
		config.HTTPClient = client
	}
}

// 默认OAuth令牌提取器，从Authorization头部提取Bearer令牌
func defaultOAuthTokenExtractor(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", ErrOAuthMissingToken
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrOAuthMissingToken
	}

	return parts[1], nil
}

// 默认OAuth错误处理器
func defaultOAuthErrorHandler(c *gin.Context, err error) {
	var message string
	var code int

	switch {
	case errors.Is(err, ErrOAuthMissingToken):
		message = "缺少OAuth令牌"
		code = ecode.AccessUnauthorized
	case errors.Is(err, ErrOAuthInvalidToken):
		message = "无效的OAuth令牌"
		code = ecode.AccessUnauthorized
	case errors.Is(err, ErrOAuthExpiredToken):
		message = "OAuth令牌已过期"
		code = ecode.AccessUnauthorized
	case errors.Is(err, ErrOAuthServerError):
		message = "OAuth服务器错误"
		code = ecode.ErrorCodeThirdParty
	case errors.Is(err, ErrOAuthIntrospectionFailed):
		message = "令牌验证失败"
		code = ecode.AccessUnauthorized
	default:
		message = "认证失败"
		code = ecode.AccessUnauthorized
	}

	response.Fail(c, code, message)
}

// TokenIntrospectionResponse 表示OAuth 2.0令牌内省响应
type TokenIntrospectionResponse struct {
	Active    bool   `json:"active"`               // 令牌是否有效
	Scope     string `json:"scope,omitempty"`      // 令牌的作用域
	ClientID  string `json:"client_id,omitempty"`  // 客户端ID
	Username  string `json:"username,omitempty"`   // 用户名
	TokenType string `json:"token_type,omitempty"` // 令牌类型
	Exp       int64  `json:"exp,omitempty"`        // 过期时间
	Iat       int64  `json:"iat,omitempty"`        // 颁发时间
	Nbf       int64  `json:"nbf,omitempty"`        // 生效时间
	Sub       string `json:"sub,omitempty"`        // 主题（通常是用户ID）
	Aud       string `json:"aud,omitempty"`        // 受众
	Iss       string `json:"iss,omitempty"`        // 颁发者
	Jti       string `json:"jti,omitempty"`        // JWT ID

	// 其他自定义字段
	Extra map[string]interface{} `json:"-"`
}

// UnmarshalJSON 实现自定义的JSON解析，将未知字段存储到Extra映射中
func (r *TokenIntrospectionResponse) UnmarshalJSON(data []byte) error {
	// 定义一个匿名结构体，包含已知字段
	type Alias TokenIntrospectionResponse

	// 解析已知字段
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 解析所有字段到map
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	// 提取未知字段到Extra
	r.Extra = make(map[string]interface{})
	for k, v := range rawMap {
		// 忽略已知字段
		switch k {
		case "active", "scope", "client_id", "username", "token_type",
			"exp", "iat", "nbf", "sub", "aud", "iss", "jti":
			continue
		default:
			r.Extra[k] = v
		}
	}

	return nil
}

// NewOAuth2Middleware 创建一个新的OAuth 2.0认证中间件
func NewOAuth2Middleware[ID comparable, Role comparable](
	userInfoExtractor func(response *TokenIntrospectionResponse) (*UserIdentity[ID, Role], error),
	options ...OAuthOption,
) gin.HandlerFunc {
	// 创建默认配置
	config := &OAuthConfig{
		TokenExtractor: defaultOAuthTokenExtractor,
		ErrorHandler:   defaultOAuthErrorHandler,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		UserInfoExtractor: userInfoExtractor,
	}

	// 应用选项
	for _, option := range options {
		option(config)
	}

	// 验证必需的配置
	if config.TokenIntrospectionEndpoint == "" {
		panic("auth: OAuth中间件需要TokenIntrospectionEndpoint选项")
	}

	if config.UserInfoExtractor == nil {
		panic("auth: OAuth中间件需要UserInfoExtractor选项")
	}

	// 返回中间件处理函数
	return func(c *gin.Context) {
		// 提取令牌
		token, err := config.TokenExtractor(c)
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		// 调用令牌内省端点
		introspectionResp, err := introspectToken(config, token)
		if err != nil {
			config.ErrorHandler(c, fmt.Errorf("%w: %v", ErrOAuthServerError, err))
			return
		}

		// 验证令牌是否有效
		if !introspectionResp.Active {
			config.ErrorHandler(c, ErrOAuthInvalidToken)
			return
		}

		// 验证令牌是否过期
		if introspectionResp.Exp > 0 && time.Now().Unix() > introspectionResp.Exp {
			config.ErrorHandler(c, ErrOAuthExpiredToken)
			return
		}

		// 提取用户信息
		if typedExtractor, ok := config.UserInfoExtractor.(func(*TokenIntrospectionResponse) (*UserIdentity[ID, Role], error)); ok {
			identity, err := typedExtractor(introspectionResp)
			if err != nil {
				config.ErrorHandler(c, fmt.Errorf("%w: %v", ErrOAuthIntrospectionFailed, err))
				return
			}

			// 将用户身份信息存储到上下文中
			if identity != nil {
				SetIdentityToContext(c, identity)
			}

			// 继续处理请求
			c.Next()
		} else {
			// 提取器类型不匹配
			panic("auth: OAuth UserInfoExtractor类型不匹配")
		}
	}
}

// introspectToken 调用令牌内省端点验证令牌
func introspectToken(config *OAuthConfig, token string) (*TokenIntrospectionResponse, error) {
	// 准备请求体
	data := fmt.Sprintf("token=%s", token)
	req, err := http.NewRequest("POST", config.TokenIntrospectionEndpoint, bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 如果提供了客户端凭证，则设置Basic认证
	if config.ClientID != "" && config.ClientSecret != "" {
		req.SetBasicAuth(config.ClientID, config.ClientSecret)
	}

	// 发送请求
	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("令牌内省请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var introspectionResp TokenIntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&introspectionResp); err != nil {
		return nil, err
	}

	return &introspectionResp, nil
}

// RequireOAuth 创建一个OAuth 2.0认证中间件的简化版本
func RequireOAuth[ID comparable, Role comparable](
	introspectionEndpoint string,
	clientID, clientSecret string,
	userInfoExtractor func(response *TokenIntrospectionResponse) (*UserIdentity[ID, Role], error),
) gin.HandlerFunc {
	return NewOAuth2Middleware(
		userInfoExtractor,
		WithTokenIntrospectionEndpoint(introspectionEndpoint),
		WithClientCredentials(clientID, clientSecret),
	)
}

// SimpleOAuthUserInfoExtractor 创建一个简单的用户信息提取器，使用预定义的字段映射
func SimpleOAuthUserInfoExtractor[ID comparable, Role comparable](
	userIDField string,
	usernameField string,
	rolesField string,
	tenantIDField string,
	roleConverter func(interface{}) []Role,
	idConverter func(interface{}) ID,
) func(response *TokenIntrospectionResponse) (*UserIdentity[ID, Role], error) {
	return func(response *TokenIntrospectionResponse) (*UserIdentity[ID, Role], error) {
		// 创建用户身份信息
		identity := &UserIdentity[ID, Role]{
			ExtraData: make(map[string]any),
		}

		// 提取用户ID
		if userIDField == "sub" && response.Sub != "" {
			if idConverter != nil {
				identity.UserID = idConverter(response.Sub)
			}
		} else if val, ok := response.Extra[userIDField]; ok && val != nil {
			if idConverter != nil {
				identity.UserID = idConverter(val)
			}
		}

		// 提取用户名
		if usernameField == "username" && response.Username != "" {
			identity.Username = response.Username
		} else if val, ok := response.Extra[usernameField]; ok && val != nil {
			if strVal, ok := val.(string); ok {
				identity.Username = strVal
			}
		}

		// 提取角色
		if rolesField != "" {
			if val, ok := response.Extra[rolesField]; ok && val != nil && roleConverter != nil {
				identity.Roles = roleConverter(val)
			}
		}

		// 提取租户ID
		if tenantIDField != "" {
			if val, ok := response.Extra[tenantIDField]; ok && val != nil {
				if strVal, ok := val.(string); ok {
					identity.TenantID = strVal
				}
			}
		}

		// 将其他字段复制到ExtraData
		for k, v := range response.Extra {
			// 忽略已处理的字段
			if k == userIDField || k == usernameField || k == rolesField || k == tenantIDField {
				continue
			}
			identity.ExtraData[k] = v
		}

		return identity, nil
	}
}
