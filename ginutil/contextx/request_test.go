package contextx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/noobtrump/go-generic-utils/ginutil/contextx"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTraceID(t *testing.T) {
	// 创建 Gin 上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	// 测试设置跟踪 ID
	contextx.SetTraceID(c, "test-trace-id")
	traceID, exists := contextx.Get[string](c, contextx.TraceIDKey)
	assert.True(t, exists, "跟踪 ID 应该存在")
	assert.Equal(t, "test-trace-id", traceID, "跟踪 ID 应该正确")

	// 测试响应头中的跟踪 ID
	assert.Equal(t, "test-trace-id", w.Header().Get(contextx.TraceIDKey), "响应头中应该包含跟踪 ID")

	// 测试获取跟踪 ID
	retrievedID := contextx.GetTraceID(c)
	assert.Equal(t, "test-trace-id", retrievedID, "获取的跟踪 ID 应该正确")

	// 测试自动生成跟踪 ID
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Request, _ = http.NewRequest("GET", "/test", nil)

	generatedID := contextx.GetTraceID(c2)
	assert.NotEmpty(t, generatedID, "应该自动生成跟踪 ID")

	// 测试从请求头获取跟踪 ID
	c3, _ := gin.CreateTestContext(httptest.NewRecorder())
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(contextx.TraceIDKey, "header-trace-id")
	c3.Request = req

	headerID := contextx.GetTraceID(c3)
	assert.Equal(t, "header-trace-id", headerID, "应该从请求头获取跟踪 ID")

	// 测试 nil 上下文
	assert.Empty(t, contextx.GetTraceID(nil), "nil 上下文应该返回空字符串")
}

func TestRequestStartTime(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)

	// 测试设置请求开始时间
	now := time.Now()
	contextx.SetRequestStartTime(c, now)

	// 测试获取请求开始时间
	startTime, exists := contextx.GetRequestStartTime(c)
	assert.True(t, exists, "请求开始时间应该存在")
	assert.Equal(t, now, startTime, "请求开始时间应该正确")

	// 测试请求持续时间
	time.Sleep(10 * time.Millisecond)
	duration, exists := contextx.GetRequestDuration(c)
	assert.True(t, exists, "请求持续时间应该存在")
	assert.True(t, duration >= 10*time.Millisecond, "请求持续时间应该至少为 10ms")

	// 测试不存在的请求开始时间
	c2, _ := gin.CreateTestContext(nil)
	_, exists = contextx.GetRequestStartTime(c2)
	assert.False(t, exists, "请求开始时间不应该存在")

	_, exists = contextx.GetRequestDuration(c2)
	assert.False(t, exists, "请求持续时间不应该存在")

	// 测试 nil 上下文
	_, exists = contextx.GetRequestStartTime(nil)
	assert.False(t, exists, "nil 上下文应该返回 false")
}

func TestClientIP(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.RemoteAddr = "192.168.1.1:1234"

	// 测试设置客户端 IP
	contextx.SetClientIP(c, "10.0.0.1")

	// 测试获取客户端 IP
	ip := contextx.GetClientIP(c)
	assert.Equal(t, "10.0.0.1", ip, "应该返回设置的客户端 IP")

	// 测试从 c.ClientIP() 获取 IP
	c2, _ := gin.CreateTestContext(nil)
	c2.Request, _ = http.NewRequest("GET", "/test", nil)
	c2.Request.RemoteAddr = "192.168.1.2:1234"

	ip = contextx.GetClientIP(c2)
	assert.Equal(t, "192.168.1.2", ip, "应该返回 c.ClientIP() 的值")

	// 测试 nil 上下文
	assert.Empty(t, contextx.GetClientIP(nil), "nil 上下文应该返回空字符串")
}

func TestRealIP(t *testing.T) {
	// 测试 X-Real-IP 头
	c1, _ := gin.CreateTestContext(nil)
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Real-IP", "10.0.0.1")
	c1.Request = req1

	assert.Equal(t, "10.0.0.1", contextx.GetRealIP(c1), "应该返回 X-Real-IP 头的值")

	// 测试 X-Forwarded-For 头
	c2, _ := gin.CreateTestContext(nil)
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "20.0.0.1, 20.0.0.2")
	c2.Request = req2

	assert.Equal(t, "20.0.0.1", contextx.GetRealIP(c2), "应该返回 X-Forwarded-For 头的第一个值")

	// 测试 ClientIP 方法
	c3, _ := gin.CreateTestContext(nil)
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "30.0.0.1:1234"
	c3.Request = req3

	assert.Equal(t, "30.0.0.1", contextx.GetRealIP(c3), "应该返回 ClientIP 方法的值")

	// 测试 nil 上下文
	assert.Empty(t, contextx.GetRealIP(nil), "nil 上下文应该返回空字符串")
}

func TestRequestInfo(t *testing.T) {
	// 创建 Gin 上下文
	c, _ := gin.CreateTestContext(nil)
	req, _ := http.NewRequest("POST", "/api/test?q=value", nil)
	req.Host = "example.com"
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Referer", "https://example.org")
	c.Request = req

	// 测试各种请求信息获取函数
	assert.Equal(t, "test-agent", contextx.GetUserAgent(c), "应该返回正确的用户代理")
	assert.Equal(t, "https://example.org", contextx.GetReferer(c), "应该返回正确的 Referer")
	assert.Equal(t, "POST", contextx.GetRequestMethod(c), "应该返回正确的请求方法")
	assert.Equal(t, "/api/test", contextx.GetRequestPath(c), "应该返回正确的请求路径")
	assert.Equal(t, "q=value", contextx.GetRequestQuery(c), "应该返回正确的查询字符串")
	assert.Equal(t, "example.com", contextx.GetRequestHost(c), "应该返回正确的主机名")
	assert.Equal(t, "http", contextx.GetRequestProtocol(c), "应该返回正确的协议")

	// 测试 nil 上下文
	assert.Empty(t, contextx.GetUserAgent(nil), "nil 上下文应该返回空字符串")
	assert.Empty(t, contextx.GetReferer(nil), "nil 上下文应该返回空字符串")
	assert.Empty(t, contextx.GetRequestMethod(nil), "nil 上下文应该返回空字符串")
	assert.Empty(t, contextx.GetRequestPath(nil), "nil 上下文应该返回空字符串")
	assert.Empty(t, contextx.GetRequestQuery(nil), "nil 上下文应该返回空字符串")
	assert.Empty(t, contextx.GetRequestHost(nil), "nil 上下文应该返回空字符串")
	assert.Empty(t, contextx.GetRequestProtocol(nil), "nil 上下文应该返回空字符串")
}
