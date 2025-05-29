// Copyright 2024 ecodeclub
//
// 本文件为 reflect.go 的测试用例，覆盖常用反射工具API，符合Go测试规范。

package reflectx

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

// 测试 IsNilValue
func TestIsNilValue(t *testing.T) {
	var nilPtr *int
	var nilSlice []int
	var nilMap map[string]int
	var nilCh chan int
	var nilFunc func()
	var nilIface interface{}
	var nilUnsafePtr unsafe.Pointer

	assert.True(t, IsNilValue(reflect.ValueOf(nilPtr)))
	assert.True(t, IsNilValue(reflect.ValueOf(nilSlice)))
	assert.True(t, IsNilValue(reflect.ValueOf(nilMap)))
	assert.True(t, IsNilValue(reflect.ValueOf(nilCh)))
	assert.True(t, IsNilValue(reflect.ValueOf(nilFunc)))
	assert.True(t, IsNilValue(reflect.ValueOf(nilIface)))
	assert.True(t, IsNilValue(reflect.ValueOf(nilUnsafePtr)))
	assert.False(t, IsNilValue(reflect.ValueOf(123)))
	assert.False(t, IsNilValue(reflect.ValueOf("abc")))
}

// 测试类型判断API
func TestTypeCheckers(t *testing.T) {
	type S struct{}
	assert.True(t, IsStruct(S{}))
	assert.True(t, IsStruct(&S{}))
	assert.True(t, IsSlice([]int{1, 2}))
	assert.True(t, IsMap(map[string]int{}))
	assert.True(t, IsPtr(&S{}))
	assert.False(t, IsStruct(123))
	assert.False(t, IsSlice(123))
	assert.False(t, IsMap(123))
	assert.False(t, IsPtr(123))
}

// 测试结构体字段名和tag
func TestGetStructFieldNamesAndTags(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	u := User{ID: 1, Name: "Tom"}
	names, err := GetStructFieldNames(u)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"ID", "Name"}, names)

	tags, err := GetStructFieldTags(u, "json")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"ID": "id", "Name": "name"}, tags)
}

// 测试结构体字段值获取与设置
func TestGetSetStructFieldValue(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}
	u := &User{ID: 1, Name: "Tom"}
	val, err := GetStructFieldValue(u, "Name")
	assert.NoError(t, err)
	assert.Equal(t, "Tom", val)

	err = SetStructFieldValue(u, "Name", "Jerry")
	assert.NoError(t, err)
	assert.Equal(t, "Jerry", u.Name)

	// 不存在字段
	_, err = GetStructFieldValue(u, "NotExist")
	assert.Error(t, err)
	err = SetStructFieldValue(u, "NotExist", 1)
	assert.Error(t, err)
}

// 测试动态调用方法
// 定义在包级作用域
type S struct{}

// 带参数和返回值的方法
func (s S) Add(a, b int) int { return a + b }

func TestCallMethod(t *testing.T) {
	s := S{}
	res, err := CallMethod(s, "Add", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, 3, res[0])

	// 方法不存在
	_, err = CallMethod(s, "NotExist")
	assert.Error(t, err)
}

// 测试动态创建切片、map、struct、指针
func TestNewSliceMapStructPtr(t *testing.T) {
	// 切片
	slice := NewSlice(reflect.TypeOf(1), 2, 4).([]int)
	assert.Len(t, slice, 2)
	// map
	m := NewMap(reflect.TypeOf(""), reflect.TypeOf(1)).(map[string]int)
	m["a"] = 1
	assert.Equal(t, 1, m["a"])
	// struct
	type User struct{ Name string }
	u := NewStruct(reflect.TypeOf(User{})).(User)
	assert.Equal(t, "", u.Name)
	// ptr
	ptr := NewPtr(reflect.TypeOf(User{})).(*User)
	assert.NotNil(t, ptr)
}

// 测试类型名称
func TestTypeName(t *testing.T) {
	assert.Equal(t, "int", TypeName(1))
	assert.Equal(t, "string", TypeName("abc"))
	type User struct{}
	assert.Equal(t, "reflectx.User", TypeName(User{}))
}
