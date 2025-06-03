package tuple

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestNewPair(t *testing.T) {
	// 测试基本类型
	p1 := NewPair("productID", 100)
	if p1.Key != "productID" || p1.Value != 100 {
		t.Errorf("NewPair 基本类型测试失败, 期望 <productID, 100>, 得到 %v", p1)
	}

	// 测试结构体类型
	type Product struct {
		ID   string
		Name string
	}
	p2 := NewPair(1, Product{ID: "001", Name: "手机"})
	if p2.Key != 1 || p2.Value.ID != "001" || p2.Value.Name != "手机" {
		t.Errorf("NewPair 结构体类型测试失败, 期望 <1, {001 手机}>, 得到 %v", p2)
	}
}

func TestPair_String(t *testing.T) {
	tests := []struct {
		name string
		pair Pair[string, int]
		want string
	}{
		{
			name: "字符串和整数",
			pair: Pair[string, int]{Key: "price", Value: 1999},
			want: "<price, 1999>",
		},
		{
			name: "空值",
			pair: Pair[string, int]{Key: "", Value: 0},
			want: "<, 0>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pair.String(); got != tt.want {
				t.Errorf("Pair.String() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestPair_Split(t *testing.T) {
	p := NewPair("sku001", 299.99)
	key, value := p.Split()

	if key != "sku001" || value != 299.99 {
		t.Errorf("Pair.Split() 返回值错误, 期望 sku001, 299.99, 得到 %v, %v", key, value)
	}
}

func TestPair_MarshalJSON(t *testing.T) {
	p := NewPair("product", 100)
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Pair.MarshalJSON() 错误 = %v", err)
	}

	// 验证JSON格式是否正确
	expected := `{"key":"product","value":100}`
	if string(data) != expected {
		t.Errorf("Pair.MarshalJSON() = %v, 期望 %v", string(data), expected)
	}
}

// 测试Pair的JSON反序列化
func TestPair_UnmarshalJSON(t *testing.T) {
	jsonStr := `{"key":"product","value":100}`
	var p Pair[string, int]

	err := json.Unmarshal([]byte(jsonStr), &p)
	if err != nil {
		t.Fatalf("Pair.UnmarshalJSON() 错误 = %v", err)
	}

	// 验证反序列化结果
	if p.Key != "product" || p.Value != 100 {
		t.Errorf("Pair.UnmarshalJSON() 结果错误, 期望 {product 100}, 得到 %v", p)
	}
}

func TestNewPairs(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		values    []int
		wantPairs []Pair[string, int]
		wantErr   bool
	}{
		{
			name:   "正常情况",
			keys:   []string{"a", "b", "c"},
			values: []int{1, 2, 3},
			wantPairs: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "b", Value: 2},
				{Key: "c", Value: 3},
			},
			wantErr: false,
		},
		{
			name:      "长度不匹配",
			keys:      []string{"a", "b"},
			values:    []int{1, 2, 3},
			wantPairs: nil,
			wantErr:   true,
		},
		{
			name:      "空切片",
			keys:      []string{},
			values:    []int{},
			wantPairs: []Pair[string, int]{},
			wantErr:   false,
		},
		{
			name:      "nil切片",
			keys:      nil,
			values:    nil,
			wantPairs: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPairs, err := NewPairs(tt.keys, tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPairs() 错误 = %v, 期望错误 %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPairs, tt.wantPairs) {
				t.Errorf("NewPairs() = %v, 期望 %v", gotPairs, tt.wantPairs)
			}
		})
	}
}

func TestSplitPairs(t *testing.T) {
	tests := []struct {
		name     string
		pairs    []Pair[string, int]
		wantKeys []string
		wantVals []int
	}{
		{
			name: "正常情况",
			pairs: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "b", Value: 2},
				{Key: "c", Value: 3},
			},
			wantKeys: []string{"a", "b", "c"},
			wantVals: []int{1, 2, 3},
		},
		{
			name:     "空切片",
			pairs:    []Pair[string, int]{},
			wantKeys: []string{},
			wantVals: []int{},
		},
		{
			name:     "nil切片",
			pairs:    nil,
			wantKeys: nil,
			wantVals: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeys, gotVals := SplitPairs(tt.pairs)
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("SplitPairs() keys = %v, 期望 %v", gotKeys, tt.wantKeys)
			}
			if !reflect.DeepEqual(gotVals, tt.wantVals) {
				t.Errorf("SplitPairs() values = %v, 期望 %v", gotVals, tt.wantVals)
			}
		})
	}
}

func TestFlattenPairs(t *testing.T) {
	tests := []struct {
		name       string
		pairs      []Pair[string, int]
		wantResult []any
	}{
		{
			name: "正常情况",
			pairs: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "b", Value: 2},
			},
			wantResult: []any{"a", 1, "b", 2},
		},
		{
			name:       "空切片",
			pairs:      []Pair[string, int]{},
			wantResult: []any{},
		},
		{
			name:       "nil切片",
			pairs:      nil,
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := FlattenPairs(tt.pairs)
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FlattenPairs() = %v, 期望 %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestPackPairs(t *testing.T) {
	tests := []struct {
		name       string
		flatPairs  []any
		wantResult []Pair[string, int]
		wantPanic  bool
	}{
		{
			name:      "正常情况",
			flatPairs: []any{"a", 1, "b", 2},
			wantResult: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "b", Value: 2},
			},
			wantPanic: false,
		},
		{
			name:       "空切片",
			flatPairs:  []any{},
			wantResult: []Pair[string, int]{},
			wantPanic:  false,
		},
		{
			name:       "nil切片",
			flatPairs:  nil,
			wantResult: nil,
			wantPanic:  false,
		},
		{
			name:       "奇数长度",
			flatPairs:  []any{"a", 1, "b"},
			wantResult: nil,
			wantPanic:  true,
		},
		{
			name:       "类型不匹配",
			flatPairs:  []any{"a", "not-an-int", "b", 2},
			wantResult: nil,
			wantPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("PackPairs() 期望 panic, 但没有发生")
					}
				}()
			}

			gotResult := PackPairs[string, int](tt.flatPairs)
			if !tt.wantPanic && !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("PackPairs() = %v, 期望 %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestNewTriple(t *testing.T) {
	// 测试创建三元组
	triple := NewTriple("productID", "手机", 1999.99)
	if triple.First != "productID" || triple.Second != "手机" || triple.Third != 1999.99 {
		t.Errorf("NewTriple 测试失败, 期望 <productID, 手机, 1999.99>, 得到 %v", triple)
	}
}

func TestTriple_String(t *testing.T) {
	tests := []struct {
		name   string
		triple Triple[string, string, float64]
		want   string
	}{
		{
			name:   "基本类型",
			triple: Triple[string, string, float64]{First: "productID", Second: "手机", Third: 1999.99},
			want:   "<productID, 手机, 1999.99>",
		},
		{
			name:   "空值",
			triple: Triple[string, string, float64]{First: "", Second: "", Third: 0},
			want:   "<, , 0>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.triple.String(); got != tt.want {
				t.Errorf("Triple.String() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestTriple_Split(t *testing.T) {
	triple := NewTriple("productID", "手机", 1999.99)
	first, second, third := triple.Split()

	if first != "productID" || second != "手机" || third != 1999.99 {
		t.Errorf("Triple.Split() 返回值错误, 期望 productID, 手机, 1999.99, 得到 %v, %v, %v", first, second, third)
	}
}

func TestTriple_MarshalJSON(t *testing.T) {
	triple := NewTriple("user", 123, true)
	data, err := json.Marshal(triple)
	if err != nil {
		t.Fatalf("Triple.MarshalJSON() 错误 = %v", err)
	}

	// 验证JSON格式是否正确
	expected := `{"first":"user","second":123,"third":true}`
	if string(data) != expected {
		t.Errorf("Triple.MarshalJSON() = %v, 期望 %v", string(data), expected)
	}
}

// 测试Triple的JSON反序列化
func TestTriple_UnmarshalJSON(t *testing.T) {
	jsonStr := `{"first":"user","second":123,"third":true}`
	var triple Triple[string, int, bool]

	err := json.Unmarshal([]byte(jsonStr), &triple)
	if err != nil {
		t.Fatalf("Triple.UnmarshalJSON() 错误 = %v", err)
	}

	// 验证反序列化结果
	if triple.First != "user" || triple.Second != 123 || triple.Third != true {
		t.Errorf("Triple.UnmarshalJSON() 结果错误, 期望 {user 123 true}, 得到 %v", triple)
	}
}

func TestNewKeyValue(t *testing.T) {
	// 测试创建键值对
	kv := NewKeyValue("color", "红色")
	if kv.Key != "color" || kv.Value != "红色" {
		t.Errorf("NewKeyValue 测试失败, 期望 <color, 红色>, 得到 %v", kv)
	}
}

func TestKeyValue_String(t *testing.T) {
	tests := []struct {
		name string
		kv   KeyValue[string, string]
		want string
	}{
		{
			name: "基本类型",
			kv:   KeyValue[string, string]{Key: "color", Value: "红色"},
			want: "color: 红色",
		},
		{
			name: "空值",
			kv:   KeyValue[string, string]{Key: "", Value: ""},
			want: ": ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kv.String(); got != tt.want {
				t.Errorf("KeyValue.String() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestKeyValue_MarshalJSON(t *testing.T) {
	kv := NewKeyValue("color", "红色")
	data, err := json.Marshal(kv)
	if err != nil {
		t.Fatalf("KeyValue.MarshalJSON() 错误 = %v", err)
	}

	// 验证JSON格式是否正确
	expected := `{"key":"color","value":"红色"}`
	if string(data) != expected {
		t.Errorf("KeyValue.MarshalJSON() = %v, 期望 %v", string(data), expected)
	}
}

func TestMapFromPairs(t *testing.T) {
	tests := []struct {
		name  string
		pairs []Pair[string, int]
		want  map[string]int
	}{
		{
			name: "正常情况",
			pairs: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "b", Value: 2},
				{Key: "c", Value: 3},
			},
			want: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
		},
		{
			name:  "空切片",
			pairs: []Pair[string, int]{},
			want:  map[string]int{},
		},
		{
			name: "重复键",
			pairs: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "a", Value: 2}, // 重复键，后者覆盖前者
			},
			want: map[string]int{
				"a": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapFromPairs(tt.pairs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapFromPairs() = %v, 期望 %v", got, tt.want)
			}
		})
	}
}

func TestPairsFromMap(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]int
		want []Pair[string, int]
	}{
		{
			name: "正常情况",
			m: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			want: []Pair[string, int]{
				{Key: "a", Value: 1},
				{Key: "b", Value: 2},
				{Key: "c", Value: 3},
			},
		},
		{
			name: "空map",
			m:    map[string]int{},
			want: []Pair[string, int]{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PairsFromMap(tt.m)

			// 由于map迭代顺序不确定，需要先转换回map再比较
			gotMap := MapFromPairs(got)
			wantMap := MapFromPairs(tt.want)

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("PairsFromMap() 转回map后 = %v, 期望 %v", gotMap, wantMap)
			}

			// 检查长度是否一致
			if len(got) != len(tt.want) {
				t.Errorf("PairsFromMap() 长度 = %v, 期望长度 %v", len(got), len(tt.want))
			}
		})
	}
}

func TestRange(t *testing.T) {
	pairs := []Pair[string, int]{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
		{Key: "c", Value: 3},
	}

	// 测试正常情况
	sum := 0
	err := Range(pairs, func(k string, v int) error {
		sum += v
		return nil
	})

	if err != nil {
		t.Errorf("Range() 错误 = %v, 期望无错误", err)
	}
	if sum != 6 {
		t.Errorf("Range() 求和结果 = %v, 期望 6", sum)
	}

	// 测试错误情况
	expectedErr := errors.New("测试错误")
	err = Range(pairs, func(k string, v int) error {
		if k == "b" {
			return expectedErr
		}
		return nil
	})

	if err != expectedErr {
		t.Errorf("Range() 错误 = %v, 期望 %v", err, expectedErr)
	}
}

func TestFilter(t *testing.T) {
	pairs := []Pair[string, int]{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
		{Key: "c", Value: 3},
		{Key: "d", Value: 4},
	}

	// 过滤出偶数值
	filtered := Filter(pairs, func(k string, v int) bool {
		return v%2 == 0
	})

	expected := []Pair[string, int]{
		{Key: "b", Value: 2},
		{Key: "d", Value: 4},
	}

	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("Filter() = %v, 期望 %v", filtered, expected)
	}

	// 测试空切片
	emptyFiltered := Filter([]Pair[string, int]{}, func(k string, v int) bool {
		return true
	})
	if len(emptyFiltered) != 0 {
		t.Errorf("Filter() 空切片结果 = %v, 期望空切片", emptyFiltered)
	}
}

func TestMap(t *testing.T) {
	pairs := []Pair[string, int]{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
		{Key: "c", Value: 3},
	}

	// 将值翻倍，键转为大写
	mapped := Map(pairs, func(k string, v int) (string, int) {
		return k + "_mapped", v * 2
	})

	expected := []Pair[string, int]{
		{Key: "a_mapped", Value: 2},
		{Key: "b_mapped", Value: 4},
		{Key: "c_mapped", Value: 6},
	}

	if !reflect.DeepEqual(mapped, expected) {
		t.Errorf("Map() = %v, 期望 %v", mapped, expected)
	}

	// 测试空切片
	emptyMapped := Map([]Pair[string, int]{}, func(k string, v int) (string, int) {
		return k, v
	})
	if len(emptyMapped) != 0 {
		t.Errorf("Map() 空切片结果 = %v, 期望空切片", emptyMapped)
	}
}

func TestReduce(t *testing.T) {
	pairs := []Pair[string, int]{
		{Key: "a", Value: 1},
		{Key: "b", Value: 2},
		{Key: "c", Value: 3},
	}

	// 求和
	sum := Reduce(pairs, 0, func(acc int, k string, v int) int {
		return acc + v
	})

	if sum != 6 {
		t.Errorf("Reduce() 求和 = %v, 期望 6", sum)
	}

	// 拼接键
	concat := Reduce(pairs, "", func(acc string, k string, v int) string {
		if acc == "" {
			return k
		}
		return acc + "," + k
	})

	if concat != "a,b,c" {
		t.Errorf("Reduce() 拼接 = %v, 期望 a,b,c", concat)
	}

	// 测试空切片
	emptySum := Reduce([]Pair[string, int]{}, 10, func(acc int, k string, v int) int {
		return acc + v
	})
	if emptySum != 10 {
		t.Errorf("Reduce() 空切片结果 = %v, 期望 10", emptySum)
	}
}
