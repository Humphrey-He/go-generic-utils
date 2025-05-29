package binding

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// Decoder defines the interface for decoding data from gin.Context to an object
type Decoder interface {
	// Decode extracts data from gin.Context and populates the target object
	Decode(c *gin.Context, obj interface{}) error
}

// JSONDecoder decodes JSON data from request body
type JSONDecoder struct{}

// Decode implements the Decoder interface for JSON data
func (d JSONDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindJSON(obj)
}

// QueryDecoder decodes data from URL query parameters
type QueryDecoder struct{}

// Decode implements the Decoder interface for query parameters
func (d QueryDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindQuery(obj)
}

// URIDecoder decodes data from URI parameters
type URIDecoder struct{}

// Decode implements the Decoder interface for URI parameters
func (d URIDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindUri(obj)
}

// FormDecoder decodes data from form submissions
type FormDecoder struct{}

// Decode implements the Decoder interface for form data
func (d FormDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Form)
}

// MultipartFormDecoder decodes data from multipart form submissions
type MultipartFormDecoder struct {
	MaxMemory int64 // Maximum memory for multipart form parsing
}

// Decode implements the Decoder interface for multipart form data
func (d MultipartFormDecoder) Decode(c *gin.Context, obj interface{}) error {
	// Set default max memory if not specified
	maxMemory := d.MaxMemory
	if maxMemory <= 0 {
		maxMemory = 32 << 20 // 32 MB default
	}

	// Parse the multipart form
	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}

	return c.ShouldBindWith(obj, binding.FormMultipart)
}

// XMLDecoder decodes XML data from request body
type XMLDecoder struct{}

// Decode implements the Decoder interface for XML data
func (d XMLDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindXML(obj)
}

// HeaderDecoder decodes data from HTTP headers
type HeaderDecoder struct{}

// Decode implements the Decoder interface for HTTP headers
func (d HeaderDecoder) Decode(c *gin.Context, obj interface{}) error {
	return c.ShouldBindHeader(obj)
}

// concurrentDecoderRegistry is a concurrent-safe registry for decoders
type concurrentDecoderRegistry struct {
	decoders map[string]Decoder
	mu       sync.RWMutex
}

// Load retrieves a decoder from the registry
func (r *concurrentDecoderRegistry) Load(contentType string) (Decoder, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	decoder, ok := r.decoders[contentType]
	return decoder, ok
}

// Store adds or updates a decoder in the registry
func (r *concurrentDecoderRegistry) Store(contentType string, decoder Decoder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.decoders[contentType] = decoder
}

// Global decoder registry
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
func GetDecoder(contentType string) (Decoder, bool) {
	return decoderRegistry.Load(contentType)
}

// RegisterDecoder registers a decoder for a specific content type
func RegisterDecoder(contentType string, decoder Decoder) {
	decoderRegistry.Store(contentType, decoder)
}
