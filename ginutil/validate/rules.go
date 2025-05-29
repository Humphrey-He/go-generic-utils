// Package validate 提供基于 go-playground/validator 的增强校验功能。
package validate

import (
	"database/sql"
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// 注册内置的自定义校验规则
func registerBuiltinRules(v *ValidatorV10) {
	// 注册所有内置的自定义校验规则
	_ = v.RegisterValidation("no-special-chars", noSpecialChars)
	_ = v.RegisterValidation("is-safe-html", isSafeHTML)
	_ = v.RegisterValidation("is-valid-country-code", isValidCountryCode)
	_ = v.RegisterValidation("is-valid-phone-for-region", isValidPhoneForRegion)
}

// RegisterRule 注册一个自定义校验规则。
// 参数：
//   - tag: 校验标签，如 "no-special-chars"
//   - fn: 校验函数
//   - callValidationEvenIfNull: 是否在字段为 nil 时也调用校验函数
//
// 返回值：
//   - error: 注册错误
func RegisterRule(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	return defaultValidator.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

// RegisterRuleWithValidator 使用指定的校验器注册一个自定义校验规则。
func RegisterRuleWithValidator(v *ValidatorV10, tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	return v.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

// noSpecialChars 校验字段是否不包含特殊字符。
// 特殊字符包括：!@#$%^&*()_+{}|:"<>?[]\;',./~`
func noSpecialChars(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 定义特殊字符正则表达式
	specialChars := regexp.MustCompile(`[!@#$%^&*()_+{}\|:"<>?\[\]\\;',./~` + "`" + `]`)
	return !specialChars.MatchString(value)
}

// isSafeHTML 校验 HTML 内容是否安全。
// 使用 html.EscapeString 进行简单的 HTML 转义检查。
// 如果转义后的内容与原始内容不同，则认为包含不安全的 HTML。
func isSafeHTML(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 使用 html.EscapeString 进行 HTML 转义
	escaped := html.EscapeString(value)

	// 如果转义后的内容与原始内容相同，则认为是安全的
	// 这种方法比较简单，只能检测基本的 HTML 标签
	return escaped == value
}

// isValidCountryCode 校验国家代码是否有效。
// 目前支持 ISO 3166-1 alpha-2 国家代码（两位字母）。
func isValidCountryCode(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 检查长度是否为 2
	if len(value) != 2 {
		return false
	}

	// 检查是否全部为大写字母
	for _, r := range value {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return false
		}
	}

	// 这里可以添加对有效国家代码的检查
	// 为简化实现，我们只检查格式，不检查具体的国家代码是否存在
	return true
}

// isValidPhoneForRegion 校验手机号是否符合特定地区的格式。
// 参数格式：is-valid-phone-for-region=CN
func isValidPhoneForRegion(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// 获取地区代码参数
	regionCode := fl.Param()
	if regionCode == "" {
		return false
	}

	// 根据不同地区应用不同的校验规则
	switch strings.ToUpper(regionCode) {
	case "CN": // 中国大陆手机号
		// 中国大陆手机号格式：1[3-9]\d{9}
		match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, value)
		return match
	case "HK": // 香港手机号
		// 香港手机号格式：[5-9]\d{7}
		match, _ := regexp.MatchString(`^[5-9]\d{7}$`, value)
		return match
	case "TW": // 台湾手机号
		// 台湾手机号格式：09\d{8}
		match, _ := regexp.MatchString(`^09\d{8}$`, value)
		return match
	case "US": // 美国手机号
		// 美国手机号格式：\d{3}-\d{3}-\d{4} 或 \(\d{3}\) \d{3}-\d{4}
		match, _ := regexp.MatchString(`^(\d{3}-\d{3}-\d{4}|\(\d{3}\) \d{3}-\d{4})$`, value)
		return match
	default:
		// 对于其他地区，暂时返回 true
		return true
	}
}

// IsUniqueInDB 创建一个校验函数，检查字段值在数据库中是否唯一。
// 参数：
//   - db: 数据库连接
//   - table: 表名
//   - column: 列名
//   - excludeIDColumn: 排除的 ID 列名（可选）
//
// 返回值：
//   - validator.Func: 校验函数
//
// 使用示例：
//
//	_ = v.RegisterValidation("is-unique-in-db", validate.IsUniqueInDB(db, "users", "email", "id"))
//
//	type User struct {
//	    ID    int    `json:"id"`
//	    Email string `json:"email" validate:"required,email,is-unique-in-db"`
//	}
func IsUniqueInDB(db *sql.DB, table, column, excludeIDColumn string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}

		var query string
		var args []interface{}

		// 如果提供了排除 ID 列，则构建排除当前记录的查询
		if excludeIDColumn != "" {
			// 尝试从结构体中获取 ID 值
			idField := fl.Parent().FieldByName(strings.Title(excludeIDColumn))
			if idField.IsValid() {
				id := idField.Interface()
				query = fmt.Sprintf("SELECT 1 FROM %s WHERE %s = ? AND %s != ? LIMIT 1", table, column, excludeIDColumn)
				args = []interface{}{value, id}
			} else {
				// 如果找不到 ID 字段，则不排除任何记录
				query = fmt.Sprintf("SELECT 1 FROM %s WHERE %s = ? LIMIT 1", table, column)
				args = []interface{}{value}
			}
		} else {
			// 简单地检查值是否存在
			query = fmt.Sprintf("SELECT 1 FROM %s WHERE %s = ? LIMIT 1", table, column)
			args = []interface{}{value}
		}

		// 执行查询
		var exists int
		err := db.QueryRow(query, args...).Scan(&exists)
		if err == sql.ErrNoRows {
			// 没有找到记录，值是唯一的
			return true
		}

		// 如果有错误或找到了记录，则值不是唯一的
		return err != nil
	}
}

// IsUniqueInSlice 创建一个校验函数，检查字段值在切片中是否唯一。
// 这个函数用于校验结构体中的切片字段，确保切片中的所有元素都是唯一的。
//
// 使用示例：
//
//	_ = v.RegisterValidation("unique-slice", validate.IsUniqueInSlice())
//
//	type Form struct {
//	    Tags []string `json:"tags" validate:"unique-slice"`
//	}
func IsUniqueInSlice() validator.Func {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()

		// 检查字段是否是切片
		if field.Kind() != reflect.Slice {
			return true
		}

		// 使用 map 检查唯一性
		seen := make(map[interface{}]bool)
		for i := 0; i < field.Len(); i++ {
			val := field.Index(i).Interface()
			if seen[val] {
				return false
			}
			seen[val] = true
		}

		return true
	}
}
