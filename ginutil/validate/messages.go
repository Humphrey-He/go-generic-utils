// Package validate 提供基于 go-playground/validator 的增强校验功能。
package validate

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// MessageStore 用于存储和检索校验规则对应的错误消息模板。
type MessageStore struct {
	// 消息模板映射，格式为：locale.tag.field -> message
	// 例如：zh.required.Username -> "用户名是必填的"
	// 或者：zh.required -> "此字段是必填的"
	templates map[string]string
}

// NewMessageStore 创建一个新的消息存储。
func NewMessageStore() *MessageStore {
	return &MessageStore{
		templates: make(map[string]string),
	}
}

// NewDefaultMessageStore 创建一个带有默认中文错误消息的消息存储。
func NewDefaultMessageStore() *MessageStore {
	ms := NewMessageStore()

	// 注册默认的中文错误消息
	ms.RegisterMessage("zh", "required", "此字段是必填的")
	ms.RegisterMessage("zh", "email", "请输入有效的电子邮箱地址")
	ms.RegisterMessage("zh", "min", "此字段的值必须大于或等于 %s")
	ms.RegisterMessage("zh", "max", "此字段的值必须小于或等于 %s")
	ms.RegisterMessage("zh", "len", "此字段的长度必须等于 %s")
	ms.RegisterMessage("zh", "eq", "此字段的值必须等于 %s")
	ms.RegisterMessage("zh", "ne", "此字段的值不能等于 %s")
	ms.RegisterMessage("zh", "lt", "此字段的值必须小于 %s")
	ms.RegisterMessage("zh", "lte", "此字段的值必须小于或等于 %s")
	ms.RegisterMessage("zh", "gt", "此字段的值必须大于 %s")
	ms.RegisterMessage("zh", "gte", "此字段的值必须大于或等于 %s")
	ms.RegisterMessage("zh", "alpha", "此字段只能包含字母")
	ms.RegisterMessage("zh", "alphanum", "此字段只能包含字母和数字")
	ms.RegisterMessage("zh", "numeric", "此字段只能包含数字")
	ms.RegisterMessage("zh", "number", "此字段必须是有效的数字")
	ms.RegisterMessage("zh", "hexadecimal", "此字段必须是有效的十六进制数")
	ms.RegisterMessage("zh", "hexcolor", "此字段必须是有效的十六进制颜色")
	ms.RegisterMessage("zh", "rgb", "此字段必须是有效的 RGB 颜色")
	ms.RegisterMessage("zh", "rgba", "此字段必须是有效的 RGBA 颜色")
	ms.RegisterMessage("zh", "hsl", "此字段必须是有效的 HSL 颜色")
	ms.RegisterMessage("zh", "hsla", "此字段必须是有效的 HSLA 颜色")
	ms.RegisterMessage("zh", "url", "此字段必须是有效的 URL")
	ms.RegisterMessage("zh", "uri", "此字段必须是有效的 URI")
	ms.RegisterMessage("zh", "uuid", "此字段必须是有效的 UUID")
	ms.RegisterMessage("zh", "uuid3", "此字段必须是有效的 UUID v3")
	ms.RegisterMessage("zh", "uuid4", "此字段必须是有效的 UUID v4")
	ms.RegisterMessage("zh", "uuid5", "此字段必须是有效的 UUID v5")
	ms.RegisterMessage("zh", "datetime", "此字段必须是有效的日期时间")
	ms.RegisterMessage("zh", "date", "此字段必须是有效的日期")
	ms.RegisterMessage("zh", "time", "此字段必须是有效的时间")
	ms.RegisterMessage("zh", "timezone", "此字段必须是有效的时区")
	ms.RegisterMessage("zh", "json", "此字段必须是有效的 JSON")
	ms.RegisterMessage("zh", "ip", "此字段必须是有效的 IP 地址")
	ms.RegisterMessage("zh", "ipv4", "此字段必须是有效的 IPv4 地址")
	ms.RegisterMessage("zh", "ipv6", "此字段必须是有效的 IPv6 地址")
	ms.RegisterMessage("zh", "mac", "此字段必须是有效的 MAC 地址")
	ms.RegisterMessage("zh", "contains", "此字段必须包含文本 '%s'")
	ms.RegisterMessage("zh", "containsany", "此字段必须包含以下字符中的至少一个: '%s'")
	ms.RegisterMessage("zh", "containsrune", "此字段必须包含字符 '%s'")
	ms.RegisterMessage("zh", "excludes", "此字段不能包含文本 '%s'")
	ms.RegisterMessage("zh", "excludesall", "此字段不能包含以下任何字符: '%s'")
	ms.RegisterMessage("zh", "excludesrune", "此字段不能包含字符 '%s'")
	ms.RegisterMessage("zh", "startswith", "此字段必须以 '%s' 开头")
	ms.RegisterMessage("zh", "endswith", "此字段必须以 '%s' 结尾")
	ms.RegisterMessage("zh", "unique", "此字段的所有值必须是唯一的")
	ms.RegisterMessage("zh", "oneof", "此字段必须是以下值之一: %s")
	ms.RegisterMessage("zh", "file", "此字段必须是有效的文件路径")
	ms.RegisterMessage("zh", "image", "此字段必须是有效的图片")
	ms.RegisterMessage("zh", "isdefault", "此字段必须是默认值")

	// 注册自定义校验规则的错误消息
	ms.RegisterMessage("zh", "no-special-chars", "此字段不能包含特殊字符")
	ms.RegisterMessage("zh", "is-safe-html", "此字段包含不安全的 HTML 内容")
	ms.RegisterMessage("zh", "is-valid-country-code", "此字段必须是有效的国家代码")
	ms.RegisterMessage("zh", "is-valid-phone-for-region", "此字段必须是有效的手机号码格式")
	ms.RegisterMessage("zh", "is-unique-in-db", "此字段的值已存在")

	return ms
}

// RegisterMessage 注册一个错误消息模板。
// 参数：
//   - locale: 语言环境，如 "zh", "en"
//   - tag: 校验标签，如 "required", "email"
//   - message: 错误消息模板
//   - field: 可选的字段名，如果提供，则只对该字段应用此消息
func (ms *MessageStore) RegisterMessage(locale, tag, message string, field ...string) {
	if ms.templates == nil {
		ms.templates = make(map[string]string)
	}

	// 如果提供了字段名，则注册字段特定的消息
	if len(field) > 0 && field[0] != "" {
		key := fmt.Sprintf("%s.%s.%s", locale, tag, field[0])
		ms.templates[key] = message
		return
	}

	// 否则注册通用消息
	key := fmt.Sprintf("%s.%s", locale, tag)
	ms.templates[key] = message
}

// GetMessage 获取校验错误的本地化消息。
// 参数：
//   - fieldError: 校验错误
//   - locale: 语言环境，如 "zh", "en"
//
// 返回值：
//   - 本地化的错误消息，如果没有找到对应的消息，则返回空字符串
func (ms *MessageStore) GetMessage(fieldError validator.FieldError, locale string) string {
	if ms.templates == nil {
		return ""
	}

	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	// 首先尝试获取字段特定的消息
	key := fmt.Sprintf("%s.%s.%s", locale, tag, field)
	if message, ok := ms.templates[key]; ok {
		// 如果有参数，则格式化消息
		if param != "" {
			return fmt.Sprintf(message, param)
		}
		return message
	}

	// 然后尝试获取通用消息
	key = fmt.Sprintf("%s.%s", locale, tag)
	if message, ok := ms.templates[key]; ok {
		// 如果有参数，则格式化消息
		if param != "" {
			return fmt.Sprintf(message, param)
		}
		return message
	}

	// 如果找不到对应的消息，则返回空字符串
	return ""
}

// RegisterTranslations 注册翻译函数到 validator.Validate 实例。
// 这是一个高级功能，通常不需要直接调用。
func (ms *MessageStore) RegisterTranslations(v *validator.Validate, locale string) error {
	// 这里可以实现更复杂的翻译注册逻辑
	// 目前我们使用简单的 GetMessage 方法，所以这个函数暂时不做任何事情
	return nil
}

// FormatFieldName 格式化字段名，使其更易读。
// 例如，将 "UserName" 转换为 "用户名"。
func FormatFieldName(field string, locale string) string {
	// 这里可以实现字段名的本地化
	// 目前简单地返回原始字段名
	return field
}

// FormatError 格式化错误消息，替换占位符。
// 参数：
//   - message: 错误消息模板
//   - fieldError: 校验错误
//
// 返回值：
//   - 格式化后的错误消息
func FormatError(message string, fieldError validator.FieldError) string {
	// 替换 {field} 占位符
	message = strings.ReplaceAll(message, "{field}", fieldError.Field())

	// 替换 {param} 占位符
	message = strings.ReplaceAll(message, "{param}", fieldError.Param())

	// 替换 {value} 占位符
	value := fmt.Sprintf("%v", fieldError.Value())
	message = strings.ReplaceAll(message, "{value}", value)

	return message
}
