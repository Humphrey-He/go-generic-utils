# Gin 渲染工具示例

本目录包含了 `ginutil/render` 包的使用示例，展示了如何在 Gin 框架中使用标准化的响应渲染功能。

## 目录结构

- `basic/` - 基础用法示例
  - `main.go` - 主程序，展示基本的 JSON、XML 和 HTML 响应渲染
  - `templates/` - HTML 模板文件
- `advanced/` - 高级用法示例
  - `main.go` - 主程序，展示高级功能，如链式 API、错误处理、中间件等
  - `data/` - 示例数据文件

## 基础用法示例

基础示例展示了 `render` 包的核心功能：

1. JSON 响应渲染
   - 成功响应 (`Success`)
   - 错误响应 (`Error`, `BadRequest`, `NotFound` 等)
   - 分页响应 (`Paginated`)

2. XML 响应渲染
   - XML 格式的标准响应

3. HTML 模板渲染
   - 页面渲染 (`HTML`)
   - 错误页面渲染 (`HTMLErrorPage`)

### 运行基础示例

```bash
cd basic
go run main.go
```

然后访问 http://localhost:8080 查看示例。

## 高级用法示例

高级示例展示了更复杂的功能：

1. 链式 API
   - 使用 `Resp()` 构建流式响应

2. 错误处理
   - 自定义错误 (`NewError`)
   - 错误包装 (`WrapError`)
   - 错误处理中间件

3. 安全特性
   - XSS 防护
   - CSRF 保护
   - 限流

4. 其他高级功能
   - 文件下载
   - HTML 辅助函数
   - 认证中间件

### 运行高级示例

```bash
cd advanced
go run main.go
```

然后访问 http://localhost:8081 查看示例。

## 注意事项

1. 示例代码仅用于演示目的，不包含完整的错误处理和安全措施。
2. 在实际项目中，请确保添加适当的日志记录、错误处理和安全检查。
3. 示例中的数据存储使用内存变量，实际应用中应使用数据库或其他持久化存储。 