// Copyright 2024 Humphrey-He
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package list

import (
	"fmt"
)

// NewIndexOutOfRangeError 创建一个详细的索引越界错误
// 包含列表长度和访问的索引信息
func NewIndexOutOfRangeError(length, index int) error {
	return fmt.Errorf("ggu: 下标超出范围，长度 %d, 下标 %d", length, index)
}
