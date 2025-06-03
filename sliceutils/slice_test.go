// Copyright 2024 Humphrey-He
// sliceutils 包的测试文件。

package sliceutils

import (
	// 导入 errors 包，用于创建错误对象

	"math/rand"
	"reflect" // 导入 reflect 包，用于运行时反射操作，如 DeepEqual
	"sort"    // 导入 sort 包，用于切片排序
	"strconv"
	"strings"
	"sync" // 导入 sync 包，提供同步原语，如 Mutex
	"testing"
	"time"
)

// Helper: testStruct 是用于测试泛型函数的自定义结构体。
type testStruct struct {
	ID   int
	Name string
}

// Helper: testStructEqual 根据 ID 和 Name 为 testStruct 实例提供自定义的相等比较。
func testStructEqual(a, b testStruct) bool {
	return a.ID == b.ID && a.Name == b.Name
}

// Helper: testStructIDMatcher 返回一个 matchFunc 用于通过 ID 查找 testStruct。
func testStructIDMatcher(id int) matchFunc[testStruct] {
	return func(s testStruct) bool {
		return s.ID == id
	}
}

// Helper: compareSlices 是一个通用的切片比较工具函数，用于测试。
// 适用于元素顺序重要且元素可比较的切片。
func compareSlices[T comparable](t *testing.T, got, want []T, context string) {
	t.Helper() // 标记为测试辅助函数
	if len(got) != len(want) {
		t.Errorf("%s: 长度不匹配: got %d, want %d. Got: %v, Want: %v", context, len(got), len(want), got, want)
		return
	}
	// 处理 nil 与空切片：在许多测试用例中视作相等。
	if (got == nil && len(want) == 0) || (want == nil && len(got) == 0) {
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s: 索引 %d 处元素不匹配: got %v, want %v. GotSlice: %v, WantSlice: %v", context, i, got[i], want[i], got, want)
			return
		}
	}
}

// Helper: compareSlicesFunc 是一个通用的切片比较工具函数，用于测试，使用自定义的相等比较函数。
func compareSlicesFunc[T any](t *testing.T, got, want []T, eq equalFunc[T], context string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: 长度不匹配: got %d, want %d. Got: %v, Want: %v", context, len(got), len(want), got, want)
		return
	}
	if (got == nil && len(want) == 0) || (want == nil && len(got) == 0) {
		return
	}
	for i := range got {
		if !eq(got[i], want[i]) {
			t.Errorf("%s: 索引 %d 处元素使用自定义 equalFunc 比较不匹配: got %v, want %v. GotSlice: %v, WantSlice: %v", context, i, got[i], want[i], got, want)
			return
		}
	}
}

// Helper: compareSlicesIgnoreOrder 用于比较顺序无关的切片（例如集合操作的结果）。
// 它在比较前对两个切片的副本进行排序。仅适用于可比较类型。
func compareSlicesIgnoreOrder[T interface {
	~int | ~string | ~float64 // 根据需要添加其他有序类型
}](t *testing.T, got, want []T, context string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: 长度不匹配 (忽略顺序): got %d, want %d. Got: %v, Want: %v", context, len(got), len(want), got, want)
		return
	}
	if (got == nil && len(want) == 0) || (want == nil && len(got) == 0) {
		return
	}

	gotCopy := append([]T(nil), got...)
	wantCopy := append([]T(nil), want...)

	// 基于类型的通用排序
	switch any(gotCopy).(type) {
	case []int:
		sort.Ints(any(gotCopy).([]int))
		sort.Ints(any(wantCopy).([]int))
	case []string:
		sort.Strings(any(gotCopy).([]string))
		sort.Strings(any(wantCopy).([]string))
	case []float64:
		sort.Float64s(any(gotCopy).([]float64))
		sort.Float64s(any(wantCopy).([]float64))
	default:
		// 对于其他类型，如果顺序真的不重要，reflect.DeepEqual 可能适用，
		// 或者需要更复杂的比较逻辑。为简单起见，此辅助函数功能受限。
		// 对其他类型回退到 reflect.DeepEqual，但这不会对它们进行排序。
		if !reflect.DeepEqual(gotCopy, wantCopy) && len(gotCopy) > 0 { // 仅当尚未排序时
			t.Logf("%s: 切片副本在没有特定排序的情况下无法直接比较。如果长度匹配，则依赖元素存在性进行判断。", context)
			// 这个简化版本检查 got 中的所有元素是否都在 want 中，反之亦然。
			// 更健壮的方法是将两者都转换为元素计数的 map。
			for _, gVal := range gotCopy {
				found := false
				for _, wVal := range wantCopy {
					if reflect.DeepEqual(gVal, wVal) { // 假设 T 是可比较的或具有深层相等的意义
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: got 中的元素 %v 在 want 中未找到 (忽略顺序). Got: %v, Want: %v", context, gVal, got, want)
					return
				}
			}
			return // 如果所有元素都找到且长度匹配
		}
	}

	if !reflect.DeepEqual(gotCopy, wantCopy) {
		t.Errorf("%s: 排序后不匹配 (忽略顺序): got_sorted %v, want_sorted %v. Original Got: %v, Original Want: %v", context, gotCopy, wantCopy, got, want)
	}
}

// --- ThreadSafeSlice 测试 ---

func TestNewThreadSafeSlice(t *testing.T) {
	t.Run("nil 输入", func(t *testing.T) {
		s := NewThreadSafeSlice[int](nil)
		if s.slice == nil { // NewThreadSafeSlice 会初始化为空切片，而不是 nil
			t.Errorf("对于 nil 输入，期望内部切片非 nil，实际为 nil")
		}
		if len(s.slice) != 0 {
			t.Errorf("对于 nil 输入，期望内部切片为空，实际为 %v", s.slice)
		}
	})

	t.Run("空切片输入", func(t *testing.T) {
		s := NewThreadSafeSlice([]int{})
		if len(s.slice) != 0 {
			t.Errorf("对于空切片输入，期望内部切片为空，实际为 %v", s.slice)
		}
	})

	t.Run("带初始元素", func(t *testing.T) {
		init := []int{1, 2, 3}
		s := NewThreadSafeSlice(init)
		compareSlices(t, s.slice, init, "NewThreadSafeSlice with elements")
		// 确保是副本
		init[0] = 99
		if s.slice[0] == 99 {
			t.Error("NewThreadSafeSlice 未创建副本")
		}
	})
}

func TestThreadSafeSlice_Append(t *testing.T) {
	s := NewThreadSafeSlice([]int{1})
	s.Append(2, 3)
	want := []int{1, 2, 3}
	compareSlices(t, s.AsSlice(), want, "Append basic")

	// 并发追加
	sc := NewThreadSafeSlice[int](nil)
	var wg sync.WaitGroup
	numAppends := 100
	for i := 0; i < numAppends; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			sc.Append(val)
		}(i)
	}
	wg.Wait()
	if sc.Len() != numAppends {
		t.Errorf("并发 Append: 期望长度 %d, 实际为 %d", numAppends, sc.Len())
	}
}

func TestThreadSafeSlice_Delete(t *testing.T) {
	testCases := []struct {
		name      string
		initial   []string
		index     int
		wantSlice []string
		wantErr   bool
	}{
		{"有效中间索引", []string{"a", "b", "c"}, 1, []string{"a", "c"}, false},
		{"有效起始索引", []string{"a", "b", "c"}, 0, []string{"b", "c"}, false},
		{"有效末尾索引", []string{"a", "b", "c"}, 2, []string{"a", "b"}, false},
		{"越界负索引", []string{"a"}, -1, []string{"a"}, true},
		{"越界大索引", []string{"a"}, 1, []string{"a"}, true},
		{"空切片", []string{}, 0, []string{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewThreadSafeSlice(tc.initial)
			err := s.Delete(tc.index)
			if (err != nil) != tc.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				compareSlices(t, s.AsSlice(), tc.wantSlice, "Delete 内容检查")
			} else { // 如果期望错误，切片应保持不变
				compareSlices(t, s.AsSlice(), tc.initial, "Delete 错误时内容检查")
			}
		})
	}
}

func TestThreadSafeSlice_Get(t *testing.T) {
	s := NewThreadSafeSlice([]int{10, 20, 30})
	val, err := s.Get(1)
	if err != nil || val != 20 {
		t.Errorf("Get(1) 期望 (20, nil错误), 实际为 (%d, %v)", val, err)
	}

	_, err = s.Get(3) // 越界
	if err == nil {
		t.Error("Get() 越界索引期望错误，实际为 nil")
	}
	if !strings.Contains(err.Error(), "下标越界") {
		t.Errorf("Get() 期望错误信息包含 '下标越界', 实际为 '%s'", err.Error())
	}

	_, err = s.Get(-1) // 越界
	if err == nil {
		t.Error("Get() 负索引期望错误，实际为 nil")
	}
}

func TestThreadSafeSlice_Set(t *testing.T) {
	s := NewThreadSafeSlice([]string{"x", "y", "z"})
	err := s.Set(1, "Y")
	if err != nil {
		t.Errorf("Set(1, 'Y') 返回非期望错误: %v", err)
	}
	compareSlices(t, s.AsSlice(), []string{"x", "Y", "z"}, "Set 内容检查")

	err = s.Set(3, "W") // 越界
	if err == nil {
		t.Error("Set() 越界索引期望错误，实际为 nil")
	}
	// 确保错误时切片未被修改
	compareSlices(t, s.AsSlice(), []string{"x", "Y", "z"}, "Set 错误时内容检查")
}

func TestThreadSafeSlice_AsSlice(t *testing.T) {
	initial := []int{1, 2, 3}
	s := NewThreadSafeSlice(initial)
	copied := s.AsSlice()
	compareSlices(t, copied, initial, "AsSlice 内容")

	// 确保是真正的副本
	if len(copied) > 0 {
		copied[0] = 100
		val, _ := s.Get(0)
		if val == 100 {
			t.Error("AsSlice 未返回副本；修改影响了原始 ThreadSafeSlice。")
		}
	}
	sEmpty := NewThreadSafeSlice[int](nil)
	if sEmpty.AsSlice() == nil { // AsSlice 对于 nil 内部切片应返回空切片，而非 nil
		t.Error("AsSlice 对于 nil 内部切片应返回空切片，而非 nil")
	}
}

func TestThreadSafeSlice_Len(t *testing.T) {
	if NewThreadSafeSlice[int](nil).Len() != 0 {
		t.Error("Len() 对于 nil 初始化期望 0")
	}
	if NewThreadSafeSlice([]int{}).Len() != 0 {
		t.Error("Len() 对于空切片初始化期望 0")
	}
	if NewThreadSafeSlice([]int{1, 2}).Len() != 2 {
		t.Error("Len() 对于非空初始化期望 2")
	}
}

// TestThreadSafeSlice_ConcurrentOps 并发执行多种操作。
func TestThreadSafeSlice_ConcurrentOps(t *testing.T) {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck // 在测试中使用 math/rand 的 Seed 是可接受的
	s := NewThreadSafeSlice[int](nil)
	var wg sync.WaitGroup
	numGoroutines := 50
	opsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				op := rand.Intn(5) // 0:Append, 1:Get, 2:Set, 3:Delete, 4:Len/AsSlice
				switch op {
				case 0:
					s.Append(routineID*1000 + j)
				case 1:
					l := s.Len()
					if l > 0 {
						_, _ = s.Get(rand.Intn(l))
					}
				case 2:
					l := s.Len()
					if l > 0 {
						_ = s.Set(rand.Intn(l), -(routineID*1000 + j))
					}
				case 3:
					l := s.Len()
					if l > 0 {
						_ = s.Delete(rand.Intn(l))
					}
				case 4:
					_ = s.Len()
					_ = s.AsSlice()
				}
			}
		}(i)
	}
	wg.Wait()
	// 由于随机性，不针对最终状态进行特定检查，
	// 此测试主要旨在捕获竞争条件（使用 -race 运行）。
	t.Logf("ThreadSafeSlice_ConcurrentOps 完成。最终长度: %d", s.Len())
}

// --- 查找操作测试 ---

func TestFind(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}
	t.Run("找到元素", func(t *testing.T) {
		val, found := Find(src, func(x int) bool { return x == 3 })
		if !found || val != 3 {
			t.Errorf("Find 期望 (3, true), 实际为 (%d, %v)", val, found)
		}
	})
	t.Run("未找到元素", func(t *testing.T) {
		_, found := Find(src, func(x int) bool { return x == 6 })
		if found {
			t.Error("Find 期望未找到 (false), 但实际找到了")
		}
	})
	t.Run("空切片", func(t *testing.T) {
		_, found := Find([]int{}, func(x int) bool { return x == 1 })
		if found {
			t.Error("Find 对空切片操作期望未找到 (false)")
		}
	})
}

func TestFindAll(t *testing.T) {
	src := []int{1, 2, 3, 2, 4, 2}
	want := []int{2, 2, 2}
	got := FindAll(src, func(x int) bool { return x == 2 })
	compareSlices(t, got, want, "FindAll 查找所有 2")

	t.Run("未找到", func(t *testing.T) {
		gotNone := FindAll(src, func(x int) bool { return x == 5 })
		if len(gotNone) != 0 {
			t.Errorf("FindAll 未找到时期望空切片, 实际为 %v", gotNone)
		}
	})
	t.Run("源切片为空", func(t *testing.T) {
		if len(FindAll([]int{}, func(x int) bool { return x == 1 })) != 0 {
			t.Error("FindAll 对空源切片操作应返回空切片")
		}
	})
}

func TestIndex(t *testing.T) {
	src := []string{"a", "b", "c", "b", "d"}
	if Index(src, "b") != 1 {
		t.Error("Index('b') 失败")
	}
	if Index(src, "d") != 4 {
		t.Error("Index('d') 失败")
	}
	if Index(src, "x") != -1 {
		t.Error("Index('x') (不存在的元素) 失败")
	}
	if Index([]string{}, "a") != -1 {
		t.Error("Index 对空切片操作失败")
	}
}

func TestIndexFunc(t *testing.T) {
	src := []testStruct{{1, "A"}, {2, "B"}, {3, "C"}}
	if IndexFunc(src, testStructIDMatcher(2)) != 1 {
		t.Error("IndexFunc 根据 ID 2 查找失败")
	}
	if IndexFunc(src, testStructIDMatcher(4)) != -1 {
		t.Error("IndexFunc 查找不存在的 ID 失败")
	}
}

func TestLastIndex(t *testing.T) {
	src := []int{1, 2, 3, 2, 1}
	if LastIndex(src, 2) != 3 {
		t.Error("LastIndex(2) 失败")
	}
	if LastIndex(src, 1) != 4 {
		t.Error("LastIndex(1) 失败")
	}
	if LastIndex(src, 5) != -1 {
		t.Error("LastIndex 查找不存在的元素失败")
	}
}

func TestLastIndexFunc(t *testing.T) {
	src := []string{"apple", "banana", "apricot", "avocado"}
	match := func(s string) bool { return strings.HasPrefix(s, "ap") }
	if LastIndexFunc(src, match) != 2 { // "apricot"
		t.Errorf("LastIndexFunc 查找 'ap' 前缀失败, got %d, want 2", LastIndexFunc(src, match))
	}
}

func TestIndexAll(t *testing.T) {
	src := []int{1, 2, 1, 3, 1, 4}
	want := []int{0, 2, 4}
	got := IndexAll(src, 1)
	compareSlices(t, got, want, "IndexAll 查找所有 1")
	if len(IndexAll(src, 5)) != 0 {
		t.Error("IndexAll 查找不存在的元素失败")
	}
}

func TestIndexAllFunc(t *testing.T) {
	src := []testStruct{{1, "A"}, {2, "B"}, {1, "C"}, {3, "D"}}
	want := []int{0, 2}
	got := IndexAllFunc(src, testStructIDMatcher(1))
	compareSlices(t, got, want, "IndexAllFunc 查找所有 ID 为 1 的元素")
}

// --- 普通切片的增删改测试 ---

func TestAdd(t *testing.T) {
	testCases := []struct {
		name    string
		src     []int
		element int
		index   int
		want    []int
		wantErr bool
	}{
		{"中间插入", []int{1, 3}, 2, 1, []int{1, 2, 3}, false},
		{"开头插入", []int{2, 3}, 1, 0, []int{1, 2, 3}, false},
		{"末尾插入", []int{1, 2}, 3, 2, []int{1, 2, 3}, false}, // Add 到 len(src) 是有效的
		{"向空切片插入", []int{}, 1, 0, []int{1}, false},
		{"错误：负索引", []int{1}, 0, -1, nil, true},
		{"错误：索引过大", []int{1}, 0, 2, nil, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 复制 src，因为 Add 可能在容量允许时修改输入，尽管它返回可能是新的切片
			srcCopy := append([]int(nil), tc.src...)
			got, err := Add(srcCopy, tc.element, tc.index)
			if (err != nil) != tc.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				compareSlices(t, got, tc.want, "Add 内容检查")
			}
		})
	}
}

func TestRegularDelete(t *testing.T) { // 重命名以避免与 ThreadSafeSlice 的 Delete 冲突
	// 逻辑类似于 ThreadSafeSlice.Delete，但用于标准切片
	src := []int{1, 2, 3}
	got, err := Delete(src, 1)
	if err != nil {
		t.Fatalf("Delete 出现非期望错误: %v", err)
	}
	compareSlices(t, got, []int{1, 3}, "Delete middle")

	_, err = Delete(src, 3) // 对于原始 src，索引越界
	if err == nil {
		t.Error("Delete 越界索引期望错误")
	}
}

func TestFilterDelete(t *testing.T) {
	src := []int{1, 2, 3, 4, 5, 6}
	// 删除偶数
	got := FilterDelete(src, func(idx int, val int) bool { return val%2 == 0 })
	want := []int{1, 3, 5}
	compareSlices(t, got, want, "FilterDelete 删除偶数")

	// 测试原始 src 的修改（FilterDelete 原地修改并返回子切片）
	// 如果原始 src 是 {1,2,3,4,5,6}, FilterDelete 后可能变成 {1,3,5,4,5,6}
	// 而 got 会是一个指向前3个元素的切片头。
	// 这个检查取决于对原地修改的理解。
	// 如果想检查原始切片是否保持不变，更安全的测试是向 FilterDelete 传递副本。
	srcCopy := []int{1, 2, 3, 4, 5, 6}
	gotFromCopy := FilterDelete(srcCopy, func(idx int, val int) bool { return val%2 == 0 })
	compareSlices(t, gotFromCopy, want, "FilterDelete 从副本中删除偶数")
	// srcCopy 本身会被修改: {1, 3, 5, ...}
	// compareSlices(t, srcCopy[:len(want)], want, "FilterDelete 副本原地修改检查")

	t.Run("删除所有", func(t *testing.T) {
		s := []int{2, 4, 6}
		res := FilterDelete(s, func(_ int, v int) bool { return true }) // 删除所有
		if len(res) != 0 {
			t.Errorf("FilterDelete 删除所有元素，期望空切片, 实际为 %v", res)
		}
	})
}

func TestRegularSet(t *testing.T) { // 重命名以避免冲突
	src := []string{"a", "b", "c"}
	// Set 原地修改 src 并返回它
	modifiedSrc, err := Set(src, 1, "B")
	if err != nil {
		t.Fatalf("Set 出现非期望错误: %v", err)
	}
	want := []string{"a", "B", "c"}
	compareSlices(t, modifiedSrc, want, "Set middle")
	compareSlices(t, src, want, "Set src 原地修改检查") // src 本身被修改

	_, err = Set(src, 3, "X")
	if err == nil {
		t.Error("Set 越界索引期望错误")
	}
}

// --- 包含/去重 相关测试 ---

func TestContains(t *testing.T) {
	src := []int{1, 2, 3}
	if !Contains(src, 2) {
		t.Error("Contains(2) 失败")
	}
	if Contains(src, 4) {
		t.Error("Contains(4) (不存在的元素) 失败")
	}
}

func TestContainsFunc(t *testing.T) {
	src := []testStruct{{1, "A"}, {2, "B"}}
	if !ContainsFunc(src, testStructIDMatcher(1)) {
		t.Error("ContainsFunc 查找 ID 1 失败")
	}
	if ContainsFunc(src, testStructIDMatcher(3)) {
		t.Error("ContainsFunc 查找不存在的 ID 失败")
	}
}

func TestContainsAny(t *testing.T) {
	src := []int{1, 2, 3, 4}
	if !ContainsAny(src, []int{5, 6, 2}) {
		t.Error("ContainsAny({5,6,2}) 失败")
	}
	if ContainsAny(src, []int{5, 6, 7}) {
		t.Error("ContainsAny({5,6,7}) (不存在的元素) 失败")
	}
	if ContainsAny([]int{}, []int{1}) {
		t.Error("ContainsAny 对空 src 操作失败")
	}
}

func TestContainsAll(t *testing.T) {
	src := []int{1, 2, 3, 4}
	if !ContainsAll(src, []int{2, 4}) {
		t.Error("ContainsAll({2,4}) 失败")
	}
	if ContainsAll(src, []int{2, 5}) {
		t.Error("ContainsAll({2,5}) (部分不存在的元素) 失败")
	}
	if !ContainsAll(src, []int{}) { // 包含空集合的所有元素应为 true
		t.Error("ContainsAll 对空 dst 操作失败")
	}
}

func TestDeduplicate(t *testing.T) {
	//注意：deduplicate 使用 map，因此结果顺序不保证。
	//通过比较排序后的切片或检查元素存在性来测试。
	src := []int{1, 2, 2, 3, 1, 4}
	wantUnordered := []int{1, 2, 3, 4}
	got := deduplicate(src)
	compareSlicesIgnoreOrder(t, got, wantUnordered, "Deduplicate basic")

	t.Run("空切片", func(t *testing.T) {
		if len(deduplicate([]int{})) != 0 {
			t.Error("Deduplicate 空切片失败")
		}
	})
	t.Run("全部唯一", func(t *testing.T) {
		s := []int{1, 2, 3}
		compareSlicesIgnoreOrder(t, deduplicate(s), s, "Deduplicate 全部唯一")
	})
}

func TestDeduplicateFunc(t *testing.T) {
	// 提供的 deduplicateFunc 保留等效元素的 *最后*一次出现。
	// 相应地调整 'want' 值。
	t.Run("结构体按 ID 去重 (保留最后一个)", func(t *testing.T) {
		src := []testStruct{{1, "FirstA"}, {2, "B"}, {1, "LastA"}, {3, "C"}}
		// 期望: {2,"B"}, {1,"LastA"}, {3,"C"} (最后唯一元素出现的顺序)
		// 这些最后唯一元素的相对顺序应从原始切片中保留。
		want := []testStruct{{2, "B"}, {1, "LastA"}, {3, "C"}}
		equal := func(s1, s2 testStruct) bool { return s1.ID == s2.ID }
		got := deduplicateFunc(src, equal)
		compareSlicesFunc(t, got, want, testStructEqual, "DeduplicateFunc 结构体")
	})

	t.Run("字符串忽略大小写去重 (保留最后一个)", func(t *testing.T) {
		src := []string{"apple", "Banana", "APPLE", "banana"} // APPLE 是最后一个 "apple", banana 是最后一个 "banana"
		want := []string{"APPLE", "banana"}                   // 这些最后唯一元素的顺序很重要
		equal := func(s1, s2 string) bool { return strings.ToLower(s1) == strings.ToLower(s2) }
		got := deduplicateFunc(src, equal)
		compareSlices(t, got, want, "DeduplicateFunc 字符串忽略大小写")
	})

	t.Run("空切片", func(t *testing.T) {
		src := []int{}
		want := []int{}
		got := deduplicateFunc(src, func(a, b int) bool { return a == b })
		compareSlices(t, got, want, "DeduplicateFunc 空切片")
	})
}

// --- 映射/聚合 相关测试 ---

func TestMap(t *testing.T) {
	src := []int{1, 2, 3}
	want := []string{"1_str", "2_str", "3_str"}
	got := Map(src, func(idx int, val int) string { return strconv.Itoa(val) + "_str" })
	compareSlices(t, got, want, "Map int to string")
	if len(Map([]int{}, func(_ int, v int) int { return v })) != 0 {
		t.Error("Map 空切片失败")
	}
}

func TestFilterMap(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}
	// 过滤偶数并将它们乘以 10
	want := []string{"20", "40"}
	got := FilterMap(src, func(idx int, val int) (string, bool) {
		if val%2 == 0 {
			return strconv.Itoa(val * 10), true
		}
		return "", false
	})
	compareSlices(t, got, want, "FilterMap 过滤偶数并乘以10")
	if len(FilterMap([]int{}, func(_ int, v int) (int, bool) { return v, true })) != 0 {
		t.Error("FilterMap 空切片失败")
	}
}

func TestToMap(t *testing.T) {
	src := []testStruct{{1, "A"}, {2, "B"}, {1, "C"}} // ID 1 重复，最后一个 ("C") 生效
	want := map[int]testStruct{
		1: {1, "C"},
		2: {2, "B"},
	}
	got := ToMap(src, func(e testStruct) int { return e.ID })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ToMap 失败: got %v, want %v", got, want)
	}
	if len(ToMap([]testStruct{}, func(e testStruct) int { return e.ID })) != 0 {
		t.Error("ToMap 空切片失败")
	}
}

func TestToMapV(t *testing.T) {
	// 问题：原测试用例中，src 只包含 ID 为 1 的 "Alice" 和 ID 为 2 的 "Bob"
	// 但期望 map 中 "Alice" 对应的值是 3，这与测试数据不匹配
	//
	// 修复思路：添加 ID 为 3 的 "Alice" 元素到测试数据中
	// 这样 ToMapV 函数会将最后一个 "Alice" (ID 为 3) 的值保存到结果 map 中
	src := []testStruct{{1, "Alice"}, {2, "Bob"}, {3, "Alice"}} // 添加 ID 为 3 的 "Alice"
	// 按 Name (string) 映射到 ID (int)
	// "Alice" 重复，键 "Alice" 对应的最后一个 (ID 3) 生效
	want := map[string]int{
		"Alice": 3, // 来自 {3, "Alice"}
		"Bob":   2, // 来自 {2, "Bob"}
	}
	got := ToMapV(src, func(e testStruct) (string, int) { return e.Name, e.ID })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ToMapV 失败: got %v, want %v", got, want)
	}
}

// --- 集合运算测试 ---
// 由于使用 map，集合运算结果通常是顺序无关的。
// 我们使用 compareSlicesIgnoreOrder。

func TestUnionSet(t *testing.T) {
	s1, s2 := []int{1, 2, 3}, []int{3, 4, 5}
	want := []int{1, 2, 3, 4, 5}
	got := UnionSet(s1, s2)
	compareSlicesIgnoreOrder(t, got, want, "UnionSet 基本操作")
	compareSlicesIgnoreOrder(t, UnionSet([]int{}, []int{1}), []int{1}, "UnionSet 与空切片")
}

func TestIntersectSet(t *testing.T) {
	s1, s2 := []int{1, 2, 3, 4}, []int{3, 4, 5, 6}
	want := []int{3, 4}
	got := IntersectSet(s1, s2)
	compareSlicesIgnoreOrder(t, got, want, "IntersectSet 基本操作")
	if len(IntersectSet([]int{1}, []int{2})) != 0 {
		t.Error("IntersectSet 无交集失败")
	}
	// 测试输入中包含重复项，确保输出已去重
	s3, s4 := []int{1, 2, 2, 3}, []int{2, 2, 3, 4}
	wantDup := []int{2, 3}
	gotDup := IntersectSet(s3, s4)
	compareSlicesIgnoreOrder(t, gotDup, wantDup, "IntersectSet 输入含重复项")
}

func TestDiffSet(t *testing.T) {
	s1, s2 := []int{1, 2, 3, 4}, []int{3, 4, 5, 6} // s1 - s2
	want := []int{1, 2}
	got := DiffSet(s1, s2)
	compareSlicesIgnoreOrder(t, got, want, "DiffSet 基本操作")
	compareSlicesIgnoreOrder(t, DiffSet([]int{1, 2}, []int{}), []int{1, 2}, "DiffSet dst 为空")
}

func TestSymmetricDiffSet(t *testing.T) {
	s1, s2 := []int{1, 2, 3, 4}, []int{3, 4, 5, 6}
	want := []int{1, 2, 5, 6}
	got := SymmetricDiffSet(s1, s2)
	compareSlicesIgnoreOrder(t, got, want, "SymmetricDiffSet 基本操作")
	compareSlicesIgnoreOrder(t, SymmetricDiffSet([]int{1}, []int{1}), []int{}, "SymmetricDiffSet 相同集合")
}

// --- 聚合函数测试 ---

func TestMax(t *testing.T) {
	// 根据函数注释，假设切片非空
	if Max([]int{1, 5, 2}) != 5 {
		t.Error("Max int 失败")
	}
	if Max([]float64{1.1, -0.5, 5.5, 2.2}) != 5.5 {
		t.Error("Max float64 失败")
	}
	if Max([]uint{1, 0, 5}) != 5 {
		t.Error("Max uint 失败")
	}
	// 测试空切片 panic (可选，取决于期望的行为文档)
	// defer func() {
	// 	if r := recover(); r == nil {
	// 		t.Errorf("Max 空切片未 panic")
	// 	}
	// }()
	// _ = Max([]int{})
}

func TestMin(t *testing.T) {
	// 假设切片非空
	if Min([]int{1, -5, 2}) != -5 {
		t.Error("Min int 失败")
	}
	if Min([]float64{1.1, -0.5, -5.5, 2.2}) != -5.5 {
		t.Error("Min float64 失败")
	}
}

func TestSum(t *testing.T) {
	if Sum([]int{1, 2, 3}) != 6 {
		t.Error("Sum int 失败")
	}
	if Sum([]float64{1.5, 2.5}) != 4.0 {
		t.Error("Sum float64 失败")
	}
	if Sum([]int{}) != 0 {
		t.Error("Sum 空切片失败")
	}
}

// --- 反转相关测试 ---

func TestReverse(t *testing.T) {
	src := []int{1, 2, 3, 4}
	want := []int{4, 3, 2, 1}
	originalSrc := append([]int(nil), src...) // 复制以检查原始切片是否被修改

	got := Reverse(src)
	compareSlices(t, got, want, "Reverse 基本操作")
	compareSlices(t, src, originalSrc, "Reverse 确保原始切片未被修改") // 检查原始切片

	if len(Reverse([]int{})) != 0 {
		t.Error("Reverse 空切片失败")
	}
}

func TestReverseSelf(t *testing.T) {
	src := []int{1, 2, 3, 4, 5}
	want := []int{5, 4, 3, 2, 1}
	ReverseSelf(src) // 原地修改
	compareSlices(t, src, want, "ReverseSelf 基本操作")

	empty := []int{}
	ReverseSelf(empty) // 不应 panic
	if len(empty) != 0 {
		t.Error("ReverseSelf 空切片失败")
	}

	nilSlice := []int(nil)
	ReverseSelf(nilSlice) // 不应 panic
	if nilSlice != nil {
		t.Error("ReverseSelf nil 切片失败，期望 nil")
	}
}
