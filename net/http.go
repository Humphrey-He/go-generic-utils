package net

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTP相关错误定义
var (
	ErrInvalidURL        = errors.New("ggu: 无效的URL")
	ErrRequestTimeout    = errors.New("ggu: 请求超时")
	ErrMaxRetriesReached = errors.New("ggu: 达到最大重试次数")
	ErrResponseDecode    = errors.New("ggu: 响应解码失败")
	ErrInvalidResponse   = errors.New("ggu: 无效的响应")
)

// HTTPClient 封装HTTP客户端，适用于电商平台后端开发
type HTTPClient struct {
	// 底层HTTP客户端
	client *http.Client
	// 基础URL
	baseURL string
	// 默认头信息
	defaultHeaders map[string]string
	// 最大重试次数
	maxRetries int
	// 重试间隔
	retryInterval time.Duration
	// 请求超时时间
	timeout time.Duration
}

// HTTPClientOption HTTP客户端配置选项
type HTTPClientOption func(*HTTPClient)

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(options ...HTTPClientOption) *HTTPClient {
	client := &HTTPClient{
		client:         &http.Client{},
		defaultHeaders: make(map[string]string),
		maxRetries:     3,
		retryInterval:  500 * time.Millisecond,
		timeout:        5 * time.Second,
	}

	// 应用配置选项
	for _, opt := range options {
		opt(client)
	}

	// 设置默认请求超时
	client.client.Timeout = client.timeout

	return client
}

// WithBaseURL 设置基础URL
func WithBaseURL(baseURL string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.baseURL = baseURL
	}
}

// WithTimeout 设置请求超时时间
func WithTimeout(timeout time.Duration) HTTPClientOption {
	return func(c *HTTPClient) {
		c.timeout = timeout
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) HTTPClientOption {
	return func(c *HTTPClient) {
		if maxRetries > 0 {
			c.maxRetries = maxRetries
		}
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) HTTPClientOption {
	return func(c *HTTPClient) {
		if interval > 0 {
			c.retryInterval = interval
		}
	}
}

// WithDefaultHeader 设置默认请求头
func WithDefaultHeader(key, value string) HTTPClientOption {
	return func(c *HTTPClient) {
		c.defaultHeaders[key] = value
	}
}

// WithDefaultHeaders 批量设置默认请求头
func WithDefaultHeaders(headers map[string]string) HTTPClientOption {
	return func(c *HTTPClient) {
		for k, v := range headers {
			c.defaultHeaders[k] = v
		}
	}
}

// buildURL 构建完整URL
func (c *HTTPClient) buildURL(path string) (string, error) {
	if path == "" {
		return "", ErrInvalidURL
	}

	// 如果path已经是完整URL，则直接返回
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	// 确保baseURL非空
	if c.baseURL == "" {
		return "", ErrInvalidURL
	}

	// 处理path前缀的斜杠
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 处理baseURL后缀的斜杠
	baseURL := strings.TrimSuffix(c.baseURL, "/")

	return baseURL + path, nil
}

// Get 发送GET请求
func (c *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

// Post 发送POST请求
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, headers)
}

// Put 发送PUT请求
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, headers)
}

// Delete 发送DELETE请求
func (c *HTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, headers)
}

// doRequest 执行HTTP请求
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	// 构建完整URL
	fullURL, err := c.buildURL(path)
	if err != nil {
		return nil, err
	}

	// 准备请求体
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case url.Values:
			bodyReader = strings.NewReader(v.Encode())
			// 默认设置表单内容类型
			if headers == nil {
				headers = make(map[string]string)
			}
			if _, exists := headers["Content-Type"]; !exists {
				headers["Content-Type"] = "application/x-www-form-urlencoded"
			}
		default:
			// 默认为JSON
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal request body failed: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
			// 默认设置JSON内容类型
			if headers == nil {
				headers = make(map[string]string)
			}
			if _, exists := headers["Content-Type"]; !exists {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置默认头信息
	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}

	// 设置自定义头信息
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 执行请求并重试
	return c.doRequestWithRetry(req)
}

// doRequestWithRetry 执行HTTP请求并在失败时重试
func (c *HTTPClient) doRequestWithRetry(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	for i := 0; i <= c.maxRetries; i++ {
		resp, err = c.client.Do(req)
		if err == nil {
			return resp, nil
		}

		// 检查是否超时
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrRequestTimeout
		}

		// 检查是否已经是最后一次尝试
		if i == c.maxRetries {
			return nil, fmt.Errorf("%w: %v", ErrMaxRetriesReached, err)
		}

		// 等待重试
		time.Sleep(c.retryInterval)
	}

	return nil, fmt.Errorf("unexpected error: %w", err)
}

// GetJSON 发送GET请求并解析JSON响应
func (c *HTTPClient) GetJSON(ctx context.Context, path string, result interface{}, headers map[string]string) error {
	resp, err := c.Get(ctx, path, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return c.parseJSONResponse(resp, result)
}

// PostJSON 发送POST请求并解析JSON响应
func (c *HTTPClient) PostJSON(ctx context.Context, path string, body, result interface{}, headers map[string]string) error {
	resp, err := c.Post(ctx, path, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return c.parseJSONResponse(resp, result)
}

// PutJSON 发送PUT请求并解析JSON响应
func (c *HTTPClient) PutJSON(ctx context.Context, path string, body, result interface{}, headers map[string]string) error {
	resp, err := c.Put(ctx, path, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return c.parseJSONResponse(resp, result)
}

// DeleteJSON 发送DELETE请求并解析JSON响应
func (c *HTTPClient) DeleteJSON(ctx context.Context, path string, result interface{}, headers map[string]string) error {
	resp, err := c.Delete(ctx, path, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return c.parseJSONResponse(resp, result)
}

// parseJSONResponse 解析JSON响应
func (c *HTTPClient) parseJSONResponse(resp *http.Response, result interface{}) error {
	// 检查响应状态
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status code: %d, body: %s", ErrInvalidResponse, resp.StatusCode, string(bodyBytes))
	}

	// 解析JSON响应
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("%w: %v", ErrResponseDecode, err)
	}

	return nil
}

// PostForm 发送表单POST请求
func (c *HTTPClient) PostForm(ctx context.Context, path string, formData url.Values, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	return c.Post(ctx, path, formData, headers)
}

// PostFormJSON 发送表单POST请求并解析JSON响应
func (c *HTTPClient) PostFormJSON(ctx context.Context, path string, formData url.Values, result interface{}, headers map[string]string) error {
	resp, err := c.PostForm(ctx, path, formData, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return c.parseJSONResponse(resp, result)
}

// ECommerceAPIClient 电商平台API客户端
type ECommerceAPIClient struct {
	httpClient *HTTPClient
	apiKey     string
	secretKey  string
}

// NewECommerceAPIClient 创建电商平台API客户端
func NewECommerceAPIClient(baseURL, apiKey, secretKey string, timeout time.Duration) *ECommerceAPIClient {
	return &ECommerceAPIClient{
		httpClient: NewHTTPClient(
			WithBaseURL(baseURL),
			WithTimeout(timeout),
			WithMaxRetries(3),
			WithDefaultHeader("User-Agent", "ECommerceSDK/1.0"),
		),
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// GetProduct 获取商品信息
func (c *ECommerceAPIClient) GetProduct(ctx context.Context, productID string) (map[string]interface{}, error) {
	// 设置认证头
	headers := map[string]string{
		"X-API-Key": c.apiKey,
	}

	var result map[string]interface{}
	err := c.httpClient.GetJSON(ctx, fmt.Sprintf("/products/%s", productID), &result, headers)
	return result, err
}

// CreateOrder 创建订单
func (c *ECommerceAPIClient) CreateOrder(ctx context.Context, orderData map[string]interface{}) (map[string]interface{}, error) {
	// 设置认证头
	headers := map[string]string{
		"X-API-Key": c.apiKey,
	}

	var result map[string]interface{}
	err := c.httpClient.PostJSON(ctx, "/orders", orderData, &result, headers)
	return result, err
}

// GetOrders 获取订单列表
func (c *ECommerceAPIClient) GetOrders(ctx context.Context, page, pageSize int) (map[string]interface{}, error) {
	// 设置认证头
	headers := map[string]string{
		"X-API-Key": c.apiKey,
	}

	path := fmt.Sprintf("/orders?page=%d&pageSize=%d", page, pageSize)
	var result map[string]interface{}
	err := c.httpClient.GetJSON(ctx, path, &result, headers)
	return result, err
}

// UpdateInventory 更新库存
func (c *ECommerceAPIClient) UpdateInventory(ctx context.Context, productID string, quantity int) (map[string]interface{}, error) {
	// 设置认证头
	headers := map[string]string{
		"X-API-Key": c.apiKey,
	}

	data := map[string]interface{}{
		"quantity": quantity,
	}

	var result map[string]interface{}
	err := c.httpClient.PutJSON(ctx, fmt.Sprintf("/products/%s/inventory", productID), data, &result, headers)
	return result, err
}
