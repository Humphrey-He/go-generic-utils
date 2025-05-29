package security

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSPDirective 是内容安全策略指令的类型。
type CSPDirective string

// 内容安全策略指令常量。
const (
	CSPDefaultSrc              CSPDirective = "default-src"
	CSPScriptSrc               CSPDirective = "script-src"
	CSPStyleSrc                CSPDirective = "style-src"
	CSPImgSrc                  CSPDirective = "img-src"
	CSPConnectSrc              CSPDirective = "connect-src"
	CSPFontSrc                 CSPDirective = "font-src"
	CSPObjectSrc               CSPDirective = "object-src"
	CSPMediaSrc                CSPDirective = "media-src"
	CSPFrameSrc                CSPDirective = "frame-src"
	CSPChildSrc                CSPDirective = "child-src"
	CSPFrameAncestors          CSPDirective = "frame-ancestors"
	CSPFormAction              CSPDirective = "form-action"
	CSPBaseURI                 CSPDirective = "base-uri"
	CSPManifestSrc             CSPDirective = "manifest-src"
	CSPWorkerSrc               CSPDirective = "worker-src"
	CSPReportURI               CSPDirective = "report-uri"
	CSPReportTo                CSPDirective = "report-to"
	CSPSandbox                 CSPDirective = "sandbox"
	CSPUpgradeInsecureRequests CSPDirective = "upgrade-insecure-requests"
	CSPBlockAllMixedContent    CSPDirective = "block-all-mixed-content"
	CSPRequireSriFor           CSPDirective = "require-sri-for"
	CSPTrustedTypes            CSPDirective = "trusted-types"
)

// CSPSourceValue 是内容安全策略源值的类型。
type CSPSourceValue string

// 内容安全策略源值常量。
const (
	CSPSelf           CSPSourceValue = "'self'"
	CSPNone           CSPSourceValue = "'none'"
	CSPUnsafeInline   CSPSourceValue = "'unsafe-inline'"
	CSPUnsafeEval     CSPSourceValue = "'unsafe-eval'"
	CSPStrictDynamic  CSPSourceValue = "'strict-dynamic'"
	CSPUnsafeHashes   CSPSourceValue = "'unsafe-hashes'"
	CSPReportSample   CSPSourceValue = "'report-sample'"
	CSPWasmUnsafeEval CSPSourceValue = "'wasm-unsafe-eval'"
	CSPAll            CSPSourceValue = "*"
	CSPHTTP           CSPSourceValue = "http:"
	CSPHTTPS          CSPSourceValue = "https:"
	CSPData           CSPSourceValue = "data:"
	CSPMediaStream    CSPSourceValue = "mediastream:"
	CSPBlob           CSPSourceValue = "blob:"
	CSPFilesystem     CSPSourceValue = "filesystem:"
)

// CSPDirectiveValues 表示 CSP 指令及其值的映射。
type CSPDirectiveValues map[CSPDirective][]CSPSourceValue

// CSPOptions 是内容安全策略选项。
type CSPOptions struct {
	// Directives 是 CSP 指令及其值的映射。
	Directives CSPDirectiveValues

	// ReportOnly 设置是否使用 Content-Security-Policy-Report-Only 头部。
	ReportOnly bool

	// ReportURI 是 CSP 违规报告的 URI。
	ReportURI string
}

// SecurityHeadersOptions 是安全头部选项。
type SecurityHeadersOptions struct {
	// CSP 是内容安全策略选项。
	CSP *CSPOptions

	// XFrameOptions 设置 X-Frame-Options 头部。
	// 可能的值：DENY, SAMEORIGIN, ALLOW-FROM uri
	XFrameOptions string

	// XContentTypeOptions 设置 X-Content-Type-Options 头部。
	// 可能的值：nosniff
	XContentTypeOptions string

	// XXSSProtection 设置 X-XSS-Protection 头部。
	// 可能的值：0, 1, 1; mode=block, 1; report=uri
	XXSSProtection string

	// ReferrerPolicy 设置 Referrer-Policy 头部。
	ReferrerPolicy string

	// PermissionsPolicy 设置 Permissions-Policy 头部。
	PermissionsPolicy string

	// StrictTransportSecurity 设置 Strict-Transport-Security 头部。
	// 例如："max-age=31536000; includeSubDomains; preload"
	StrictTransportSecurity string
}

// DefaultCSPOptions 返回默认的内容安全策略选项。
func DefaultCSPOptions() *CSPOptions {
	return &CSPOptions{
		Directives: CSPDirectiveValues{
			CSPDefaultSrc: {CSPSelf},
			CSPScriptSrc:  {CSPSelf},
			CSPStyleSrc:   {CSPSelf},
			CSPImgSrc:     {CSPSelf, CSPData},
			CSPFontSrc:    {CSPSelf, CSPData},
			CSPConnectSrc: {CSPSelf},
			CSPObjectSrc:  {CSPNone},
			CSPFrameSrc:   {CSPNone},
		},
		ReportOnly: false,
	}
}

// DefaultSecurityHeadersOptions 返回默认的安全头部选项。
func DefaultSecurityHeadersOptions() *SecurityHeadersOptions {
	return &SecurityHeadersOptions{
		CSP:                     DefaultCSPOptions(),
		XFrameOptions:           "SAMEORIGIN",
		XContentTypeOptions:     "nosniff",
		XXSSProtection:          "1; mode=block",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
	}
}

// GenerateCSPHeader 生成内容安全策略头部值。
func GenerateCSPHeader(options *CSPOptions) string {
	if options == nil || len(options.Directives) == 0 {
		return ""
	}

	var directives []string

	// 添加每个指令
	for directive, values := range options.Directives {
		if len(values) == 0 {
			continue
		}

		var sources []string
		for _, value := range values {
			sources = append(sources, string(value))
		}

		directives = append(directives, string(directive)+" "+strings.Join(sources, " "))
	}

	// 添加 report-uri 指令
	if options.ReportURI != "" {
		directives = append(directives, string(CSPReportURI)+" "+options.ReportURI)
	}

	return strings.Join(directives, "; ")
}

// SecurityHeaders 返回一个 Gin 中间件，用于设置安全相关的 HTTP 头部。
func SecurityHeaders(options ...*SecurityHeadersOptions) gin.HandlerFunc {
	// 使用默认配置
	opts := DefaultSecurityHeadersOptions()

	// 应用自定义配置
	if len(options) > 0 && options[0] != nil {
		opts = options[0]
	}

	return func(c *gin.Context) {
		// 设置内容安全策略头部
		if opts.CSP != nil {
			cspHeader := GenerateCSPHeader(opts.CSP)
			if cspHeader != "" {
				if opts.CSP.ReportOnly {
					c.Header("Content-Security-Policy-Report-Only", cspHeader)
				} else {
					c.Header("Content-Security-Policy", cspHeader)
				}
			}
		}

		// 设置其他安全头部
		if opts.XFrameOptions != "" {
			c.Header("X-Frame-Options", opts.XFrameOptions)
		}
		if opts.XContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", opts.XContentTypeOptions)
		}
		if opts.XXSSProtection != "" {
			c.Header("X-XSS-Protection", opts.XXSSProtection)
		}
		if opts.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", opts.ReferrerPolicy)
		}
		if opts.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", opts.PermissionsPolicy)
		}
		if opts.StrictTransportSecurity != "" {
			c.Header("Strict-Transport-Security", opts.StrictTransportSecurity)
		}

		c.Next()
	}
}

// NoCache 返回一个 Gin 中间件，用于设置禁止缓存的 HTTP 头部。
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// CSPBuilder 用于构建内容安全策略。
type CSPBuilder struct {
	options *CSPOptions
}

// NewCSPBuilder 创建一个新的内容安全策略构建器。
func NewCSPBuilder() *CSPBuilder {
	return &CSPBuilder{
		options: DefaultCSPOptions(),
	}
}

// Add 向指定指令添加源值。
func (b *CSPBuilder) Add(directive CSPDirective, values ...CSPSourceValue) *CSPBuilder {
	if b.options.Directives == nil {
		b.options.Directives = make(CSPDirectiveValues)
	}
	b.options.Directives[directive] = append(b.options.Directives[directive], values...)
	return b
}

// Set 设置指定指令的源值，覆盖现有值。
func (b *CSPBuilder) Set(directive CSPDirective, values ...CSPSourceValue) *CSPBuilder {
	if b.options.Directives == nil {
		b.options.Directives = make(CSPDirectiveValues)
	}
	b.options.Directives[directive] = values
	return b
}

// ReportOnly 设置是否使用 Content-Security-Policy-Report-Only 头部。
func (b *CSPBuilder) ReportOnly(reportOnly bool) *CSPBuilder {
	b.options.ReportOnly = reportOnly
	return b
}

// ReportURI 设置 CSP 违规报告的 URI。
func (b *CSPBuilder) ReportURI(uri string) *CSPBuilder {
	b.options.ReportURI = uri
	return b
}

// Build 构建内容安全策略选项。
func (b *CSPBuilder) Build() *CSPOptions {
	return b.options
}

// BuildMiddleware 构建内容安全策略中间件。
func (b *CSPBuilder) BuildMiddleware() gin.HandlerFunc {
	return SecurityHeaders(&SecurityHeadersOptions{
		CSP: b.options,
	})
}

// 辅助函数

// EscapeHTML 转义 HTML 特殊字符。
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// UnescapeHTML 反转义 HTML 特殊字符。
func UnescapeHTML(s string) string {
	return html.UnescapeString(s)
}

// EscapeJS 转义 JavaScript 字符串中的特殊字符。
func EscapeJS(s string) string {
	// 将字符串转换为 JSON 字符串，这会转义 JavaScript 特殊字符
	b, err := json.Marshal(s)
	if err != nil {
		return s
	}
	// 去掉引号
	return string(b[1 : len(b)-1])
}

// SafeResponse 返回一个安全的响应，添加安全头部。
func SafeResponse(c *gin.Context, code int, obj interface{}) {
	// 设置基本的安全头部
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("X-Frame-Options", "DENY")
	c.JSON(code, obj)
}

// CSPNonce 生成一个 CSP nonce 并将其添加到上下文中。
func CSPNonce(c *gin.Context) string {
	nonce := generateCSRFToken(16) // 重用 CSRF 令牌生成函数
	c.Set("csp_nonce", nonce)
	return nonce
}

// GetCSPNonce 从 Gin 上下文中获取 CSP nonce。
func GetCSPNonce(c *gin.Context) string {
	if v, exists := c.Get("csp_nonce"); exists {
		if nonce, ok := v.(string); ok {
			return nonce
		}
	}
	return ""
}

// WithCSPNonce 返回一个 Gin 中间件，用于在每个请求中生成 CSP nonce，
// 并为指定的指令添加 nonce 源值。
func WithCSPNonce(directives ...CSPDirective) gin.HandlerFunc {
	// 如果没有指定指令，则默认为 script-src 和 style-src
	if len(directives) == 0 {
		directives = []CSPDirective{CSPScriptSrc, CSPStyleSrc}
	}

	return func(c *gin.Context) {
		// 生成 nonce
		nonce := CSPNonce(c)
		nonceValue := CSPSourceValue("'nonce-" + nonce + "'")

		// 获取或创建 CSP 选项
		var cspOptions *CSPOptions
		if v, exists := c.Get("csp_options"); exists {
			if opts, ok := v.(*CSPOptions); ok {
				cspOptions = opts
			}
		}
		if cspOptions == nil {
			cspOptions = DefaultCSPOptions()
			c.Set("csp_options", cspOptions)
		}

		// 为指定的指令添加 nonce 源值
		for _, directive := range directives {
			if cspOptions.Directives == nil {
				cspOptions.Directives = make(CSPDirectiveValues)
			}
			cspOptions.Directives[directive] = append(cspOptions.Directives[directive], nonceValue)
		}

		// 设置 CSP 头部
		cspHeader := GenerateCSPHeader(cspOptions)
		if cspHeader != "" {
			if cspOptions.ReportOnly {
				c.Header("Content-Security-Policy-Report-Only", cspHeader)
			} else {
				c.Header("Content-Security-Policy", cspHeader)
			}
		}

		c.Next()
	}
}

// CSPMiddleware 是 SecurityHeaders 的别名，用于向后兼容。
func CSPMiddleware(options ...*CSPOptions) gin.HandlerFunc {
	var opts *SecurityHeadersOptions
	if len(options) > 0 && options[0] != nil {
		opts = &SecurityHeadersOptions{
			CSP: options[0],
		}
	}
	return SecurityHeaders(opts)
}

// SetSafeResponseHeaders 设置基本的安全响应头部。
func SetSafeResponseHeaders(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("X-Frame-Options", "DENY")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
}

// WithSecureHeaders 返回一个 Gin 中间件，用于设置基本的安全响应头部。
func WithSecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetSafeResponseHeaders(c)
		c.Next()
	}
}

// RenderScriptTag 生成一个带有 nonce 的 script 标签。
func RenderScriptTag(c *gin.Context, js string) string {
	nonce := GetCSPNonce(c)
	if nonce != "" {
		return `<script nonce="` + nonce + `">` + js + `</script>`
	}
	return `<script>` + js + `</script>`
}

// RenderStyleTag 生成一个带有 nonce 的 style 标签。
func RenderStyleTag(c *gin.Context, css string) string {
	nonce := GetCSPNonce(c)
	if nonce != "" {
		return `<style nonce="` + nonce + `">` + css + `</style>`
	}
	return `<style>` + css + `</style>`
}

// RenderExternalScriptTag 生成一个带有 nonce 的外部 script 标签。
func RenderExternalScriptTag(c *gin.Context, src string) string {
	nonce := GetCSPNonce(c)
	if nonce != "" {
		return `<script src="` + src + `" nonce="` + nonce + `"></script>`
	}
	return `<script src="` + src + `"></script>`
}

// RenderExternalStyleTag 生成一个带有 nonce 的外部 style 标签。
func RenderExternalStyleTag(c *gin.Context, href string) string {
	nonce := GetCSPNonce(c)
	if nonce != "" {
		return `<link rel="stylesheet" href="` + href + `" nonce="` + nonce + `">`
	}
	return `<link rel="stylesheet" href="` + href + `">`
}

// SetSafeCookie 设置一个安全的 Cookie。
func SetSafeCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", true, true)
}

// SetSecureFlag 设置响应头的安全标志。
func SetSecureFlag(c *gin.Context) {
	c.Writer.Header().Set("X-Secured", "1")
}

// IsSecureRequest 检查请求是否安全（使用 HTTPS）。
func IsSecureRequest(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	return c.GetHeader("X-Forwarded-Proto") == "https"
}
