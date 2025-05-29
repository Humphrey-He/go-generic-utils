// Copyright 2024 ecodeclub
//
// 本文件为 mapx.go 的示例文件，演示泛型Map、线程安全Map、链表Map、TreeMap、多值Map等常用用法。

package maputils

import (
	"fmt"
	"sort"
)

// ExampleGenericMap 演示泛型Map的基本用法
func ExampleGenericMap() {
	m := NewGenericMap[string, int]()
	m.Set("apple", 10)
	m.Set("banana", 20)
	val, ok := m.Get("apple")
	fmt.Println("apple:", val, ok)
	m.Delete("banana")
	fmt.Println("keys:", m.Keys())
	fmt.Println("values:", m.Values())
	// Output:
	// apple: 10 true
	// keys: [apple]
	// values: [10]
}

// ExampleSyncMap 演示线程安全Map的基本用法
func ExampleSyncMap() {
	m := NewSyncMap[int, string]()
	m.Set(1, "one")
	m.Set(2, "two")
	val, ok := m.Get(2)
	fmt.Println("2:", val, ok)
	m.Delete(1)
	fmt.Println("keys:", m.Keys())
	// Output:
	// 2: two true
	// keys: [2]
}

// ExampleLinkedMap 演示链表Map的插入顺序
func ExampleLinkedMap() {
	m := NewLinkedMap[string, int]()
	m.Set("x", 100)
	m.Set("y", 200)
	m.Set("z", 300)
	fmt.Println("keys:", m.Keys())
	m.Delete("y")
	fmt.Println("keys after delete:", m.Keys())
	// Output:
	// keys: [x y z]
	// keys after delete: [x z]
}

// ExampleTreeMap 演示TreeMap的有序特性
func ExampleTreeMap() {
	m := NewTreeMap[int, string]()
	m.Set(3, "c")
	m.Set(1, "a")
	m.Set(2, "b")
	keys := m.Keys()
	sort.Ints(keys)
	for _, k := range keys {
		v, _ := m.Get(k)
		fmt.Printf("%d:%s ", k, v)
	}
	fmt.Println()
	// Output:
	// 1:a 2:b 3:c
}

// ExampleMultiMap 演示多值Map的用法
func ExampleMultiMap() {
	m := NewMultiMap[string, int]()
	m.Add("fruit", 1)
	m.Add("fruit", 2)
	m.Add("veg", 3)
	fmt.Println("fruit:", m.Get("fruit"))
	fmt.Println("veg:", m.Get("veg"))
	// Output:
	// fruit: [1 2]
	// veg: [3]
}
