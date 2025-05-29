// Package ratelimit 提供了一个用于 Gin 框架的请求限流中间件。
//
// 该中间件可以限制客户端的请求频率，防止恶意请求或过度使用 API，
// 支持多种存储后端（内存、Redis 等）和灵活的限流策略配置。
package ratelimit

import (
	"ggu/ginutil/response"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// 错误码常量
const (
	// TooManyRequests 表示请求过于频繁
	TooManyRequests = http.StatusTooManyRequests
)

// LimiterStore 是限流状态存储的接口。
type LimiterStore interface {
	// AllowN 检查是否允许 n 个事件通过。
	// 返回是否允许以及建议的重试等待时间。
	AllowN(key string, limit rate.Limit, burst int, n int) (bool, time.Duration)

	// Close 关闭存储并释放资源。
	Close() error
}

// KeyFunc 是从 gin.Context 中提取限流键的函数类型。
type KeyFunc func(c *gin.Context) string

// ErrorHandler 是处理限流错误的函数类型。
type ErrorHandler func(c *gin.Context, retryAfter time.Duration)

// SkipperFunc 是判断是否跳过限流的函数类型。
type SkipperFunc func(c *gin.Context) bool

// Config 定义了限流中间件的配置选项。
type Config struct {
	// Store 是限流状态的存储，默认为内存存储。
	Store LimiterStore

	// KeyFunc 是从 gin.Context 中提取限流键的函数，默认使用客户端 IP。
	KeyFunc KeyFunc

	// ErrorHandler 是处理限流错误的函数。
	// 如果不提供，将使用默认的错误处理函数。
	ErrorHandler ErrorHandler

	// Skipper 是判断是否跳过限流的函数。
	// 如果不提供，将对所有请求进行限流。
	Skipper SkipperFunc

	// Limit 是每秒允许的请求数，默认为 10。
	Limit rate.Limit

	// Burst 是允许的突发请求数，默认为 20。
	Burst int

	// TokensPerRequest 是每个请求消耗的令牌数，默认为 1。
	TokensPerRequest int

	// DisableHeaders 表示是否禁用限流相关的响应头，默认为 false。
	DisableHeaders bool
}

// Option 是配置限流中间件的函数选项。
type Option func(*Config)

// WithStore 设置限流状态的存储。
func WithStore(store LimiterStore) Option {
	return func(c *Config) {
		c.Store = store
	}
}

// WithKeyFunc 设置从 gin.Context 中提取限流键的函数。
func WithKeyFunc(keyFunc KeyFunc) Option {
	return func(c *Config) {
		c.KeyFunc = keyFunc
	}
}

// WithErrorHandler 设置处理限流错误的函数。
func WithErrorHandler(handler ErrorHandler) Option {
	return func(c *Config) {
		c.ErrorHandler = handler
	}
}

// WithSkipper 设置判断是否跳过限流的函数。
func WithSkipper(skipper SkipperFunc) Option {
	return func(c *Config) {
		c.Skipper = skipper
	}
}

// WithLimit 设置每秒允许的请求数。
func WithLimit(limit rate.Limit) Option {
	return func(c *Config) {
		c.Limit = limit
	}
}

// WithBurst 设置允许的突发请求数。
func WithBurst(burst int) Option {
	return func(c *Config) {
		c.Burst = burst
	}
}

// WithTokensPerRequest 设置每个请求消耗的令牌数。
func WithTokensPerRequest(tokens int) Option {
	return func(c *Config) {
		c.TokensPerRequest = tokens
	}
}

// WithDisableHeaders 设置是否禁用限流相关的响应头。
func WithDisableHeaders(disable bool) Option {
	return func(c *Config) {
		c.DisableHeaders = disable
	}
}

// DefaultConfig 返回默认的限流中间件配置。
func DefaultConfig() *Config {
	return &Config{
		Store:            NewMemoryStore(time.Minute * 5),
		KeyFunc:          defaultKeyFunc,
		ErrorHandler:     defaultErrorHandler,
		Skipper:          nil,
		Limit:            10,
		Burst:            20,
		TokensPerRequest: 1,
		DisableHeaders:   false,
	}
}

// 默认的从 gin.Context 中提取限流键的函数
func defaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// 默认的处理限流错误的函数
func defaultErrorHandler(c *gin.Context, retryAfter time.Duration) {
	if !c.IsAborted() {
		// 设置 Retry-After 头部
		if retryAfter > 0 {
			c.Header("Retry-After", retryAfter.String())
		}

		// 使用 response 包返回统一的错误响应
		response.Fail(c, TooManyRequests, "请求过于频繁，请稍后重试")
	}
}

// New 使用默认配置创建一个新的限流中间件。
func New() gin.HandlerFunc {
	return NewWithConfig()
}

// NewWithConfig 使用自定义配置创建一个新的限流中间件。
func NewWithConfig(options ...Option) gin.HandlerFunc {
	config := DefaultConfig()
	for _, option := range options {
		option(config)
	}
	return newRateLimitHandler(config)
}

// newRateLimitHandler 创建一个处理限流的 Gin 中间件。
func newRateLimitHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过限流
		if config.Skipper != nil && config.Skipper(c) {
			c.Next()
			return
		}

		// 获取限流键
		key := config.KeyFunc(c)

		// 检查是否允许请求
		allowed, retryAfter := config.Store.AllowN(key, config.Limit, config.Burst, config.TokensPerRequest)

		// 如果不允许请求
		if !allowed {
			// 调用错误处理函数
			config.ErrorHandler(c, retryAfter)
			return
		}

		// 如果不禁用头部，设置 RateLimit 相关头部
		if !config.DisableHeaders {
			c.Header("X-RateLimit-Limit", formatRateLimit(config.Limit))
			c.Header("X-RateLimit-Remaining", formatRateLimitRemaining(config.Limit, config.Burst, config.TokensPerRequest))
			c.Header("X-RateLimit-Reset", formatRateLimitReset(config.Limit))
		}

		// 继续处理请求
		c.Next()
	}
}

// formatRateLimit 格式化 RateLimit 头部的值。
func formatRateLimit(limit rate.Limit) string {
	return formatInt(int(limit))
}

// formatRateLimitRemaining 格式化 RateLimit-Remaining 头部的值。
func formatRateLimitRemaining(limit rate.Limit, burst, tokens int) string {
	// 这里简化处理，实际上应该根据当前令牌桶中的令牌数计算
	remaining := burst - tokens
	if remaining < 0 {
		remaining = 0
	}
	return formatInt(remaining)
}

// formatRateLimitReset 格式化 RateLimit-Reset 头部的值。
func formatRateLimitReset(limit rate.Limit) string {
	// 计算令牌桶重新填满的时间（秒）
	resetTime := time.Now().Add(time.Second * time.Duration(1/float64(limit)))
	return formatInt(int(resetTime.Unix()))
}

// formatInt 将整数转换为字符串。
func formatInt(n int) string {
	return string([]byte(string('0' + rune(n))))
}

// memoryStoreEntry 是内存存储的条目。
type memoryStoreEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// MemoryStore 是基于内存的限流状态存储。
type MemoryStore struct {
	entries  map[string]*memoryStoreEntry
	mu       sync.Mutex
	janitor  *time.Ticker
	stopChan chan struct{}
	ttl      time.Duration
}

// NewMemoryStore 创建一个新的内存存储。
func NewMemoryStore(ttl time.Duration) *MemoryStore {
	store := &MemoryStore{
		entries:  make(map[string]*memoryStoreEntry),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}

	// 启动清理 goroutine
	store.janitor = time.NewTicker(ttl)
	go store.cleanupLoop()

	return store
}

// AllowN 检查是否允许 n 个事件通过。
func (s *MemoryStore) AllowN(key string, limit rate.Limit, burst int, n int) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取或创建限流器
	entry, exists := s.entries[key]
	if !exists || entry.limiter.Limit() != limit || entry.limiter.Burst() != burst {
		entry = &memoryStoreEntry{
			limiter:  rate.NewLimiter(limit, burst),
			lastSeen: time.Now(),
		}
		s.entries[key] = entry
	} else {
		entry.lastSeen = time.Now()
	}

	// 检查是否允许请求
	now := time.Now()
	reservation := entry.limiter.ReserveN(now, n)
	if !reservation.OK() {
		// 不允许请求，计算需要等待的时间
		retryAfter := entry.limiter.Reserve().Delay()
		return false, retryAfter
	}

	// 允许请求，但可能需要等待
	delay := reservation.DelayFrom(now)
	if delay > 0 {
		// 需要等待，取消预约并返回需要等待的时间
		reservation.Cancel()
		return false, delay
	}

	// 允许请求，不需要等待
	return true, 0
}

// Close 关闭内存存储并释放资源。
func (s *MemoryStore) Close() error {
	s.janitor.Stop()
	close(s.stopChan)
	return nil
}

// cleanupLoop 定期清理过期的限流器。
func (s *MemoryStore) cleanupLoop() {
	for {
		select {
		case <-s.janitor.C:
			s.cleanup()
		case <-s.stopChan:
			return
		}
	}
}

// cleanup 清理过期的限流器。
func (s *MemoryStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.entries {
		if now.Sub(entry.lastSeen) > s.ttl {
			delete(s.entries, key)
		}
	}
}

// IPKeyFunc 返回一个使用客户端 IP 作为限流键的函数。
func IPKeyFunc() KeyFunc {
	return func(c *gin.Context) string {
		return c.ClientIP()
	}
}

// PathKeyFunc 返回一个使用请求路径作为限流键的函数。
func PathKeyFunc() KeyFunc {
	return func(c *gin.Context) string {
		return c.Request.URL.Path
	}
}

// MethodKeyFunc 返回一个使用请求方法作为限流键的函数。
func MethodKeyFunc() KeyFunc {
	return func(c *gin.Context) string {
		return c.Request.Method
	}
}

// CombinedKeyFunc 返回一个组合多个键函数的函数。
func CombinedKeyFunc(keyFuncs ...KeyFunc) KeyFunc {
	return func(c *gin.Context) string {
		var key string
		for _, kf := range keyFuncs {
			key += kf(c) + ":"
		}
		return key
	}
}

// HeaderKeyFunc 返回一个使用请求头作为限流键的函数。
func HeaderKeyFunc(header string) KeyFunc {
	return func(c *gin.Context) string {
		return c.GetHeader(header)
	}
}

// QueryKeyFunc 返回一个使用查询参数作为限流键的函数。
func QueryKeyFunc(param string) KeyFunc {
	return func(c *gin.Context) string {
		return c.Query(param)
	}
}

// ParamKeyFunc 返回一个使用路径参数作为限流键的函数。
func ParamKeyFunc(param string) KeyFunc {
	return func(c *gin.Context) string {
		return c.Param(param)
	}
}

// UserKeyFunc 返回一个使用用户标识作为限流键的函数。
// 用户标识应该存储在 gin.Context 中，键为 "user_id"。
func UserKeyFunc() KeyFunc {
	return func(c *gin.Context) string {
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok {
				return id
			}
		}
		return c.ClientIP()
	}
}

// WhitelistSkipper 返回一个跳过白名单 IP 的函数。
func WhitelistSkipper(ips ...string) SkipperFunc {
	whitelist := make(map[string]struct{}, len(ips))
	for _, ip := range ips {
		whitelist[ip] = struct{}{}
	}

	return func(c *gin.Context) bool {
		clientIP := c.ClientIP()
		_, exists := whitelist[clientIP]
		return exists
	}
}

// MethodSkipper 返回一个跳过指定 HTTP 方法的函数。
func MethodSkipper(methods ...string) SkipperFunc {
	skipMethods := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		skipMethods[method] = struct{}{}
	}

	return func(c *gin.Context) bool {
		_, exists := skipMethods[c.Request.Method]
		return exists
	}
}

// PathSkipper 返回一个跳过指定路径的函数。
func PathSkipper(paths ...string) SkipperFunc {
	skipPaths := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		skipPaths[path] = struct{}{}
	}

	return func(c *gin.Context) bool {
		_, exists := skipPaths[c.Request.URL.Path]
		return exists
	}
}

// CombinedSkipper 返回一个组合多个跳过函数的函数。
// 如果任何一个跳过函数返回 true，则跳过限流。
func CombinedSkipper(skippers ...SkipperFunc) SkipperFunc {
	return func(c *gin.Context) bool {
		for _, skipper := range skippers {
			if skipper(c) {
				return true
			}
		}
		return false
	}
}

// RateLimit 使用指定的限流参数创建一个限流中间件。
func RateLimit(limit rate.Limit, burst int) gin.HandlerFunc {
	return NewWithConfig(WithLimit(limit), WithBurst(burst))
}

// RateLimitPerIP 使用指定的限流参数创建一个基于 IP 的限流中间件。
func RateLimitPerIP(limit rate.Limit, burst int) gin.HandlerFunc {
	return NewWithConfig(WithLimit(limit), WithBurst(burst), WithKeyFunc(IPKeyFunc()))
}

// RateLimitPerPath 使用指定的限流参数创建一个基于路径的限流中间件。
func RateLimitPerPath(limit rate.Limit, burst int) gin.HandlerFunc {
	return NewWithConfig(WithLimit(limit), WithBurst(burst), WithKeyFunc(PathKeyFunc()))
}

// RateLimitPerUser 使用指定的限流参数创建一个基于用户的限流中间件。
func RateLimitPerUser(limit rate.Limit, burst int) gin.HandlerFunc {
	return NewWithConfig(WithLimit(limit), WithBurst(burst), WithKeyFunc(UserKeyFunc()))
}
