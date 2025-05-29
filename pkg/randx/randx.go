package randx

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"
)

// 错误定义
var (
	// 注意：保留rand_code.go中已有的错误定义，这里只添加新的错误
	ErrInvalidRange          = errors.New("ggu: 无效的范围")
	ErrInvalidCharset        = errors.New("ggu: 无效的字符集")
	ErrInsufficientEntropy   = errors.New("ggu: 系统熵不足")
	ErrInvalidDistribution   = errors.New("ggu: 无效的分布")
	ErrInvalidProbability    = errors.New("ggu: 无效的概率值")
	ErrInvalidParameterValue = errors.New("ggu: 无效的参数值")
)

const (
	// 电商平台常用字符集
	CharsetSKU     = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"   // 不包含容易混淆的I和O
	CharsetCoupon  = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"   // 用于优惠券码生成
	CharsetOrderID = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" // 用于订单ID生成
)

// safeRand 并发安全的随机源
type safeRand struct {
	src  io.Reader
	lock sync.Mutex
}

// 随机源实例，使用互斥锁保证并发安全
var globalRand = &safeRand{src: rand.Reader}

// generateSecureRandom 使用加密安全的随机源生成随机字符串
func generateSecureRandom(length int, charset string) (string, error) {
	charsetLength := len(charset)
	maxNum := big.NewInt(int64(charsetLength))
	result := make([]byte, length)

	globalRand.lock.Lock()
	defer globalRand.lock.Unlock()

	for i := 0; i < length; i++ {
		// 生成随机索引
		randomIndex, err := rand.Int(globalRand.src, maxNum)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrInsufficientEntropy, err)
		}
		result[i] = charset[randomIndex.Int64()]
	}

	return string(result), nil
}

// RandInt 生成指定范围内的随机整数 [min, max)
func RandInt(min, max int) (int, error) {
	if min >= max {
		return 0, ErrInvalidRange
	}

	globalRand.lock.Lock()
	defer globalRand.lock.Unlock()

	// 计算范围
	delta := max - min
	maxBig := big.NewInt(int64(delta))

	// 生成随机数
	n, err := rand.Int(globalRand.src, maxBig)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInsufficientEntropy, err)
	}

	return min + int(n.Int64()), nil
}

// RandFloat64 生成 [0.0, 1.0) 范围内的随机浮点数
func RandFloat64() (float64, error) {
	// 生成 0 到 2^53-1 之间的随机整数
	maxBig := big.NewInt(1)
	maxBig.Lsh(maxBig, 53) // 2^53

	globalRand.lock.Lock()
	defer globalRand.lock.Unlock()

	n, err := rand.Int(globalRand.src, maxBig)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInsufficientEntropy, err)
	}

	// 转换为 [0.0, 1.0) 范围内的浮点数
	return float64(n.Int64()) / float64(maxBig.Int64()), nil
}

// RandFloat64Range 生成 [min, max) 范围内的随机浮点数
func RandFloat64Range(min, max float64) (float64, error) {
	if min >= max {
		return 0, ErrInvalidRange
	}

	r, err := RandFloat64()
	if err != nil {
		return 0, err
	}

	return min + r*(max-min), nil
}

// ---- 电商平台常用随机函数 ----

// RandProductID 生成随机商品ID
// prefix: 商品ID前缀，例如 "P"
// length: 数字部分长度
func RandProductID(prefix string, length int) (string, error) {
	if length <= 0 {
		return "", ErrInvalidParameterValue
	}

	// 使用数字字符集生成随机字符串
	digits, err := generateSecureRandom(length, "0123456789")
	if err != nil {
		return "", err
	}

	return prefix + digits, nil
}

// RandSKU 生成随机SKU编码
// format: {prefix}-{randomPart}-{suffix}
func RandSKU(prefix string, randomLength int, suffix string) (string, error) {
	if randomLength <= 0 {
		return "", ErrInvalidParameterValue
	}

	randomPart, err := generateSecureRandom(randomLength, CharsetSKU)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s", prefix, randomPart, suffix), nil
}

// RandOrderID 生成订单ID，格式为：前缀 + 日期时间 + 随机数字
func RandOrderID(prefix string, randomLength int) (string, error) {
	if randomLength <= 0 {
		return "", ErrInvalidParameterValue
	}

	// 获取当前时间戳，格式：yyMMddHHmmss
	timestamp := time.Now().Format("060102150405")

	// 生成随机数字部分
	randomPart, err := generateSecureRandom(randomLength, "0123456789")
	if err != nil {
		return "", err
	}

	return prefix + timestamp + randomPart, nil
}

// RandCouponCode 生成优惠券码
// 可选格式：
// 1. 纯随机字符（默认）
// 2. 分段式：XXXX-XXXX-XXXX-XXXX
func RandCouponCode(length int, segmented bool) (string, error) {
	if length <= 0 {
		return "", ErrInvalidParameterValue
	}

	if segmented {
		// 确保长度是4的倍数，便于分段
		if length%4 != 0 {
			length = ((length + 3) / 4) * 4
		}

		segmentLength := length / 4
		segments := make([]string, 4)

		for i := 0; i < 4; i++ {
			code, err := generateSecureRandom(segmentLength, CharsetCoupon)
			if err != nil {
				return "", err
			}
			segments[i] = code
		}

		return strings.Join(segments, "-"), nil
	}

	// 不分段的情况
	return generateSecureRandom(length, CharsetCoupon)
}

// RandPrice 生成指定范围内的随机价格（保留两位小数）
func RandPrice(min, max float64) (float64, error) {
	if min >= max {
		return 0, ErrInvalidRange
	}

	// 生成随机浮点数
	price, err := RandFloat64Range(min, max)
	if err != nil {
		return 0, err
	}

	// 保留两位小数
	return math.Floor(price*100) / 100, nil
}

// RandPhone 生成随机中国大陆手机号
func RandPhone() (string, error) {
	// 手机号前缀
	prefixes := []string{"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"150", "151", "152", "153", "155", "156", "157", "158", "159",
		"170", "176", "177", "178",
		"180", "181", "182", "183", "184", "185", "186", "187", "188", "189",
		"199", "198", "166"}

	// 随机选择前缀
	prefixIndex, err := RandInt(0, len(prefixes))
	if err != nil {
		return "", err
	}

	// 生成8位随机数字
	rest, err := generateSecureRandom(8, "0123456789")
	if err != nil {
		return "", err
	}

	return prefixes[prefixIndex] + rest, nil
}

// RandEmail 生成随机电子邮件地址
func RandEmail(domains ...string) (string, error) {
	if len(domains) == 0 {
		domains = []string{"gmail.com", "outlook.com", "yahoo.com", "icloud.com", "163.com", "qq.com"}
	}

	// 随机选择域名
	domainIndex, err := RandInt(0, len(domains))
	if err != nil {
		return "", err
	}

	// 生成用户名（8-12位字母和数字）
	usernameLength, err := RandInt(8, 13)
	if err != nil {
		return "", err
	}

	// 使用字母和数字字符集
	charset := "0123456789" + "abcdefghijklmnopqrstuvwxyz"
	username, err := generateSecureRandom(usernameLength, charset)
	if err != nil {
		return "", err
	}

	return username + "@" + domains[domainIndex], nil
}

// Shuffle 打乱切片中元素的顺序（Fisher-Yates算法）
func Shuffle[T any](slice []T) ([]T, error) {
	result := make([]T, len(slice))
	copy(result, slice)

	for i := len(result) - 1; i > 0; i-- {
		// 生成 [0, i] 范围内的随机索引
		j, err := RandInt(0, i+1)
		if err != nil {
			return nil, err
		}

		// 交换元素
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// WeightedChoice 根据权重随机选择一个元素
// weights 切片的元素之和必须大于0
func WeightedChoice[T any](items []T, weights []float64) (T, error) {
	var zero T

	if len(items) != len(weights) {
		return zero, ErrInvalidParameterValue
	}

	if len(items) == 0 {
		return zero, ErrInvalidParameterValue
	}

	// 计算权重总和
	var sum float64
	for _, w := range weights {
		if w < 0 {
			return zero, ErrInvalidProbability
		}
		sum += w
	}

	if sum <= 0 {
		return zero, ErrInvalidProbability
	}

	// 生成 [0, sum) 范围内的随机值
	r, err := RandFloat64Range(0, sum)
	if err != nil {
		return zero, err
	}

	// 根据随机值选择元素
	var cumSum float64
	for i, w := range weights {
		cumSum += w
		if r < cumSum {
			return items[i], nil
		}
	}

	// 由于浮点数精度问题，可能走到这里，直接返回最后一个元素
	return items[len(items)-1], nil
}

// RandBool 根据给定概率生成随机布尔值
// probability: 生成true的概率，范围[0, 1]
func RandBool(probability float64) (bool, error) {
	if probability < 0 || probability > 1 {
		return false, ErrInvalidProbability
	}

	r, err := RandFloat64()
	if err != nil {
		return false, err
	}

	return r < probability, nil
}

// RandDate 生成指定时间范围内的随机日期时间
func RandDate(start, end time.Time) (time.Time, error) {
	if start.After(end) {
		return time.Time{}, ErrInvalidRange
	}

	delta := end.Unix() - start.Unix()
	if delta <= 0 {
		return start, nil
	}

	// 生成随机秒数偏移
	randDelta, err := RandInt(0, int(delta)+1)
	if err != nil {
		return time.Time{}, err
	}

	return start.Add(time.Duration(randDelta) * time.Second), nil
}

// RandUUID 生成随机UUID（v4版本）
func RandUUID() (string, error) {
	uuid := make([]byte, 16)

	globalRand.lock.Lock()
	defer globalRand.lock.Unlock()

	_, err := rand.Read(uuid)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInsufficientEntropy, err)
	}

	// 设置版本为4（随机生成的UUID）
	uuid[6] = (uuid[6] & 0x0F) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3F) | 0x80 // Variant RFC4122

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16]), nil
}
