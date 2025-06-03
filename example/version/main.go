// Copyright 2024 Humphrey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 版本信息示例
// Version information example
package main

import (
	"fmt"

	ggu "github.com/Humphrey-He/go-generic-utils"
)

func main() {
	// 获取版本信息
	// Get version information
	version := ggu.GetVersion()
	fmt.Printf("GGU 版本: %s\n", version)

	// 获取完整版本信息
	// Get full version information
	fullVersion := ggu.GetFullVersionInfo()
	fmt.Printf("完整版本信息: %s\n", fullVersion)

	// 输出版本号组件
	// Output version components
	fmt.Printf("主版本号: %d\n", ggu.Major)
	fmt.Printf("次版本号: %d\n", ggu.Minor)
	fmt.Printf("修订号: %d\n", ggu.Patch)

	if ggu.Pre != "" {
		fmt.Printf("预发布标签: %s\n", ggu.Pre)
	} else {
		fmt.Println("这是一个正式发布版本")
	}

	if ggu.Build != "" {
		fmt.Printf("构建元数据: %s\n", ggu.Build)
	}
}
