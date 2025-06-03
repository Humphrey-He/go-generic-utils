# 变更日志 (Changelog)

本文档记录 GGU (Go Generic Utils) 库的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [1.0.0] - 2024-05-15

### 新增 (Added)
- 数据结构模块 (dataStructures)
  - tuple: 元组实现
  - set: 集合实现
  - queue: 队列实现
  - list: 链表实现
  - maputils: 映射工具
- ginutil 模块: Gin 框架增强工具
  - binding: 请求绑定增强
  - render: 响应渲染工具
  - register: 路由注册工具
  - security: 安全相关工具
  - validate: 请求验证
  - paginator: 分页工具
  - contextx: 上下文增强
  - middleware: 中间件集合
  - response: 响应格式化
  - ecode: 错误码管理
- sliceutils: 切片操作工具
- net: 网络相关工具
- tree: 树数据结构
- syncx: 同步原语增强
- reflect: 反射工具
- pool: 对象池实现
- retry: 重试机制
- web: Web 开发工具
- pkg: 通用包

### 变更 (Changed)
- 首次稳定版本发布，API 已稳定

### 修复 (Fixed)
- 修复了绑定模块中的验证错误处理问题
- 修复了切片工具中的索引越界问题

### 安全 (Security)
- 增强了 gin 安全中间件 