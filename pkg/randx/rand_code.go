package randx

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

// 类型定义
type Type int

// 字符集类型常量
const (
	TypeDigit Type = 1 << iota
	TypeLowerCase
	TypeUpperCase
	TypeSpecial
	TypeMixed        = TypeDigit | TypeLowerCase | TypeUpperCase | TypeSpecial
	TypeAlphanumeric = TypeDigit | TypeLowerCase | TypeUpperCase
)

// 错误定义
var (
	ErrTypeNotSupported   = errors.New("ggu: 不支持的类型")
	ErrLengthLessThanZero = errors.New("ggu: 长度小于0")
)

// 默认字符集
const (
	digitChar   = "0123456789"
	lowerChar   = "abcdefghijklmnopqrstuvwxyz"
	upperChar   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChar = "!@#$%^&*()_+-=[]{}|;:,.<>?/"
)

var globalRandSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandCode 生成指定长度和类型的随机字符串
func RandCode(length int, typ Type) (string, error) {
	if length < 0 {
		return "", ErrLengthLessThanZero
	}

	if length == 0 {
		return "", nil
	}

	if typ <= 0 || typ > TypeMixed {
		return "", ErrTypeNotSupported
	}

	var charset strings.Builder

	if typ&TypeDigit != 0 {
		charset.WriteString(digitChar)
	}

	if typ&TypeLowerCase != 0 {
		charset.WriteString(lowerChar)
	}

	if typ&TypeUpperCase != 0 {
		charset.WriteString(upperChar)
	}

	if typ&TypeSpecial != 0 {
		charset.WriteString(specialChar)
	}

	return RandStrByCharset(length, charset.String())
}

// RandStrByCharset 根据给定的字符集生成随机字符串
func RandStrByCharset(length int, charset string) (string, error) {
	if length < 0 {
		return "", ErrLengthLessThanZero
	}

	if length == 0 {
		return "", nil
	}

	if charset == "" {
		return "", ErrInvalidCharset
	}

	charsetRunes := []rune(charset)
	charsetLen := len(charsetRunes)

	result := make([]rune, length)
	for i := 0; i < length; i++ {
		result[i] = charsetRunes[globalRandSource.Intn(charsetLen)]
	}

	return string(result), nil
}
