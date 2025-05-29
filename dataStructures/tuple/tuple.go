// Package tuple 提供了一系列通用的元组数据结构和操作函数。
// 元组可以用于在Go语言中表示不同类型的数据对，特别适合电商平台等场景中表示数据关联关系。
package tuple

import (
	"encoding/json"
	"fmt"
)

// Pair 表示一个键值对，可以容纳任意类型的键和值
// 适用于：商品与价格的映射、用户与订单的关联、商品与库存的对应关系等
type Pair[K any, V any] struct {
	Key   K
	Value V
}

// NewPair 创建一个新的键值对
func NewPair[K any, V any](key K, value V) Pair[K, V] {
	return Pair[K, V]{
		Key:   key,
		Value: value,
	}
}

// String 返回Pair的字符串表示
func (p Pair[K, V]) String() string {
	return fmt.Sprintf("<%v, %v>", p.Key, p.Value)
}

// Split 方法将Key, Value作为返回参数传出
func (p Pair[K, V]) Split() (K, V) {
	return p.Key, p.Value
}

// MarshalJSON 实现JSON序列化接口
func (p Pair[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"key":   p.Key,
		"value": p.Value,
	})
}

// NewPairs 从键数组和值数组创建键值对数组
// 例如：将商品ID列表和对应的价格列表转换为商品-价格对列表
func NewPairs[K any, V any](keys []K, values []V) ([]Pair[K, V], error) {
	if keys == nil || values == nil {
		return nil, fmt.Errorf("keys与values均不可为nil")
	}

	n := len(keys)
	if n != len(values) {
		return nil, fmt.Errorf("keys与values的长度不同, len(keys)=%d, len(values)=%d", n, len(values))
	}

	pairs := make([]Pair[K, V], n)
	for i := 0; i < n; i++ {
		pairs[i] = NewPair(keys[i], values[i])
	}
	return pairs, nil
}

// SplitPairs 将键值对数组拆分为键数组和值数组
// 例如：将商品-价格对列表拆分为商品ID列表和价格列表
func SplitPairs[K any, V any](pairs []Pair[K, V]) (keys []K, values []V) {
	if pairs == nil {
		return nil, nil
	}

	n := len(pairs)
	keys = make([]K, n)
	values = make([]V, n)

	for i, pair := range pairs {
		keys[i], values[i] = pair.Split()
	}
	return
}

// FlattenPairs 将键值对数组展平为一个扁平数组
func FlattenPairs[K any, V any](pairs []Pair[K, V]) []any {
	if pairs == nil {
		return nil
	}

	n := len(pairs)
	flatPairs := make([]any, 0, n*2)

	for _, pair := range pairs {
		flatPairs = append(flatPairs, pair.Key, pair.Value)
	}
	return flatPairs
}

// PackPairs 将扁平数组重新打包为键值对数组
func PackPairs[K any, V any](flatPairs []any) []Pair[K, V] {
	if flatPairs == nil {
		return nil
	}

	if len(flatPairs)%2 != 0 {
		panic("扁平数组长度必须为偶数")
	}

	n := len(flatPairs) / 2
	pairs := make([]Pair[K, V], n)

	for i := 0; i < n; i++ {
		k, ok1 := flatPairs[i*2].(K)
		v, ok2 := flatPairs[i*2+1].(V)

		if !ok1 || !ok2 {
			panic(fmt.Sprintf("类型转换失败，位置 %d: 无法将 %T 转换为 %T 或将 %T 转换为 %T",
				i, flatPairs[i*2], *new(K), flatPairs[i*2+1], *new(V)))
		}

		pairs[i] = NewPair(k, v)
	}
	return pairs
}

// Triple 表示一个三元组，可以容纳三个任意类型的值
// 适用于：商品ID-名称-价格、订单ID-用户ID-状态等多维度数据表示
type Triple[T1 any, T2 any, T3 any] struct {
	First  T1
	Second T2
	Third  T3
}

// NewTriple 创建一个新的三元组
func NewTriple[T1, T2, T3 any](first T1, second T2, third T3) Triple[T1, T2, T3] {
	return Triple[T1, T2, T3]{
		First:  first,
		Second: second,
		Third:  third,
	}
}

// String 返回Triple的字符串表示
func (t Triple[T1, T2, T3]) String() string {
	return fmt.Sprintf("<%v, %v, %v>", t.First, t.Second, t.Third)
}

// Split 方法将三个值作为返回参数传出
func (t Triple[T1, T2, T3]) Split() (T1, T2, T3) {
	return t.First, t.Second, t.Third
}

// MarshalJSON 实现JSON序列化接口
func (t Triple[T1, T2, T3]) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"first":  t.First,
		"second": t.Second,
		"third":  t.Third,
	})
}

// KeyValue 是Pair的一个别名，语义上更适合表示键值对
// 适用于：配置项的键值对、HTTP头部等
type KeyValue[K comparable, V any] Pair[K, V]

// NewKeyValue 创建一个新的键值对
func NewKeyValue[K comparable, V any](key K, value V) KeyValue[K, V] {
	return KeyValue[K, V]{
		Key:   key,
		Value: value,
	}
}

// String 返回KeyValue的字符串表示
func (kv KeyValue[K, V]) String() string {
	return fmt.Sprintf("%v: %v", kv.Key, kv.Value)
}

// MarshalJSON 实现JSON序列化接口
func (kv KeyValue[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"key":   kv.Key,
		"value": kv.Value,
	})
}

// MapFromPairs 从键值对数组创建map
// 例如：将商品ID-价格对列表转换为商品ID到价格的映射
func MapFromPairs[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	result := make(map[K]V, len(pairs))
	for _, p := range pairs {
		result[p.Key] = p.Value
	}
	return result
}

// PairsFromMap 从map创建键值对数组
// 例如：将商品ID到价格的映射转换为商品ID-价格对列表
func PairsFromMap[K comparable, V any](m map[K]V) []Pair[K, V] {
	result := make([]Pair[K, V], 0, len(m))
	for k, v := range m {
		result = append(result, NewPair(k, v))
	}
	return result
}

// Range 遍历键值对数组并对每个元素执行指定函数
func Range[K any, V any](pairs []Pair[K, V], fn func(K, V) error) error {
	for _, p := range pairs {
		if err := fn(p.Key, p.Value); err != nil {
			return err
		}
	}
	return nil
}

// Filter 过滤键值对数组，只保留满足条件的元素
func Filter[K any, V any](pairs []Pair[K, V], predicate func(K, V) bool) []Pair[K, V] {
	result := make([]Pair[K, V], 0)
	for _, p := range pairs {
		if predicate(p.Key, p.Value) {
			result = append(result, p)
		}
	}
	return result
}

// Map 对键值对数组中的每个元素应用转换函数，返回转换后的新数组
func Map[K1 any, V1 any, K2 any, V2 any](pairs []Pair[K1, V1], mapper func(K1, V1) (K2, V2)) []Pair[K2, V2] {
	result := make([]Pair[K2, V2], len(pairs))
	for i, p := range pairs {
		k2, v2 := mapper(p.Key, p.Value)
		result[i] = NewPair(k2, v2)
	}
	return result
}

// Reduce 对键值对数组执行归约操作，将多个键值对合并为单个结果
func Reduce[K any, V any, R any](pairs []Pair[K, V], initialValue R, reducer func(R, K, V) R) R {
	result := initialValue
	for _, p := range pairs {
		result = reducer(result, p.Key, p.Value)
	}
	return result
}
