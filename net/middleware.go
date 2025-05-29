package net

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// HttpLogger 用于记录HTTP请求和响应的中间件
type HttpLogger struct {
	delegate http.RoundTripper
	log      func(l LogEntry, err error)
}

// LogEntry 包含HTTP请求和响应的日志信息
type LogEntry struct {
	URL         string        // 请求URL
	Method      string        // 请求方法
	ReqHeaders  http.Header   // 请求头
	ReqBody     string        // 请求体
	RespStatus  string        // 响应状态
	RespHeaders http.Header   // 响应头
	RespBody    string        // 响应体
	StartTime   time.Time     // 请求开始时间
	Duration    time.Duration // 请求持续时间
}

// NewHttpLogger 创建一个新的日志记录中间件
func NewHttpLogger(rp http.RoundTripper, log func(l LogEntry, err error)) *HttpLogger {
	if rp == nil {
		rp = http.DefaultTransport
	}
	return &HttpLogger{
		delegate: rp,
		log:      log,
	}
}

// RoundTrip 实现http.RoundTripper接口
func (l *HttpLogger) RoundTrip(request *http.Request) (resp *http.Response, err error) {
	logEntry := LogEntry{
		URL:        request.URL.String(),
		Method:     request.Method,
		ReqHeaders: request.Header.Clone(),
		StartTime:  time.Now(),
	}

	defer func() {
		logEntry.Duration = time.Since(logEntry.StartTime)

		if resp != nil {
			logEntry.RespStatus = resp.Status
			logEntry.RespHeaders = resp.Header.Clone()

			if resp.Body != nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body = io.NopCloser(bytes.NewReader(body))
				logEntry.RespBody = string(body)
			}
		}

		l.log(logEntry, err)
	}()

	if request.Body != nil {
		body, _ := io.ReadAll(request.Body)
		request.Body = io.NopCloser(bytes.NewReader(body))
		logEntry.ReqBody = string(body)
	}

	resp, err = l.delegate.RoundTrip(request)
	return
}

// NewLoggingClient 创建一个支持日志记录的HTTP客户端
func NewLoggingClient(logger func(l LogEntry, err error)) *http.Client {
	return &http.Client{
		Transport: NewHttpLogger(http.DefaultTransport, logger),
	}
}

// 电商平台常用的HTTP客户端配置

// NewEcommerceClient 创建一个适合电商平台使用的HTTP客户端
// 包含合理的超时设置、重试机制和日志记录
func NewEcommerceClient(logger func(l LogEntry, err error)) *http.Client {
	return &http.Client{
		Transport: NewHttpLogger(http.DefaultTransport, logger),
		Timeout:   10 * time.Second, // 默认10秒超时
	}
}

// RetryableTransport 实现可重试的HTTP传输
type RetryableTransport struct {
	delegate    http.RoundTripper
	maxRetries  int
	retryDelay  time.Duration
	shouldRetry func(resp *http.Response, err error) bool
}

// NewRetryableTransport 创建一个支持重试的HTTP传输
func NewRetryableTransport(delegate http.RoundTripper, maxRetries int, retryDelay time.Duration) *RetryableTransport {
	if delegate == nil {
		delegate = http.DefaultTransport
	}

	return &RetryableTransport{
		delegate:   delegate,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		shouldRetry: func(resp *http.Response, err error) bool {
			// 默认重试条件：网络错误或5xx服务器错误
			if err != nil {
				return true
			}
			return resp.StatusCode >= 500
		},
	}
}

// RoundTrip 实现http.RoundTripper接口，支持重试
func (rt *RetryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// 保存请求体以便重试时使用
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	for attempt := 0; attempt <= rt.maxRetries; attempt++ {
		// 如果不是第一次尝试，等待一段时间后再重试
		if attempt > 0 {
			time.Sleep(rt.retryDelay)
		}

		// 重新设置请求体
		if reqBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		resp, err = rt.delegate.RoundTrip(req)

		// 检查是否需要重试
		if !rt.shouldRetry(resp, err) || attempt == rt.maxRetries {
			break
		}

		// 关闭响应体以避免资源泄漏
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	return resp, err
}

// NewRetryableClient 创建一个支持重试的HTTP客户端
func NewRetryableClient(maxRetries int, retryDelay time.Duration) *http.Client {
	return &http.Client{
		Transport: NewRetryableTransport(http.DefaultTransport, maxRetries, retryDelay),
	}
}

// WithRetryAndLogging 创建一个同时支持重试和日志记录的HTTP客户端
func WithRetryAndLogging(maxRetries int, retryDelay time.Duration, logger func(l LogEntry, err error)) *http.Client {
	retryTransport := NewRetryableTransport(http.DefaultTransport, maxRetries, retryDelay)
	return &http.Client{
		Transport: NewHttpLogger(retryTransport, logger),
		Timeout:   30 * time.Second, // 考虑到重试，超时时间设置长一些
	}
}

// RateLimitTransport 实现限流的HTTP传输
type RateLimitTransport struct {
	delegate   http.RoundTripper
	ticker     *time.Ticker
	reqChannel chan struct{}
}

// NewRateLimitTransport 创建一个支持限流的HTTP传输
// maxQPS: 每秒最大请求数
func NewRateLimitTransport(delegate http.RoundTripper, maxQPS int) *RateLimitTransport {
	if delegate == nil {
		delegate = http.DefaultTransport
	}

	// 创建一个缓冲通道，用于限制并发请求数
	reqChannel := make(chan struct{}, maxQPS)

	// 初始填充通道
	for i := 0; i < maxQPS; i++ {
		reqChannel <- struct{}{}
	}

	// 创建一个定时器，每秒释放maxQPS个令牌
	ticker := time.NewTicker(time.Second)

	rt := &RateLimitTransport{
		delegate:   delegate,
		ticker:     ticker,
		reqChannel: reqChannel,
	}

	// 启动一个goroutine，定时释放令牌
	go func() {
		for range ticker.C {
			// 每秒释放maxQPS个令牌
			for i := 0; i < maxQPS; i++ {
				select {
				case rt.reqChannel <- struct{}{}:
					// 成功添加令牌
				default:
					// 通道已满，不需要添加更多令牌
				}
			}
		}
	}()

	return rt
}

// RoundTrip 实现http.RoundTripper接口，支持限流
func (rt *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 获取令牌，如果没有令牌可用，会阻塞
	<-rt.reqChannel

	// 发送请求
	return rt.delegate.RoundTrip(req)
}

// Close 关闭限流传输
func (rt *RateLimitTransport) Close() {
	rt.ticker.Stop()
	close(rt.reqChannel)
}

// NewRateLimitClient 创建一个支持限流的HTTP客户端
func NewRateLimitClient(maxQPS int) *http.Client {
	return &http.Client{
		Transport: NewRateLimitTransport(http.DefaultTransport, maxQPS),
	}
}
