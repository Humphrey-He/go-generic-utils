// Copyright 2024 Humphrey-He
//
// 本文件为 mapx.go 的测试用例，覆盖泛型Map、线程安全Map、链表Map、TreeMap、多值Map等主要功能，符合Go测试规范。

package maputils

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试泛型Map基本功能
func TestGenericMap_Basic(t *testing.T) {
	m := NewGenericMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	val, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)
	m.Delete("a")
	_, ok = m.Get("a")
	assert.False(t, ok)
	assert.ElementsMatch(t, []string{"b"}, m.Keys())
	assert.ElementsMatch(t, []int{2}, m.Values())
	assert.Equal(t, 1, m.Len())
}

// 测试泛型Map扩展功能
func TestGenericMap_Extended(t *testing.T) {
	m := NewGenericMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("d", 4)

	// 测试Map函数
	mappedInterface := m.Map(func(k string, v int) (string, int) {
		return k + "_mapped", v * 2
	})

	// 将接口转换为具体类型
	mapped, ok := mappedInterface.(*GenericMap[string, int])
	assert.True(t, ok)

	// 验证原始map不变
	assert.Equal(t, 4, m.Len())
	v, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, v)

	// 验证新map
	assert.Equal(t, 4, mapped.Len())
	v, ok = mapped.Get("a_mapped")
	assert.True(t, ok)
	assert.Equal(t, 2, v)

	// 测试Filter函数
	filtered := m.Filter(func(k string, v int) bool {
		return v%2 == 0 // 只保留偶数
	})

	// 验证过滤结果
	assert.Equal(t, 2, filtered.Len())
	_, ok = filtered.Get("a")
	assert.False(t, ok) // a=1 应该被过滤
	v, ok = filtered.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, v)

	// 测试ForEach
	sum := 0
	m.ForEach(func(k string, v int) {
		sum += v
	})
	assert.Equal(t, 10, sum) // 1+2+3+4=10

	// 测试Merge
	other := NewGenericMap[string, int]()
	other.Set("c", 30) // 覆盖c
	other.Set("e", 5)  // 新增e

	merged := m.Merge(other)
	assert.Equal(t, 5, merged.Len())
	v, _ = merged.Get("c")
	assert.Equal(t, 30, v) // 应该是other中的值
	v, _ = merged.Get("e")
	assert.Equal(t, 5, v)

	// 测试Clear
	m.Clear()
	assert.Equal(t, 0, m.Len())
}

// 测试线程安全Map并发读写
func TestSyncMap_Concurrent(t *testing.T) {
	m := NewSyncMap[int, int]()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			m.Set(val, val*10)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, m.Len())
	for i := 0; i < 100; i++ {
		v, ok := m.Get(i)
		assert.True(t, ok)
		assert.Equal(t, i*10, v)
	}
}

// 测试链表Map插入顺序
func TestLinkedMap_Order(t *testing.T) {
	m := NewLinkedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	assert.Equal(t, []string{"a", "b", "c"}, m.Keys())
	m.Delete("b")
	assert.Equal(t, []string{"a", "c"}, m.Keys())
	val, ok := m.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, val)
}

// 测试TreeMap基本功能
func TestTreeMap_Basic(t *testing.T) {
	m := NewTreeMap[int, string]()
	m.Set(2, "b")
	m.Set(1, "a")
	m.Set(3, "c")
	val, ok := m.Get(2)
	assert.True(t, ok)
	assert.Equal(t, "b", val)
	// keys顺序未保证，手动排序
	keys := m.Keys()
	sort.Ints(keys)
	assert.Equal(t, []int{1, 2, 3}, keys)
	m.Delete(2)
	_, ok = m.Get(2)
	assert.False(t, ok)
}

// 测试TreeMap顺序和完整功能
func TestTreeMap_Order(t *testing.T) {
	m := NewTreeMap[int, string]()

	// 测试插入顺序
	m.Set(5, "e")
	m.Set(3, "c")
	m.Set(1, "a")
	m.Set(4, "d")
	m.Set(2, "b")

	// 验证自动排序
	assert.Equal(t, []int{1, 2, 3, 4, 5}, m.Keys())

	// 测试更新值
	m.Set(3, "cc")
	val, ok := m.Get(3)
	assert.True(t, ok)
	assert.Equal(t, "cc", val)

	// 测试删除
	m.Delete(3)
	assert.Equal(t, []int{1, 2, 4, 5}, m.Keys())
	_, ok = m.Get(3)
	assert.False(t, ok)

	// 测试清空
	m.Clear()
	assert.Equal(t, 0, m.Len())
	assert.Empty(t, m.Keys())
}

// 测试多值Map
func TestMultiMap_Basic(t *testing.T) {
	m := NewMultiMap[string, int]()
	m.Add("a", 1)
	m.Add("a", 2)
	m.Add("b", 3)
	assert.ElementsMatch(t, []int{1, 2}, m.Get("a"))
	assert.ElementsMatch(t, []int{3}, m.Get("b"))
	m.Delete("a")
	assert.Empty(t, m.Get("a"))
	assert.ElementsMatch(t, []string{"b"}, m.Keys())
	assert.Equal(t, 1, m.Len())
}

// 测试多值Map的扩展功能
func TestMultiMap_Extended(t *testing.T) {
	m := NewMultiMap[string, int]()

	// 测试添加多个值
	m.Add("fruits", 1)     // apple
	m.Add("fruits", 2)     // banana
	m.Add("fruits", 3)     // cherry
	m.Add("vegetables", 4) // carrot
	m.Add("vegetables", 5) // broccoli

	// 测试获取所有值
	assert.ElementsMatch(t, []int{1, 2, 3}, m.Get("fruits"))

	// 测试删除单个值
	removed := m.RemoveValue("fruits", 2)
	assert.True(t, removed)
	assert.ElementsMatch(t, []int{1, 3}, m.Get("fruits"))

	// 测试不存在的值
	removed = m.RemoveValue("fruits", 10)
	assert.False(t, removed)

	// 测试获取所有值的平铺列表
	allValues := m.AllValues()
	assert.ElementsMatch(t, []int{1, 3, 4, 5}, allValues)

	// 测试检查键是否包含特定值
	contains := m.Contains("fruits", 3)
	assert.True(t, contains)
	contains = m.Contains("fruits", 2)
	assert.False(t, contains)

	// 测试清空
	m.Clear()
	assert.Equal(t, 0, m.Len())
	assert.Empty(t, m.Get("fruits"))
	assert.Empty(t, m.Get("vegetables"))
}
