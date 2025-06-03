package render_test

import (
	"fmt"
	"net/http"

	"github.com/Humphrey-He/go-generic-utils/ginutil/render"

	"github.com/gin-gonic/gin"
)

// 示例用户结构体
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 示例：配置渲染器
func ExampleConfigure() {
	// 设置渲染配置
	render.Configure(render.Config{
		JSONPrettyPrint:          true,
		HTMLTemplateDir:          "views/*",
		DefaultHTMLErrorTemplate: "views/error.html",
	})

	// 输出：
	// 配置已应用
}

// 示例：发送成功的 JSON 响应
func ExampleSuccess() {
	// 创建 Gin 引擎
	r := gin.New()

	// 注册路由
	r.GET("/api/users/:id", func(c *gin.Context) {
		// 获取用户数据
		user := User{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
		}

		// 发送成功响应
		render.Success(c, user)
	})

	// 输出:
	// {
	//   "code": 0,
	//   "message": "操作成功",
	//   "data": {
	//     "id": 1,
	//     "username": "admin",
	//     "email": "admin@example.com"
	//   },
	//   "trace_id": "",
	//   "server_time": "2024-09-01T12:00:00Z"
	// }
}

// 示例：发送错误的 JSON 响应
func ExampleError() {
	// 创建 Gin 引擎
	r := gin.New()

	// 注册路由
	r.GET("/api/users/:id", func(c *gin.Context) {
		// 发送错误响应
		render.Error(c, render.CodeNotFound, "用户不存在")
	})

	// 输出:
	// {
	//   "code": 1004,
	//   "message": "用户不存在",
	//   "data": {},
	//   "trace_id": "",
	//   "server_time": "2024-09-01T12:00:00Z"
	// }
}

// 示例：发送分页的 JSON 响应
func ExamplePaginated() {
	// 创建 Gin 引擎
	r := gin.New()

	// 注册路由
	r.GET("/api/users", func(c *gin.Context) {
		// 用户列表
		users := []User{
			{ID: 1, Username: "user1", Email: "user1@example.com"},
			{ID: 2, Username: "user2", Email: "user2@example.com"},
		}

		// 发送分页响应
		render.Paginated(c, users, 100, 1, 10)
	})

	// 输出:
	// {
	//   "code": 0,
	//   "message": "操作成功",
	//   "data": {
	//     "list": [
	//       {"id": 1, "username": "user1", "email": "user1@example.com"},
	//       {"id": 2, "username": "user2", "email": "user2@example.com"}
	//     ],
	//     "total": 100,
	//     "page_num": 1,
	//     "page_size": 10,
	//     "pages": 10,
	//     "has_next": true,
	//     "has_prev": false
	//   },
	//   "trace_id": "",
	//   "server_time": "2024-09-01T12:00:00Z"
	// }
}

// 示例：发送 XML 响应
func ExampleSuccessXML() {
	// 创建 Gin 引擎
	r := gin.New()

	// 注册路由
	r.GET("/api/users/:id", func(c *gin.Context) {
		// 获取用户数据
		user := User{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
		}

		// 发送 XML 响应
		render.SuccessXML(c, user)
	})

	// 输出:
	// <StandardResponse>
	//   <code>0</code>
	//   <message>操作成功</message>
	//   <data>
	//     <id>1</id>
	//     <username>admin</username>
	//     <email>admin@example.com</email>
	//   </data>
	//   <trace_id></trace_id>
	//   <server_time>2024-09-01T12:00:00Z</server_time>
	// </StandardResponse>
}

// 示例：渲染 HTML 模板
func ExampleHTML() {
	// 创建 Gin 引擎
	r := gin.New()

	// 注册模板
	render.RegisterTemplates(r, "templates/*")

	// 注册辅助函数
	render.AddUserDataHelper()
	render.AddCSRFHelper()

	// 注册路由
	r.GET("/users/:id", func(c *gin.Context) {
		// 获取用户数据
		user := User{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
		}

		// 设置用户 ID
		c.Set(render.ContextKeyUserID, user.ID)

		// 渲染模板
		render.HTML(c, http.StatusOK, "user.html", gin.H{
			"title": "用户详情",
			"user":  user,
		})
	})

	// 输出:
	// <!DOCTYPE html>
	// <html>
	// <head>
	//   <title>用户详情</title>
	// </head>
	// <body>
	//   <h1>用户详情</h1>
	//   <p>ID: 1</p>
	//   <p>用户名: admin</p>
	//   <p>邮箱: admin@example.com</p>
	// </body>
	// </html>
}

// 示例：渲染错误页面
func ExampleHTMLErrorPage() {
	// 创建 Gin 引擎
	r := gin.New()

	// 注册模板
	render.RegisterTemplates(r, "templates/*")

	// 注册路由
	r.GET("/users/:id", func(c *gin.Context) {
		// 渲染错误页面
		render.NotFoundPage(c, "用户不存在")
	})

	// 输出:
	// <!DOCTYPE html>
	// <html>
	// <head>
	//   <title>错误 - 资源不存在</title>
	// </head>
	// <body>
	//   <h1>错误</h1>
	//   <p>错误码: 1004</p>
	//   <p>错误信息: 用户不存在</p>
	// </body>
	// </html>
}

// 示例：完整的 Web 应用
func Example() {
	// 创建 Gin 引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 配置渲染器
	render.Configure(render.Config{
		JSONPrettyPrint:          true,
		HTMLTemplateDir:          "templates/*",
		DefaultHTMLErrorTemplate: "error.html",
	})

	// 注册模板
	render.RegisterTemplates(r)

	// 注册辅助函数
	render.AddUserDataHelper()
	render.AddCSRFHelper()
	render.AddFlashMessageHelper()

	// 添加 API 路由
	api := r.Group("/api")
	{
		api.GET("/users", func(c *gin.Context) {
			// 模拟分页数据
			users := []User{
				{ID: 1, Username: "user1", Email: "user1@example.com"},
				{ID: 2, Username: "user2", Email: "user2@example.com"},
			}

			render.Paginated(c, users, 100, 1, 10)
		})

		api.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id != "1" && id != "2" {
				render.NotFound(c, fmt.Sprintf("用户 %s 不存在", id))
				return
			}

			user := User{
				ID:       1,
				Username: "admin",
				Email:    "admin@example.com",
			}

			render.Success(c, user)
		})

		api.POST("/users", func(c *gin.Context) {
			var user User
			if err := c.ShouldBindJSON(&user); err != nil {
				render.BadRequest(c, "无效的用户数据")
				return
			}

			// 模拟创建用户
			user.ID = 3

			render.Success(c, user, "用户创建成功")
		})
	}

	// 添加 Web 路由
	r.GET("/", func(c *gin.Context) {
		render.HTMLSuccessPage(c, "index.html", gin.H{
			"title": "首页",
		})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id != "1" && id != "2" {
			render.NotFoundPage(c, fmt.Sprintf("用户 %s 不存在", id))
			return
		}

		user := User{
			ID:       1,
			Username: "admin",
			Email:    "admin@example.com",
		}

		render.HTML(c, http.StatusOK, "user.html", gin.H{
			"title": "用户详情",
			"user":  user,
		})
	})

	// 启动服务器
	// r.Run(":8080")

	// 输出:
	// [GIN-debug] GET    /api/users              --> main.Example.func1 (1 handlers)
	// [GIN-debug] GET    /api/users/:id          --> main.Example.func2 (1 handlers)
	// [GIN-debug] POST   /api/users              --> main.Example.func3 (1 handlers)
	// [GIN-debug] GET    /                       --> main.Example.func4 (1 handlers)
	// [GIN-debug] GET    /users/:id              --> main.Example.func5 (1 handlers)
}
