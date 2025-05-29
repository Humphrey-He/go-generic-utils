package queue

import (
	"container/heap"
	"errors"
	"sync"
	"time"
)

///////////////////// 队列接口 /////////////////////

// Queue 队列通用接口
type Queue[T any] interface {
	Enqueue(val T) error // 入队
	Dequeue() (T, error) // 出队
	Len() int            // 队列长度
	IsEmpty() bool       // 是否为空
}

///////////////////// 并发安全数组阻塞队列 /////////////////////

// ConcurrentArrayBlockingQueue 并发安全的有界阻塞队列（环形数组实现）
type ConcurrentArrayBlockingQueue[T any] struct {
	mu    sync.Mutex
	cond  *sync.Cond
	data  []T
	front int
	rear  int
	size  int
	cap   int
}

// NewConcurrentArrayBlockingQueue 创建一个有界阻塞队列
func NewConcurrentArrayBlockingQueue[T any](capacity int) *ConcurrentArrayBlockingQueue[T] {
	q := &ConcurrentArrayBlockingQueue[T]{
		data: make([]T, capacity),
		cap:  capacity,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue 入队，队满时阻塞
func (q *ConcurrentArrayBlockingQueue[T]) Enqueue(val T) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.size == q.cap {
		q.cond.Wait()
	}
	q.data[q.rear] = val
	q.rear = (q.rear + 1) % q.cap
	q.size++
	q.cond.Signal()
	return nil
}

// Dequeue 出队，队空时阻塞
func (q *ConcurrentArrayBlockingQueue[T]) Dequeue() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.size == 0 {
		q.cond.Wait()
	}
	val := q.data[q.front]
	q.front = (q.front + 1) % q.cap
	q.size--
	q.cond.Signal()
	return val, nil
}

func (q *ConcurrentArrayBlockingQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

func (q *ConcurrentArrayBlockingQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

///////////////////// 并发安全链表阻塞队列 /////////////////////

// node 单向链表节点
type node[T any] struct {
	val  T
	next *node[T]
}

// ConcurrentLinkedBlockingQueue 并发安全的链表阻塞队列
type ConcurrentLinkedBlockingQueue[T any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	head *node[T]
	tail *node[T]
	size int
}

// NewConcurrentLinkedBlockingQueue 创建链表阻塞队列
func NewConcurrentLinkedBlockingQueue[T any]() *ConcurrentLinkedBlockingQueue[T] {
	n := &node[T]{}
	q := &ConcurrentLinkedBlockingQueue[T]{
		head: n,
		tail: n,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *ConcurrentLinkedBlockingQueue[T]) Enqueue(val T) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := &node[T]{val: val}
	q.tail.next = n
	q.tail = n
	q.size++
	q.cond.Signal()
	return nil
}

func (q *ConcurrentLinkedBlockingQueue[T]) Dequeue() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.head.next == nil {
		q.cond.Wait()
	}
	n := q.head.next
	q.head.next = n.next
	if q.tail == n {
		q.tail = q.head
	}
	q.size--
	return n.val, nil
}

func (q *ConcurrentLinkedBlockingQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}

func (q *ConcurrentLinkedBlockingQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

///////////////////// 并发安全优先级队列 /////////////////////

// priorityItem 优先级队列元素
type priorityItem[T any] struct {
	value    T
	priority int
	index    int
}

// priorityQueueHeap 实现heap.Interface
type priorityQueueHeap[T any] []*priorityItem[T]

func (h priorityQueueHeap[T]) Len() int           { return len(h) }
func (h priorityQueueHeap[T]) Less(i, j int) bool { return h[i].priority > h[j].priority }
func (h priorityQueueHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *priorityQueueHeap[T]) Push(x any)        { *h = append(*h, x.(*priorityItem[T])) }
func (h *priorityQueueHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// ConcurrentPriorityQueue 并发安全优先级队列
type ConcurrentPriorityQueue[T any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	pq   priorityQueueHeap[T]
}

// NewConcurrentPriorityQueue 创建优先级队列
func NewConcurrentPriorityQueue[T any]() *ConcurrentPriorityQueue[T] {
	q := &ConcurrentPriorityQueue[T]{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue 按优先级入队
func (q *ConcurrentPriorityQueue[T]) EnqueueWithPriority(val T, priority int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	item := &priorityItem[T]{value: val, priority: priority}
	heap.Push(&q.pq, item)
	q.cond.Signal()
	return nil
}

// Dequeue 按优先级出队
func (q *ConcurrentPriorityQueue[T]) Dequeue() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.pq.Len() == 0 {
		q.cond.Wait()
	}
	item := heap.Pop(&q.pq).(*priorityItem[T])
	return item.value, nil
}

func (q *ConcurrentPriorityQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pq.Len()
}

func (q *ConcurrentPriorityQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

///////////////////// 并发安全延迟队列 /////////////////////

// DelayItem 延迟队列元素
type DelayItem[T any] struct {
	Value    T
	ExpireAt time.Time
	index    int
}

// delayQueueHeap 实现heap.Interface
type delayQueueHeap[T any] []*DelayItem[T]

func (h delayQueueHeap[T]) Len() int           { return len(h) }
func (h delayQueueHeap[T]) Less(i, j int) bool { return h[i].ExpireAt.Before(h[j].ExpireAt) }
func (h delayQueueHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *delayQueueHeap[T]) Push(x any)        { *h = append(*h, x.(*DelayItem[T])) }
func (h *delayQueueHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// DelayQueue 并发安全延迟队列
type DelayQueue[T any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	pq   delayQueueHeap[T]
}

// NewDelayQueue 创建延迟队列
func NewDelayQueue[T any]() *DelayQueue[T] {
	q := &DelayQueue[T]{}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue 将一个元素添加到延迟队列中，并指定其到期时间。
// val 是要添加的元素值。
// expireAt 是元素的绝对到期时间 (UTC 时间或带时区的时间)。
// 此方法是线程安全的。
func (q *DelayQueue[T]) Enqueue(val T, expireAt time.Time) error {
	q.mu.Lock()         // 获取锁以保护共享资源 pq
	defer q.mu.Unlock() // 确保在函数退出时释放锁

	// if q.closed { // (可选) 如果队列已关闭，则可以拒绝新元素
	// 	return errors.New("队列已关闭，无法入队")
	// }

	item := &DelayItem[T]{Value: val, ExpireAt: expireAt}
	heap.Push(&q.pq, item) // 将元素推入堆中，heap 包会自动调用 h.Push 并维护堆属性
	q.cond.Signal()        // 通知一个可能因队列为空或等待特定到期时间而阻塞的 Dequeue goroutine
	return nil
}

// Dequeue 从延迟队列中获取一个已到期的元素。
// 如果队列为空，此方法会阻塞，直到有新元素入队。
// 如果队首元素尚未到期，此方法会阻塞，直到该元素到期或被一个更早到期的新入队元素取代。
// 返回取出的元素值和 nil 错误。在当前实现中，除非将来添加关闭逻辑，否则一般不返回错误。
// 此方法是线程安全的。
func (q *DelayQueue[T]) Dequeue() (T, error) {
	q.mu.Lock()         // 获取锁以保护共享资源 pq
	defer q.mu.Unlock() // 确保在函数所有返回路径上都释放锁

	for { // 使用无限循环，直到成功取出一个元素
		// if q.closed && q.pq.Len() == 0 { // (可选) 检查队列是否已关闭且为空，用于优雅退出
		// 	var zero T
		// 	return zero, errors.New("队列已关闭且为空")
		// }

		if q.pq.Len() == 0 {
			// 队列为空，阻塞等待 Enqueue 操作的信号
			q.cond.Wait() // Wait 会自动释放 q.mu 锁，并在被唤醒后尝试重新获取锁
			// 被唤醒后，重新开始外层 for 循环，检查队列状态
			continue
		}

		// 队列非空，获取队首元素 (但不立即从堆中移除)
		item := q.pq[0] // pq[0] 是堆顶元素，即最早到期的元素
		now := time.Now()

		if now.Before(item.ExpireAt) {
			// 队首元素尚未到期，需要等待
			waitTime := item.ExpireAt.Sub(now) // 计算还需等待的时间

			// 为了使 q.cond.Wait() 能够被超时唤醒，我们需要一种机制。
			// 标准库的 cond 没有 WaitTimeout。一个常见模式是：
			// 启动一个辅助 goroutine，在超时后调用 q.cond.Signal()。
			// 这个辅助 goroutine 需要能够被取消，以避免泄漏。

			done := make(chan struct{}) // 用于通知辅助 goroutine 停止
			go func() {
				// 这个 goroutine 的职责是在 waitTime 之后，或者在被取消之前，发出信号
				select {
				case <-time.After(waitTime): // 等待指定时间
					q.mu.Lock()     // 在操作条件变量前获取锁
					q.cond.Signal() // 超时后，发送信号唤醒等待的 Dequeue
					q.mu.Unlock()   // 释放锁
				case <-done: // Dequeue 被其他方式唤醒，或者队首元素已改变
					return // 辅助 goroutine 退出
				}
			}()

			// 在条件变量上等待。这个等待可能被以下任一情况唤醒：
			// 1. Enqueue 操作加入新元素并调用 q.cond.Signal()。
			// 2. 上面创建的辅助 timer goroutine 在超时后调用 q.cond.Signal()。
			q.cond.Wait() // Wait 会释放 q.mu，被唤醒后重新获取 q.mu

			// 清理辅助 timer goroutine。
			// close(done) 会让 select 中的 <-done 被选中（如果它还没退出），从而使 goroutine 结束。
			close(done)

			// 被唤醒后，无论原因如何，都应重新评估队列的整体状态（队首元素可能已变，或已到期）
			continue // 回到外层 for 循环的开始
		}

		// 队首元素已到期，可以从堆中取出并返回
		poppedItem := heap.Pop(&q.pq).(*DelayItem[T]) // Pop 会调用 h.Pop 并调整堆
		return poppedItem.Value, nil
	}
	// 此处代码不可达，因为外层是 for {} 无限循环，且总有返回路径（return 或 continue）
}

func (q *DelayQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.pq.Len()
}

func (q *DelayQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

///////////////////// 通用错误 /////////////////////

var (
	ErrQueueEmpty = errors.New("队列为空")
	ErrQueueFull  = errors.New("队列已满")
)
