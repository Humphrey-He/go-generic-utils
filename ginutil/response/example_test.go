package response_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/noobtrump/go-generic-utils/ginutil/ecode"
	"github.com/noobtrump/go-generic-utils/ginutil/response"

	"github.com/gin-gonic/gin"
)

func init() {
	// 设置固定的服务器时间戳，用于测试
	response.SetFixedServerTimeForTest(1620000000000)
}

// 用户模型示例
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 这个示例展示如何使用OK函数发送成功响应
func Example_ok() {
	// 创建Gin路由
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 定义处理函数
	r.GET("/users/:id", func(c *gin.Context) {
		// 模拟获取用户
		user := User{
			ID:    1,
			Name:  "张三",
			Email: "zhangsan@example.com",
		}

		// 设置模拟的TraceID
		c.Set(response.GinTraceIDKey, "test-trace-id")

		// 发送成功响应
		response.OK(c, user)
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 200
	// Response Body: {"code":0,"message":"操作成功","data":{"id":1,"name":"张三","email":"zhangsan@example.com"},"trace_id":"test-trace-id","server_time":1620000000000}
}

// 这个示例展示如何使用BadRequest函数发送错误响应
func Example_badRequest() {
	// 创建Gin路由
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 定义处理函数
	r.POST("/users", func(c *gin.Context) {
		// 模拟验证错误
		validationErrors := []map[string]string{
			{"field": "name", "message": "姓名不能为空"},
			{"field": "email", "message": "邮箱格式不正确"},
		}

		// 设置模拟的TraceID
		c.Set(response.GinTraceIDKey, "test-trace-id")

		// 发送错误响应
		response.BadRequest(c, "请求参数错误", validationErrors)
	})

	// 创建测试请求
	req := httptest.NewRequest("POST", "/users", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 400
	// Response Body: {"code":40001,"message":"请求参数错误","data":[{"field":"name","message":"姓名不能为空"},{"field":"email","message":"邮箱格式不正确"}],"trace_id":"test-trace-id","server_time":1620000000000}
}

// 这个示例展示如何使用RespondPaginated函数发送分页响应
func Example_respondPaginated() {
	// 创建Gin路由
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 定义处理函数
	r.GET("/users", func(c *gin.Context) {
		// 模拟分页数据
		users := []User{
			{ID: 1, Name: "张三", Email: "zhangsan@example.com"},
			{ID: 2, Name: "李四", Email: "lisi@example.com"},
		}
		totalCount := int64(100)
		pageNum := 1
		pageSize := 10

		// 设置模拟的TraceID
		c.Set(response.GinTraceIDKey, "test-trace-id")

		// 发送分页响应
		response.RespondPaginated(c, users, totalCount, pageNum, pageSize)
	})

	// 创建测试请求
	req := httptest.NewRequest("GET", "/users?pageNum=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 200
	// Response Body: {"code":0,"message":"操作成功","data":{"list":[{"id":1,"name":"张三","email":"zhangsan@example.com"},{"id":2,"name":"李四","email":"lisi@example.com"}],"total":100,"pageNum":1,"pageSize":10,"totalPages":10},"trace_id":"test-trace-id","server_time":1620000000000}
}

// 这个示例展示如何使用NoContent函数发送无内容响应
func Example_noContent() {
	// 创建Gin路由
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 定义处理函数
	r.DELETE("/users/:id", func(c *gin.Context) {
		// 设置模拟的TraceID
		c.Set(response.GinTraceIDKey, "test-trace-id")

		// 发送无内容响应
		response.NoContent(c)
	})

	// 创建测试请求
	req := httptest.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Headers:")
	fmt.Println("  X-Trace-ID:", w.Header().Get("X-Trace-ID"))
	fmt.Println("  X-Server-Time:", w.Header().Get("X-Server-Time"))
	fmt.Println("Response Body Length:", len(w.Body.String()))

	// Output:
	// Status Code: 204
	// Headers:
	//   X-Trace-ID: test-trace-id
	//   X-Server-Time: 1620000000000
	// Response Body Length: 0
}

// 这个示例展示如何注册和使用自定义错误码
func Example_customErrorCode() {
	// 注册自定义错误码消息
	const ErrorCodeOutOfStock = 40050
	ecode.RegisterMessage(ErrorCodeOutOfStock, "商品库存不足")

	// 创建Gin路由
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 定义处理函数
	r.POST("/orders", func(c *gin.Context) {
		// 模拟库存不足错误
		// 设置模拟的TraceID
		c.Set(response.GinTraceIDKey, "test-trace-id")

		// 发送自定义错误响应
		response.Fail(c, ErrorCodeOutOfStock, ecode.GetMessage(ErrorCodeOutOfStock))
	})

	// 创建测试请求
	req := httptest.NewRequest("POST", "/orders", nil)
	w := httptest.NewRecorder()

	// 执行请求
	r.ServeHTTP(w, req)

	// 输出响应
	fmt.Println("Status Code:", w.Code)
	fmt.Println("Response Body:", w.Body.String())

	// Output:
	// Status Code: 400
	// Response Body: {"code":40050,"message":"商品库存不足","trace_id":"test-trace-id","server_time":1620000000000}
}
