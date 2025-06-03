package binding

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

// 测试数据结构
// Test data structures
type TestDecoderData struct {
	Name  string `json:"name" form:"name" xml:"name"`
	Value int    `json:"value" form:"value" xml:"value"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

// 测试JSONDecoder
// Test JSONDecoder
func TestJSONDecoder(t *testing.T) {
	// 准备测试数据
	// Prepare test data
	jsonData := `{"name":"test", "value":123}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 创建decoder
	// Create decoder
	decoder := JSONDecoder{}
	var data TestDecoderData

	// 测试解码
	// Test decoding
	err := decoder.Decode(c, &data)
	assert.NoError(t, err, "JSON decoding should not error")
	assert.Equal(t, "test", data.Name, "Name should be decoded correctly")
	assert.Equal(t, 123, data.Value, "Value should be decoded correctly")

	// 测试无效JSON
	// Test invalid JSON
	invalidJSON := `{"name":"test", value:123}`
	req = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	err = decoder.Decode(c, &data)
	assert.Error(t, err, "Invalid JSON should cause error")
}

// 测试QueryDecoder
// Test QueryDecoder
func TestQueryDecoder(t *testing.T) {
	// 准备测试数据
	// Prepare test data
	req := httptest.NewRequest("GET", "/test?name=test&value=123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 创建decoder
	// Create decoder
	decoder := QueryDecoder{}
	var data TestDecoderData

	// 测试解码
	// Test decoding
	err := decoder.Decode(c, &data)
	assert.NoError(t, err, "Query decoding should not error")
	assert.Equal(t, "test", data.Name, "Name should be decoded correctly")
	assert.Equal(t, 123, data.Value, "Value should be decoded correctly")
}

// 测试URIDecoder
// Test URIDecoder
func TestURIDecoder(t *testing.T) {
	// 创建路由器
	// Create router
	router := gin.New()
	// 使用自定义结构体，确保URI参数能正确绑定
	// Use custom struct to ensure URI parameters can be bound correctly
	type URIParams struct {
		Name  string `uri:"name"`
		Value int    `uri:"value"`
	}

	router.GET("/test/:name/:value", func(c *gin.Context) {
		var params URIParams
		decoder := URIDecoder{}
		err := decoder.Decode(c, &params)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, params)
	})

	// 发送请求
	// Send request
	req := httptest.NewRequest("GET", "/test/testname/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查结果
	// Check result
	assert.Equal(t, http.StatusOK, w.Code, "Request should succeed")
	assert.Contains(t, w.Body.String(), `"Name":"testname"`, "Name should be in response")
	assert.Contains(t, w.Body.String(), `"Value":123`, "Value should be in response")
}

// 测试FormDecoder
// Test FormDecoder
func TestFormDecoder(t *testing.T) {
	// 准备测试数据
	// Prepare test data
	formData := "name=test&value=123"
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 创建decoder
	// Create decoder
	decoder := FormDecoder{}
	var data TestDecoderData

	// 测试解码
	// Test decoding
	err := decoder.Decode(c, &data)
	assert.NoError(t, err, "Form decoding should not error")
	assert.Equal(t, "test", data.Name, "Name should be decoded correctly")
	assert.Equal(t, 123, data.Value, "Value should be decoded correctly")
}

// 测试XMLDecoder
// Test XMLDecoder
func TestXMLDecoder(t *testing.T) {
	// 准备测试数据
	// Prepare test data
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<TestDecoderData>
	<name>test</name>
	<value>123</value>
</TestDecoderData>`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(xmlData))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 创建decoder
	// Create decoder
	decoder := XMLDecoder{}
	var data TestDecoderData

	// 测试解码
	// Test decoding
	err := decoder.Decode(c, &data)
	assert.NoError(t, err, "XML decoding should not error")
	assert.Equal(t, "test", data.Name, "Name should be decoded correctly")
	assert.Equal(t, 123, data.Value, "Value should be decoded correctly")

	// 测试无效XML
	// Test invalid XML
	invalidXML := `<TestDecoderData><name>test</name><value>invalid</value></TestDecoderData>`
	req = httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidXML))
	req.Header.Set("Content-Type", "application/xml")
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = req

	err = decoder.Decode(c, &data)
	assert.Error(t, err, "Invalid XML should cause error")
}

// 测试HeaderDecoder
// Test HeaderDecoder
func TestHeaderDecoder(t *testing.T) {
	// 准备测试数据
	// Prepare test data
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Name", "test")
	req.Header.Set("X-Value", "123")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 创建header binding结构体
	// Create header binding struct
	type HeaderData struct {
		Name  string `header:"X-Name"`
		Value int    `header:"X-Value"`
	}

	// 创建decoder
	// Create decoder
	decoder := HeaderDecoder{}
	var data HeaderData

	// 测试解码
	// Test decoding
	err := decoder.Decode(c, &data)
	assert.NoError(t, err, "Header decoding should not error")
	assert.Equal(t, "test", data.Name, "Name should be decoded correctly")
	assert.Equal(t, 123, data.Value, "Value should be decoded correctly")
}

// 测试MultipartFormDecoder
// Test MultipartFormDecoder
func TestMultipartFormDecoder(t *testing.T) {
	// 准备测试数据 - 多部分表单数据需要特殊处理
	// 这里我们模拟Gin的测试方式
	// Prepare test data - multipart form requires special handling
	// Here we simulate Gin's testing approach
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 创建一个multipart/form-data请求
	// Create a multipart/form-data request
	c.Request = &http.Request{
		Method: "POST",
		Header: http.Header{},
	}
	c.Request.Header.Set("Content-Type", binding.MIMEMultipartPOSTForm)

	// 设置表单值
	// Set form values
	form := map[string][]string{
		"name":  {"test"},
		"value": {"123"},
	}
	c.Request.MultipartForm = &multipart.Form{
		Value: form,
	}

	// 创建decoder
	// Create decoder
	decoder := MultipartFormDecoder{}
	var data TestDecoderData

	// 测试解码
	// Test decoding
	err := decoder.Decode(c, &data)
	assert.NoError(t, err, "Multipart form decoding should not error")
	assert.Equal(t, "test", data.Name, "Name should be decoded correctly")
	assert.Equal(t, 123, data.Value, "Value should be decoded correctly")
}

// 测试GetDecoder函数
// Test GetDecoder function
func TestGetDecoder(t *testing.T) {
	// 测试获取已注册的解码器
	// Test getting registered decoders
	decoder, ok := GetDecoder(binding.MIMEJSON)
	assert.True(t, ok, "Should find JSON decoder")
	assert.IsType(t, JSONDecoder{}, decoder, "Should return correct decoder type")

	decoder, ok = GetDecoder(binding.MIMEXML)
	assert.True(t, ok, "Should find XML decoder")
	assert.IsType(t, XMLDecoder{}, decoder, "Should return correct decoder type")

	decoder, ok = GetDecoder(binding.MIMEPlain)
	assert.True(t, ok, "Should find Plain decoder")
	assert.IsType(t, FormDecoder{}, decoder, "Should return correct decoder type")

	// 测试获取未注册的解码器
	// Test getting unregistered decoder
	decoder, ok = GetDecoder("application/unknown")
	assert.False(t, ok, "Should not find unknown decoder")
	assert.Nil(t, decoder, "Should return nil for unknown decoder")
}

// 测试RegisterDecoder函数
// Test RegisterDecoder function
func TestRegisterDecoder(t *testing.T) {
	// 创建自定义解码器
	// Create custom decoder
	customDecoder := &testCustomDecoder{}

	// 注册自定义解码器
	// Register custom decoder
	customContentType := "application/custom"
	RegisterDecoder(customContentType, customDecoder)

	// 测试获取已注册的自定义解码器
	// Test getting registered custom decoder
	decoder, ok := GetDecoder(customContentType)
	assert.True(t, ok, "Should find custom decoder")
	assert.IsType(t, &testCustomDecoder{}, decoder, "Should return correct custom decoder type")
}

// 自定义解码器用于测试
// Custom decoder for testing
type testCustomDecoder struct{}

func (d *testCustomDecoder) Decode(c *gin.Context, obj interface{}) error {
	return nil
}
