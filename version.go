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

package ggu

import (
	"fmt"
	"runtime"
	"strings"
)

// 版本信息
// Version information
const (
	// Major 主版本号
	Major = 1
	// Minor 次版本号
	Minor = 0
	// Patch 修订号
	Patch = 0
	// Pre 预发布标签，如 "alpha.1", "beta.2", "rc.1"，正式版为空
	// Pre-release tag, e.g., "alpha.1", "beta.2", "rc.1", empty for stable releases
	Pre = ""
	// Build 构建元数据，如构建日期或哈希值
	// Build metadata, such as build date or hash
	Build = ""
)

var (
	// Version 完整版本字符串
	// Complete version string
	Version = makeVersion()
)

// makeVersion 生成完整的版本字符串
// makeVersion generates the complete version string
func makeVersion() string {
	version := fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)

	if Pre != "" {
		version = fmt.Sprintf("%s-%s", version, Pre)
	}

	if Build != "" {
		version = fmt.Sprintf("%s+%s", version, Build)
	}

	return version
}

// GetVersion 返回当前版本信息
// GetVersion returns current version information
func GetVersion() string {
	return Version
}

// GetFullVersionInfo 返回完整的版本信息，包括Go版本
// GetFullVersionInfo returns complete version information including Go version
func GetFullVersionInfo() string {
	return fmt.Sprintf("GGU %s (Go %s)", Version, strings.TrimPrefix(runtime.Version(), "go"))
}
