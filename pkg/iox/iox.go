package iox

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
)

// ================== 多段字节流 ==================

// MultipleBytesReader 支持顺序读取多个[]byte的io.Reader
type MultipleBytesReader struct {
	buffers [][]byte
	index   int
	offset  int
}

// NewMultipleBytesReader 创建多段字节流读取器
func NewMultipleBytesReader(buffers ...[]byte) *MultipleBytesReader {
	return &MultipleBytesReader{buffers: buffers}
}

// Read 实现io.Reader接口，顺序读取所有[]byte
func (r *MultipleBytesReader) Read(p []byte) (n int, err error) {
	for r.index < len(r.buffers) && n < len(p) {
		buf := r.buffers[r.index]
		if r.offset >= len(buf) {
			r.index++
			r.offset = 0
			continue
		}
		copied := copy(p[n:], buf[r.offset:])
		n += copied
		r.offset += copied
		if r.offset == len(buf) {
			r.index++
			r.offset = 0
		}
	}
	if n == 0 && r.index >= len(r.buffers) {
		return 0, io.EOF
	}
	return n, nil
}

// ================== 并发安全多段字节流 ==================

// ConcurrentMultipleBytesReader 并发安全的多段字节流读取器
type ConcurrentMultipleBytesReader struct {
	mu       sync.Mutex
	delegate *MultipleBytesReader
}

// NewConcurrentMultipleBytesReader 创建并发安全多段字节流读取器
func NewConcurrentMultipleBytesReader(buffers ...[]byte) *ConcurrentMultipleBytesReader {
	return &ConcurrentMultipleBytesReader{
		delegate: NewMultipleBytesReader(buffers...),
	}
}

// Read 并发安全读取
func (r *ConcurrentMultipleBytesReader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.delegate.Read(p)
}

// ================== JSON流读取工具 ==================

// JSONReader 支持流式读取JSON对象的io.Reader
type JSONReader struct {
	enc *json.Encoder
	buf *bytes.Buffer
}

// NewJSONReader 创建JSON流读取器
// v: 任意可序列化为JSON的对象
func NewJSONReader(v any) *JSONReader {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	_ = enc.Encode(v)
	return &JSONReader{
		enc: enc,
		buf: buf,
	}
}

// Read 实现io.Reader接口，读取JSON字节流
func (r *JSONReader) Read(p []byte) (n int, err error) {
	return r.buf.Read(p)
}
