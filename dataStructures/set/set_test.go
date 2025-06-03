package set

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetx_Add(t *testing.T) {
	Addvals := []int{1, 2, 3, 1}
	s := NewMapSet[int](10)
	t.Run("Add", func(t *testing.T) {
		for _, val := range Addvals {
			s.Add(val)
		}
		assert.Equal(t, s.m, map[int]struct{}{
			1: struct{}{},
			2: struct{}{},
			3: struct{}{},
		})
	})
}

func TestSetx_Delete(t *testing.T) {
	testcases := []struct {
		name    string
		delVal  int
		setSet  map[int]struct{}
		wantSet map[int]struct{}
		isExist bool
	}{
		{
			name:   "delete val ",
			delVal: 2,
			setSet: map[int]struct{}{
				2: struct{}{},
			},
			wantSet: map[int]struct{}{},
			isExist: true,
		},
		{
			name:   "deleted val not found",
			delVal: 3,
			setSet: map[int]struct{}{
				2: struct{}{},
			},
			wantSet: map[int]struct{}{
				2: struct{}{},
			},
			isExist: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMapSet[int](10)
			s.m = tc.setSet
			s.Delete(tc.delVal)
			assert.Equal(t, tc.wantSet, s.m)
		})
	}
}

func TestSetx_IsExist(t *testing.T) {
	s := NewMapSet[int](10)
	s.Add(1)
	testcases := []struct {
		name    string
		val     int
		isExist bool
	}{
		{
			name:    "found",
			val:     1,
			isExist: true,
		},
		{
			name:    "not fonud",
			val:     2,
			isExist: false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ok := s.Exist(tc.val)
			assert.Equal(t, tc.isExist, ok)
		})
	}
}

func TestSetx_Values(t *testing.T) {
	s := NewMapSet[int](10)
	testcases := []struct {
		name    string
		setSet  map[int]struct{}
		wantval map[int]struct{}
	}{
		{
			name: "found values",
			setSet: map[int]struct{}{
				1: struct{}{},
				2: struct{}{},
				3: struct{}{},
			},
			wantval: map[int]struct{}{
				1: struct{}{},
				2: struct{}{},
				3: struct{}{},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s.m = tc.setSet
			vals := s.Keys()
			ok := equal(vals, tc.wantval)
			assert.Equal(t, true, ok)
		})
	}
}

func TestMapSet(t *testing.T) {
	t.Run("初始化与基本操作", func(t *testing.T) {
		// 创建集合
		set := NewMapSet[string](0)
		assert.NotNil(t, set, "NewMapSet应返回非nil的Set实例")
		assert.Equal(t, 0, set.Len(), "新集合的长度应为0")

		// 添加元素
		set.Add("one")
		set.Add("two")
		set.Add("three")
		set.Add("one") // 重复添加

		// 检查长度
		assert.Equal(t, 3, set.Len(), "集合应包含3个元素")

		// 检查存在
		assert.True(t, set.Exist("one"), "元素'one'应存在")
		assert.True(t, set.Exist("two"), "元素'two'应存在")
		assert.True(t, set.Exist("three"), "元素'three'应存在")
		assert.False(t, set.Exist("four"), "元素'four'不应存在")

		// 获取所有元素
		keys := set.Keys()
		assert.Equal(t, 3, len(keys), "Keys()应返回3个元素")
		assert.ElementsMatch(t, []string{"one", "two", "three"}, keys, "Keys()应返回所有添加的元素")

		// 删除元素
		set.Delete("two")
		assert.Equal(t, 2, set.Len(), "删除后集合应包含2个元素")
		assert.False(t, set.Exist("two"), "元素'two'应已被删除")

		// 清空集合
		set.Clear()
		assert.Equal(t, 0, set.Len(), "Clear()后集合应为空")
		assert.False(t, set.Exist("one"), "Clear()后元素'one'不应存在")
	})

	t.Run("AddIfNotExist", func(t *testing.T) {
		set := NewMapSet[int](0)

		// 添加不存在的元素
		added := set.AddIfNotExist(1)
		assert.True(t, added, "AddIfNotExist应返回true表示元素被添加")
		assert.True(t, set.Exist(1), "元素1应存在")

		// 添加已存在的元素
		added = set.AddIfNotExist(1)
		assert.False(t, added, "AddIfNotExist应返回false表示元素未被添加")
		assert.Equal(t, 1, set.Len(), "集合长度应仍为1")
	})

	t.Run("ForEach", func(t *testing.T) {
		set := NewMapSet[int](0)
		set.Add(1)
		set.Add(2)
		set.Add(3)

		sum := 0
		set.ForEach(func(key int) bool {
			sum += key
			return true // 继续遍历
		})
		assert.Equal(t, 6, sum, "ForEach应遍历所有元素")

		count := 0
		set.ForEach(func(key int) bool {
			count++
			return key != 2 // 遇到2时停止
		})
		assert.True(t, count < 3, "ForEach应在返回false时停止遍历")
	})

	t.Run("集合操作", func(t *testing.T) {
		set1 := NewMapSet[int](0)
		set2 := NewMapSet[int](0)

		// 添加元素
		for i := 1; i <= 5; i++ {
			set1.Add(i)
		}

		for i := 3; i <= 7; i++ {
			set2.Add(i)
		}

		// 并集
		union := set1.Union(set2)
		assert.Equal(t, 7, union.Len(), "并集应包含7个元素")
		for i := 1; i <= 7; i++ {
			assert.True(t, union.Exist(i), "并集应包含元素"+string(rune(i+'0')))
		}

		// 交集
		intersect := set1.Intersect(set2)
		assert.Equal(t, 3, intersect.Len(), "交集应包含3个元素")
		for i := 3; i <= 5; i++ {
			assert.True(t, intersect.Exist(i), "交集应包含元素"+string(rune(i+'0')))
		}

		// 差集
		diff := set1.Difference(set2)
		assert.Equal(t, 2, diff.Len(), "差集应包含2个元素")
		assert.True(t, diff.Exist(1), "差集应包含元素1")
		assert.True(t, diff.Exist(2), "差集应包含元素2")

		// 子集判断
		assert.False(t, set1.IsSubsetOf(set2), "set1不应是set2的子集")
		assert.False(t, set2.IsSubsetOf(set1), "set2不应是set1的子集")

		subset := NewMapSet[int](0)
		subset.Add(1)
		subset.Add(2)
		assert.True(t, subset.IsSubsetOf(set1), "subset应是set1的子集")
	})

	t.Run("ToSlice和ToSortedSlice", func(t *testing.T) {
		set := NewMapSet[int](0)
		set.Add(3)
		set.Add(1)
		set.Add(2)

		// ToSlice
		slice := set.ToSlice()
		assert.Equal(t, 3, len(slice), "ToSlice应返回所有元素")
		assert.ElementsMatch(t, []int{1, 2, 3}, slice, "ToSlice应包含所有元素")

		// ToSortedSlice
		sortedSlice := ToSortedSlice(set)
		assert.Equal(t, 3, len(sortedSlice), "ToSortedSlice应返回所有元素")
		assert.Equal(t, []int{1, 2, 3}, sortedSlice, "ToSortedSlice应返回排序后的元素")
	})
}

func TestTreeSet(t *testing.T) {
	t.Run("创建与基本操作", func(t *testing.T) {
		// 创建比较器
		intComparator := ComparatorRealNumber[int]()

		// 创建TreeSet
		set, err := NewTreeSet[int](intComparator)
		assert.NoError(t, err, "NewTreeSet不应返回错误")
		assert.NotNil(t, set, "NewTreeSet应返回非nil的TreeSet实例")

		// 空比较器测试
		_, err = NewTreeSet[int](nil)
		assert.Error(t, err, "空比较器应返回错误")
		assert.Equal(t, ErrNilComparator, err, "应返回ErrNilComparator错误")

		// 添加元素
		set.Add(3)
		set.Add(1)
		set.Add(5)
		set.Add(2)
		set.Add(4)

		// 检查元素顺序和数量
		assert.Equal(t, 5, set.Len(), "集合长度应为5")

		// Keys应返回有序数组
		keys := set.Keys()
		assert.Equal(t, []int{1, 2, 3, 4, 5}, keys, "Keys应返回有序元素")

		// 检查元素存在
		assert.True(t, set.Exist(3), "元素3应存在")
		assert.False(t, set.Exist(6), "元素6不应存在")

		// 删除元素
		set.Delete(3)
		assert.Equal(t, 4, set.Len(), "删除后长度应为4")
		assert.False(t, set.Exist(3), "元素3应已被删除")

		// 清空集合
		set.Clear()
		assert.Equal(t, 0, set.Len(), "Clear后集合应为空")
	})

	t.Run("AddIfNotExist", func(t *testing.T) {
		comparator := ComparatorString()
		set, _ := NewTreeSet[string](comparator)

		added := set.AddIfNotExist("one")
		assert.True(t, added, "首次添加应返回true")

		added = set.AddIfNotExist("one")
		assert.False(t, added, "重复添加应返回false")

		assert.Equal(t, 1, set.Len(), "集合长度应为1")
	})

	t.Run("ForEach", func(t *testing.T) {
		comparator := ComparatorRealNumber[int]()
		set, _ := NewTreeSet[int](comparator)

		set.Add(3)
		set.Add(1)
		set.Add(2)

		// 集合遍历
		elements := make([]int, 0)
		set.ForEach(func(key int) bool {
			elements = append(elements, key)
			return true
		})

		assert.Equal(t, []int{1, 2, 3}, elements, "ForEach应按顺序遍历元素")

		// 提前终止遍历
		limitedElements := make([]int, 0)
		set.ForEach(func(key int) bool {
			limitedElements = append(limitedElements, key)
			return key != 2 // 当遇到2时停止
		})

		assert.Equal(t, []int{1, 2}, limitedElements, "ForEach应在返回false时停止遍历")
	})
}

func TestConcurrentSet(t *testing.T) {
	t.Run("创建与基本操作", func(t *testing.T) {
		set := NewConcurrentSet[string](0)
		assert.NotNil(t, set, "NewConcurrentSet应返回非nil的ConcurrentSet实例")

		// 添加元素
		set.Add("a")
		set.Add("b")
		set.Add("c")

		// 检查长度
		assert.Equal(t, 3, set.Len(), "集合长度应为3")

		// 检查元素存在
		assert.True(t, set.Exist("a"), "元素'a'应存在")
		assert.False(t, set.Exist("d"), "元素'd'不应存在")

		// 获取所有元素
		keys := set.Keys()
		assert.Len(t, keys, 3, "应返回3个元素")
		assert.ElementsMatch(t, []string{"a", "b", "c"}, keys, "Keys应返回所有元素")

		// 删除元素
		set.Delete("b")
		assert.Equal(t, 2, set.Len(), "删除后长度应为2")
		assert.False(t, set.Exist("b"), "元素'b'应已被删除")

		// 清空集合
		set.Clear()
		assert.Equal(t, 0, set.Len(), "Clear后集合应为空")
	})

	t.Run("AddIfNotExist", func(t *testing.T) {
		set := NewConcurrentSet[int](0)

		added := set.AddIfNotExist(1)
		assert.True(t, added, "首次添加应返回true")

		added = set.AddIfNotExist(1)
		assert.False(t, added, "重复添加应返回false")
	})

	t.Run("ForEach", func(t *testing.T) {
		set := NewConcurrentSet[int](0)
		for i := 1; i <= 5; i++ {
			set.Add(i)
		}

		// 集合遍历
		sum := 0
		set.ForEach(func(key int) bool {
			sum += key
			return true
		})

		assert.Equal(t, 15, sum, "ForEach应遍历所有元素")

		// 提前终止遍历
		visited := make([]int, 0)
		stopValue := 3 // 当遇到这个值时停止

		set.ForEach(func(key int) bool {
			visited = append(visited, key)
			return key != stopValue // 当遇到stopValue时停止
		})

		// 验证stopValue应该在visited中，且应该是最后一个元素
		assert.Contains(t, visited, stopValue, "应该访问到停止值")
		assert.Equal(t, stopValue, visited[len(visited)-1], "停止值应该是最后一个访问的元素")

		// 验证没有遍历完所有元素
		assert.Less(t, len(visited), 5, "ForEach应在返回false时停止遍历")
	})
}

func TestExpirableSet(t *testing.T) {
	t.Run("创建与基本操作", func(t *testing.T) {
		set := NewExpirableSet[string](500 * time.Millisecond)
		assert.NotNil(t, set, "NewExpirableSet应返回非nil的ExpirableSet实例")

		// 添加元素(永不过期)
		set.Add("a")
		set.Add("b")

		// 添加带TTL的元素
		set.AddWithTTL("c", 100*time.Millisecond)

		// 检查长度
		assert.Equal(t, 3, set.Len(), "集合长度应为3")

		// 检查元素存在
		assert.True(t, set.Exist("a"), "元素'a'应存在")
		assert.True(t, set.Exist("c"), "元素'c'应存在")

		// 等待短暂过期的元素过期
		time.Sleep(150 * time.Millisecond)

		// 再次检查
		assert.False(t, set.Exist("c"), "元素'c'应已过期")
		assert.Equal(t, 2, set.Len(), "过期后集合长度应为2")

		// GetTTL测试
		ttl := set.GetTTL("a")
		assert.True(t, ttl > 0, "永不过期元素的TTL应大于0")

		ttl = set.GetTTL("not_exist")
		assert.Equal(t, time.Duration(-1), ttl, "不存在元素的TTL应为-1")

		// 删除元素
		set.Delete("a")
		assert.False(t, set.Exist("a"), "元素'a'应已被删除")

		// 清空集合
		set.Clear()
		assert.Equal(t, 0, set.Len(), "Clear后集合应为空")

		// 关闭清理协程
		set.Close()
	})

	t.Run("AddIfNotExist", func(t *testing.T) {
		set := NewExpirableSet[int](500 * time.Millisecond)

		added := set.AddIfNotExist(1)
		assert.True(t, added, "首次添加应返回true")

		added = set.AddIfNotExist(1)
		assert.False(t, added, "重复添加应返回false")

		// 测试过期后再添加
		set.AddWithTTL(2, 100*time.Millisecond)
		time.Sleep(150 * time.Millisecond)

		added = set.AddIfNotExist(2)
		assert.True(t, added, "元素过期后再添加应返回true")

		set.Close()
	})

	t.Run("ForEach", func(t *testing.T) {
		set := NewExpirableSet[int](500 * time.Millisecond)

		// 添加不同过期时间的元素
		set.Add(1)                              // 永不过期
		set.Add(2)                              // 永不过期
		set.AddWithTTL(3, 100*time.Millisecond) // 快速过期
		set.AddWithTTL(4, 1*time.Hour)          // 长时间过期

		// 立即遍历
		count := 0
		set.ForEach(func(key int) bool {
			count++
			return true
		})
		assert.Equal(t, 4, count, "ForEach应遍历所有未过期元素")

		// 等待快速过期的元素过期
		time.Sleep(150 * time.Millisecond)

		// 再次遍历
		countAfterExpire := 0
		set.ForEach(func(key int) bool {
			countAfterExpire++
			return true
		})
		assert.Equal(t, 3, countAfterExpire, "ForEach应只遍历未过期元素")

		set.Close()
	})
}

func equal(nums []int, m map[int]struct{}) bool {
	for _, num := range nums {
		_, ok := m[num]
		if !ok {
			return false
		}
		delete(m, num)
	}
	return true && len(m) == 0
}

// goos: linux
// goarch: amd64
// pkg: github.com/gotomicro/ggu/set
// cpu: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz
// BenchmarkSet/set_add-8            178898              6504 ns/op             210 B/op          5 allocs/op
// BenchmarkSet/map_add-8            176377              6446 ns/op             210 B/op          5 allocs/op
// BenchmarkSet/set_del-8            271983              4437 ns/op               0 B/op          0 allocs/op
// BenchmarkSet/map_del-8            289152              4143 ns/op               0 B/op          0 allocs/op
// BenchmarkSet/set_exist-8          348619              3408 ns/op               0 B/op          0 allocs/op
// BenchmarkSet/map_exist-8          403066              3061 ns/op               0 B/op          0 allocs/op

func BenchmarkSet(b *testing.B) {
	b.Run("set_add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			s := NewMapSet[int](100)
			b.StartTimer()
			setadd(s)
		}
	})
	b.Run("map_add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := make(map[int]struct{}, 100)
			b.StartTimer()
			mapadd(m)
		}
	})
	b.Run("set_del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			s := NewMapSet[int](100)
			setadd(s)
			b.StartTimer()
			setdel(s)
		}
	})
	b.Run("map_del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := make(map[int]struct{}, 100)
			mapadd(m)
			b.StartTimer()
			mapdel(m)
		}
	})
	b.Run("set_exist", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			s := NewMapSet[int](100)
			setadd(s)
			b.StartTimer()
			setGet(s)
		}
	})
	b.Run("map_exist", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			m := make(map[int]struct{}, 100)
			mapadd(m)
			b.StartTimer()
			mapGet(m)
		}
	})

}

func setadd(s Set[int]) {
	for i := 0; i < 100; i++ {
		s.Add(i)
	}
}

func mapadd(m map[int]struct{}) {
	for i := 0; i < 100; i++ {
		m[i] = struct{}{}
	}
}

func setdel(s Set[int]) {
	for i := 0; i < 100; i++ {
		s.Delete(i)
	}
}

func mapdel(m map[int]struct{}) {
	for i := 0; i < 100; i++ {
		delete(m, i)
	}
}
func setGet(s Set[int]) {
	for i := 0; i < 100; i++ {
		_ = s.Exist(i)
	}
}

func mapGet(s map[int]struct{}) {
	for i := 0; i < 100; i++ {
		_ = s[i]
	}
}
