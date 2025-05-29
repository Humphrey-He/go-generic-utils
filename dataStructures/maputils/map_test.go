// Copyright 2024 ecodeclub
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
