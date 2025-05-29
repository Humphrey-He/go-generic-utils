package tree

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试库存管理器基础功能
func TestInventoryManager_Basic(t *testing.T) {
	// 初始化库存管理器
	im := NewInventoryManager()

	// 测试添加SKU
	err := im.AddSku("SKU001", 100, 20, "WH-001")
	assert.NoError(t, err, "添加SKU应成功")

	// 测试获取库存
	stock, err := im.GetStock("SKU001")
	assert.NoError(t, err, "获取库存应成功")
	assert.Equal(t, 100, stock, "初始库存应为100")

	// 测试不存在的SKU
	_, err = im.GetStock("SKU999")
	assert.ErrorIs(t, err, ErrSkuNotFound, "获取不存在的SKU应返回错误")

	// 测试负库存
	err = im.AddSku("SKU002", -10, 5, "WH-001")
	assert.ErrorIs(t, err, ErrInvalidQuantity, "添加负库存应返回错误")
}

// 测试库存预留与提交流程
func TestInventoryManager_ReserveCommitFlow(t *testing.T) {
	im := NewInventoryManager()
	im.AddSku("SKU001", 100, 20, "WH-001")

	// 测试预留库存
	err := im.Reserve("SKU001", 30)
	assert.NoError(t, err, "预留库存应成功")

	// 检查可用库存
	available, err := im.GetStock("SKU001")
	assert.NoError(t, err)
	assert.Equal(t, 70, available, "预留后可用库存应为70")

	// 测试库存不足情况
	err = im.Reserve("SKU001", 80)
	assert.ErrorIs(t, err, ErrStockShortage, "超出可用库存应返回错误")

	// 测试确认库存扣减
	err = im.Commit("SKU001", 20)
	assert.NoError(t, err, "确认扣减库存应成功")

	// 检查预留和总库存
	item, err := im.inventory.Get("SKU001")
	assert.NoError(t, err)
	assert.Equal(t, 80, item.Stock, "扣减后总库存应为80")
	assert.Equal(t, 10, item.Reserved, "扣减后预留库存应为10")

	// 测试释放预留库存
	err = im.Release("SKU001", 10)
	assert.NoError(t, err, "释放预留库存应成功")

	// 检查预留库存
	item, err = im.inventory.Get("SKU001")
	assert.NoError(t, err)
	assert.Equal(t, 0, item.Reserved, "释放后预留库存应为0")
}

// 测试补充库存和批量更新
func TestInventoryManager_RestockAndBatch(t *testing.T) {
	im := NewInventoryManager()
	im.AddSku("SKU001", 100, 20, "WH-001")
	im.AddSku("SKU002", 50, 10, "WH-001")

	// 测试补充库存
	err := im.Restock("SKU001", 50)
	assert.NoError(t, err, "补充库存应成功")

	stock, _ := im.GetStock("SKU001")
	assert.Equal(t, 150, stock, "补充后库存应为150")

	// 测试批量更新
	updates := map[string]int{
		"SKU001": 10,  // 增加10
		"SKU002": -20, // 减少20
	}

	results := im.BatchUpdateStock(updates)
	assert.NoError(t, results["SKU001"], "SKU001更新应成功")
	assert.NoError(t, results["SKU002"], "SKU002更新应成功")

	// 检查更新结果
	stock1, _ := im.GetStock("SKU001")
	stock2, _ := im.GetStock("SKU002")
	assert.Equal(t, 160, stock1, "SKU001库存应为160")
	assert.Equal(t, 30, stock2, "SKU002库存应为30")
}

// 测试仓库和低库存查询
func TestInventoryManager_WarehouseAndLowStock(t *testing.T) {
	im := NewInventoryManager()
	im.AddSku("SKU001", 100, 20, "WH-001")
	im.AddSku("SKU002", 50, 10, "WH-001")
	im.AddSku("SKU003", 5, 10, "WH-002") // 低于安全库存

	// 测试获取指定仓库商品
	wh001Items := im.GetItemsByWarehouse("WH-001")
	assert.Len(t, wh001Items, 2, "WH-001仓库应有2个商品")

	// 测试获取低库存商品
	lowStockItems := im.GetLowStockItems()
	assert.Len(t, lowStockItems, 1, "应有1个低库存商品")
	assert.Equal(t, "SKU003", lowStockItems[0].Sku, "SKU003应为低库存")
}

// 测试库存事件处理
func TestInventoryManager_EventHandling(t *testing.T) {
	im := NewInventoryManager()

	// 设置事件接收通道
	eventCh := make(chan string, 10)
	im.AddEventHandler(func(item *InventoryItem, action string, quantity int) {
		eventCh <- action
	})

	// 触发各种操作
	im.AddSku("SKU001", 100, 20, "WH-001")
	im.Reserve("SKU001", 30)
	im.Commit("SKU001", 20)
	im.Release("SKU001", 10)
	im.Restock("SKU001", 50)

	// 验证事件
	expectedEvents := []string{"add", "reserve", "commit", "release", "restock"}
	receivedEvents := make([]string, 0, 5)

	// 收集事件
	for i := 0; i < 5; i++ {
		select {
		case event := <-eventCh:
			receivedEvents = append(receivedEvents, event)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("超时等待事件, 已收到 %d 个事件", len(receivedEvents))
		}
	}

	// 验证事件数量和类型(不关心顺序)
	assert.Len(t, receivedEvents, 5, "应触发5个事件")
	for _, expected := range expectedEvents {
		found := false
		for _, received := range receivedEvents {
			if received == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "应接收到 %s 事件", expected)
	}
}

// 测试库存管理器并发安全性
func TestInventoryManager_ConcurrentSafety(t *testing.T) {
	im := NewInventoryManager()
	im.AddSku("SKU001", 1000, 20, "WH-001")

	// 并发预留和释放
	var wg sync.WaitGroup
	concurrency := 100
	wg.Add(concurrency * 2)

	// 并发预留
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			err := im.Reserve("SKU001", 1)
			assert.NoError(t, err)
		}()
	}

	// 并发释放
	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			// 等待片刻确保预留已完成
			time.Sleep(time.Duration(idx%10) * time.Millisecond)
			err := im.Release("SKU001", 1)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// 检查最终状态
	item, err := im.inventory.Get("SKU001")
	assert.NoError(t, err)
	assert.Equal(t, 1000, item.Stock, "总库存应保持不变")
	assert.Equal(t, 0, item.Reserved, "预留库存应为0")
}

// 测试价格管理器基础功能
func TestPriceManager_Basic(t *testing.T) {
	pm := NewPriceManager()
	now := time.Now()

	// 测试添加商品
	sku1 := &ProductSku{
		ID:            "SKU001",
		ProductID:     "P001",
		Price:         99.99,
		OriginalPrice: 129.99,
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := pm.AddProduct(sku1)
	assert.NoError(t, err, "添加商品应成功")

	// 测试获取商品
	product, err := pm.GetProduct("SKU001")
	assert.NoError(t, err, "获取商品应成功")
	assert.Equal(t, 99.99, product.Price, "价格应为99.99")

	// 测试更新价格
	err = pm.UpdatePrice("SKU001", 89.99)
	assert.NoError(t, err, "更新价格应成功")

	// 验证更新后的价格
	product, _ = pm.GetProduct("SKU001")
	assert.Equal(t, 89.99, product.Price, "新价格应为89.99")
	assert.Equal(t, 99.99, product.OriginalPrice, "原价应为99.99")

	// 测试无效价格
	err = pm.UpdatePrice("SKU001", -10.0)
	assert.ErrorIs(t, err, ErrInvalidPrice, "负价格应返回错误")

	// 测试不存在的商品
	_, err = pm.GetProduct("SKU999")
	assert.ErrorIs(t, err, ErrSkuNotFound, "获取不存在的商品应返回错误")
}

// 测试价格区间查询和排序
func TestPriceManager_RangeAndSort(t *testing.T) {
	pm := NewPriceManager()
	now := time.Now()

	// 添加多个商品
	skus := []*ProductSku{
		{ID: "SKU001", ProductID: "P001", Price: 99.99, Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "SKU002", ProductID: "P002", Price: 199.99, Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "SKU003", ProductID: "P003", Price: 299.99, Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "SKU004", ProductID: "P004", Price: 399.99, Status: "inactive", CreatedAt: now, UpdatedAt: now}, // 非活跃
		{ID: "SKU005", ProductID: "P005", Price: 149.99, Status: "active", CreatedAt: now, UpdatedAt: now},
	}

	for _, sku := range skus {
		pm.AddProduct(sku)
	}

	// 测试价格区间查询
	products, err := pm.GetProductsInPriceRange(100, 300)
	assert.NoError(t, err, "价格区间查询应成功")
	assert.Len(t, products, 2, "100-300区间应有2个活跃商品")

	// 检查查询结果中没有非活跃商品
	for _, p := range products {
		assert.Equal(t, "active", p.Status, "结果应只包含活跃商品")
		assert.GreaterOrEqual(t, p.Price, 100.0, "价格应大于等于100")
		assert.Less(t, p.Price, 300.0, "价格应小于300")
	}

	// 测试按价格升序排序
	ascending := pm.GetProductsSortedByPrice(true)
	assert.Len(t, ascending, 4, "应有4个活跃商品")
	for i := 0; i < len(ascending)-1; i++ {
		assert.LessOrEqual(t, ascending[i].Price, ascending[i+1].Price, "价格应升序排列")
	}

	// 测试按价格降序排序
	descending := pm.GetProductsSortedByPrice(false)
	assert.Len(t, descending, 4, "应有4个活跃商品")
	for i := 0; i < len(descending)-1; i++ {
		assert.GreaterOrEqual(t, descending[i].Price, descending[i+1].Price, "价格应降序排列")
	}
}

// 测试批量价格更新
func TestPriceManager_BatchUpdate(t *testing.T) {
	pm := NewPriceManager()
	now := time.Now()

	// 添加多个商品
	skus := []*ProductSku{
		{ID: "SKU001", ProductID: "P001", Price: 99.99, Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "SKU002", ProductID: "P002", Price: 199.99, Status: "active", CreatedAt: now, UpdatedAt: now},
		{ID: "SKU003", ProductID: "P003", Price: 299.99, Status: "active", CreatedAt: now, UpdatedAt: now},
	}

	for _, sku := range skus {
		pm.AddProduct(sku)
	}

	// 批量更新价格
	updates := map[string]float64{
		"SKU001": 89.99,
		"SKU002": 179.99,
		"SKU003": 269.99,
		"SKU999": 999.99, // 不存在的SKU
	}

	results := pm.BatchUpdatePrices(updates)

	// 验证结果
	assert.NoError(t, results["SKU001"], "SKU001更新应成功")
	assert.NoError(t, results["SKU002"], "SKU002更新应成功")
	assert.NoError(t, results["SKU003"], "SKU003更新应成功")
	assert.ErrorIs(t, results["SKU999"], ErrSkuNotFound, "不存在的SKU应返回错误")

	// 检查更新后的价格
	product1, _ := pm.GetProduct("SKU001")
	product2, _ := pm.GetProduct("SKU002")
	product3, _ := pm.GetProduct("SKU003")

	assert.Equal(t, 89.99, product1.Price, "SKU001新价格应为89.99")
	assert.Equal(t, 179.99, product2.Price, "SKU002新价格应为179.99")
	assert.Equal(t, 269.99, product3.Price, "SKU003新价格应为269.99")
}

// 测试价格管理器并发安全性
func TestPriceManager_ConcurrentSafety(t *testing.T) {
	pm := NewPriceManager()
	now := time.Now()

	// 添加商品
	sku := &ProductSku{
		ID:            "SKU001",
		ProductID:     "P001",
		Price:         100.0,
		OriginalPrice: 100.0,
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	pm.AddProduct(sku)

	// 并发更新价格
	var wg sync.WaitGroup
	concurrency := 100
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			// 价格在99-101之间波动
			newPrice := 99.0 + float64(idx%3)
			err := pm.UpdatePrice("SKU001", newPrice)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// 检查最终价格在允许范围内
	product, err := pm.GetProduct("SKU001")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, product.Price, 99.0, "最终价格应不小于99.0")
	assert.LessOrEqual(t, product.Price, 101.0, "最终价格应不大于101.0")
}

// 测试搜索引擎基础功能
func TestSearchEngine_Basic(t *testing.T) {
	se := NewSearchEngine()

	// 测试索引商品
	se.IndexProduct("P001", []string{"苹果", "手机", "iPhone"})
	se.IndexProduct("P002", []string{"华为", "手机", "Mate"})
	se.IndexProduct("P003", []string{"小米", "手机", "Redmi"})

	// 测试搜索
	results := se.Search("手机", 10)
	assert.Len(t, results, 1, "应找到1个匹配词")
	assert.Equal(t, "手机", results[0].Term, "匹配词应为'手机'")
	assert.Len(t, results[0].ProductIDs, 3, "应匹配3个商品")
	assert.Equal(t, 3, results[0].Score, "分数应为3")

	// 测试精确搜索
	results = se.Search("iPhone", 10)
	assert.Len(t, results, 1, "应找到1个匹配词")
	assert.Equal(t, "iPhone", results[0].Term, "匹配词应为'iPhone'")
	assert.Len(t, results[0].ProductIDs, 1, "应匹配1个商品")
	assert.Equal(t, "P001", results[0].ProductIDs[0], "应匹配P001")

	// 测试部分匹配
	results = se.Search("苹", 10)
	assert.Len(t, results, 1, "应找到1个匹配词")
	assert.Equal(t, "苹果", results[0].Term, "匹配词应为'苹果'")

	// 测试无匹配
	results = se.Search("笔记本", 10)
	assert.Len(t, results, 0, "应找不到匹配")
}

// 测试自动补全功能
func TestSearchEngine_AutoComplete(t *testing.T) {
	se := NewSearchEngine()

	// 索引商品
	se.IndexProduct("P001", []string{"苹果手机", "iPhone", "智能手机"})
	se.IndexProduct("P002", []string{"苹果平板", "iPad", "智能平板"})
	se.IndexProduct("P003", []string{"华为手机", "Mate", "智能手机"})

	// 测试前缀补全
	completions := se.AutoComplete("苹", 5)
	assert.Len(t, completions, 2, "应有2个补全结果")
	assert.Contains(t, completions, "苹果手机", "补全应包含'苹果手机'")
	assert.Contains(t, completions, "苹果平板", "补全应包含'苹果平板'")

	// 测试限制数量
	completions = se.AutoComplete("智", 1)
	assert.Len(t, completions, 1, "应限制为1个补全结果")

	// 测试无匹配
	completions = se.AutoComplete("笔记本", 5)
	assert.Len(t, completions, 0, "应找不到匹配")
}

// 测试热门搜索词功能
func TestSearchEngine_TopSearchTerms(t *testing.T) {
	se := NewSearchEngine()

	// 索引商品并创建不同频率
	se.IndexProduct("P001", []string{"手机", "智能手机", "5G"})
	se.IndexProduct("P002", []string{"手机", "智能手机", "5G"})
	se.IndexProduct("P003", []string{"手机", "智能手机"})
	se.IndexProduct("P004", []string{"平板", "智能平板"})
	se.IndexProduct("P005", []string{"平板"})

	// 获取热门词
	topTerms := se.GetTopSearchTerms(3)
	require.Len(t, topTerms, 3, "应返回3个热门词")

	// 验证顺序（按频率降序）
	assert.Equal(t, "手机", topTerms[0], "最热门词应为'手机'")
	assert.Equal(t, "智能手机", topTerms[1], "第二热门词应为'智能手机'")
	assert.Equal(t, "5G", topTerms[2], "第三热门词应为'5G'")

	// 测试不同限制
	topTerms = se.GetTopSearchTerms(10)
	assert.Len(t, topTerms, 5, "应返回所有5个词")

	topTerms = se.GetTopSearchTerms(0)
	assert.Len(t, topTerms, 0, "限制为0应返回空列表")
}

// 测试搜索引擎并发安全性
func TestSearchEngine_ConcurrentSafety(t *testing.T) {
	se := NewSearchEngine()

	// 并发索引和搜索
	var wg sync.WaitGroup
	concurrency := 50
	wg.Add(concurrency * 2)

	// 并发索引
	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			productID := fmt.Sprintf("P%03d", idx)
			terms := []string{
				fmt.Sprintf("商品%d", idx),
				fmt.Sprintf("类别%d", idx%5),
				"测试商品",
			}
			se.IndexProduct(productID, terms)
		}(i)
	}

	// 并发搜索
	results := make([][]SearchResult, concurrency)
	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			time.Sleep(time.Duration(idx%10) * time.Millisecond)
			results[idx] = se.Search("测试", 10)
		}(i)
	}

	wg.Wait()

	// 验证搜索结果非空
	for i, result := range results {
		if len(result) == 0 {
			t.Errorf("第%d个搜索结果为空", i)
		}
	}

	// 验证最终状态
	finalResults := se.Search("测试", 100)
	assert.NotEmpty(t, finalResults, "最终搜索结果不应为空")
	if len(finalResults) > 0 {
		assert.Equal(t, "测试商品", finalResults[0].Term, "应匹配'测试商品'")
		assert.Len(t, finalResults[0].ProductIDs, concurrency, "应匹配所有商品")
	}
}

// 测试三个组件的集成场景
func TestECommerceIntegration(t *testing.T) {
	// 初始化三个组件
	inventory := NewInventoryManager()
	priceManager := NewPriceManager()
	searchEngine := NewSearchEngine()

	// 创建测试数据
	now := time.Now()
	sku1 := &ProductSku{
		ID:            "SKU001",
		ProductID:     "P001",
		Price:         999.99,
		OriginalPrice: 1299.99,
		Attributes: map[string]string{
			"颜色": "黑色",
			"型号": "X100",
		},
		Stock:     100,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 1. 添加到价格管理器
	err := priceManager.AddProduct(sku1)
	assert.NoError(t, err, "添加商品到价格管理器应成功")

	// 2. 添加到库存管理器
	err = inventory.AddSku(sku1.ID, sku1.Stock, 20, "WH-001")
	assert.NoError(t, err, "添加商品到库存管理器应成功")

	// 3. 添加到搜索引擎
	searchEngine.IndexProduct(sku1.ProductID, []string{
		"X100", "高端", "黑色", "旗舰",
	})

	// 模拟购买流程

	// 1. 搜索商品
	results := searchEngine.Search("X100", 10)
	assert.NotEmpty(t, results, "搜索应返回结果")

	// 2. 查询价格
	product, err := priceManager.GetProduct("SKU001")
	assert.NoError(t, err, "查询价格应成功")
	assert.Equal(t, 999.99, product.Price, "价格应为999.99")

	// 3. 检查库存
	stock, err := inventory.GetStock("SKU001")
	assert.NoError(t, err, "查询库存应成功")
	assert.Equal(t, 100, stock, "库存应为100")

	// 4. 添加到购物车(预留)
	err = inventory.Reserve("SKU001", 2)
	assert.NoError(t, err, "预留库存应成功")

	// 5. 确认购买(提交)
	err = inventory.Commit("SKU001", 2)
	assert.NoError(t, err, "确认库存扣减应成功")

	// 6. 验证最终库存
	finalStock, err := inventory.GetStock("SKU001")
	assert.NoError(t, err, "查询最终库存应成功")
	assert.Equal(t, 98, finalStock, "最终库存应为98")

	// 7. 促销价格更新
	err = priceManager.UpdatePrice("SKU001", 899.99)
	assert.NoError(t, err, "更新价格应成功")

	updatedProduct, _ := priceManager.GetProduct("SKU001")
	assert.Equal(t, 899.99, updatedProduct.Price, "更新后价格应为899.99")
	assert.Equal(t, 999.99, updatedProduct.OriginalPrice, "原价应为999.99")
}
