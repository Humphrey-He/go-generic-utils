package register

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResourceGroup 是一个增强的路由组，提供更便捷的 RESTful 资源路由注册。
type ResourceGroup struct {
	*gin.RouterGroup
	basePath string
}

// NewResourceGroup 创建一个新的资源路由组。
func NewResourceGroup(group *gin.RouterGroup) *ResourceGroup {
	return &ResourceGroup{
		RouterGroup: group,
		basePath:    group.BasePath(),
	}
}

// Group 创建一个新的路由组，并返回 ResourceGroup 实例。
func (g *ResourceGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *ResourceGroup {
	return NewResourceGroup(g.RouterGroup.Group(relativePath, handlers...))
}

// Resource 注册一个资源控制器，自动创建 RESTful 路由。
// 参数：
//   - name: 资源名称，用于构建路由路径
//   - controller: 实现了 RESTController 接口的控制器
//   - middleware: 可选的中间件，应用于所有路由
func (g *ResourceGroup) Resource(name string, controller RESTController, middleware ...gin.HandlerFunc) *ResourceGroup {
	// 创建资源路由组
	resourceGroup := g.Group(name)

	// 注册 RESTful 路由
	RegisterRESTRoutesWithMiddleware(resourceGroup.RouterGroup, controller, middleware...)

	return resourceGroup
}

// NestedResource 注册一个嵌套的资源控制器，如 /posts/:postId/comments。
// 参数：
//   - parentName: 父资源名称，如 "posts"
//   - parentParam: 父资源参数名，如 "postId"
//   - name: 资源名称，如 "comments"
//   - controller: 实现了 RESTController 接口的控制器
//   - middleware: 可选的中间件，应用于所有路由
func (g *ResourceGroup) NestedResource(parentName, parentParam, name string, controller RESTController, middleware ...gin.HandlerFunc) *ResourceGroup {
	// 创建嵌套路由路径，如 /posts/:postId/comments
	nestedPath := path.Join(parentName, ":"+parentParam, name)

	// 创建资源路由组
	resourceGroup := g.Group(nestedPath)

	// 注册 RESTful 路由
	RegisterRESTRoutesWithMiddleware(resourceGroup.RouterGroup, controller, middleware...)

	return resourceGroup
}

// ResourceWithID 注册一个资源控制器，使用自定义的 ID 参数名。
// 参数：
//   - name: 资源名称，用于构建路由路径
//   - idParam: ID 参数名，默认为 "id"
//   - controller: 实现了 RESTController 接口的控制器
//   - middleware: 可选的中间件，应用于所有路由
func (g *ResourceGroup) ResourceWithID(name, idParam string, controller RESTController, middleware ...gin.HandlerFunc) *ResourceGroup {
	// 创建资源路由组
	resourceGroup := g.Group(name)

	// 注册自定义 ID 参数的 RESTful 路由
	rg := resourceGroup.RouterGroup

	// 应用中间件
	handlers := middleware

	// 注册路由
	rg.GET("", append(handlers, controller.Index)...)
	rg.GET("/:"+idParam, append(handlers, controller.Show)...)
	rg.POST("", append(handlers, controller.Create)...)
	rg.PUT("/:"+idParam, append(handlers, controller.Update)...)
	rg.DELETE("/:"+idParam, append(handlers, controller.Delete)...)

	return resourceGroup
}

// Mount 将另一个路由组挂载到当前路由组。
func (g *ResourceGroup) Mount(relativePath string, mountedGroup *ResourceGroup) *ResourceGroup {
	// 获取被挂载组的所有路由
	mountedRoutes := extractRoutesFromGroup(mountedGroup.RouterGroup)

	// 在当前组中注册这些路由
	for _, route := range mountedRoutes {
		// 构建新的路径
		newPath := path.Join(relativePath, strings.TrimPrefix(route.Path, mountedGroup.basePath))

		// 注册路由
		switch route.Method {
		case http.MethodGet:
			g.GET(newPath, route.Handlers...)
		case http.MethodPost:
			g.POST(newPath, route.Handlers...)
		case http.MethodPut:
			g.PUT(newPath, route.Handlers...)
		case http.MethodDelete:
			g.DELETE(newPath, route.Handlers...)
		case http.MethodPatch:
			g.PATCH(newPath, route.Handlers...)
		case http.MethodHead:
			g.HEAD(newPath, route.Handlers...)
		case http.MethodOptions:
			g.OPTIONS(newPath, route.Handlers...)
		}
	}

	return g
}

// ResourceWithActions 注册一个资源控制器，并添加自定义操作。
// 参数：
//   - name: 资源名称，用于构建路由路径
//   - controller: 实现了 RESTController 接口的控制器
//   - actions: 自定义操作映射，键为操作名，值为处理函数
//   - middleware: 可选的中间件，应用于所有路由
func (g *ResourceGroup) ResourceWithActions(name string, controller RESTController, actions map[string]gin.HandlerFunc, middleware ...gin.HandlerFunc) *ResourceGroup {
	// 先注册标准 RESTful 路由
	resourceGroup := g.Resource(name, controller, middleware...)

	// 添加自定义操作
	for actionName, handler := range actions {
		// 对集合的操作 (如 /posts/export)
		if strings.HasPrefix(actionName, "collection:") {
			action := strings.TrimPrefix(actionName, "collection:")
			parts := strings.SplitN(action, ":", 2)
			method := strings.ToUpper(parts[0])
			actionPath := ""
			if len(parts) > 1 {
				actionPath = parts[1]
			}

			// 注册路由
			registerRouteWithMethod(resourceGroup.RouterGroup, method, actionPath, append(middleware, handler)...)
		} else {
			// 对资源实例的操作 (如 /posts/:id/publish)
			parts := strings.SplitN(actionName, ":", 2)
			method := strings.ToUpper(parts[0])
			actionPath := ""
			if len(parts) > 1 {
				actionPath = parts[1]
			}

			// 构建路径 /:id/action
			fullPath := path.Join("/:id", actionPath)

			// 注册路由
			registerRouteWithMethod(resourceGroup.RouterGroup, method, fullPath, append(middleware, handler)...)
		}
	}

	return resourceGroup
}

// RouteInfo 存储路由信息
type GroupRouteInfo struct {
	Method   string
	Path     string
	Handlers []gin.HandlerFunc
}

// extractRoutesFromGroup 从路由组中提取所有路由。
// 注意：这是一个实验性功能，可能不适用于所有 Gin 版本。
func extractRoutesFromGroup(group *gin.RouterGroup) []GroupRouteInfo {
	// 这个函数在实际应用中可能需要根据 Gin 的内部实现进行调整
	// 由于 Gin 没有提供直接访问路由组中已注册路由的公共 API，
	// 这里提供一个简化的实现，实际使用时可能需要修改

	// 在实际应用中，可以考虑在注册路由时同时记录路由信息
	return []GroupRouteInfo{}
}

// registerRouteWithMethod 根据 HTTP 方法注册路由。
func registerRouteWithMethod(group *gin.RouterGroup, method, path string, handlers ...gin.HandlerFunc) {
	switch method {
	case http.MethodGet:
		group.GET(path, handlers...)
	case http.MethodPost:
		group.POST(path, handlers...)
	case http.MethodPut:
		group.PUT(path, handlers...)
	case http.MethodDelete:
		group.DELETE(path, handlers...)
	case http.MethodPatch:
		group.PATCH(path, handlers...)
	case http.MethodHead:
		group.HEAD(path, handlers...)
	case http.MethodOptions:
		group.OPTIONS(path, handlers...)
	default:
		group.Any(path, handlers...)
	}
}

// NewResource 创建一个新的资源路由组。
func NewResource(engine *gin.Engine, name string) *ResourceGroup {
	return NewResourceGroup(engine.Group(name))
}

// APIGroup 是一个便捷的 API 路由组构建器。
type APIGroup struct {
	*ResourceGroup
	version string
}

// NewAPIGroup 创建一个新的 API 路由组。
// 参数：
//   - engine: Gin 引擎
//   - basePath: API 基础路径，如 "/api"
//   - version: API 版本，如 "v1"
func NewAPIGroup(engine *gin.Engine, basePath, version string) *APIGroup {
	// 构建完整路径，如 /api/v1
	fullPath := path.Join(basePath, version)

	return &APIGroup{
		ResourceGroup: NewResourceGroup(engine.Group(fullPath)),
		version:       version,
	}
}

// Version 返回 API 版本。
func (a *APIGroup) Version() string {
	return a.version
}

// ResourceWithPrefix 注册一个带前缀的资源控制器。
// 参数：
//   - prefix: 资源前缀，如 "admin"
//   - name: 资源名称，如 "users"
//   - controller: 实现了 RESTController 接口的控制器
func (a *APIGroup) ResourceWithPrefix(prefix, name string, controller RESTController, middleware ...gin.HandlerFunc) *ResourceGroup {
	// 构建路径，如 /admin/users
	fullPath := path.Join(prefix, name)

	return a.Resource(fullPath, controller, middleware...)
}

// Namespace 创建一个命名空间，用于组织相关的资源。
// 参数：
//   - name: 命名空间名称
//   - fn: 配置函数，用于在命名空间内注册资源
func (a *APIGroup) Namespace(name string, fn func(ns *ResourceGroup)) *APIGroup {
	// 创建命名空间路由组
	ns := a.Group(name)

	// 调用配置函数
	fn(ns)

	return a
}

// RouteBuilder 是一个流式 API 路由构建器。
type RouteBuilder struct {
	group    *gin.RouterGroup
	path     string
	handlers []gin.HandlerFunc
}

// NewRouteBuilder 创建一个新的路由构建器。
func NewRouteBuilder(group *gin.RouterGroup) *RouteBuilder {
	return &RouteBuilder{
		group:    group,
		handlers: []gin.HandlerFunc{},
	}
}

// Path 设置路由路径。
func (b *RouteBuilder) Path(path string) *RouteBuilder {
	b.path = path
	return b
}

// Use 添加中间件。
func (b *RouteBuilder) Use(handlers ...gin.HandlerFunc) *RouteBuilder {
	b.handlers = append(b.handlers, handlers...)
	return b
}

// GET 注册 GET 路由。
func (b *RouteBuilder) GET(handlers ...gin.HandlerFunc) *gin.RouterGroup {
	b.group.GET(b.path, append(b.handlers, handlers...)...)
	return b.group
}

// POST 注册 POST 路由。
func (b *RouteBuilder) POST(handlers ...gin.HandlerFunc) *gin.RouterGroup {
	b.group.POST(b.path, append(b.handlers, handlers...)...)
	return b.group
}

// PUT 注册 PUT 路由。
func (b *RouteBuilder) PUT(handlers ...gin.HandlerFunc) *gin.RouterGroup {
	b.group.PUT(b.path, append(b.handlers, handlers...)...)
	return b.group
}

// DELETE 注册 DELETE 路由。
func (b *RouteBuilder) DELETE(handlers ...gin.HandlerFunc) *gin.RouterGroup {
	b.group.DELETE(b.path, append(b.handlers, handlers...)...)
	return b.group
}

// PATCH 注册 PATCH 路由。
func (b *RouteBuilder) PATCH(handlers ...gin.HandlerFunc) *gin.RouterGroup {
	b.group.PATCH(b.path, append(b.handlers, handlers...)...)
	return b.group
}

// Route 是 gin.RouterGroup 的扩展，提供流式 API。
type Route struct {
	*gin.RouterGroup
}

// NewRoute 创建一个新的路由。
func NewRoute(group *gin.RouterGroup) *Route {
	return &Route{RouterGroup: group}
}

// At 创建一个路由构建器。
func (r *Route) At(path string) *RouteBuilder {
	return NewRouteBuilder(r.RouterGroup).Path(path)
}

// Resource 注册一个资源控制器。
func (r *Route) Resource(name string, controller RESTController, middleware ...gin.HandlerFunc) *ResourceGroup {
	return NewResourceGroup(r.RouterGroup).Resource(name, controller, middleware...)
}

// Group 创建一个新的路由组。
func (r *Route) Group(relativePath string, handlers ...gin.HandlerFunc) *Route {
	return NewRoute(r.RouterGroup.Group(relativePath, handlers...))
}

// NewRouter 创建一个新的路由。
func NewRouter(engine *gin.Engine) *Route {
	return NewRoute(engine.Group("/"))
}
