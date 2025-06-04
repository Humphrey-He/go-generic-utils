package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Humphrey-He/go-generic-utils/dataStructures/list"
	"github.com/Humphrey-He/go-generic-utils/dataStructures/queue"
	"github.com/Humphrey-He/go-generic-utils/dataStructures/set"
	"github.com/Humphrey-He/go-generic-utils/pkg/randx"
	"github.com/Humphrey-He/go-generic-utils/pool"
	"github.com/Humphrey-He/go-generic-utils/retry"
	"github.com/Humphrey-He/go-generic-utils/sliceutils"
	"github.com/Humphrey-He/go-generic-utils/tree"
)

// 模拟连接对象
type Connection struct {
	ID string
}

func main() {
	fmt.Println("=== GGU 泛型工具库使用示例 ===")

	// 1. 数据结构示例
	fmt.Println("\n=== 数据结构示例 ===")
	demoDataStructures()
	time.Sleep(100 * time.Millisecond)

	// 2. 集合操作示例
	fmt.Println("\n=== 集合操作示例 ===")
	demoCollections()
	time.Sleep(100 * time.Millisecond)

	// 3. 并发工具示例
	fmt.Println("\n=== 并发工具示例 ===")
	demoConcurrency()
	time.Sleep(100 * time.Millisecond)

	// 4. Set集合示例
	fmt.Println("\n=== Set集合示例 ===")
	demoSets()
	time.Sleep(100 * time.Millisecond)

	// 5. Queue队列示例
	fmt.Println("\n=== Queue队列示例 ===")
	demoQueues()
	time.Sleep(100 * time.Millisecond)

	// 6. Tree树结构示例
	fmt.Println("\n=== Tree树结构示例 ===")
	demoTrees()
	time.Sleep(100 * time.Millisecond)

	// 7. 对象池示例
	fmt.Println("\n=== 对象池示例 ===")
	demoPool()
	time.Sleep(100 * time.Millisecond)

	// 8. 重试机制示例
	fmt.Println("\n=== 重试机制示例 ===")
	demoRetry()
	time.Sleep(100 * time.Millisecond)

	// 9. HTTP工具示例
	fmt.Println("\n=== HTTP工具示例 ===")
	demoHTTPUtil()
	time.Sleep(100 * time.Millisecond)

	// 10. 随机工具示例
	fmt.Println("\n=== 随机工具示例 ===")
	demoRandom()
	time.Sleep(100 * time.Millisecond)

	// 11. Gin工具示例
	fmt.Println("\n=== Gin工具示例 ===")
	fmt.Println("运行 RunGinDemo() 函数可以启动一个完整的Gin API服务器")
	fmt.Println("请查看 gin_demo.go 文件了解完整的Gin工具包使用示例")
	fmt.Println("Gin工具包主要功能:")
	fmt.Println("- 标准化API响应: response包")
	fmt.Println("- 统一错误码: ecode包")
	fmt.Println("- 请求绑定与验证: binding和validate包")
	fmt.Println("- 路由自动注册: register包")
	fmt.Println("- 分页支持: paginator包")
	fmt.Println("- 中间件: 认证、CORS、日志、限流、恢复")
	fmt.Println("- 上下文扩展: contextx包")
	demoGinFunctions()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n=== 示例结束 ===")
}

// 演示数据结构
func demoDataStructures() {
	// ArrayList 示例
	fmt.Println("--- ArrayList ---")
	intList := list.NewArrayList[int](10)
	_ = intList.Append(1, 2, 3, 4, 5)
	fmt.Printf("ArrayList内容: %v\n", intList.AsSlice())
	val, err := intList.Get(2)
	if err == nil {
		fmt.Printf("获取索引2的元素: %d\n", val)
	}
	_ = intList.Set(2, 30)
	fmt.Printf("设置索引2为30后: %v\n", intList.AsSlice())
	removed, _ := intList.Delete(1)
	fmt.Printf("删除索引1的元素(%d)后: %v\n", removed, intList.AsSlice())
	fmt.Printf("列表长度: %d\n", intList.Len())

	// LinkedList 示例
	fmt.Println("\n--- LinkedList ---")
	strList := list.NewLinkedList[string]()
	_ = strList.Append("Go")
	_ = strList.Add(1, "泛型")
	_ = strList.Add(2, "工具库")
	fmt.Printf("LinkedList内容: %v\n", strList.AsSlice())
	strVal, err := strList.Get(1)
	if err == nil {
		fmt.Printf("获取索引1的元素: %s\n", strVal)
	}
	_ = strList.Add(1, "高性能")
	fmt.Printf("在索引1插入'高性能'后: %v\n", strList.AsSlice())

	// ConcurrentList 示例
	fmt.Println("\n--- ConcurrentList ---")
	concList := list.NewConcurrentList[float64](10)
	_ = concList.Append(1.1)
	_ = concList.Add(1, 2.2)
	_ = concList.Add(2, 3.3)
	fmt.Printf("ConcurrentList内容: %v\n", concList.AsSlice())
	_ = concList.Set(0, 9.9)
	fmt.Printf("设置索引0为9.9后: %v\n", concList.AsSlice())
}

// 演示集合操作
func demoCollections() {
	// 过滤操作
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	evens := sliceutils.FindAll(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("偶数过滤结果: %v\n", evens)

	// 映射操作
	doubled := sliceutils.Map(numbers, func(idx int, n int) int {
		return n * 2
	})
	fmt.Printf("数字翻倍结果: %v\n", doubled)

	// 归约操作
	sum := sliceutils.Sum(numbers)
	fmt.Printf("求和结果: %d\n", sum)

	// 查找操作
	found, ok := sliceutils.Find(numbers, func(n int) bool {
		return n > 5 && n < 8
	})
	fmt.Printf("找到大于5小于8的数: %v (找到: %v)\n", found, ok)

	// 分组操作
	evenNumbers := sliceutils.FindAll(numbers, func(n int) bool {
		return n%2 == 0
	})
	oddNumbers := sliceutils.FindAll(numbers, func(n int) bool {
		return n%2 != 0
	})
	groups := map[string][]int{
		"even": evenNumbers,
		"odd":  oddNumbers,
	}
	fmt.Printf("奇偶分组结果: %v\n", groups)
}

// 演示并发工具
func demoConcurrency() {
	// RWMutex 示例
	fmt.Println("--- RWMutex ---")
	var rwm sync.RWMutex
	var value int = 100

	// 读取值
	fmt.Printf("初始值: %d\n", value)

	// 使用读锁
	rwm.RLock()
	fmt.Printf("读锁中的值: %d\n", value)
	rwm.RUnlock()

	// 使用写锁
	rwm.Lock()
	value = 200
	fmt.Printf("更新后的值: %d\n", value)
	rwm.Unlock()

	// 安全地执行操作
	rwm.RLock()
	result := value * 2
	rwm.RUnlock()
	fmt.Printf("通过读锁访问器计算的结果: %d\n", result)

	// 原子更新
	rwm.Lock()
	value += 50
	rwm.Unlock()
	fmt.Printf("原子更新后的值: %d\n", value)
}

// 演示Set集合
func demoSets() {
	// 创建一个整数集合
	intSet := set.NewMapSet[int](10)

	// 添加元素
	intSet.Add(1)
	intSet.Add(2)
	intSet.Add(3)
	intSet.Add(4)
	intSet.Add(5)
	intSet.Add(3) // 重复元素会被忽略
	intSet.Add(4)
	intSet.Add(5)
	intSet.Add(6)
	intSet.Add(7)

	fmt.Printf("整数集合内容: %v\n", intSet.ToSlice())
	fmt.Printf("集合大小: %d\n", intSet.Len())
	fmt.Printf("集合是否包含4: %v\n", intSet.Exist(4))
	fmt.Printf("集合是否包含10: %v\n", intSet.Exist(10))

	// 创建另一个集合
	anotherSet := set.NewMapSet[int](10)
	anotherSet.Add(5)
	anotherSet.Add(6)
	anotherSet.Add(7)
	anotherSet.Add(8)
	anotherSet.Add(9)

	// 集合操作
	unionSet := intSet.Union(anotherSet)
	fmt.Printf("并集: %v\n", unionSet.ToSlice())

	intersectionSet := intSet.Intersect(anotherSet)
	fmt.Printf("交集: %v\n", intersectionSet.ToSlice())

	differenceSet := intSet.Difference(anotherSet)
	fmt.Printf("差集: %v\n", differenceSet.ToSlice())

	// 移除元素
	intSet.Delete(1)
	fmt.Printf("移除1后的集合: %v\n", intSet.ToSlice())

	// 清空集合
	intSet.Clear()
	fmt.Printf("清空后的集合大小: %d\n", intSet.Len())
}

// 演示Queue队列
func demoQueues() {
	// 创建一个字符串队列
	strQueue := queue.NewConcurrentLinkedBlockingQueue[string]()

	// 入队操作
	strQueue.Enqueue("第一个")
	strQueue.Enqueue("第二个")
	strQueue.Enqueue("第三个")

	fmt.Printf("队列大小: %d\n", strQueue.Len())

	// 查看队首元素但不移除
	if front, err := strQueue.Peek(); err == nil {
		fmt.Printf("队首元素: %s\n", front)
	}

	// 出队操作
	if value, err := strQueue.Dequeue(); err == nil {
		fmt.Printf("出队元素: %s\n", value)
	}

	fmt.Printf("出队后队列大小: %d\n", strQueue.Len())

	// 创建优先级队列
	pq := queue.NewConcurrentPriorityQueue[int]()

	// 添加元素
	pq.Enqueue(3)
	pq.Enqueue(1)
	pq.Enqueue(5)
	pq.Enqueue(2)

	fmt.Printf("优先队列大小: %d\n", pq.Len())

	// 出队操作
	if value, err := pq.Dequeue(); err == nil {
		fmt.Printf("优先队列出队元素: %d\n", value)
	}

	fmt.Printf("出队后优先队列大小: %d\n", pq.Len())
}

// 演示Tree树结构
func demoTrees() {
	// 创建二叉搜索树
	intComparator := func(a, b int) int {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}

	bst, _ := tree.NewAVLTree[int, string](intComparator)

	// 插入键值对
	bst.Put(5, "五")
	bst.Put(3, "三")
	bst.Put(8, "八")
	bst.Put(1, "一")
	bst.Put(4, "四")

	fmt.Printf("树的大小: %d\n", bst.Size())

	// 查找
	if value, err := bst.Get(3); err == nil {
		fmt.Printf("键3对应的值: %s\n", value)
	}

	// 判断键是否存在
	fmt.Printf("键5是否存在: %v\n", bst.Contains(5))
	fmt.Printf("键10是否存在: %v\n", bst.Contains(10))

	// 遍历树 - 中序遍历应该得到排序结果
	var result []int
	bst.ForEach(func(k int, v string) bool {
		result = append(result, k)
		return true
	})
	fmt.Printf("中序遍历结果(排序): %v\n", result)

	// 删除节点
	bst.Remove(3)
	fmt.Printf("删除键3后，树的大小: %d\n", bst.Size())

	// 再次遍历
	result = []int{}
	bst.ForEach(func(k int, v string) bool {
		result = append(result, k)
		return true
	})
	fmt.Printf("删除后的中序遍历结果: %v\n", result)
}

// 演示对象池功能
func demoPool() {
	// 创建一个简单对象池
	connectionPool := pool.NewSimpleObjectPool(func() *Connection {
		// 模拟创建连接
		conn := &Connection{
			ID: fmt.Sprintf("conn-%d", time.Now().UnixNano()),
		}
		fmt.Printf("创建新连接: %s\n", conn.ID)
		return conn
	})

	// 从池中获取连接
	conn1 := connectionPool.Get()
	fmt.Printf("获取连接1: %s\n", conn1.ID)

	conn2 := connectionPool.Get()
	fmt.Printf("获取连接2: %s\n", conn2.ID)

	// 归还连接到池中
	connectionPool.Put(conn1)
	fmt.Println("归还连接1到池中")

	// 再次获取连接（应该复用之前归还的连接）
	conn3 := connectionPool.Get()
	fmt.Printf("获取连接3: %s (可能等于连接1)\n", conn3.ID)

	// 使用任务池
	taskPool := pool.NewFixedTaskPool(5)
	fmt.Println("创建了固定大小(5)的任务池")

	// 提交任务
	_ = taskPool.Submit(func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("任务1执行完成")
	})

	fmt.Printf("任务池容量: %d\n", taskPool.Cap())

	// 关闭任务池
	taskPool.Shutdown()
	fmt.Println("任务池已关闭")
}

// 模拟可能失败的操作
func unstableOperation() (string, error) {
	// 模拟随机失败
	if time.Now().UnixNano()%3 == 0 {
		return "", errors.New("操作暂时失败")
	}
	return "操作成功", nil
}

// 演示重试机制
func demoRetry() {
	// 创建指数退避重试策略
	strategy, _ := retry.NewExponentialBackoffRetryStrategy(
		100*time.Millisecond, // 初始等待时间
		1*time.Second,        // 最大等待时间
		5,                    // 最多尝试5次
	)

	fmt.Println("创建了指数退避重试策略:")
	fmt.Println("- 初始等待时间: 100ms")
	fmt.Println("- 最大等待时间: 1s")
	fmt.Println("- 最大尝试次数: 5次")

	// 使用重试机制执行不稳定操作
	ctx := context.Background()
	err := retry.Retry(ctx, strategy, func() error {
		fmt.Println("尝试执行操作...")
		_, err := unstableOperation()
		return err
	})

	if err != nil {
		fmt.Printf("最终操作失败: %v\n", err)
	} else {
		fmt.Printf("最终操作成功!\n")
	}

	// 使用固定间隔重试策略
	fixedStrategy, _ := retry.NewFixedIntervalRetryStrategy(
		200*time.Millisecond, // 固定间隔
		3,                    // 最多尝试3次
	)

	fmt.Println("\n创建了固定间隔重试策略:")
	fmt.Println("- 固定间隔: 200ms")
	fmt.Println("- 最大尝试次数: 3次")

	err = retry.Retry(ctx, fixedStrategy, func() error {
		fmt.Println("使用固定间隔策略尝试操作...")
		return errors.New("模拟持续失败") // 模拟固定失败
	})

	fmt.Printf("固定间隔重试结果: %v\n", err)
}

// 演示HTTP工具
func demoHTTPUtil() {
	// 简单展示net包的HTTP相关功能
	fmt.Println("net包HTTP工具功能:")
	fmt.Println("- 自定义中间件支持")
	fmt.Println("- 请求重试和超时控制")
	fmt.Println("- 限流和断路器集成")

	// 展示请求构建
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)

	// 添加请求头
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", "demo-key")

	fmt.Println("\n请求详情:")
	fmt.Printf("- 方法: %s\n", req.Method)
	fmt.Printf("- URL: %s\n", req.URL)
	fmt.Printf("- 请求头:")
	for k, v := range req.Header {
		fmt.Printf("\n  %s: %s", k, v)
	}

	// 展示net包中间件功能
	fmt.Println("\n\n中间件功能:")
	fmt.Println("- 日志记录")
	fmt.Println("- 请求计时")
	fmt.Println("- 链路追踪")
	fmt.Println("- IP限流")
	fmt.Println("- 用户认证")
}

// 演示随机工具
func demoRandom() {
	// 生成随机字符串
	randomStr, _ := randx.RandCouponCode(10, false)
	fmt.Printf("随机优惠券码(10位): %s\n", randomStr)

	// 生成随机数字
	randomNum, _ := randx.RandInt(1, 100)
	fmt.Printf("随机数字(1-100): %d\n", randomNum)

	// 生成随机商品ID
	randomProdID, _ := randx.RandProductID("P", 8)
	fmt.Printf("随机商品ID: %s\n", randomProdID)

	// 生成随机SKU
	randomSKU, _ := randx.RandSKU("ITEM", 6, "RED")
	fmt.Printf("随机SKU: %s\n", randomSKU)

	// 生成随机订单ID
	randomOrderID, _ := randx.RandOrderID("ORD", 4)
	fmt.Printf("随机订单ID: %s\n", randomOrderID)

	// 生成随机价格
	randomPrice, _ := randx.RandPrice(9.99, 99.99)
	fmt.Printf("随机价格: %.2f\n", randomPrice)

	// 生成随机手机号
	randomPhone, _ := randx.RandPhone()
	fmt.Printf("随机手机号: %s\n", randomPhone)

	// 生成随机邮箱
	randomEmail, _ := randx.RandEmail("example.com", "test.com")
	fmt.Printf("随机邮箱: %s\n", randomEmail)

	// 生成随机UUID
	randomUUID, _ := randx.RandUUID()
	fmt.Printf("随机UUID: %s\n", randomUUID)
}

// 演示Gin函数
func demoGinFunctions() {
	fmt.Println("\n--- Gin工具包结构 ---")
	fmt.Println("1. 响应处理 (response)")
	fmt.Println("   - StandardResponse: 统一的API响应格式")
	fmt.Println("   - Success/Fail: 成功和失败响应辅助函数")
	fmt.Println("   - SuccessWithPagination: 分页数据响应")

	fmt.Println("\n2. 错误码 (ecode)")
	fmt.Println("   - 标准化错误码体系")
	fmt.Println("   - 用户错误: 4xxxx")
	fmt.Println("   - 系统错误: 5xxxx")
	fmt.Println("   - 第三方服务错误: 7xxxx")

	fmt.Println("\n3. 请求绑定 (binding)")
	fmt.Println("   - JSON、Query、URI、表单数据绑定")
	fmt.Println("   - 自定义解码器注册")

	fmt.Println("\n4. 路由注册 (register)")
	fmt.Println("   - 自动注册控制器路由")
	fmt.Println("   - REST风格API支持")
	fmt.Println("   - 基于标签的路由定义")

	fmt.Println("\n5. 中间件 (middleware)")
	fmt.Println("   - 认证 (JWT)")
	fmt.Println("   - CORS跨域")
	fmt.Println("   - 日志记录")
	fmt.Println("   - 请求限流")
	fmt.Println("   - 错误恢复")
	fmt.Println("   - 请求超时")

	fmt.Println("\n6. 分页 (paginator)")
	fmt.Println("   - 偏移量分页")
	fmt.Println("   - 游标分页")
	fmt.Println("   - 自动绑定分页参数")

	fmt.Println("\n7. 上下文扩展 (contextx)")
	fmt.Println("   - 用户身份管理")
	fmt.Println("   - 请求ID跟踪")
	fmt.Println("   - 上下文数据存取")

	fmt.Println("\n完整示例见 gin_demo.go 文件")
}
