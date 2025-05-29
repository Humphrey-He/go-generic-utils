package ecode

// 定义标准错误码
// 错误码规则：
// - 0: 成功
// - 4xxxx: A类错误，用户错误（参数错误、权限错误等）
// - 5xxxx: B类错误，系统内部错误
// - 7xxxx: C类错误，第三方服务错误

const (
	// OK 表示成功
	OK = 0
	// SuccessMessage 默认成功消息
	SuccessMessage = "操作成功"
	// DefaultSuccessMessage 默认成功消息（与SuccessMessage相同，为了命名一致性）
	DefaultSuccessMessage = SuccessMessage
)

// A类错误：用户错误（4xxxx）
const (
	// ErrorCodeBadRequest 请求参数错误
	ErrorCodeBadRequest = 40001
	// AccessUnauthorized 未授权
	AccessUnauthorized = 40100
	// AccessPermissionDenied 无权限
	AccessPermissionDenied = 40300
	// ErrorCodeNotFound 资源不存在
	ErrorCodeNotFound = 40400
	// ErrorCodeUserInputInvalid 用户输入无效
	ErrorCodeUserInputInvalid = 40002
	// ErrorCodeUserExists 用户已存在
	ErrorCodeUserExists = 40003
	// ErrorCodeResourceExists 资源已存在
	ErrorCodeResourceExists = 40004
	// ErrorCodeTooManyRequests 请求过多
	ErrorCodeTooManyRequests = 40029
)

// B类错误：系统内部错误（5xxxx）
const (
	// ErrorCodeInternal 系统内部错误
	ErrorCodeInternal = 50000
	// ErrorCodeDatabase 数据库错误
	ErrorCodeDatabase = 50001
	// ErrorCodeCache 缓存错误
	ErrorCodeCache = 50002
	// ErrorCodeConfig 配置错误
	ErrorCodeConfig = 50003
)

// C类错误：第三方服务错误（7xxxx）
const (
	// ErrorCodeThirdParty 第三方服务错误
	ErrorCodeThirdParty = 70000
	// ErrorCodeThirdPartyTimeout 第三方服务超时
	ErrorCodeThirdPartyTimeout = 70001
	// ErrorCodeThirdPartyUnavailable 第三方服务不可用
	ErrorCodeThirdPartyUnavailable = 70002
)

// 错误码对应的默认消息
var codeMessages = map[int]string{
	OK:                             SuccessMessage,
	ErrorCodeBadRequest:            "请求参数错误",
	AccessUnauthorized:             "未授权，请先登录",
	AccessPermissionDenied:         "无权限访问该资源",
	ErrorCodeNotFound:              "请求的资源不存在",
	ErrorCodeUserInputInvalid:      "用户输入无效",
	ErrorCodeUserExists:            "用户已存在",
	ErrorCodeResourceExists:        "资源已存在",
	ErrorCodeTooManyRequests:       "请求过于频繁，请稍后再试",
	ErrorCodeInternal:              "系统内部错误",
	ErrorCodeDatabase:              "数据库操作失败",
	ErrorCodeCache:                 "缓存操作失败",
	ErrorCodeConfig:                "系统配置错误",
	ErrorCodeThirdParty:            "第三方服务调用失败",
	ErrorCodeThirdPartyTimeout:     "第三方服务调用超时",
	ErrorCodeThirdPartyUnavailable: "第三方服务不可用",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// RegisterMessage 注册自定义错误码消息
func RegisterMessage(code int, message string) {
	codeMessages[code] = message
}
