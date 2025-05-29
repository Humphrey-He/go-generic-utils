// Package ratelimit 提供了一个用于 Gin 框架的请求限流中间件。
package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RedisStore 是基于 Redis 的限流状态存储。
// 使用 Redis 的 INCR 和 EXPIRE 命令实现简单的计数器限流。
// 对于更高级的令牌桶或滑动窗口算法，通常需要使用 Lua 脚本确保原子性。
type RedisStore struct {
	client         *redis.Client
	keyPrefix      string
	defaultTTL     time.Duration
	useTokenBucket bool
	script         *redis.Script
}

// RedisStoreOption 是配置 RedisStore 的函数选项。
type RedisStoreOption func(*RedisStore)

// WithRedisKeyPrefix 设置 Redis 键前缀。
func WithRedisKeyPrefix(prefix string) RedisStoreOption {
	return func(s *RedisStore) {
		s.keyPrefix = prefix
	}
}

// WithRedisDefaultTTL 设置 Redis 键的默认过期时间。
func WithRedisDefaultTTL(ttl time.Duration) RedisStoreOption {
	return func(s *RedisStore) {
		s.defaultTTL = ttl
	}
}

// WithRedisTokenBucket 设置是否使用令牌桶算法。
func WithRedisTokenBucket(use bool) RedisStoreOption {
	return func(s *RedisStore) {
		s.useTokenBucket = use
	}
}

// NewRedisStore 创建一个新的 Redis 存储。
func NewRedisStore(client *redis.Client, options ...RedisStoreOption) *RedisStore {
	store := &RedisStore{
		client:         client,
		keyPrefix:      "ratelimit:",
		defaultTTL:     time.Hour,
		useTokenBucket: false,
	}

	for _, option := range options {
		option(store)
	}

	// 如果使用令牌桶算法，初始化 Lua 脚本
	if store.useTokenBucket {
		store.script = redis.NewScript(`
			local key = KEYS[1]
			local limit = tonumber(ARGV[1])
			local burst = tonumber(ARGV[2])
			local period = tonumber(ARGV[3])
			local cost = tonumber(ARGV[4])
			local now = tonumber(ARGV[5])
			
			-- 获取当前令牌桶状态
			local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
			local tokens = tonumber(bucket[1] or burst)
			local last_time = tonumber(bucket[2] or 0)
			
			-- 计算从上次请求到现在新增的令牌数
			local elapsed = math.max(0, now - last_time)
			local new_tokens = math.min(burst, tokens + (elapsed / period) * limit)
			
			-- 检查是否有足够的令牌
			local allowed = new_tokens >= cost
			local new_tokens_after_request = new_tokens
			local retry_after = 0
			
			if allowed then
				new_tokens_after_request = new_tokens - cost
			else
				-- 计算需要等待的时间
				retry_after = math.ceil((cost - new_tokens) * period / limit)
			end
			
			-- 更新令牌桶状态
			redis.call('HMSET', key, 'tokens', new_tokens_after_request, 'last_time', now)
			redis.call('EXPIRE', key, period * 2)
			
			return {allowed and 1 or 0, retry_after}
		`)
	}

	return store
}

// AllowN 检查是否允许 n 个事件通过。
func (s *RedisStore) AllowN(key string, limit rate.Limit, burst int, n int) (bool, time.Duration) {
	ctx := context.Background()
	redisKey := s.keyPrefix + key

	// 如果使用令牌桶算法
	if s.useTokenBucket {
		return s.allowNTokenBucket(ctx, redisKey, limit, burst, n)
	}

	// 否则使用简单的计数器限流
	return s.allowNCounter(ctx, redisKey, limit, burst, n)
}

// allowNTokenBucket 使用令牌桶算法检查是否允许 n 个事件通过。
func (s *RedisStore) allowNTokenBucket(ctx context.Context, key string, limit rate.Limit, burst int, n int) (bool, time.Duration) {
	// 计算限流周期（秒）
	period := int64(1 / float64(limit))
	if period < 1 {
		period = 1
	}

	// 执行 Lua 脚本
	result, err := s.script.Eval(ctx, s.client, []string{key}, []interface{}{
		int64(limit),
		burst,
		period,
		n,
		time.Now().Unix(),
	}).Result()

	if err != nil {
		// 如果发生错误，默认允许请求
		return true, 0
	}

	// 解析结果
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 2 {
		// 如果结果格式不正确，默认允许请求
		return true, 0
	}

	allowed, _ := resultSlice[0].(int64)
	retryAfter, _ := resultSlice[1].(int64)

	return allowed == 1, time.Duration(retryAfter) * time.Second
}

// allowNCounter 使用简单的计数器限流检查是否允许 n 个事件通过。
func (s *RedisStore) allowNCounter(ctx context.Context, key string, limit rate.Limit, burst int, n int) (bool, time.Duration) {
	// 计算限流周期（秒）
	period := int64(1 / float64(limit))
	if period < 1 {
		period = 1
	}

	// 计算当前时间窗口的键
	windowKey := fmt.Sprintf("%s:%d", key, time.Now().Unix()/period)

	// 获取当前计数
	count, err := s.client.Get(ctx, windowKey).Int64()
	if err != nil && err != redis.Nil {
		// 如果发生错误（不是键不存在的错误），默认允许请求
		return true, 0
	}

	// 检查是否超过突发限制
	if count >= int64(burst) {
		// 计算需要等待的时间
		ttl, err := s.client.TTL(ctx, windowKey).Result()
		if err != nil || ttl < 0 {
			ttl = time.Duration(period) * time.Second
		}
		return false, ttl
	}

	// 增加计数
	_, err = s.client.IncrBy(ctx, windowKey, int64(n)).Result()
	if err != nil {
		// 如果发生错误，默认允许请求
		return true, 0
	}

	// 设置过期时间
	s.client.Expire(ctx, windowKey, time.Duration(period)*time.Second)

	return true, 0
}

// Close 关闭 Redis 存储并释放资源。
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// SlidingWindowRedisStore 是基于 Redis 的滑动窗口限流存储。
type SlidingWindowRedisStore struct {
	client    *redis.Client
	keyPrefix string
	script    *redis.Script
}

// NewSlidingWindowRedisStore 创建一个新的滑动窗口 Redis 存储。
func NewSlidingWindowRedisStore(client *redis.Client, keyPrefix string) *SlidingWindowRedisStore {
	store := &SlidingWindowRedisStore{
		client:    client,
		keyPrefix: keyPrefix,
	}

	// 初始化滑动窗口 Lua 脚本
	store.script = redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local current = tonumber(ARGV[3])
		
		-- 移除过期的请求
		redis.call('ZREMRANGEBYSCORE', key, 0, current - window)
		
		-- 获取当前窗口内的请求数
		local count = redis.call('ZCARD', key)
		
		-- 检查是否超过限制
		if count >= limit then
			-- 获取最早的请求时间
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			if #oldest > 0 then
				return {0, oldest[2] + window - current}
			else
				return {0, window}
			end
		end
		
		-- 添加当前请求
		redis.call('ZADD', key, current, current .. ':' .. math.random())
		redis.call('EXPIRE', key, window)
		
		return {1, 0}
	`)

	return store
}

// AllowN 检查是否允许 n 个事件通过。
func (s *SlidingWindowRedisStore) AllowN(key string, limit rate.Limit, burst int, n int) (bool, time.Duration) {
	ctx := context.Background()
	redisKey := s.keyPrefix + key

	// 计算窗口大小（秒）
	window := int64(1 / float64(limit))
	if window < 1 {
		window = 1
	}

	// 执行 Lua 脚本
	result, err := s.script.Eval(ctx, s.client, []string{redisKey}, []interface{}{
		burst,
		window,
		time.Now().Unix(),
	}).Result()

	if err != nil {
		// 如果发生错误，默认允许请求
		return true, 0
	}

	// 解析结果
	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) != 2 {
		// 如果结果格式不正确，默认允许请求
		return true, 0
	}

	allowed, _ := resultSlice[0].(int64)
	retryAfter, _ := resultSlice[1].(int64)

	return allowed == 1, time.Duration(retryAfter) * time.Second
}

// Close 关闭 Redis 存储并释放资源。
func (s *SlidingWindowRedisStore) Close() error {
	return s.client.Close()
}

// FixedWindowRedisStore 是基于 Redis 的固定窗口限流存储。
type FixedWindowRedisStore struct {
	client    *redis.Client
	keyPrefix string
	window    time.Duration
}

// NewFixedWindowRedisStore 创建一个新的固定窗口 Redis 存储。
func NewFixedWindowRedisStore(client *redis.Client, keyPrefix string, window time.Duration) *FixedWindowRedisStore {
	return &FixedWindowRedisStore{
		client:    client,
		keyPrefix: keyPrefix,
		window:    window,
	}
}

// AllowN 检查是否允许 n 个事件通过。
func (s *FixedWindowRedisStore) AllowN(key string, limit rate.Limit, burst int, n int) (bool, time.Duration) {
	ctx := context.Background()

	// 计算当前时间窗口
	now := time.Now()
	windowStart := now.Truncate(s.window)
	windowKey := s.keyPrefix + key + ":" + strconv.FormatInt(windowStart.Unix(), 10)

	// 获取当前计数
	count, err := s.client.Get(ctx, windowKey).Int64()
	if err != nil && err != redis.Nil {
		// 如果发生错误（不是键不存在的错误），默认允许请求
		return true, 0
	}

	// 检查是否超过限制
	if count >= int64(burst) {
		// 计算需要等待的时间
		nextWindow := windowStart.Add(s.window)
		retryAfter := nextWindow.Sub(now)
		return false, retryAfter
	}

	// 增加计数
	_, err = s.client.IncrBy(ctx, windowKey, int64(n)).Result()
	if err != nil {
		// 如果发生错误，默认允许请求
		return true, 0
	}

	// 设置过期时间
	s.client.Expire(ctx, windowKey, s.window)

	return true, 0
}

// Close 关闭 Redis 存储并释放资源。
func (s *FixedWindowRedisStore) Close() error {
	return s.client.Close()
}
