package binding

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Decoder defines the interface for decoding data from gin.Context to an object
// Decoder 定义了从 gin.Context 解码数据到对象的接口
type Decoder interface {
	// Decode extracts data from gin.Context and populates the target object
	// Decode 从 gin.Context 提取数据并填充目标对象
	Decode(c *gin.Context, obj interface{}) error
}

// JSONDecoder decodes JSON data from request body
// JSONDecoder 从请求体解码 JSON 数据
type JSONDecoder struct{}

// Decode implements the Decoder interface for JSON data
// Decode 实现了 JSON 数据的 Decoder 接口
func (d JSONDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

// QueryDecoder decodes data from URL query parameters
// QueryDecoder 从 URL 查询参数解码数据
type QueryDecoder struct{}

// Decode implements the Decoder interface for query parameters
// Decode 实现了查询参数的 Decoder 接口
func (d QueryDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
}

// URIDecoder decodes data from URI parameters
// URIDecoder 从 URI 参数解码数据
type URIDecoder struct{}

// Decode implements the Decoder interface for URI parameters
// Decode 实现了 URI 参数的 Decoder 接口
func (d URIDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindUri(obj)
}

// FormDecoder decodes data from form submissions
// FormDecoder 从表单提交解码数据
type FormDecoder struct{}

// Decode implements the Decoder interface for form data
// Decode 实现了表单数据的 Decoder 接口
func (d FormDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Form)
}

// MultipartFormDecoder decodes data from multipart form submissions
// MultipartFormDecoder 从多部分表单提交解码数据
type MultipartFormDecoder struct {
	MaxMemory int64 // Maximum memory for multipart form parsing 多部分表单解析的最大内存
}

// Decode implements the Decoder interface for multipart form data
// Decode 实现了多部分表单数据的 Decoder 接口
func (d MultipartFormDecoder) Decode(c *gin.Context, obj interface{}) error {
	// Set default max memory if not specified
	// 如果未指定，设置默认最大内存
	maxMemory := d.MaxMemory
	if maxMemory <= 0 {
		maxMemory = 32 << 20 // 32 MB default 默认 32 MB
	}

	// Parse the multipart form
	// 解析多部分表单
	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}

	return c.ShouldBindWith(obj, binding.FormMultipart)
}

// XMLDecoder decodes XML data from request body
// XMLDecoder 从请求体解码 XML 数据
type XMLDecoder struct{}

// Decode implements the Decoder interface for XML data
// Decode 实现了 XML 数据的 Decoder 接口
func (d XMLDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindXML(obj)
}

// HeaderDecoder decodes data from HTTP headers
// HeaderDecoder 从 HTTP 标头解码数据
type HeaderDecoder struct{}

// Decode implements the Decoder interface for HTTP headers
// Decode 实现了 HTTP 标头的 Decoder 接口
func (d HeaderDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindHeader(obj)
}

// concurrentDecoderRegistry is a concurrent-safe registry for decoders
// concurrentDecoderRegistry 是一个线程安全的解码器注册表
type concurrentDecoderRegistry struct {
	decoders map[string]Decoder
	mu       sync.RWMutex
}

// Load retrieves a decoder from the registry
// Load 从注册表中检索解码器
func (r *concurrentDecoderRegistry) Load(contentType string) (Decoder, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	decoder, ok := r.decoders[contentType]
	return decoder, ok
}

// Store adds or updates a decoder in the registry
// Store 在注册表中添加或更新解码器
func (r *concurrentDecoderRegistry) Store(contentType string, decoder Decoder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.decoders[contentType] = decoder
}

// Global decoder registry
// 全局解码器注册表
var decoderRegistry = &concurrentDecoderRegistry{
	decoders: map[string]Decoder{
		binding.MIMEJSON:              JSONDecoder{},
		binding.MIMEXML:               XMLDecoder{},
		binding.MIMEXML2:              XMLDecoder{},
		binding.MIMEPlain:             FormDecoder{},
		binding.MIMEPOSTForm:          FormDecoder{},
		binding.MIMEMultipartPOSTForm: MultipartFormDecoder{},
	},
}

// GetDecoder returns a decoder based on the content type
// GetDecoder 根据内容类型返回解码器
func GetDecoder(contentType string) (Decoder, bool) {
	return decoderRegistry.Load(contentType)
}

// RegisterDecoder registers a decoder for a specific content type
// RegisterDecoder 为特定内容类型注册解码器
func RegisterDecoder(contentType string, decoder Decoder) {
	decoderRegistry.Store(contentType, decoder)
}
