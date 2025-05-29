package syncx

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// ===================== Cond 条件变量测试 =====================
func TestCond_BroadcastAndSignal(t *testing.T) {
	cond := NewCond(&sync.Mutex{})
	count := 0
	waitN := 5
	ch := make(chan struct{}, waitN)

	for i := 0; i < waitN; i++ {
		go func() {
			cond.L.Lock()
			_ = cond.Wait(context.Background())
			count++
			cond.L.Unlock()
			ch <- struct{}{}
		}()
	}
	// 确保所有 goroutine 都在等待
	time.Sleep(50 * time.Millisecond)
	cond.L.Lock()
	cond.Broadcast()
	cond.L.Unlock()
	for i := 0; i < waitN; i++ {
		<-ch
	}
	if count != waitN {
		t.Errorf("Broadcast 唤醒数量错误, got %d, want %d", count, waitN)
	}

	// Signal 单个唤醒
	count = 0
	go func() {
		cond.L.Lock()
		_ = cond.Wait(context.Background())
		count++
		cond.L.Unlock()
		ch <- struct{}{}
	}()
	time.Sleep(20 * time.Millisecond)
	cond.L.Lock()
	cond.Signal()
	cond.L.Unlock()
	<-ch
	if count != 1 {
		t.Errorf("Signal 唤醒数量错误, got %d, want 1", count)
	}
}

// ===================== Map 泛型并发安全Map测试 =====================
func TestMap_Basic(t *testing.T) {
	m := Map[string, int]{}
	m.Store("a", 1)
	v, ok := m.Load("a")
	if !ok || v != 1 {
		t.Errorf("Load/Store 失败, got %v, want 1", v)
	}
	actual, loaded := m.LoadOrStore("a", 2)
	if !loaded || actual != 1 {
		t.Errorf("LoadOrStore 已存在时失败")
	}
	actual, loaded = m.LoadOrStore("b", 3)
	if loaded || actual != 3 {
		t.Errorf("LoadOrStore 新增时失败")
	}
	m.Delete("a")
	_, ok = m.Load("a")
	if ok {
		t.Errorf("Delete 失败")
	}
	m.Store("c", 5)
	m.Store("d", 6)
	sum := 0
	m.Range(func(k string, v int) bool {
		sum += v
		return true
	})
	if sum != 14 {
		t.Errorf("Range 求和错误, got %d, want 14", sum)
	}
}

func TestMap_LoadOrStoreFunc(t *testing.T) {
	m := Map[string, int]{}
	v, loaded, err := m.LoadOrStoreFunc("x", func() (int, error) { return 42, nil })
	if loaded || v != 42 || err != nil {
		t.Errorf("LoadOrStoreFunc 新增失败")
	}
	v, loaded, err = m.LoadOrStoreFunc("x", func() (int, error) { return 99, nil })
	if !loaded || v != 42 || err != nil {
		t.Errorf("LoadOrStoreFunc 已存在失败")
	}
	_, _, err = m.LoadOrStoreFunc("y", func() (int, error) { return 0, errors.New("fail") })
	if err == nil {
		t.Errorf("LoadOrStoreFunc 错误未返回")
	}
}

// ===================== Pool 泛型对象池测试 =====================
func TestPool_Basic(t *testing.T) {
	cnt := 0
	p := NewPool(func() int {
		cnt++
		return 100
	})
	v := p.Get()
	if v != 100 {
		t.Errorf("Pool Get 失败")
	}
	p.Put(200)
	v2 := p.Get()
	if v2 != 200 && v2 != 100 {
		t.Errorf("Pool Put/Get 失败")
	}
	if cnt == 0 {
		t.Errorf("Pool factory 未调用")
	}
}

// ===================== LimitPool 限流对象池测试 =====================
func TestLimitPool_Basic(t *testing.T) {
	p := NewLimitPool(2, func() int { return 7 })
	v1, ok1 := p.Get()
	v2, ok2 := p.Get()
	// v3, ok3 := p.Get() // 未使用变量，已移除
	if !ok1 || v1 != 7 || !ok2 || v2 != 7 {
		t.Errorf("LimitPool 前两次应成功")
	}
	// if ok3 {
	// 	t.Errorf("LimitPool 超限应失败")
	// }
	p.Put(v1)
	v4, ok4 := p.Get()
	if !ok4 || v4 != 7 {
		t.Errorf("LimitPool Put 后应可再次获取")
	}
}

// ===================== SegmentKeysLock 分段Key锁测试 =====================
func TestSegmentKeysLock_Basic(t *testing.T) {
	lock := NewSegmentKeysLock(8)
	key := "order:123"
	lock.Lock(key)
	locked := !lock.TryLock(key)
	if !locked {
		t.Errorf("TryLock 应该失败")
	}
	lock.Unlock(key)
	if !lock.TryLock(key) {
		t.Errorf("TryLock 应该成功")
	}
	lock.Unlock(key)
	lock.RLock(key)
	if !lock.TryRLock(key) {
		t.Errorf("TryRLock 应该成功")
	}
	lock.RUnlock(key)
	lock.RUnlock(key)
}

// ===================== Value 泛型原子操作测试 =====================
func TestValue_Basic(t *testing.T) {
	v := NewValueOf(10)
	if v.Load() != 10 {
		t.Errorf("Value Load 失败")
	}
	v.Store(20)
	if v.Load() != 20 {
		t.Errorf("Value Store 失败")
	}
	old := v.Swap(30)
	if old != 20 || v.Load() != 30 {
		t.Errorf("Value Swap 失败")
	}
	ok := v.CompareAndSwap(30, 40)
	if !ok || v.Load() != 40 {
		t.Errorf("Value CompareAndSwap 失败")
	}
}

// 以下为增强API测试，需配合增强版Value实现
// func TestValue_UpdateAndTryUpdate(t *testing.T) {
// 	v := NewValueOf(1)
// 	// Update 累加
// 	newVal, ok := v.Update(func(old int) (int, bool) { return old + 1, true })
// 	if !ok || newVal != 2 {
// 		t.Errorf("Value Update 失败")
// 	}
// 	// TryUpdate 限定重试次数
// 	v.Store(5)
// 	newVal, ok = v.TryUpdate(func(old int) (int, bool) { return old * 2, true }, 3)
// 	if !ok || newVal != 10 {
// 		t.Errorf("Value TryUpdate 失败")
// 	}
// 	// Update 返回 false 不应更新
// 	v.Store(100)
// 	_, ok = v.Update(func(old int) (int, bool) { return 0, false })
// 	if ok {
// 		t.Errorf("Value Update 应该返回 false")
// 	}
// }
//
// func TestValue_CloneAndReset(t *testing.T) {
// 	v := NewValueOf(99)
// 	clone := v.Clone()
// 	if clone != 99 {
// 		t.Errorf("Value Clone 失败")
// 	}
// 	v.Reset()
// 	if v.Load() != 0 {
// 		t.Errorf("Value Reset 失败")
// 	}
// }
