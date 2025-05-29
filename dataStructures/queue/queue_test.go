// queue_test.go
package queue

import (
	"math/rand" // 导入 math/rand 包，用于生成随机数 (例如，在并发测试中)
	"reflect"
	"sort"    // 导入 sort 包，用于切片排序 (例如，验证并发测试结果)
	"sync"    // 导入 sync 包，提供同步原语
	"testing" // 导入 testing 包，提供Go语言的测试功能
	"time"    // 导入 time 包，用于处理时间相关的操作，如延迟队列
)

// TestConcurrentArrayBlockingQueue 对并发安全数组阻塞队列进行测试。
func TestConcurrentArrayBlockingQueue(t *testing.T) {
	capacity := 3                                       // 定义队列容量
	q := NewConcurrentArrayBlockingQueue[int](capacity) // 创建队列实例

	t.Run("初始化检查", func(t *testing.T) {
		if q.Len() != 0 {
			t.Errorf("新队列长度应为 0, 实际为 %d", q.Len())
		}
		if !q.IsEmpty() {
			t.Error("新队列应为空")
		}
		if q.cap != capacity {
			t.Errorf("队列容量应为 %d, 实际为 %d", capacity, q.cap)
		}
	})

	t.Run("基本入队和出队", func(t *testing.T) {
		// 入队元素
		err := q.Enqueue(1)
		if err != nil {
			t.Fatalf("Enqueue(1) 失败: %v", err)
		}
		err = q.Enqueue(2)
		if err != nil {
			t.Fatalf("Enqueue(2) 失败: %v", err)
		}

		if q.Len() != 2 {
			t.Errorf("入队两个元素后长度应为 2, 实际为 %d", q.Len())
		}

		// 出队元素并验证顺序
		val, err := q.Dequeue()
		if err != nil || val != 1 {
			t.Fatalf("Dequeue() 期望得到 1, 实际得到 %d, 错误: %v", val, err)
		}
		val, err = q.Dequeue()
		if err != nil || val != 2 {
			t.Fatalf("Dequeue() 期望得到 2, 实际得到 %d, 错误: %v", val, err)
		}

		if !q.IsEmpty() {
			t.Error("出队所有元素后队列应为空")
		}
	})

	t.Run("队列满时阻塞Enqueue", func(t *testing.T) {
		qFull := NewConcurrentArrayBlockingQueue[int](1) // 容量为1的队列
		qFull.Enqueue(10)                                // 入队一个元素使其变满

		enqueueDone := make(chan bool) // 用于通知 Enqueue 操作完成的 channel
		go func() {
			qFull.Enqueue(20) // 此操作应阻塞，因为队列已满
			enqueueDone <- true
		}()

		select {
		case <-enqueueDone:
			t.Fatal("Enqueue 在队列满时不应立即返回")
		case <-time.After(100 * time.Millisecond): // 等待一段时间以确认阻塞
			// 预期行为：Enqueue 阻塞
		}

		// 出队一个元素以腾出空间
		val, _ := qFull.Dequeue()
		if val != 10 {
			t.Errorf("出队的元素应为 10, 实际为 %d", val)
		}

		select {
		case <-enqueueDone: // 现在 Enqueue(20) 应该可以完成了
			// 预期行为：阻塞的 Enqueue 完成
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Enqueue 在队列有空间后未能完成")
		}
		val, _ = qFull.Dequeue()
		if val != 20 {
			t.Errorf("后续出队的元素应为 20, 实际为 %d", val)
		}
	})

	t.Run("队列空时阻塞Dequeue", func(t *testing.T) {
		qEmpty := NewConcurrentArrayBlockingQueue[int](1) // 创建一个队列
		dequeueDone := make(chan int)                     // 用于接收 Dequeue 结果的 channel

		go func() {
			val, _ := qEmpty.Dequeue() // 此操作应阻塞，因为队列为空
			dequeueDone <- val
		}()

		select {
		case <-dequeueDone:
			t.Fatal("Dequeue 在队列空时不应立即返回")
		case <-time.After(100 * time.Millisecond): // 等待一段时间以确认阻塞
			// 预期行为：Dequeue 阻塞
		}

		// 入队一个元素
		qEmpty.Enqueue(100)

		select {
		case val := <-dequeueDone: // 现在 Dequeue 应该可以完成了
			if val != 100 {
				t.Errorf("出队的元素应为 100, 实际为 %d", val)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Dequeue 在队列有元素后未能完成")
		}
	})

	t.Run("并发读写", func(t *testing.T) {
		qConc := NewConcurrentArrayBlockingQueue[int](100) // 足够大的容量以减少阻塞概率，主要测试并发安全
		numGoroutines := 10
		itemsPerGoroutine := 10
		var wg sync.WaitGroup
		totalItems := numGoroutines * itemsPerGoroutine

		// 生产者
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(offset int) {
				defer wg.Done()
				for j := 0; j < itemsPerGoroutine; j++ {
					item := offset*itemsPerGoroutine + j
					qConc.Enqueue(item)
				}
			}(i)
		}

		// 消费者
		results := make(chan int, totalItems)
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < itemsPerGoroutine; j++ {
					val, _ := qConc.Dequeue()
					results <- val
				}
			}()
		}

		wg.Wait()      // 等待所有 goroutine 完成
		close(results) // 关闭 results channel

		if qConc.Len() != 0 {
			t.Errorf("并发操作后队列长度应为 0, 实际为 %d", qConc.Len())
		}

		// 验证所有元素是否都被正确处理（不保证顺序，但应全部存在）
		sum := 0
		count := 0
		expectedSum := 0
		for i := 0; i < totalItems; i++ {
			expectedSum += i
		}
		for val := range results {
			sum += val
			count++
		}
		if count != totalItems {
			t.Errorf("并发操作后取出的元素数量应为 %d, 实际为 %d", totalItems, count)
		}
		if sum != expectedSum {
			t.Errorf("并发操作后取出元素的总和不匹配, 期望 %d, 实际 %d", expectedSum, sum)
		}
	})
}

// TestConcurrentLinkedBlockingQueue 对并发安全链表阻塞队列进行测试。
func TestConcurrentLinkedBlockingQueue(t *testing.T) {
	q := NewConcurrentLinkedBlockingQueue[string]() // 创建队列实例

	t.Run("初始化检查", func(t *testing.T) {
		if q.Len() != 0 {
			t.Errorf("新队列长度应为 0, 实际为 %d", q.Len())
		}
		if !q.IsEmpty() {
			t.Error("新队列应为空")
		}
	})

	t.Run("基本入队和出队", func(t *testing.T) {
		q.Enqueue("hello")
		q.Enqueue("world")

		if q.Len() != 2 {
			t.Errorf("入队两个元素后长度应为 2, 实际为 %d", q.Len())
		}

		val, _ := q.Dequeue()
		if val != "hello" {
			t.Errorf("Dequeue() 期望 'hello', 实际 '%s'", val)
		}
		val, _ = q.Dequeue()
		if val != "world" {
			t.Errorf("Dequeue() 期望 'world', 实际 '%s'", val)
		}

		if !q.IsEmpty() {
			t.Error("出队所有元素后队列应为空")
		}
	})

	t.Run("队列空时阻塞Dequeue", func(t *testing.T) {
		qEmpty := NewConcurrentLinkedBlockingQueue[int]()
		dequeueDone := make(chan int)

		go func() {
			val, _ := qEmpty.Dequeue() // 应阻塞
			dequeueDone <- val
		}()

		select {
		case <-dequeueDone:
			t.Fatal("Dequeue 在队列空时不应立即返回")
		case <-time.After(100 * time.Millisecond):
			// 预期行为
		}

		qEmpty.Enqueue(99)

		select {
		case val := <-dequeueDone:
			if val != 99 {
				t.Errorf("出队的元素应为 99, 实际为 %d", val)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Dequeue 在队列有元素后未能完成")
		}
	})

	t.Run("并发读写", func(t *testing.T) {
		qConc := NewConcurrentLinkedBlockingQueue[int]()
		numGoroutines := 10
		itemsPerGoroutine := 10
		var wg sync.WaitGroup
		totalItems := numGoroutines * itemsPerGoroutine

		// 生产者
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(offset int) {
				defer wg.Done()
				for j := 0; j < itemsPerGoroutine; j++ {
					qConc.Enqueue(offset*itemsPerGoroutine + j)
				}
			}(i)
		}

		// 消费者
		results := make([]int, 0, totalItems)
		var muResults sync.Mutex // 保护 results 切片
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < itemsPerGoroutine; j++ {
					val, _ := qConc.Dequeue()
					muResults.Lock()
					results = append(results, val)
					muResults.Unlock()
				}
			}()
		}

		wg.Wait()

		if qConc.Len() != 0 {
			t.Errorf("并发操作后队列长度应为 0, 实际为 %d", qConc.Len())
		}
		if len(results) != totalItems {
			t.Errorf("并发操作后取出的元素数量应为 %d, 实际为 %d", totalItems, len(results))
		}

		// 验证元素（由于并发，顺序不定，但所有元素都应该存在）
		sort.Ints(results) // 对结果排序以便于比较
		expected := make([]int, totalItems)
		for i := 0; i < totalItems; i++ {
			expected[i] = i
		}
		if !reflect.DeepEqual(results, expected) {
			t.Errorf("并发操作后取出的元素不匹配或丢失")
		}
	})
}

// TestConcurrentPriorityQueue 对并发安全优先级队列进行测试。
func TestConcurrentPriorityQueue(t *testing.T) {
	q := NewConcurrentPriorityQueue[string]() // 创建队列实例

	t.Run("初始化检查", func(t *testing.T) {
		if q.Len() != 0 {
			t.Errorf("新队列长度应为 0, 实际为 %d", q.Len())
		}
		if !q.IsEmpty() {
			t.Error("新队列应为空")
		}
	})

	t.Run("按优先级入队和出队", func(t *testing.T) {
		q.EnqueueWithPriority("low", 1)
		q.EnqueueWithPriority("high", 10)
		q.EnqueueWithPriority("medium", 5)

		if q.Len() != 3 {
			t.Errorf("入队三个元素后长度应为 3, 实际为 %d", q.Len())
		}

		// 出队并验证优先级顺序 (高优先级先出)
		val, _ := q.Dequeue()
		if val != "high" {
			t.Errorf("Dequeue() 期望 'high', 实际 '%s'", val)
		}
		val, _ = q.Dequeue()
		if val != "medium" {
			t.Errorf("Dequeue() 期望 'medium', 实际 '%s'", val)
		}
		val, _ = q.Dequeue()
		if val != "low" {
			t.Errorf("Dequeue() 期望 'low', 实际 '%s'", val)
		}

		if !q.IsEmpty() {
			t.Error("出队所有元素后队列应为空")
		}
	})

	t.Run("队列空时阻塞Dequeue", func(t *testing.T) {
		qEmpty := NewConcurrentPriorityQueue[int]()
		dequeueDone := make(chan int)

		go func() {
			val, _ := qEmpty.Dequeue() // 应阻塞
			dequeueDone <- val
		}()

		select {
		case <-dequeueDone:
			t.Fatal("Dequeue 在队列空时不应立即返回")
		case <-time.After(100 * time.Millisecond):
			// 预期行为
		}

		qEmpty.EnqueueWithPriority(100, 1)

		select {
		case val := <-dequeueDone:
			if val != 100 {
				t.Errorf("出队的元素应为 100, 实际为 %d", val)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Dequeue 在队列有元素后未能完成")
		}
	})

	t.Run("并发优先级处理", func(t *testing.T) {
		qConc := NewConcurrentPriorityQueue[int]()
		var wg sync.WaitGroup
		numItems := 30 // 总共 30 个项目，10 个高优，10 个中优，10 个低优

		// 生产者 (随机顺序入队不同优先级的元素)
		items := make([]struct {
			val  int
			prio int
		}, numItems)
		for i := 0; i < numItems/3; i++ {
			items[i] = struct {
				val  int
				prio int
			}{val: 100 + i, prio: 10} // 高
			items[i+numItems/3] = struct {
				val  int
				prio int
			}{val: 200 + i, prio: 5} // 中
			items[i+numItems*2/3] = struct {
				val  int
				prio int
			}{val: 300 + i, prio: 1} // 低
		}
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })

		for _, item := range items {
			wg.Add(1)
			go func(it struct {
				val  int
				prio int
			}) {
				defer wg.Done()
				qConc.EnqueueWithPriority(it.val, it.prio)
			}(item)
		}

		// 等待所有入队完成
		wg.Wait()

		// 消费者 (按顺序出队并验证优先级)
		var results []int
		var priorities []int
		for i := 0; i < numItems; i++ {
			val, _ := qConc.Dequeue()
			results = append(results, val)
			// 为了验证优先级，我们需要知道原始优先级，这里我们根据值范围推断
			if val >= 100 && val < 200 {
				priorities = append(priorities, 10)
			} else if val >= 200 && val < 300 {
				priorities = append(priorities, 5)
			} else {
				priorities = append(priorities, 1)
			}
		}

		if qConc.Len() != 0 {
			t.Errorf("并发优先级操作后队列长度应为 0, 实际为 %d", qConc.Len())
		}

		// 验证出队顺序是否符合优先级 (高优先级在前)
		for i := 0; i < numItems-1; i++ {
			if priorities[i] < priorities[i+1] {
				t.Errorf("优先级顺序错误: 在索引 %d 处, 优先级 %d 小于后续的优先级 %d。结果: %v", i, priorities[i], priorities[i+1], results)
				break
			}
		}
	})
}

// TestDelayQueue 对并发安全延迟队列进行测试。
func TestDelayQueue(t *testing.T) {
	q := NewDelayQueue[string]() // 创建队列实例

	t.Run("初始化检查", func(t *testing.T) {
		if q.Len() != 0 {
			t.Errorf("新队列长度应为 0, 实际为 %d", q.Len())
		}
		if !q.IsEmpty() {
			t.Error("新队列应为空")
		}
	})

	t.Run("基本延迟入队和出队", func(t *testing.T) {
		now := time.Now()
		q.Enqueue("item1_later", now.Add(200*time.Millisecond)) // 200ms后到期
		q.Enqueue("item0_sooner", now.Add(50*time.Millisecond)) // 50ms后到期 (应先出队)

		if q.Len() != 2 {
			t.Errorf("入队两个元素后长度应为 2, 实际为 %d", q.Len())
		}

		// 出队并验证顺序和延迟
		startTime := time.Now()
		val1, _ := q.Dequeue() // 应阻塞直到 item0_sooner 到期
		duration1 := time.Since(startTime)

		if val1 != "item0_sooner" {
			t.Errorf("Dequeue() 期望 'item0_sooner', 实际 '%s'", val1)
		}
		if duration1 < 40*time.Millisecond || duration1 > 150*time.Millisecond { // 允许一些误差
			t.Errorf("item0_sooner 的出队延迟不在预期范围内 (50ms): %v", duration1)
		}

		startTime = time.Now()
		val2, _ := q.Dequeue() // 应阻塞直到 item1_later 到期
		duration2 := time.Since(startTime)

		if val2 != "item1_later" {
			t.Errorf("Dequeue() 期望 'item1_later', 实际 '%s'", val2)
		}
		// duration2 的计算起点是 val1 出队后，所以其期望延迟是 (200ms - duration1)。
		// 更简单的测试是检查它是否在初始入队时间后约 200ms 后出队。
		// 此处我们期望它在 item0_sooner 出队后再等待约 150ms (200-50)。
		if duration2 < (150-50)*time.Millisecond || duration2 > (150+100)*time.Millisecond { // 允许一些误差
			t.Errorf("item1_later 的出队延迟不在预期范围内 (约150ms after first): %v", duration2)
		}

		if !q.IsEmpty() {
			t.Error("出队所有元素后队列应为空")
		}
	})

	t.Run("队列空时阻塞Dequeue", func(t *testing.T) {
		qEmpty := NewDelayQueue[int]()
		dequeueDone := make(chan bool)

		go func() {
			qEmpty.Dequeue() // 应阻塞
			dequeueDone <- true
		}()

		select {
		case <-dequeueDone:
			t.Fatal("Dequeue 在队列空时不应立即返回")
		case <-time.After(100 * time.Millisecond):
			// 预期行为
		}

		// 入队一个元素后，Dequeue 仍然可能阻塞，直到该元素到期
		qEmpty.Enqueue(1, time.Now().Add(200*time.Millisecond))

		select {
		case <-dequeueDone: // 如果元素立即到期（不太可能），或者逻辑有误
			t.Fatal("Dequeue 在元素未到期时不应立即返回")
		case <-time.After(100 * time.Millisecond): // 100ms 后元素仍未到期
			// 预期行为，Dequeue 仍在阻塞等待元素到期
		}

		// 等待元素到期
		time.Sleep(150 * time.Millisecond) // 总共等待 100 + 150 = 250ms > 200ms

		select {
		case <-dequeueDone:
			// 预期行为，元素到期后 Dequeue 完成
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Dequeue 在元素到期后未能完成")
		}
	})

	t.Run("并发延迟处理", func(t *testing.T) {
		qConc := NewDelayQueue[int]()
		var wg sync.WaitGroup
		numItems := 5
		delays := []time.Duration{300 * time.Millisecond, 100 * time.Millisecond, 500 * time.Millisecond, 50 * time.Millisecond, 200 * time.Millisecond}

		startTime := time.Now() // 记录测试开始时间

		// 生产者
		for i := 0; i < numItems; i++ {
			wg.Add(1)
			go func(val int, delay time.Duration) {
				defer wg.Done()
				qConc.Enqueue(val, startTime.Add(delay)) // 使用 startTime + delay 作为到期时间
			}(i, delays[i])
		}

		// 等待所有入队完成 (goroutine 启动也需要时间)
		wg.Wait()

		// 消费者
		var results []int
		for i := 0; i < numItems; i++ {
			val, _ := qConc.Dequeue() // 将会按照到期时间顺序阻塞并出队
			results = append(results, val)
			t.Logf("出队: %d, 耗时: %v", val, time.Since(startTime))
		}

		if qConc.Len() != 0 {
			t.Errorf("并发延迟操作后队列长度应为 0, 实际为 %d", qConc.Len())
		}

		// 期望的出队顺序是根据 delays 排序后的原始索引
		// delays: [300, 100, 500, 50, 200] -> 对应的原始值(索引): [0, 1, 2, 3, 4]
		// 排序后: [50, 100, 200, 300, 500] -> 对应的原始值(索引): [3, 1, 4, 0, 2]
		expectedOrder := []int{3, 1, 4, 0, 2}
		if !reflect.DeepEqual(results, expectedOrder) {
			t.Errorf("并发延迟操作后出队顺序不正确。期望: %v, 实际: %v", expectedOrder, results)
		}
	})
}
