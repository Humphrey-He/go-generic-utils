package security

import (
	"bytes"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

// SanitizePolicy 定义了输入清理的策略。
type SanitizePolicy int

const (
	// PolicyNone 不进行任何清理。
	PolicyNone SanitizePolicy = iota

	// PolicyStrict 严格模式，移除所有 HTML 标签和危险字符。
	PolicyStrict

	// PolicyBasic 基本模式，允许一些基本的格式化标签，如 <b>, <i>, <u>。
	PolicyBasic

	// PolicyRelaxed 宽松模式，允许更多的标签和属性，适用于富文本编辑器。
	PolicyRelaxed
)

// disallowedTagsStrict 是严格模式下禁止的 HTML 标签的正则表达式。
var disallowedTagsStrict = regexp.MustCompile(`<[^>]*>`)

// disallowedTagsBasic 是基本模式下禁止的 HTML 标签的正则表达式。
// 允许 <b>, <i>, <u>, <strong>, <em>, <br>, <p>, <div>, <span> 标签。
var disallowedTagsBasic = regexp.MustCompile(`<(?!/?(?:b|i|u|strong|em|br|p|div|span)(?:\s|>))[^>]*>`)

// dangerousAttributes 是危险属性的正则表达式。
var dangerousAttributes = regexp.MustCompile(`(?i)(on\w+|style|class|id)=["'][^"']*["']`)

// scriptTags 是脚本标签的正则表达式。
var scriptTags = regexp.MustCompile(`(?i)<script\b[^>]*>.*?</script>`)

// styleTags 是样式标签的正则表达式。
var styleTags = regexp.MustCompile(`(?i)<style\b[^>]*>.*?</style>`)

// iframeTags 是 iframe 标签的正则表达式。
var iframeTags = regexp.MustCompile(`(?i)<iframe\b[^>]*>.*?</iframe>`)

// objectTags 是 object 标签的正则表达式。
var objectTags = regexp.MustCompile(`(?i)<object\b[^>]*>.*?</object>`)

// embedTags 是 embed 标签的正则表达式。
var embedTags = regexp.MustCompile(`(?i)<embed\b[^>]*>.*?</embed>`)

// jsProtocol 是 JavaScript 协议的正则表达式。
var jsProtocol = regexp.MustCompile(`(?i)javascript:`)

// controlChars 是控制字符的正则表达式。
var controlChars = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

// multipleSpaces 是多个空格的正则表达式。
var multipleSpaces = regexp.MustCompile(`\s+`)

// SanitizeHTML 根据指定的策略清理 HTML 字符串。
func SanitizeHTML(html string, policy SanitizePolicy) string {
	if html == "" {
		return ""
	}

	switch policy {
	case PolicyNone:
		return html
	case PolicyStrict:
		// 移除所有 HTML 标签
		html = disallowedTagsStrict.ReplaceAllString(html, "")
	case PolicyBasic:
		// 移除脚本、样式和其他危险标签
		html = scriptTags.ReplaceAllString(html, "")
		html = styleTags.ReplaceAllString(html, "")
		html = iframeTags.ReplaceAllString(html, "")
		html = objectTags.ReplaceAllString(html, "")
		html = embedTags.ReplaceAllString(html, "")
		// 移除危险属性
		html = dangerousAttributes.ReplaceAllString(html, "")
		// 移除 JavaScript 协议
		html = jsProtocol.ReplaceAllString(html, "")
		// 移除不允许的标签
		html = disallowedTagsBasic.ReplaceAllString(html, "")
	case PolicyRelaxed:
		// 移除脚本和其他危险标签
		html = scriptTags.ReplaceAllString(html, "")
		html = iframeTags.ReplaceAllString(html, "")
		html = objectTags.ReplaceAllString(html, "")
		html = embedTags.ReplaceAllString(html, "")
		// 移除 JavaScript 协议
		html = jsProtocol.ReplaceAllString(html, "")
	}

	// 移除控制字符
	html = controlChars.ReplaceAllString(html, "")

	return html
}

// SanitizeText 清理纯文本字符串，移除控制字符并规范化空白。
func SanitizeText(text string) string {
	if text == "" {
		return ""
	}

	// 移除控制字符
	text = controlChars.ReplaceAllString(text, "")

	// 规范化空白
	text = multipleSpaces.ReplaceAllString(strings.TrimSpace(text), " ")

	return text
}

// TruncateText 截断文本到指定长度，可选择添加省略号。
func TruncateText(text string, maxLength int, addEllipsis bool) string {
	if len(text) <= maxLength {
		return text
	}

	truncated := text[:maxLength]
	if addEllipsis {
		return truncated + "..."
	}

	return truncated
}

// RemoveControlChars 移除字符串中的控制字符。
func RemoveControlChars(s string) string {
	return controlChars.ReplaceAllString(s, "")
}

// NormalizeWhitespace 规范化字符串中的空白。
func NormalizeWhitespace(s string) string {
	return multipleSpaces.ReplaceAllString(strings.TrimSpace(s), " ")
}

// RemoveHTML 完全移除 HTML 标签。
func RemoveHTML(html string) string {
	return disallowedTagsStrict.ReplaceAllString(html, "")
}

// StripTags 是 RemoveHTML 的别名。
func StripTags(html string) string {
	return RemoveHTML(html)
}

// SanitizeStringField 清理结构体中的字符串字段。
// 由于使用了反射，不建议在性能敏感的代码中使用。
func SanitizeStringField(field reflect.Value, policy SanitizePolicy) {
	if field.Kind() == reflect.String {
		sanitized := SanitizeHTML(field.String(), policy)
		if field.CanSet() {
			field.SetString(sanitized)
		}
	}
}

// SanitizeAllStringFields 清理结构体中的所有字符串字段。
// 由于使用了反射，不建议在性能敏感的代码中使用。
func SanitizeAllStringFields(obj interface{}, policy SanitizePolicy) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String {
			SanitizeStringField(field, policy)
		} else if field.Kind() == reflect.Struct {
			SanitizeAllStringFields(field.Addr().Interface(), policy)
		} else if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			SanitizeAllStringFields(field.Interface(), policy)
		}
	}
}

// SanitizeStruct 是 SanitizeAllStringFields 的别名。
func SanitizeStruct(obj interface{}, policy SanitizePolicy) {
	SanitizeAllStringFields(obj, policy)
}

// IsValidEmail 检查字符串是否是有效的电子邮件地址。
func IsValidEmail(email string) bool {
	// 简单的电子邮件验证
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// IsValidURL 检查字符串是否是有效的 URL。
func IsValidURL(url string) bool {
	// 简单的 URL 验证
	pattern := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
	match, _ := regexp.MatchString(pattern, url)
	return match
}

// IsValidPhoneNumber 检查字符串是否是有效的电话号码。
func IsValidPhoneNumber(phone string) bool {
	// 简单的电话号码验证
	pattern := `^[0-9+\-() ]{8,20}$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}

// SanitizeJSON 清理 JSON 字符串中的可能导致 XSS 的内容。
func SanitizeJSON(json string) string {
	// 替换尖括号，防止 HTML 注入
	json = strings.ReplaceAll(json, "<", "\\u003c")
	json = strings.ReplaceAll(json, ">", "\\u003e")

	// 移除控制字符
	json = controlChars.ReplaceAllString(json, "")

	return json
}

// AllowedCharacters 检查字符串是否只包含允许的字符。
func AllowedCharacters(s string, allowed *regexp.Regexp) bool {
	for _, r := range s {
		if !allowed.MatchString(string(r)) {
			return false
		}
	}
	return true
}

// OnlyAlphanumeric 检查字符串是否只包含字母和数字。
func OnlyAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// OnlyASCII 检查字符串是否只包含 ASCII 字符。
func OnlyASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// SanitizeFilename 清理文件名，移除不安全字符。
func SanitizeFilename(filename string) string {
	// 移除路径分隔符和其他危险字符
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	// 移除控制字符
	filename = controlChars.ReplaceAllString(filename, "")

	// 移除前导和尾随空格
	filename = strings.TrimSpace(filename)

	// 确保文件名不为空
	if filename == "" {
		filename = "file"
	}

	return filename
}

// SanitizeQueryString 清理查询字符串，移除危险字符。
func SanitizeQueryString(query string) string {
	// 移除 <> 字符，防止 HTML 注入
	query = strings.ReplaceAll(query, "<", "")
	query = strings.ReplaceAll(query, ">", "")

	// 移除控制字符
	query = controlChars.ReplaceAllString(query, "")

	return query
}

// SanitizeSearchTerm 清理搜索词，适用于数据库搜索。
func SanitizeSearchTerm(term string) string {
	// 移除 SQL 注入字符
	term = strings.ReplaceAll(term, "'", "")
	term = strings.ReplaceAll(term, "\"", "")
	term = strings.ReplaceAll(term, ";", "")
	term = strings.ReplaceAll(term, "--", "")
	term = strings.ReplaceAll(term, "/*", "")
	term = strings.ReplaceAll(term, "*/", "")

	// 移除控制字符
	term = controlChars.ReplaceAllString(term, "")

	// 规范化空白
	term = multipleSpaces.ReplaceAllString(strings.TrimSpace(term), " ")

	return term
}

// RemoveEmoji 移除字符串中的表情符号。
func RemoveEmoji(s string) string {
	var buffer bytes.Buffer
	for _, r := range s {
		if r < 0x10000 { // 表情符号通常在 Unicode 私有区域之外
			buffer.WriteRune(r)
		}
	}
	return buffer.String()
}

// SafeHTML 根据策略安全地清理 HTML，返回清理后的 HTML 和是否被修改。
func SafeHTML(html string, policy SanitizePolicy) (string, bool) {
	sanitized := SanitizeHTML(html, policy)
	return sanitized, sanitized != html
}

// ValidateAndSanitize 验证输入并进行清理。
// 如果验证失败，返回错误消息。
func ValidateAndSanitize(input string, maxLength int, policy SanitizePolicy) (string, string) {
	if input == "" {
		return "", "输入不能为空"
	}

	if len(input) > maxLength {
		return "", "输入超过最大长度 " + string(maxLength)
	}

	sanitized := SanitizeHTML(input, policy)
	if sanitized == "" {
		return "", "清理后的输入为空"
	}

	return sanitized, ""
}

// SanitizeMap 清理 map 中的所有字符串值。
func SanitizeMap(m map[string]interface{}, policy SanitizePolicy) {
	for k, v := range m {
		if str, ok := v.(string); ok {
			m[k] = SanitizeHTML(str, policy)
		} else if subMap, ok := v.(map[string]interface{}); ok {
			SanitizeMap(subMap, policy)
		}
	}
}

// SanitizeSlice 清理切片中的所有字符串值。
func SanitizeSlice(slice []interface{}, policy SanitizePolicy) {
	for i, v := range slice {
		if str, ok := v.(string); ok {
			slice[i] = SanitizeHTML(str, policy)
		} else if subMap, ok := v.(map[string]interface{}); ok {
			SanitizeMap(subMap, policy)
		} else if subSlice, ok := v.([]interface{}); ok {
			SanitizeSlice(subSlice, policy)
		}
	}
}

// HTMLEntityEncode 将 HTML 特殊字符编码为实体。
func HTMLEntityEncode(s string) string {
	return EscapeHTML(s)
}

// HTMLAttributeEncode 将 HTML 属性中的特殊字符编码。
func HTMLAttributeEncode(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// URLEncode 对 URL 进行编码，移除不安全的字符。
func URLEncode(url string) string {
	// 移除控制字符
	url = controlChars.ReplaceAllString(url, "")

	// 移除 JavaScript 协议
	url = jsProtocol.ReplaceAllString(url, "")

	return url
}

// ContentSecurityPolicyEncode 编码 CSP 策略中的特殊字符。
func ContentSecurityPolicyEncode(s string) string {
	s = strings.ReplaceAll(s, ";", "%3B")
	return s
}
