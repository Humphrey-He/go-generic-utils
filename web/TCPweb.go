package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // 函数退出时关闭连接
	fmt.Printf("Client connected: %s\n", conn.RemoteAddr())

	// 4. 使用 bufio 读取客户端数据
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Client %s disconnected: %v\n", conn.RemoteAddr(), err)
			return
		}

		// 5. 处理数据（这里简单回显）
		message = strings.TrimSpace(message)
		fmt.Printf("Received from %s: %s\n", conn.RemoteAddr(), message)

		// 6. 返回响应
		response := fmt.Sprintf("Echo: %s\n", message)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response:", err)
			return
		}
	}
}

func main() {
	// 1. 监听端口
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listener.Close() // 确保关闭监听器
	fmt.Println("Server started, listening on :8080")

	// 2. 循环接受客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue // 继续等待新连接
		}

		// 3. 为每个连接启动一个 goroutine 处理
		go handleConnection(conn)
	}
}
