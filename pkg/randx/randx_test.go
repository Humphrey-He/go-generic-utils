// Copyright 2023 ecodeclub
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

package randx

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRandCode 测试随机字符串生成
func TestRandCode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		length    int
		typ       Type
		wantMatch string
		wantErr   error
	}{
		{
			name:      "数字验证码",
			length:    100,
			typ:       TypeDigit,
			wantMatch: "^[0-9]+$",
		},
		{
			name:      "小写字母验证码",
			length:    100,
			typ:       TypeLowerCase,
			wantMatch: "^[a-z]+$",
		},
		{
			name:      "数字+小写字母验证码",
			length:    100,
			typ:       TypeDigit | TypeLowerCase,
			wantMatch: "^[a-z0-9]+$",
		},
		{
			name:      "数字+大写字母验证码",
			length:    100,
			typ:       TypeDigit | TypeUpperCase,
			wantMatch: "^[A-Z0-9]+$",
		},
		{
			name:      "大写字母验证码",
			length:    100,
			typ:       TypeUpperCase,
			wantMatch: "^[A-Z]+$",
		},
		{
			name:      "大小写字母验证码",
			length:    100,
			typ:       TypeUpperCase | TypeLowerCase,
			wantMatch: "^[a-zA-Z]+$",
		},
		{
			name:      "字母和数字验证码",
			length:    100,
			typ:       TypeAlphanumeric,
			wantMatch: "^[0-9a-zA-Z]+$",
		},
		{
			name:      "所有类型验证",
			length:    100,
			typ:       TypeMixed,
			wantMatch: "^[\\S\\s]+$",
		},
		{
			name:      "特殊字符类型验证",
			length:    100,
			typ:       TypeSpecial,
			wantMatch: "^[^0-9a-zA-Z]+$",
		},
		{
			name:    "未定义类型(超过范围)",
			length:  100,
			typ:     TypeMixed + 1,
			wantErr: ErrTypeNotSupported,
		},
		{
			name:    "未定义类型(0)",
			length:  100,
			typ:     0,
			wantErr: ErrTypeNotSupported,
		},
		{
			name:    "长度小于0",
			length:  -1,
			typ:     TypeDigit,
			wantErr: ErrLengthLessThanZero,
		},
		{
			name:      "长度等于0",
			length:    0,
			typ:       TypeMixed,
			wantMatch: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, err := RandCode(tc.length, tc.typ)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Len(t, code, tc.length)

			if tc.length > 0 {
				matched, err := regexp.MatchString(tc.wantMatch, code)
				require.NoError(t, err)
				assert.Truef(t, matched, "expected %s but got %s", tc.wantMatch, code)
			}
		})
	}
}

// TestRandStrByCharset 测试根据字符集生成随机字符串
func TestRandStrByCharset(t *testing.T) {
	t.Parallel()

	matchFunc := func(str, charset string) bool {
		for _, c := range str {
			if !strings.Contains(charset, string(c)) {
				return false
			}
		}
		return true
	}

	testCases := []struct {
		name    string
		length  int
		charset string
		wantErr error
	}{
		{
			name:    "长度小于0",
			length:  -1,
			charset: "123",
			wantErr: ErrLengthLessThanZero,
		},
		{
			name:    "长度等于0",
			length:  0,
			charset: "123",
		},
		{
			name:    "空字符集",
			length:  10,
			charset: "",
			wantErr: ErrInvalidCharset,
		},
		{
			name:    "随机字符串测试1",
			length:  100,
			charset: "2rg248ry227t@@",
		},
		{
			name:    "随机字符串测试2",
			length:  100,
			charset: "2rg248ry227t@&*($.!",
		},
		{
			name:    "中文字符集",
			length:  10,
			charset: "你好世界中国",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, err := RandStrByCharset(tc.length, tc.charset)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Len(t, code, tc.length)

			if tc.length > 0 {
				assert.True(t, matchFunc(code, tc.charset))
			}
		})
	}
}

// TestRandInt 测试随机整数生成
func TestRandInt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		min     int
		max     int
		wantErr error
	}{
		{
			name: "min小于max",
			min:  0,
			max:  100,
		},
		{
			name:    "min等于max",
			min:     100,
			max:     100,
			wantErr: ErrInvalidRange,
		},
		{
			name:    "min大于max",
			min:     100,
			max:     0,
			wantErr: ErrInvalidRange,
		},
		{
			name: "负数范围",
			min:  -100,
			max:  -50,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// 生成100个随机数进行测试
			for i := 0; i < 100; i++ {
				n, err := RandInt(tc.min, tc.max)

				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
					return
				}

				require.NoError(t, err)
				assert.GreaterOrEqual(t, n, tc.min)
				assert.Less(t, n, tc.max)
			}
		})
	}
}

// TestRandFloat64 测试随机浮点数生成
func TestRandFloat64(t *testing.T) {
	t.Parallel()

	// 生成100个随机浮点数进行测试
	for i := 0; i < 100; i++ {
		n, err := RandFloat64()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, n, 0.0)
		assert.Less(t, n, 1.0)
	}
}

// TestRandFloat64Range 测试指定范围内的随机浮点数生成
func TestRandFloat64Range(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		min     float64
		max     float64
		wantErr error
	}{
		{
			name: "min小于max",
			min:  0,
			max:  100,
		},
		{
			name:    "min等于max",
			min:     100,
			max:     100,
			wantErr: ErrInvalidRange,
		},
		{
			name:    "min大于max",
			min:     100,
			max:     0,
			wantErr: ErrInvalidRange,
		},
		{
			name: "负数范围",
			min:  -100,
			max:  -50,
		},
		{
			name: "小数范围",
			min:  0.1,
			max:  0.2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// 生成100个随机数进行测试
			for i := 0; i < 100; i++ {
				n, err := RandFloat64Range(tc.min, tc.max)

				if tc.wantErr != nil {
					assert.ErrorIs(t, err, tc.wantErr)
					return
				}

				require.NoError(t, err)
				assert.GreaterOrEqual(t, n, tc.min)
				assert.Less(t, n, tc.max)
			}
		})
	}
}

// TestRandUUID 测试UUID生成
func TestRandUUID(t *testing.T) {
	t.Parallel()

	// UUID v4 格式正则表达式
	uuidRegex := "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"

	// 生成多个UUID并验证其唯一性
	uuids := make(map[string]struct{})

	for i := 0; i < 1000; i++ {
		uuid, err := RandUUID()
		require.NoError(t, err)

		// 验证格式
		matched, err := regexp.MatchString(uuidRegex, uuid)
		require.NoError(t, err)
		assert.True(t, matched, "不是有效的UUID格式: %s", uuid)

		// 验证唯一性
		_, exists := uuids[uuid]
		assert.False(t, exists, "UUID重复: %s", uuid)
		uuids[uuid] = struct{}{}
	}
}

// TestRandBool 测试随机布尔值生成
func TestRandBool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		probability float64
		wantErr     error
	}{
		{
			name:        "概率为0",
			probability: 0,
		},
		{
			name:        "概率为1",
			probability: 1,
		},
		{
			name:        "概率为0.5",
			probability: 0.5,
		},
		{
			name:        "概率小于0",
			probability: -0.1,
			wantErr:     ErrInvalidProbability,
		},
		{
			name:        "概率大于1",
			probability: 1.1,
			wantErr:     ErrInvalidProbability,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantErr != nil {
				_, err := RandBool(tc.probability)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			// 生成1000个随机布尔值，并检查true的比例
			// 这不是一个精确的测试，但可以大致检查概率分布
			trueCount := 0
			iterations := 1000

			for i := 0; i < iterations; i++ {
				b, err := RandBool(tc.probability)
				require.NoError(t, err)
				if b {
					trueCount++
				}
			}

			// 计算实际概率
			actualProb := float64(trueCount) / float64(iterations)

			// 根据样本大小允许一定的误差范围
			// 对于1000次测试，使用0.05的允许误差
			allowedError := 0.05
			assert.InDelta(t, tc.probability, actualProb, allowedError)
		})
	}
}

// TestRandDate 测试随机日期生成
func TestRandDate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name    string
		start   time.Time
		end     time.Time
		wantErr error
	}{
		{
			name:  "过去一周",
			start: now.AddDate(0, 0, -7),
			end:   now,
		},
		{
			name:  "过去一个月",
			start: now.AddDate(0, -1, 0),
			end:   now,
		},
		{
			name:  "过去一年",
			start: now.AddDate(-1, 0, 0),
			end:   now,
		},
		{
			name:  "未来一周",
			start: now,
			end:   now.AddDate(0, 0, 7),
		},
		{
			name:  "相同时间",
			start: now,
			end:   now,
		},
		{
			name:    "开始时间晚于结束时间",
			start:   now.AddDate(0, 0, 1),
			end:     now,
			wantErr: ErrInvalidRange,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			date, err := RandDate(tc.start, tc.end)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)

			// 特殊情况：开始和结束时间相同
			if tc.start.Equal(tc.end) {
				assert.Equal(t, tc.start, date)
				return
			}

			// 验证生成的日期在范围内
			assert.True(t, !date.Before(tc.start), "日期 %v 早于开始时间 %v", date, tc.start)
			assert.True(t, !date.After(tc.end), "日期 %v 晚于结束时间 %v", date, tc.end)
		})
	}
}

// TestShuffle 测试洗牌算法
func TestShuffle(t *testing.T) {
	t.Parallel()

	// 测试不同类型的切片
	t.Run("整数切片", func(t *testing.T) {
		t.Parallel()

		original := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

		shuffled, err := Shuffle(original)
		require.NoError(t, err)

		// 确保长度相同
		assert.Equal(t, len(original), len(shuffled))

		// 确保所有元素都存在
		for _, v := range original {
			assert.Contains(t, shuffled, v)
		}

		// 确保顺序已经改变（这个测试可能偶尔失败，但概率很低）
		different := false
		for i := range original {
			if original[i] != shuffled[i] {
				different = true
				break
			}
		}
		assert.True(t, different, "洗牌后顺序应该改变")
	})

	t.Run("字符串切片", func(t *testing.T) {
		t.Parallel()

		original := []string{"a", "b", "c", "d", "e"}

		shuffled, err := Shuffle(original)
		require.NoError(t, err)

		// 确保长度相同
		assert.Equal(t, len(original), len(shuffled))

		// 确保所有元素都存在
		for _, v := range original {
			assert.Contains(t, shuffled, v)
		}
	})

	t.Run("空切片", func(t *testing.T) {
		t.Parallel()

		var original []int

		shuffled, err := Shuffle(original)
		require.NoError(t, err)

		assert.Empty(t, shuffled)
	})
}

// TestWeightedChoice 测试带权重的随机选择
func TestWeightedChoice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		items   []string
		weights []float64
		wantErr error
	}{
		{
			name:    "正常权重",
			items:   []string{"A", "B", "C"},
			weights: []float64{1, 2, 3},
		},
		{
			name:    "空切片",
			items:   []string{},
			weights: []float64{},
			wantErr: ErrInvalidParameterValue,
		},
		{
			name:    "长度不匹配",
			items:   []string{"A", "B", "C"},
			weights: []float64{1, 2},
			wantErr: ErrInvalidParameterValue,
		},
		{
			name:    "负权重",
			items:   []string{"A", "B", "C"},
			weights: []float64{1, -1, 3},
			wantErr: ErrInvalidProbability,
		},
		{
			name:    "零权重总和",
			items:   []string{"A", "B", "C"},
			weights: []float64{0, 0, 0},
			wantErr: ErrInvalidProbability,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantErr != nil {
				_, err := WeightedChoice(tc.items, tc.weights)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			// 权重检查 - 生成足够多的样本以验证分布
			counts := make(map[string]int)
			iterations := 10000

			for i := 0; i < iterations; i++ {
				choice, err := WeightedChoice(tc.items, tc.weights)
				require.NoError(t, err)
				counts[choice]++
			}

			// 计算预期的理论分布
			var totalWeight float64
			for _, w := range tc.weights {
				totalWeight += w
			}

			// 检查每个选项的分布是否接近其权重比例
			for i, item := range tc.items {
				expectedProb := tc.weights[i] / totalWeight
				actualProb := float64(counts[item]) / float64(iterations)

				// 允许5%的误差
				assert.InDelta(t, expectedProb, actualProb, 0.05,
					"项目 %s 的观察到的概率 %.4f 与预期概率 %.4f 相差太大", item, actualProb, expectedProb)
			}
		})
	}
}

// TestRandProductID 测试商品ID生成
func TestRandProductID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		prefix  string
		length  int
		wantErr error
	}{
		{
			name:   "正常参数",
			prefix: "P",
			length: 6,
		},
		{
			name:   "空前缀",
			prefix: "",
			length: 6,
		},
		{
			name:   "长前缀",
			prefix: "PRODUCT_",
			length: 6,
		},
		{
			name:    "零长度",
			prefix:  "P",
			length:  0,
			wantErr: ErrLengthLessThanZero,
		},
		{
			name:    "负长度",
			prefix:  "P",
			length:  -1,
			wantErr: ErrLengthLessThanZero,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id, err := RandProductID(tc.prefix, tc.length)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)

			// 验证前缀
			assert.True(t, strings.HasPrefix(id, tc.prefix))

			// 验证数字部分
			numPart := id[len(tc.prefix):]
			assert.Len(t, numPart, tc.length)

			// 验证数字部分只包含数字
			matched, err := regexp.MatchString("^[0-9]+$", numPart)
			require.NoError(t, err)
			assert.True(t, matched)
		})
	}
}

// 更多电商相关函数的测试在 ecommerce_test.go 中
