package timeout_test

import (
	"ggu/ginutil/middleware/timeout"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 测试正常请求在超时时间内完成
func TestTimeoutNormal(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用超时中间件，设置超时时间为 1 秒
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(1 * time.Second),
	))

	// 添加一个处理时间短于超时时间的路由
	r.GET("/normal", func(c *gin.Context) {
		// 模拟处理时间为 100 毫秒
		time.Sleep(100 * time.Millisecond)
		c.String(http.StatusOK, "正常响应")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/normal", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "正常响应", w.Body.String())
}

// 测试请求超时
func TestTimeoutExceeded(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用超时中间件，设置超时时间为 100 毫秒
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100 * time.Millisecond),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/timeout", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.String(http.StatusOK, "这个响应不应该被发送")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/timeout", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	// 检查响应体是否包含默认的超时消息
	assert.Contains(t, w.Body.String(), "请求处理超时")
}

// 测试自定义超时处理函数
func TestCustomTimeoutHandler(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 自定义超时处理函数
	customHandler := func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "自定义超时错误",
			"code":  503,
		})
	}

	// 使用超时中间件，设置超时时间为 100 毫秒和自定义处理函数
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutHandler(customHandler),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/custom-timeout", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.String(http.StatusOK, "这个响应不应该被发送")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/custom-timeout", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "自定义超时错误")
}

// 测试自定义超时 JSON 响应
func TestCustomTimeoutJSON(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 自定义超时 JSON 响应
	customJSON := gin.H{
		"error":   "请求超时",
		"code":    504,
		"timeout": true,
	}

	// 使用超时中间件，设置超时时间为 100 毫秒和自定义 JSON 响应
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutJSON(customJSON),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/json-timeout", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.String(http.StatusOK, "这个响应不应该被发送")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/json-timeout", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "请求超时")
	assert.Contains(t, w.Body.String(), "true")
}

// 测试 panic 传递
func TestPanicPropagation(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 添加恢复中间件来捕获 panic
	r.Use(gin.Recovery())

	// 使用超时中间件，设置超时时间为 1 秒
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(1 * time.Second),
	))

	// 添加一个会 panic 的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("测试 panic")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	// gin.Recovery() 会捕获 panic 并返回 500 状态码
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// 测试禁用超时功能
func TestDisableTimeout(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用超时中间件，但禁用超时功能
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithDisableTimeout(true),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/disabled-timeout", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 由于禁用了超时功能，这个响应应该被发送
		c.String(http.StatusOK, "正常响应")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/disabled-timeout", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "正常响应", w.Body.String())
}

// 测试自定义超时消息
func TestCustomTimeoutMessage(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用超时中间件，设置超时时间为 100 毫秒和自定义超时消息
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutMessage("自定义超时消息"),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/message-timeout", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.String(http.StatusOK, "这个响应不应该被发送")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/message-timeout", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "自定义超时消息")
}

// 测试自定义超时状态码
func TestCustomTimeoutCode(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 创建 Gin 路由
	r := gin.New()

	// 使用超时中间件，设置超时时间为 100 毫秒和自定义超时状态码
	r.Use(timeout.NewWithConfig(
		timeout.WithTimeout(100*time.Millisecond),
		timeout.WithTimeoutCode(http.StatusServiceUnavailable),
	))

	// 添加一个处理时间长于超时时间的路由
	r.GET("/code-timeout", func(c *gin.Context) {
		// 模拟处理时间为 500 毫秒
		time.Sleep(500 * time.Millisecond)
		// 这个响应不应该被发送
		c.String(http.StatusOK, "这个响应不应该被发送")
	})

	// 创建一个模拟请求
	req := httptest.NewRequest("GET", "/code-timeout", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	// 设置为测试模式
	gin.SetMode(gin.TestMode)

	// 测试 TimeoutWithHandler
	{
		r := gin.New()
		r.Use(timeout.TimeoutWithHandler(100*time.Millisecond, func(c *gin.Context) {
			c.String(http.StatusServiceUnavailable, "处理函数超时")
		}))
		r.GET("/handler", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
		})

		req := httptest.NewRequest("GET", "/handler", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Equal(t, "处理函数超时", w.Body.String())
	}

	// 测试 TimeoutWithMessage
	{
		r := gin.New()
		r.Use(timeout.TimeoutWithMessage(100*time.Millisecond, "消息超时"))
		r.GET("/message", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
		})

		req := httptest.NewRequest("GET", "/message", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
		assert.Contains(t, w.Body.String(), "消息超时")
	}

	// 测试 TimeoutWithCode
	{
		r := gin.New()
		r.Use(timeout.TimeoutWithCode(100*time.Millisecond, http.StatusBadGateway))
		r.GET("/code", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
		})

		req := httptest.NewRequest("GET", "/code", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
	}

	// 测试 TimeoutWithJSON
	{
		r := gin.New()
		r.Use(timeout.TimeoutWithJSON(100*time.Millisecond, gin.H{"error": "JSON 超时"}))
		r.GET("/json", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
		})

		req := httptest.NewRequest("GET", "/json", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
		assert.Contains(t, w.Body.String(), "JSON 超时")
	}
}
