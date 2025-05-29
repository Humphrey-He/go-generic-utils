// Package response 提供了一套标准化的HTTP响应工具，用于Gin应用程序。
//
// 本包的主要目标是统一API响应格式，简化控制器代码，并提供一致的错误处理机制。
// 它与 ginutil/ecode 包紧密配合，通过业务错误码系统实现细粒度的错误分类和处理。
//
// 主要特性:
//   - 使用泛型的标准响应结构 (StandardResponse[T])
//   - 成功响应辅助函数 (OK, Created, NoContent等)
//   - 错误响应辅助函数 (Fail, BadRequest, NotFound等)
//   - 分页数据响应支持 (PaginatedData, RespondPaginated)
//   - 业务错误码与HTTP状态码的智能映射
//
// 基本用法:
//
//	import (
//	    "ggu/ginutil/response"
//	    "github.com/gin-gonic/gin"
//	)
//
//	// 成功响应示例
//	func GetUser(c *gin.Context) {
//	    user := User{ID: 1, Name: "张三"}
//	    response.OK(c, user)
//	}
//
//	// 错误响应示例
//	func CreateUser(c *gin.Context) {
//	    var req UserRequest
//	    if err := c.ShouldBindJSON(&req); err != nil {
//	        response.BadRequest(c, "请求参数错误", err.Error())
//	        return
//	    }
//	    // 处理业务逻辑...
//	}
//
//	// 分页响应示例
//	func ListUsers(c *gin.Context) {
//	    pageInfo := response.GetPageInfo(c)
//	    users := []User{...} // 从数据库获取
//	    totalCount := int64(100) // 总记录数
//	    response.RespondPaginated(c, users, totalCount, pageInfo.PageNum, pageInfo.PageSize)
//	}
//
// 与错误码系统集成:
// 本包使用 ginutil/ecode 包中定义的错误码，通过 mapBusinessCodeToHTTPStatus 函数
// 将业务错误码映射到合适的HTTP状态码。这使得API可以在保持RESTful语义的同时，
// 提供更细粒度的错误信息。
//
// 链路追踪支持:
// 响应中包含可选的 trace_id 字段，用于跟踪请求在系统中的处理流程。
// 默认从Gin上下文中的 X-Trace-ID 键获取，通常由中间件设置。
package response
