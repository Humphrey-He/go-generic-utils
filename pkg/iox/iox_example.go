// Copyright 2024 ecodeclub
//
// 本文件为 iox.go 的示例文件，演示多段字节流、并发安全多段字节流、JSON流读取等常用用法。

package iox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// ExampleMultipleBytesReader 演示多段字节流的基本用法
func ExampleMultipleBytesReader() {
	data1 := []byte("foo")
	data2 := []byte("bar")
	r := NewMultipleBytesReader(data1, data2)
	buf := make([]byte, 6)
	n, err := r.Read(buf)
	fmt.Println(string(buf[:n]), err)
	// Output:
	// foobar <nil>
}

// ExampleConcurrentMultipleBytesReader 演示并发安全多段字节流的用法
func ExampleConcurrentMultipleBytesReader() {
	data := [][]byte{[]byte("hello,"), []byte("world")}
	r := NewConcurrentMultipleBytesReader(data...)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 11)
			n, err := r.Read(buf)
			fmt.Println(string(buf[:n]), err)
		}()
	}
	wg.Wait()
	// Output:
	// hello,world <nil>
	//  <EOF>
}

// ExampleJSONReader 演示JSON流读取的用法
func ExampleJSONReader() {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	u := User{ID: 1, Name: "Tom"}
	r := NewJSONReader(u)
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	var u2 User
	_ = json.Unmarshal(buf.Bytes(), &u2)
	fmt.Println(u2.ID, u2.Name)
	// Output:
	// 1 Tom
}
