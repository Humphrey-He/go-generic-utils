// Copyright 2024 Humphrey-He
//
// 本文件为 iox.go 的测试用例，覆盖多段字节流、并发安全多段字节流、JSON流读取等主要功能，符合Go测试规范。

package iox

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试多段字节流读取器
func TestMultipleBytesReader_Basic(t *testing.T) {
	data1 := []byte("hello,")
	data2 := []byte("world")
	r := NewMultipleBytesReader(data1, data2)
	buf := make([]byte, 11)
	n, err := r.Read(buf)
	assert.Equal(t, 11, n)
	assert.NoError(t, err)
	assert.Equal(t, "hello,world", string(buf))
	// 读到结尾应返回EOF
	n, err = r.Read(buf)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

// 测试多段字节流分多次读取
func TestMultipleBytesReader_MultiRead(t *testing.T) {
	data1 := []byte("abc")
	data2 := []byte("defg")
	r := NewMultipleBytesReader(data1, data2)
	buf := make([]byte, 2)
	var result []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
	}
	assert.Equal(t, "abcdefg", string(result))
}

// 测试并发安全多段字节流
func TestConcurrentMultipleBytesReader_Concurrent(t *testing.T) {
	data := [][]byte{[]byte("foo"), []byte("bar")}
	r := NewConcurrentMultipleBytesReader(data...)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 6)
			_, _ = r.Read(buf)
		}()
	}
	wg.Wait()
}

// 测试JSON流读取
func TestJSONReader_Basic(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	u := User{ID: 1, Name: "Tom"}
	r := NewJSONReader(u)
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	assert.NoError(t, err)
	var u2 User
	err = json.Unmarshal(buf.Bytes(), &u2)
	assert.NoError(t, err)
	assert.Equal(t, u, u2)
}
