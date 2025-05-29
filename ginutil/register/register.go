// Package register 提供 Gin 框架的路由自动注册功能。
// 允许将不同业务模块的路由定义分散到各自的包中，并能被自动发现和注册。
package register

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// 全局路由注册器实例
var (
	// 全局注册器，用于存储所有已注册的控制器
	globalRegistry = &Registry{
		controllers: make(map[string]interface{}),
	}
)

// Registry 是路由注册器，用于管理控制器的注册和路由映射。
type Registry struct {
	mu          sync.RWMutex
	controllers map[string]interface{}
}

// GetGlobalRegistry 返回全局注册器实例。
func GetGlobalRegistry() *Registry {
	return globalRegistry
}

// NewRegistry 创建一个新的注册器实例。
func NewRegistry() *Registry {
	return &Registry{
		controllers: make(map[string]interface{}),
	}
}

// Register 注册一个控制器到注册器。
// name 参数是控制器的唯一标识符，通常是控制器的名称。
// controller 参数是控制器实例，可以是任何类型。
func (r *Registry) Register(name string, controller interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.controllers[name]; exists {
		panic(fmt.Sprintf("控制器 %s 已经注册", name))
	}

	r.controllers[name] = controller
}

// RegisterController 是 Register 的别名，用于语义清晰。
func (r *Registry) RegisterController(name string, controller interface{}) {
	r.Register(name, controller)
}

// Get 获取已注册的控制器。
func (r *Registry) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	controller, exists := r.controllers[name]
	return controller, exists
}

// GetAll 获取所有已注册的控制器。
func (r *Registry) GetAll() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 创建一个副本，避免外部修改
	result := make(map[string]interface{}, len(r.controllers))
	for name, controller := range r.controllers {
		result[name] = controller
	}

	return result
}

// RegisterGlobal 注册一个控制器到全局注册器。
func RegisterGlobal(name string, controller interface{}) {
	globalRegistry.Register(name, controller)
}

// GetGlobal 从全局注册器获取已注册的控制器。
func GetGlobal(name string) (interface{}, bool) {
	return globalRegistry.Get(name)
}

// GetAllGlobal 获取全局注册器中的所有控制器。
func GetAllGlobal() map[string]interface{} {
	return globalRegistry.GetAll()
}

// Routable 是控制器需要实现的接口，用于自动注册路由。
type Routable interface {
	// RegisterRoutes 注册路由到给定的路由组。
	RegisterRoutes(group *gin.RouterGroup)
}

// RegisterRoutes 将注册器中的所有实现了 Routable 接口的控制器注册到给定的路由组。
func (r *Registry) RegisterRoutes(router *gin.Engine) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, controller := range r.controllers {
		if routable, ok := controller.(Routable); ok {
			// 使用控制器名称作为路由前缀
			group := router.Group("/" + strings.ToLower(name))
			routable.RegisterRoutes(group)
		}
	}
}

// RegisterRoutesWithGroup 将注册器中的所有实现了 Routable 接口的控制器注册到给定的路由组。
func (r *Registry) RegisterRoutesWithGroup(group *gin.RouterGroup) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, controller := range r.controllers {
		if routable, ok := controller.(Routable); ok {
			// 使用控制器名称作为路由前缀
			controllerGroup := group.Group("/" + strings.ToLower(name))
			routable.RegisterRoutes(controllerGroup)
		}
	}
}

// RegisterGlobalRoutes 将全局注册器中的所有实现了 Routable 接口的控制器注册到给定的路由引擎。
func RegisterGlobalRoutes(router *gin.Engine) {
	globalRegistry.RegisterRoutes(router)
}

// RegisterGlobalRoutesWithGroup 将全局注册器中的所有实现了 Routable 接口的控制器注册到给定的路由组。
func RegisterGlobalRoutesWithGroup(group *gin.RouterGroup) {
	globalRegistry.RegisterRoutesWithGroup(group)
}

// RouteTag 表示路由标签的配置。
type RouteTag struct {
	Method  string // HTTP 方法：GET, POST, PUT, DELETE 等
	Path    string // 路由路径
	Summary string // 路由摘要，用于文档生成
	Desc    string // 路由描述，用于文档生成
}

// parseRouteTag 解析路由标签。
// 标签格式：`route:"GET /path"`
// 或者：`route:"GET /path [summary] [description]"`
func parseRouteTag(tag string) (*RouteTag, error) {
	if tag == "" {
		return nil, fmt.Errorf("空标签")
	}

	// 匹配方法和路径
	re := regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|ANY)\s+(/[^\s\[]*)\s*(?:\[(.*?)\])?\s*(?:\((.*?)\))?$`)
	matches := re.FindStringSubmatch(tag)
	if matches == nil || len(matches) < 3 {
		return nil, fmt.Errorf("无效的路由标签格式: %s", tag)
	}

	route := &RouteTag{
		Method: matches[1],
		Path:   matches[2],
	}

	// 如果有摘要
	if len(matches) > 3 && matches[3] != "" {
		route.Summary = matches[3]
	}

	// 如果有描述
	if len(matches) > 4 && matches[4] != "" {
		route.Desc = matches[4]
	}

	return route, nil
}

// RouteInfo 存储路由方法信息
type RouteInfo struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
	Summary string
	Desc    string
}

// RegisterControllerWithTags 使用结构体标签注册控制器的路由。
// 控制器方法需要有 `route:"METHOD /path"` 标签。
func (r *Registry) RegisterControllerWithTags(router *gin.Engine, controller interface{}) error {
	return r.registerControllerWithTags(router, nil, controller)
}

// RegisterControllerWithTagsToGroup 使用结构体标签将控制器的路由注册到指定的路由组。
func (r *Registry) RegisterControllerWithTagsToGroup(group *gin.RouterGroup, controller interface{}) error {
	return r.registerControllerWithTags(nil, group, controller)
}

// registerControllerWithTags 是实际的注册函数，处理 router 和 group 两种情况。
func (r *Registry) registerControllerWithTags(router *gin.Engine, group *gin.RouterGroup, controller interface{}) error {
	controllerType := reflect.TypeOf(controller)
	controllerValue := reflect.ValueOf(controller)

	if controllerType.Kind() != reflect.Ptr || controllerType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("控制器必须是结构体指针")
	}

	// 获取控制器结构体类型
	structType := controllerType.Elem()

	// 获取结构体中的路由配置字段
	routes, err := extractRoutesFromStruct(structType)
	if err != nil {
		return err
	}

	// 遍历控制器的所有方法
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)

		// 检查方法是否在路由配置中
		routeInfo, ok := routes[method.Name]
		if !ok {
			continue
		}

		// 检查方法签名是否正确
		if !isValidHandlerMethod(method) {
			return fmt.Errorf("方法 %s 的签名不正确，应为 func(*gin.Context)", method.Name)
		}

		// 创建处理函数
		handler := createHandlerFunc(controllerValue, method.Index)
		routeInfo.Handler = handler

		// 注册路由
		if router != nil {
			registerRoute(router, routeInfo.Method, routeInfo.Path, handler)
		} else if group != nil {
			registerRouteToGroup(group, routeInfo.Method, routeInfo.Path, handler)
		} else {
			return fmt.Errorf("router 和 group 不能同时为 nil")
		}
	}

	return nil
}

// extractRoutesFromStruct 从结构体中提取路由配置。
// 路由配置可以通过结构体字段的标签定义，例如：
//
//	type UserController struct {
//	    GetUser    RouteInfo `route:"GET /users/:id [获取用户] (获取指定ID的用户信息)"`
//	    CreateUser RouteInfo `route:"POST /users [创建用户] (创建新用户)"`
//	}
func extractRoutesFromStruct(structType reflect.Type) (map[string]RouteInfo, error) {
	routes := make(map[string]RouteInfo)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// 获取字段的 route 标签
		routeTag := field.Tag.Get("route")
		if routeTag == "" {
			continue
		}

		// 解析路由标签
		route, err := parseRouteTag(routeTag)
		if err != nil {
			return nil, fmt.Errorf("解析字段 %s 的路由标签错误: %v", field.Name, err)
		}

		// 字段名应该对应于方法名
		methodName := field.Name

		// 存储路由信息
		routes[methodName] = RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Summary: route.Summary,
			Desc:    route.Desc,
		}
	}

	return routes, nil
}

// isValidHandlerMethod 检查方法签名是否为 func(*gin.Context)。
func isValidHandlerMethod(method reflect.Method) bool {
	methodType := method.Type

	// 方法应该有一个接收者和一个参数
	if methodType.NumIn() != 2 {
		return false
	}

	// 参数应该是 *gin.Context
	paramType := methodType.In(1)
	return paramType.String() == "*gin.Context"
}

// createHandlerFunc 创建一个 gin.HandlerFunc，调用控制器的指定方法。
func createHandlerFunc(controllerValue reflect.Value, methodIndex int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 调用控制器方法
		controllerValue.Method(methodIndex).Call([]reflect.Value{reflect.ValueOf(c)})
	}
}

// registerRoute 注册路由到 gin.Engine。
func registerRoute(router *gin.Engine, method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		router.GET(path, handler)
	case "POST":
		router.POST(path, handler)
	case "PUT":
		router.PUT(path, handler)
	case "DELETE":
		router.DELETE(path, handler)
	case "PATCH":
		router.PATCH(path, handler)
	case "HEAD":
		router.HEAD(path, handler)
	case "OPTIONS":
		router.OPTIONS(path, handler)
	case "ANY":
		router.Any(path, handler)
	}
}

// registerRouteToGroup 注册路由到 gin.RouterGroup。
func registerRouteToGroup(group *gin.RouterGroup, method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		group.GET(path, handler)
	case "POST":
		group.POST(path, handler)
	case "PUT":
		group.PUT(path, handler)
	case "DELETE":
		group.DELETE(path, handler)
	case "PATCH":
		group.PATCH(path, handler)
	case "HEAD":
		group.HEAD(path, handler)
	case "OPTIONS":
		group.OPTIONS(path, handler)
	case "ANY":
		group.Any(path, handler)
	}
}

// RegisterGlobalControllerWithTags 使用结构体标签将控制器注册到全局路由引擎。
func RegisterGlobalControllerWithTags(router *gin.Engine, controller interface{}) error {
	return globalRegistry.RegisterControllerWithTags(router, controller)
}

// RegisterGlobalControllerWithTagsToGroup 使用结构体标签将控制器注册到全局路由组。
func RegisterGlobalControllerWithTagsToGroup(group *gin.RouterGroup, controller interface{}) error {
	return globalRegistry.RegisterControllerWithTagsToGroup(group, controller)
}

// ControllerBase 是一个基础控制器结构体，提供通用功能。
// 其他控制器可以嵌入此结构体以获得这些功能。
type ControllerBase struct {
	// 控制器名称，用于日志和调试
	Name string
}

// SetName 设置控制器名称。
func (c *ControllerBase) SetName(name string) {
	c.Name = name
}

// GetName 获取控制器名称。
func (c *ControllerBase) GetName() string {
	return c.Name
}

// RESTController 是一个实现了 RESTful API 的控制器接口。
type RESTController interface {
	// Index 返回资源列表
	Index(c *gin.Context)
	// Show 返回单个资源
	Show(c *gin.Context)
	// Create 创建资源
	Create(c *gin.Context)
	// Update 更新资源
	Update(c *gin.Context)
	// Delete 删除资源
	Delete(c *gin.Context)
}

// RegisterRESTRoutes 为实现了 RESTController 接口的控制器注册 RESTful 路由。
func RegisterRESTRoutes(group *gin.RouterGroup, controller RESTController) {
	group.GET("", controller.Index)
	group.GET("/:id", controller.Show)
	group.POST("", controller.Create)
	group.PUT("/:id", controller.Update)
	group.DELETE("/:id", controller.Delete)
}

// RegisterRESTRoutesWithMiddleware 为实现了 RESTController 接口的控制器注册带中间件的 RESTful 路由。
func RegisterRESTRoutesWithMiddleware(group *gin.RouterGroup, controller RESTController, middleware ...gin.HandlerFunc) {
	group.GET("", append(middleware, controller.Index)...)
	group.GET("/:id", append(middleware, controller.Show)...)
	group.POST("", append(middleware, controller.Create)...)
	group.PUT("/:id", append(middleware, controller.Update)...)
	group.DELETE("/:id", append(middleware, controller.Delete)...)
}

// AutoRegisterControllers 自动注册所有实现了 Routable 接口的控制器。
// 这个函数应该在应用启动时调用。
func (r *Registry) AutoRegisterControllers(router *gin.Engine) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, controller := range r.controllers {
		if routable, ok := controller.(Routable); ok {
			// 使用控制器的名称作为路由前缀
			var name string
			if named, ok := controller.(interface{ GetName() string }); ok {
				name = named.GetName()
			} else {
				// 使用类型名作为默认名称
				name = reflect.TypeOf(controller).Elem().Name()
			}

			// 将名称转换为小写并移除 "Controller" 后缀
			name = strings.ToLower(name)
			name = strings.TrimSuffix(name, "controller")

			// 创建路由组并注册路由
			group := router.Group("/" + name)
			routable.RegisterRoutes(group)
		}
	}
}

// AutoRegisterGlobalControllers 自动注册全局注册器中的所有控制器。
func AutoRegisterGlobalControllers(router *gin.Engine) {
	globalRegistry.AutoRegisterControllers(router)
}
