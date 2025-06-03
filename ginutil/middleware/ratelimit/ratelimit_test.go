package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Humphrey-He/go-generic-utils/ginutil/middleware/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 测试内存存储的基本功能
func TestMemoryStore(t *testing.T) {
	store := ratelimit.NewMemoryStore(time.Minute)
	defer store.Close()

	// 测试允许请求
	allowed, retryAfter := store.AllowN("test-key", 10, 1, 1)
	assert.True(t, allowed)
	assert.Equal(t, time.Duration(0), retryAfter)

	// 测试突发请求
	for i := 0; i < 10; i++ {
		allowed, _ := store.AllowN("test-key", 1, 10, 1)
		assert.True(t, allowed)
	}

	// 测试超过限制
	allowed, retryAfter = store.AllowN("test-key", 1, 10, 1)
	assert.False(t, allowed)
	assert.True(t, retryAfter > 0)
}

// 测试内存存储的并发安全性
func TestMemoryStoreConcurrency(t *testing.T) {
	store := ratelimit.NewMemoryStore(time.Minute)
	defer store.Close()

	var wg sync.WaitGroup
	// 启动 10 个 goroutine 同时访问
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 每个 goroutine 发起 10 个请求
			for j := 0; j < 10; j++ {
				store.AllowN("concurrent-test", 50, 50, 1)
			}
		}()
	}
	wg.Wait()

	// 验证总共允许了 50 个请求
	allowed, _ := store.AllowN("concurrent-test", 50, 50, 1)
	assert.False(t, allowed)
}

// 测试清理过期的限流器
func TestMemoryStoreCleanup(t *testing.T) {
	// 创建一个短 TTL 的存储
	store := ratelimit.NewMemoryStore(10 * time.Millisecond)
	defer store.Close()

	// 添加一个限流器
	store.AllowN("cleanup-test", 1, 1, 1)

	// 等待 TTL 过期
	time.Sleep(20 * time.Millisecond)

	// 再次请求，应该被允许
	allowed, _ := store.AllowN("cleanup-test", 1, 1, 1)
	assert.True(t, allowed)
}

// 测试基本的限流中间件功能
func TestRateLimitMiddleware(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用限流中间件，设置每秒允许 5 个请求，突发 5 个
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(5),
		ratelimit.WithBurst(5),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 发送 5 个请求，应该都被允许
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 发送第 6 个请求，应该被限流
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "请求过于频繁")
}

// 测试基于 IP 的限流
func TestRateLimitPerIP(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用基于 IP 的限流中间件
	r.Use(ratelimit.RateLimitPerIP(2, 2))

	// 添加一个简单的 API 端点
	r.GET("/api/ip", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 发送 2 个来自同一 IP 的请求，应该都被允许
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/ip", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 发送第 3 个来自同一 IP 的请求，应该被限流
	req := httptest.NewRequest("GET", "/api/ip", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 发送来自不同 IP 的请求，应该被允许
	req = httptest.NewRequest("GET", "/api/ip", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// 测试基于路径的限流
func TestRateLimitPerPath(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用基于路径的限流中间件
	r.Use(ratelimit.RateLimitPerPath(2, 2))

	// 添加两个 API 端点
	r.GET("/api/path1", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/api/path2", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 发送 2 个请求到 path1，应该都被允许
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/path1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 发送第 3 个请求到 path1，应该被限流
	req := httptest.NewRequest("GET", "/api/path1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// 发送请求到 path2，应该被允许
	req = httptest.NewRequest("GET", "/api/path2", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// 测试自定义错误处理函数
func TestCustomErrorHandler(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用自定义错误处理函数的限流中间件
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(1),
		ratelimit.WithBurst(1),
		ratelimit.WithErrorHandler(func(c *gin.Context, retryAfter time.Duration) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":      "自定义限流错误",
				"retryAfter": retryAfter.String(),
			})
		}),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/custom-error", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 发送一个请求，应该被允许
	req := httptest.NewRequest("GET", "/api/custom-error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 发送第二个请求，应该被限流并使用自定义错误处理
	req = httptest.NewRequest("GET", "/api/custom-error", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "自定义限流错误")
}

// 测试跳过函数
func TestSkipperFunc(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用带跳过函数的限流中间件
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(1),
		ratelimit.WithBurst(1),
		ratelimit.WithSkipper(ratelimit.WhitelistSkipper("192.168.1.100")),
	))

	// 添加一个简单的 API 端点
	r.GET("/api/skip", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 发送来自白名单 IP 的多个请求，应该都被允许
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/skip", nil)
		req.RemoteAddr = "192.168.1.100:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 发送来自非白名单 IP 的请求，第一个应该被允许
	req := httptest.NewRequest("GET", "/api/skip", nil)
	req.RemoteAddr = "192.168.1.200:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 发送来自非白名单 IP 的第二个请求，应该被限流
	req = httptest.NewRequest("GET", "/api/skip", nil)
	req.RemoteAddr = "192.168.1.200:1234"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

// 测试组合键函数
func TestCombinedKeyFunc(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用组合键函数的限流中间件
	r.Use(ratelimit.NewWithConfig(
		ratelimit.WithLimit(1),
		ratelimit.WithBurst(1),
		ratelimit.WithKeyFunc(ratelimit.CombinedKeyFunc(
			ratelimit.IPKeyFunc(),
			ratelimit.PathKeyFunc(),
		)),
	))

	// 添加两个 API 端点
	r.GET("/api/combined1", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.GET("/api/combined2", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// 发送请求到 combined1，应该被允许
	req1 := httptest.NewRequest("GET", "/api/combined1", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 发送请求到 combined2，应该被允许（不同路径）
	req2 := httptest.NewRequest("GET", "/api/combined2", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 发送第二个请求到 combined1，应该被限流
	req3 := httptest.NewRequest("GET", "/api/combined1", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

// 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	// 测试 RateLimit 辅助函数
	{
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(ratelimit.RateLimit(1, 1))
		r.GET("/helper1", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		// 第一个请求应该被允许
		req := httptest.NewRequest("GET", "/helper1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 第二个请求应该被限流
		req = httptest.NewRequest("GET", "/helper1", nil)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	}

	// 测试 RateLimitPerIP 辅助函数
	{
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(ratelimit.RateLimitPerIP(1, 1))
		r.GET("/helper2", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		// 第一个请求应该被允许
		req := httptest.NewRequest("GET", "/helper2", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 第二个请求应该被限流
		req = httptest.NewRequest("GET", "/helper2", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		// 来自不同 IP 的请求应该被允许
		req = httptest.NewRequest("GET", "/helper2", nil)
		req.RemoteAddr = "192.168.1.2:1234"
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}
