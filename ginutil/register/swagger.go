package register

/*
Swagger UI 集成说明：

本文件提供了将 Swagger UI 集成到 Gin 应用的功能。

问题描述：
原代码使用 go:embed 指令嵌入 swaggerui 目录中的静态文件，但实际项目中并不存在这个目录，
导致编译错误："pattern swaggerui: no matching files found"。

修复思路：
1. 移除 go:embed 指令，改为优先从本地文件系统加载 Swagger UI 文件
2. 如果本地文件不存在，则使用 CDN 方式加载 Swagger UI
3. 保持原有的 API 兼容性，使现有代码不需要修改

使用方法：
1. 若希望使用本地文件，只需在项目根目录创建 swaggerui 目录并添加必要文件
2. 若不想维护本地文件，系统会自动使用 CDN 方式加载 Swagger UI

注意事项：
1. swagger.json 文件仍然需要在本地存在，路径由 SwaggerConfig.JSONFile 指定
2. 使用 CDN 方式可能会受到网络环境的影响，如果需要离线使用，建议提供本地文件
*/

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// SwaggerConfig 是 Swagger UI 的配置选项。
type SwaggerConfig struct {
	// BasePath 是 Swagger UI 的基础路径，默认为 "/swagger"。
	BasePath string

	// JSONPath 是 swagger.json 文件的路径，默认为 "/swagger/doc.json"。
	JSONPath string

	// JSONFile 是 swagger.json 文件的实际文件系统路径，默认为 "./docs/swagger.json"。
	JSONFile string

	// Title 是 Swagger UI 的标题。
	Title string

	// Description 是 Swagger UI 的描述。
	Description string

	// Version 是 API 的版本。
	Version string

	// DeepLinking 是否启用 Swagger UI 的深度链接功能。
	DeepLinking bool

	// PersistAuthorization 是否持久化授权信息。
	PersistAuthorization bool

	// OAuth2RedirectUrl 是 OAuth2 重定向 URL。
	OAuth2RedirectUrl string

	// DocExpansion 设置文档的默认展开级别，可选值：
	// "list" - 展开标签
	// "full" - 展开标签和操作
	// "none" - 不展开任何内容
	DocExpansion string
}

// DefaultSwaggerConfig 返回默认的 Swagger 配置。
func DefaultSwaggerConfig() *SwaggerConfig {
	return &SwaggerConfig{
		BasePath:             "/swagger",
		JSONPath:             "/swagger/doc.json",
		JSONFile:             "./docs/swagger.json",
		Title:                "API Documentation",
		Description:          "API Documentation powered by Swagger UI",
		Version:              "1.0.0",
		DeepLinking:          true,
		PersistAuthorization: true,
		DocExpansion:         "list",
	}
}

// swaggerUIFS 已移除嵌入文件系统，改为使用外部文件或 CDN
var swaggerUIFS embed.FS

// RegisterSwaggerUI 注册 Swagger UI 路由。
// 默认从本地文件系统加载 Swagger UI 文件。
// 如果本地文件不存在，将使用 CDN。
func RegisterSwaggerUI(router *gin.Engine, config *SwaggerConfig) error {
	if config == nil {
		config = DefaultSwaggerConfig()
	}

	// 注册 swagger.json 文件
	router.GET(config.JSONPath, func(c *gin.Context) {
		// 检查文件是否存在
		if _, err := os.Stat(config.JSONFile); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("Swagger JSON file not found: %s", config.JSONFile),
			})
			return
		}

		// 读取 swagger.json 文件
		c.File(config.JSONFile)
	})

	// 检查本地 Swagger UI 文件是否存在
	swaggerUIPath := "./swaggerui"
	useCDN := false

	if _, err := os.Stat(swaggerUIPath); os.IsNotExist(err) {
		// 本地文件不存在，使用 CDN
		useCDN = true
	}

	if useCDN {
		// 注册使用 CDN 的 Swagger UI 首页
		router.GET(config.BasePath, func(c *gin.Context) {
			html := generateSwaggerUIHtml(config)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, html)
		})

		// 在使用 CDN 时，我们还需要注册 index.html 路径
		router.GET(path.Join(config.BasePath, "index.html"), func(c *gin.Context) {
			html := generateSwaggerUIHtml(config)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, html)
		})
	} else {
		// 使用本地文件系统
		router.Static(config.BasePath, swaggerUIPath)

		// 注册 Swagger UI 首页
		router.GET(config.BasePath, func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, path.Join(config.BasePath, "index.html"))
		})
	}

	// 注册 Swagger UI 配置
	router.GET(path.Join(config.BasePath, "swagger-config.json"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"url":                  config.JSONPath,
			"title":                config.Title,
			"description":          config.Description,
			"version":              config.Version,
			"deepLinking":          config.DeepLinking,
			"persistAuthorization": config.PersistAuthorization,
			"oauth2RedirectUrl":    config.OAuth2RedirectUrl,
			"docExpansion":         config.DocExpansion,
		})
	})

	return nil
}

// SwaggerUIHandler 返回一个处理 Swagger UI 请求的处理函数。
func SwaggerUIHandler(config *SwaggerConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSwaggerConfig()
	}

	// 检查本地 Swagger UI 文件是否存在
	swaggerUIPath := "./swaggerui"
	useCDN := false

	if _, err := os.Stat(swaggerUIPath); os.IsNotExist(err) {
		// 本地文件不存在，使用 CDN
		useCDN = true
	}

	return func(c *gin.Context) {
		// 获取请求路径
		requestPath := c.Request.URL.Path
		basePath := config.BasePath

		// 处理 swagger-config.json 请求
		if requestPath == path.Join(basePath, "swagger-config.json") {
			c.JSON(http.StatusOK, gin.H{
				"url":                  config.JSONPath,
				"title":                config.Title,
				"description":          config.Description,
				"version":              config.Version,
				"deepLinking":          config.DeepLinking,
				"persistAuthorization": config.PersistAuthorization,
				"oauth2RedirectUrl":    config.OAuth2RedirectUrl,
				"docExpansion":         config.DocExpansion,
			})
			return
		}

		// 处理 swagger.json 请求
		if requestPath == config.JSONPath {
			// 检查文件是否存在
			if _, err := os.Stat(config.JSONFile); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("Swagger JSON file not found: %s", config.JSONFile),
				})
				return
			}

			// 读取 swagger.json 文件
			c.File(config.JSONFile)
			return
		}

		// 如果是根路径或 index.html
		if requestPath == basePath || requestPath == basePath+"/" || requestPath == path.Join(basePath, "index.html") {
			if useCDN {
				// 使用 CDN 生成 HTML
				html := generateSwaggerUIHtml(config)
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusOK, html)
				return
			} else {
				// 使用本地文件
				if requestPath != path.Join(basePath, "index.html") {
					c.Redirect(http.StatusMovedPermanently, path.Join(basePath, "index.html"))
					return
				}
				c.File(filepath.Join(swaggerUIPath, "index.html"))
				return
			}
		}

		// 如果使用 CDN，那么除了上面处理的特殊路径外，其他路径都返回 404
		if useCDN {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("File not found: %s", requestPath),
			})
			return
		}

		// 处理静态文件请求
		filePath := strings.TrimPrefix(requestPath, basePath)
		if filePath == "" || filePath == "/" {
			filePath = "index.html"
		}
		filePath = strings.TrimPrefix(filePath, "/")

		// 从本地文件系统读取
		localFilePath := filepath.Join(swaggerUIPath, filePath)
		if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("File not found: %s", filePath),
			})
			return
		}

		c.File(localFilePath)
	}
}

// getContentType 根据文件扩展名返回内容类型。
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

// SwaggerUIConfig 是 Swagger UI 的配置选项。
type SwaggerUIConfig struct {
	// URL 是 swagger.json 文件的 URL。
	URL string `json:"url"`

	// Title 是 Swagger UI 的标题。
	Title string `json:"title,omitempty"`

	// Description 是 Swagger UI 的描述。
	Description string `json:"description,omitempty"`

	// Version 是 API 的版本。
	Version string `json:"version,omitempty"`

	// DeepLinking 是否启用 Swagger UI 的深度链接功能。
	DeepLinking bool `json:"deepLinking,omitempty"`

	// PersistAuthorization 是否持久化授权信息。
	PersistAuthorization bool `json:"persistAuthorization,omitempty"`

	// OAuth2RedirectUrl 是 OAuth2 重定向 URL。
	OAuth2RedirectUrl string `json:"oauth2RedirectUrl,omitempty"`

	// DocExpansion 设置文档的默认展开级别。
	DocExpansion string `json:"docExpansion,omitempty"`
}

// SwaggerDocGenerator 是 Swagger 文档生成器的接口。
type SwaggerDocGenerator interface {
	// GenerateSwaggerJSON 生成 swagger.json 文件。
	GenerateSwaggerJSON(outputPath string) error

	// GetSwaggerJSON 返回 swagger.json 内容。
	GetSwaggerJSON() ([]byte, error)
}

// SimpleSwaggerGenerator 是一个简单的 Swagger 文档生成器实现。
type SimpleSwaggerGenerator struct {
	// Title 是 API 的标题。
	Title string

	// Description 是 API 的描述。
	Description string

	// Version 是 API 的版本。
	Version string

	// BasePath 是 API 的基础路径。
	BasePath string

	// Host 是 API 的主机名。
	Host string

	// Schemes 是 API 支持的协议。
	Schemes []string

	// Tags 是 API 的标签。
	Tags []map[string]interface{}

	// Paths 是 API 的路径。
	Paths map[string]interface{}

	// Definitions 是 API 的定义。
	Definitions map[string]interface{}
}

// NewSimpleSwaggerGenerator 创建一个新的简单 Swagger 文档生成器。
func NewSimpleSwaggerGenerator() *SimpleSwaggerGenerator {
	return &SimpleSwaggerGenerator{
		Title:       "API Documentation",
		Description: "API Documentation",
		Version:     "1.0.0",
		BasePath:    "/",
		Host:        "localhost",
		Schemes:     []string{"http"},
		Tags:        []map[string]interface{}{},
		Paths:       map[string]interface{}{},
		Definitions: map[string]interface{}{},
	}
}

// GenerateSwaggerJSON 生成 swagger.json 文件。
func (g *SimpleSwaggerGenerator) GenerateSwaggerJSON(outputPath string) error {
	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 获取 swagger.json 内容
	content, err := g.GetSwaggerJSON()
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(outputPath, content, 0644)
}

// GetSwaggerJSON 返回 swagger.json 内容。
func (g *SimpleSwaggerGenerator) GetSwaggerJSON() ([]byte, error) {
	// 构建 Swagger 规范
	swagger := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       g.Title,
			"description": g.Description,
			"version":     g.Version,
		},
		"host":     g.Host,
		"basePath": g.BasePath,
		"schemes":  g.Schemes,
		"tags":     g.Tags,
		"paths":    g.Paths,
	}

	if len(g.Definitions) > 0 {
		swagger["definitions"] = g.Definitions
	}

	// 将 Swagger 规范转换为 JSON
	return json.Marshal(swagger)
}

// AddSwaggerEndpoints 是 RegisterSwaggerUI 的简化版本，使用默认配置。
func AddSwaggerEndpoints(router *gin.Engine, jsonFilePath string) error {
	config := DefaultSwaggerConfig()
	config.JSONFile = jsonFilePath
	return RegisterSwaggerUI(router, config)
}

// RegisterSwaggerWithConfig 注册 Swagger UI 路由，使用自定义配置。
func RegisterSwaggerWithConfig(router *gin.Engine, config *SwaggerConfig) error {
	return RegisterSwaggerUI(router, config)
}

// generateSwaggerUIHtml 生成使用 CDN 的 Swagger UI HTML 页面。
func generateSwaggerUIHtml(config *SwaggerConfig) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + config.Title + `</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.1.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
        .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.1.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.1.0/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
        window.ui = SwaggerUIBundle({
            url: "` + config.JSONPath + `",
            dom_id: '#swagger-ui',
            deepLinking: ` + fmt.Sprintf("%t", config.DeepLinking) + `,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            layout: "StandaloneLayout",
            docExpansion: "` + config.DocExpansion + `",
            persistAuthorization: ` + fmt.Sprintf("%t", config.PersistAuthorization) + `
        });
    };
    </script>
</body>
</html>`
}
