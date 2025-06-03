# retry - 重试机制

`retry`包提供了灵活的重试策略和机制，适用于处理可能失败的操作，如网络请求、第三方服务调用等。该包设计了可扩展的重试策略接口，并提供了多种常用策略实现。

## 核心特性

- **多种重试策略**：固定间隔、指数退避、自适应超时等
- **线程安全**：所有策略实现都保证线程安全
- **上下文集成**：支持通过context取消重试
- **可扩展性**：基于接口设计，易于扩展自定义策略
- **状态反馈**：支持根据错误类型调整重试行为

## 重试策略

### 固定间隔重试

最简单的重试策略，每次重试间隔相同。

```go
// 创建固定间隔重试策略，每5秒重试一次，最多重试3次
strategy, _ := retry.NewFixedIntervalRetryStrategy(5*time.Second, 3)

// 使用策略执行重试
err := retry.Retry(ctx, strategy, func() error {
    return callExternalService()
})
```

### 指数退避重试

每次重试的间隔呈指数增长，避免对故障服务造成过大压力。

```go
// 创建指数退避重试策略
// 初始间隔100ms，最大间隔10s，最多重试5次
strategy, _ := retry.NewExponentialBackoffRetryStrategy(
    100*time.Millisecond,
    10*time.Second,
    5,
)

// 使用策略执行重试
err := retry.Retry(ctx, strategy, func() error {
    return callExternalService()
})
```

### 自适应超时重试

根据历史失败率动态调整重试行为。

```go
// 创建基础策略
baseStrategy, _ := retry.NewExponentialBackoffRetryStrategy(
    100*time.Millisecond,
    5*time.Second,
    3,
)

// 创建自适应策略，窗口大小64，阈值10
adaptiveStrategy := retry.NewAdaptiveTimeoutRetryStrategy(
    baseStrategy,
    64,   // 滑动窗口长度
    10,   // 超时阈值
)

// 使用策略执行重试
err := retry.Retry(ctx, adaptiveStrategy, func() error {
    return callExternalService()
})
```

## 高级用法

### 带超时的重试

```go
// 创建带10秒超时的上下文
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 创建重试策略
strategy, _ := retry.NewExponentialBackoffRetryStrategy(
    100*time.Millisecond,
    1*time.Second,
    5,
)

// 执行重试，如果超过10秒会自动取消
err := retry.Retry(ctx, strategy, func() error {
    return callExternalService()
})
```

### 自定义重试策略

实现`Strategy`接口可以创建自定义重试策略：

```go
type MyStrategy struct {
    // 自定义字段
}

// 返回下一次重试的间隔和是否继续重试
func (s *MyStrategy) Next() (time.Duration, bool) {
    // 自定义逻辑
    return 1*time.Second, true
}

// 根据错误调整策略
func (s *MyStrategy) Report(err error) retry.Strategy {
    // 根据错误类型调整策略
    return s
}

// 使用自定义策略
err := retry.Retry(ctx, &MyStrategy{}, func() error {
    return callExternalService()
})
```

## 最佳实践

- 对于网络请求，推荐使用指数退避策略，避免在服务故障时造成雪崩
- 设置合理的最大重试次数和超时时间，避免无限重试
- 对于关键业务，使用自适应策略，根据历史成功率动态调整重试行为
- 在微服务架构中，为不同的服务调用配置不同的重试策略
- 结合熔断机制使用，避免对已知故障服务持续发起请求 