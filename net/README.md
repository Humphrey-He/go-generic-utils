# net - 网络工具包

`net`包提供了增强的网络通信工具，特别是HTTP客户端的封装，简化了RESTful API调用、JSON处理和表单提交等常见网络操作。该包专为高并发、高可靠的后端服务设计，内置了重试、超时等机制。

## 核心特性

- **增强的HTTP客户端**：封装标准库的`http.Client`，提供更友好的API
- **自动重试机制**：内置请求重试逻辑，应对网络抖动
- **超时控制**：细粒度的请求超时控制
- **JSON处理**：自动处理JSON请求和响应
- **表单提交**：简化表单数据提交
- **中间件支持**：可扩展的HTTP中间件机制
- **电商API客户端**：专为电商场景优化的API客户端实现

## 使用示例

### 基本HTTP请求

```go
// 创建HTTP客户端
client := net.NewHTTPClient(
    net.WithTimeout(5*time.Second),
    net.WithMaxRetries(3),
    net.WithRetryInterval(500*time.Millisecond),
)

// 发送GET请求
resp, err := client.Get(context.Background(), "https://api.example.com/users", nil)
if err != nil {
    log.Fatalf("请求失败: %v", err)
}
defer resp.Body.Close()

// 读取响应
data, err := io.ReadAll(resp.Body)
if err != nil {
    log.Fatalf("读取响应失败: %v", err)
}
fmt.Println(string(data))
```

### JSON请求和响应

```go
// 创建HTTP客户端
client := net.NewHTTPClient(
    net.WithBaseURL("https://api.example.com"),
    net.WithDefaultHeader("Authorization", "Bearer token123"),
)

// 请求体
requestBody := map[string]interface{}{
    "name": "商品A",
    "price": 99.99,
    "stock": 100,
}

// 响应结构
var response struct {
    ID      string  `json:"id"`
    Name    string  `json:"name"`
    Price   float64 `json:"price"`
    Created string  `json:"created_at"`
}

// 发送POST请求并自动解析JSON响应
err := client.PostJSON(
    context.Background(),
    "/products",
    requestBody,
    &response,
    nil,
)

if err != nil {
    log.Fatalf("创建产品失败: %v", err)
}

fmt.Printf("创建的产品ID: %s, 名称: %s\n", response.ID, response.Name)
```

### 表单提交

```go
// 创建表单数据
formData := url.Values{}
formData.Set("username", "user123")
formData.Set("password", "pass456")

// 创建HTTP客户端
client := net.NewHTTPClient()

// 提交表单并解析JSON响应
var loginResponse struct {
    Token   string `json:"token"`
    Expires int64  `json:"expires"`
}

err := client.PostFormJSON(
    context.Background(),
    "https://api.example.com/login",
    formData,
    &loginResponse,
    nil,
)

if err != nil {
    log.Fatalf("登录失败: %v", err)
}

fmt.Printf("登录成功，令牌: %s\n", loginResponse.Token)
```

### 电商API客户端

```go
// 创建电商API客户端
apiClient := net.NewECommerceAPIClient(
    "https://api.ecommerce.com",
    "api_key_123",
    "secret_key_456",
    10*time.Second,
)

// 获取产品信息
product, err := apiClient.GetProduct(context.Background(), "prod-123")
if err != nil {
    log.Fatalf("获取产品失败: %v", err)
}
fmt.Printf("产品名称: %s, 价格: %.2f\n", product["name"], product["price"])

// 创建订单
orderData := map[string]interface{}{
    "customer_id": "cust-456",
    "items": []map[string]interface{}{
        {"product_id": "prod-123", "quantity": 2},
        {"product_id": "prod-456", "quantity": 1},
    },
}

order, err := apiClient.CreateOrder(context.Background(), orderData)
if err != nil {
    log.Fatalf("创建订单失败: %v", err)
}
fmt.Printf("订单ID: %s, 总金额: %.2f\n", order["id"], order["total_amount"])
```

## 高级用法

### 自定义请求头

```go
// 设置默认请求头
client := net.NewHTTPClient(
    net.WithDefaultHeaders(map[string]string{
        "User-Agent": "GGU Client/1.0",
        "Accept":     "application/json",
    }),
)

// 为特定请求设置头信息
headers := map[string]string{
    "X-Custom-Header": "value",
    "Authorization":   "Bearer token123",
}

resp, err := client.Get(context.Background(), "/api/resource", headers)
```

### 请求超时和重试

```go
// 创建具有超时和重试策略的客户端
client := net.NewHTTPClient(
    net.WithTimeout(3*time.Second),      // 3秒超时
    net.WithMaxRetries(5),               // 最多重试5次
    net.WithRetryInterval(200*time.Millisecond), // 重试间隔200毫秒
)

// 创建带超时的上下文
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 发送请求
resp, err := client.Get(ctx, "/api/slow-resource", nil)
```

## 最佳实践

1. **合理设置超时**：根据API预期响应时间设置合理的超时值
   ```go
   client := net.NewHTTPClient(net.WithTimeout(5*time.Second))
   ```

2. **使用基础URL**：对同一API的多个请求使用基础URL
   ```go
   client := net.NewHTTPClient(net.WithBaseURL("https://api.example.com"))
   resp, err := client.Get(ctx, "/users", nil) // 将请求 https://api.example.com/users
   ```

3. **复用HTTP客户端**：创建一个全局客户端实例，而不是每次请求都创建新客户端
   ```go
   var httpClient = net.NewHTTPClient(/* options */)
   
   func getUser() {
       httpClient.Get(/* ... */)
   }
   ```

4. **使用上下文控制请求生命周期**：通过上下文传递超时、取消信号等
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   client.Get(ctx, "/api/resource", nil)
   ```

5. **处理错误响应**：检查HTTP状态码，适当处理非2xx响应
   ```go
   resp, err := client.Get(ctx, "/api/resource", nil)
   if err != nil {
       // 处理网络错误
   } else if resp.StatusCode >= 400 {
       // 处理HTTP错误
   }
   ``` 