package render

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是响应构建器，用于链式调用构建响应
type Response struct {
	ctx         *gin.Context
	code        int
	message     string
	data        interface{}
	httpStatus  int
	contentType string
}

// NewResponse 创建一个新的响应构建器
func NewResponse(c *gin.Context) *Response {
	return &Response{
		ctx:         c,
		code:        CodeSuccess,
		message:     DefaultSuccessMessage,
		data:        gin.H{},
		httpStatus:  200,
		contentType: "application/json",
	}
}

// Code 设置业务码
func (r *Response) Code(code int) *Response {
	r.code = code
	return r
}

// Message 设置消息
func (r *Response) Message(message string) *Response {
	r.message = message
	return r
}

// Data 设置数据
func (r *Response) Data(data interface{}) *Response {
	r.data = data
	return r
}

// Status 设置 HTTP 状态码
func (r *Response) Status(status int) *Response {
	r.httpStatus = status
	return r
}

// ContentType 设置内容类型
func (r *Response) ContentType(contentType string) *Response {
	r.contentType = contentType
	return r
}

// JSON 发送 JSON 响应
func (r *Response) JSON() {
	Custom(r.ctx, r.code, r.message, r.data, r.httpStatus)
}

// XML 发送 XML 响应
func (r *Response) XML() {
	CustomXML(r.ctx, r.code, r.message, r.data, r.httpStatus)
}

// HTML 发送 HTML 响应
func (r *Response) HTML(templateName string) {
	data, ok := r.data.(gin.H)
	if !ok {
		data = gin.H{"data": r.data}
	}

	// 添加状态信息
	data["code"] = r.code
	data["message"] = r.message
	data["success"] = r.code == CodeSuccess

	HTML(r.ctx, r.httpStatus, templateName, data)
}

// Err 发送错误响应
func (r *Response) Err() {
	Error(r.ctx, r.code, r.message, r.httpStatus)
}

// Success 发送成功响应
func (r *Response) Success() {
	r.code = CodeSuccess
	if r.message == "" {
		r.message = DefaultSuccessMessage
	}
	r.JSON()
}

// Abort 终止请求处理并发送响应
func (r *Response) Abort() {
	r.ctx.Abort()
	r.JSON()
}

// File 发送文件响应
func (r *Response) File(filepath string) {
	r.ctx.File(filepath)
}

// FileAttachment 发送文件附件响应
func (r *Response) FileAttachment(filepath, filename string) {
	r.ctx.FileAttachment(filepath, filename)
}

// Stream 发送流响应
func (r *Response) Stream(step func(w io.Writer) bool) {
	r.ctx.Stream(step)
}

// Download 文件下载辅助函数
func Download(c *gin.Context, filepath, filename string) {
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.File(filepath)
}

// FileTypeMap 文件类型映射
var FileTypeMap = map[string]string{
	".html": "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".mp4":  "video/mp4",
	".mp3":  "audio/mpeg",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".zip":  "application/zip",
	".tar":  "application/x-tar",
	".gz":   "application/gzip",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".xml":  "application/xml",
}

// GetContentTypeByExt 根据文件扩展名获取内容类型
func GetContentTypeByExt(ext string) string {
	if contentType, ok := FileTypeMap[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}

// WithJSON 在上下文中设置 JSON 格式的响应
func WithJSON(c *gin.Context) *Response {
	return NewResponse(c).ContentType("application/json")
}

// WithXML 在上下文中设置 XML 格式的响应
func WithXML(c *gin.Context) *Response {
	return NewResponse(c).ContentType("application/xml")
}

// WithHTML 在上下文中设置 HTML 格式的响应
func WithHTML(c *gin.Context) *Response {
	return NewResponse(c).ContentType("text/html")
}

// Resp 是 NewResponse 的快捷方式
func Resp(c *gin.Context) *Response {
	return NewResponse(c)
}

// From 根据错误创建响应
func From(c *gin.Context, err error) *Response {
	resp := NewResponse(c)

	if err == nil {
		return resp
	}

	var errWithCode *ErrWithCode
	if errors.As(err, &errWithCode) {
		resp.Code(errWithCode.Code).Message(errWithCode.Message)
		resp.Status(MapBusinessCodeToHTTPStatus(errWithCode.Code))
	} else {
		resp.Code(CodeInternalError).Message(err.Error())
		resp.Status(http.StatusInternalServerError)
	}

	return resp
}
