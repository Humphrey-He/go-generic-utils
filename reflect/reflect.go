package reflectx

import (
	"errors"
	"reflect"
)

// IsNilValue 安全判断任意值是否为nil
// 支持chan、map、slice、interface、ptr、func、unsafe.Pointer等类型
func IsNilValue(val reflect.Value) bool {
	if !val.IsValid() {
		return true
	}
	switch val.Kind() {
	case reflect.Map, reflect.Chan, reflect.Slice, reflect.Interface, reflect.Ptr, reflect.Func, reflect.UnsafePointer:
		return val.IsNil()
	}
	return false
}

// IsStruct 判断是否为结构体类型
func IsStruct(val any) bool {
	return reflect.TypeOf(val).Kind() == reflect.Struct
}

// IsSlice 判断是否为切片类型
func IsSlice(val any) bool {
	return reflect.TypeOf(val).Kind() == reflect.Slice
}

// IsMap 判断是否为map类型
func IsMap(val any) bool {
	return reflect.TypeOf(val).Kind() == reflect.Map
}

// IsPtr 判断是否为指针类型
func IsPtr(val any) bool {
	return reflect.TypeOf(val).Kind() == reflect.Ptr
}

// GetStructFieldNames 获取结构体所有字段名
func GetStructFieldNames(val any) ([]string, error) {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("参数不是结构体或结构体指针")
	}
	var names []string
	for i := 0; i < t.NumField(); i++ {
		names = append(names, t.Field(i).Name)
	}
	return names, nil
}

// GetStructFieldTags 获取结构体所有字段的tag
func GetStructFieldTags(val any, tagKey string) (map[string]string, error) {
	t := reflect.TypeOf(val)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.New("参数不是结构体或结构体指针")
	}
	tags := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tags[field.Name] = field.Tag.Get(tagKey)
	}
	return tags, nil
}

// GetStructFieldValue 获取结构体字段值
func GetStructFieldValue(val any, fieldName string) (any, error) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("参数不是结构体或结构体指针")
	}
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, errors.New("字段不存在")
	}
	return field.Interface(), nil
}

// SetStructFieldValue 设置结构体字段值（需为指针）
// 仅支持可导出字段
func SetStructFieldValue(ptr any, fieldName string, value any) error {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("参数必须为结构体指针")
	}
	field := v.Elem().FieldByName(fieldName)
	if !field.IsValid() {
		return errors.New("字段不存在")
	}
	if !field.CanSet() {
		return errors.New("字段不可设置（可能是未导出字段）")
	}
	val := reflect.ValueOf(value)
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
		return nil
	}
	return errors.New("类型不匹配，无法赋值")
}

// CallMethod 动态调用对象方法
// obj: 对象实例，methodName: 方法名，args: 参数
// 返回值：[]any（所有返回值），error
func CallMethod(obj any, methodName string, args ...any) ([]any, error) {
	v := reflect.ValueOf(obj)
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return nil, errors.New("方法不存在")
	}
	if len(args) != method.Type().NumIn() {
		return nil, errors.New("参数数量不匹配")
	}
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	out := method.Call(in)
	result := make([]any, len(out))
	for i, v := range out {
		result[i] = v.Interface()
	}
	return result, nil
}

// NewSlice 动态创建切片
func NewSlice(elemType reflect.Type, length, cap int) any {
	return reflect.MakeSlice(reflect.SliceOf(elemType), length, cap).Interface()
}

// NewMap 动态创建map
func NewMap(keyType, elemType reflect.Type) any {
	return reflect.MakeMap(reflect.MapOf(keyType, elemType)).Interface()
}

// NewStruct 动态创建结构体实例
func NewStruct(t reflect.Type) any {
	return reflect.New(t).Elem().Interface()
}

// NewPtr 动态创建指针实例
func NewPtr(t reflect.Type) any {
	return reflect.New(t).Interface()
}

// TypeName 获取类型名称
func TypeName(val any) string {
	return reflect.TypeOf(val).String()
}
