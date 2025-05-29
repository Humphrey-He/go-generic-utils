package tree

import (
	"fmt"
	"time"
)

// ExampleInventoryManager 展示库存管理器的基本用法
func ExampleInventoryManager() {
	// 创建库存管理器
	inventory := NewInventoryManager()

	// 添加库存事件处理器
	inventory.AddEventHandler(func(item *InventoryItem, action string, quantity int) {
		fmt.Printf("库存事件: SKU=%s, 动作=%s, 数量=%d\n", item.Sku, action, quantity)
	})

	// 添加SKU到库存
	inventory.AddSku("SKU001", 100, 20, "WH-001")
	inventory.AddSku("SKU002", 50, 10, "WH-001")
	inventory.AddSku("SKU003", 30, 5, "WH-002")

	// 查询库存
	stock, _ := inventory.GetStock("SKU001")
	fmt.Println("SKU001当前库存:", stock)

	// 预留库存(下单)
	inventory.Reserve("SKU001", 10)
	availableStock, _ := inventory.GetStock("SKU001")
	fmt.Println("SKU001可用库存:", availableStock)

	// 确认库存扣减(付款)
	inventory.Commit("SKU001", 8)

	// 释放预留库存(取消部分)
	inventory.Release("SKU001", 2)

	// 补充库存
	inventory.Restock("SKU002", 20)

	// 批量更新库存
	updates := map[string]int{
		"SKU001": 10, // 增加10个
		"SKU002": -5, // 减少5个
	}
	inventory.BatchUpdateStock(updates)

	// 获取低库存商品
	lowStockItems := inventory.GetLowStockItems()
	fmt.Println("低库存商品数量:", len(lowStockItems))

	// 获取指定仓库的商品
	wh001Items := inventory.GetItemsByWarehouse("WH-001")
	fmt.Println("WH-001仓库商品数量:", len(wh001Items))

	// Output:
	// SKU001当前库存: 100
	// SKU001可用库存: 90
	// 低库存商品数量: 0
	// WH-001仓库商品数量: 2
}

// ExamplePriceManager 展示价格管理器的基本用法
func ExamplePriceManager() {
	// 创建价格管理器
	priceManager := NewPriceManager()

	// 添加商品
	now := time.Now()

	priceManager.AddProduct(&ProductSku{
		ID:            "SKU001",
		ProductID:     "P001",
		Price:         99.99,
		OriginalPrice: 129.99,
		Attributes: map[string]string{
			"颜色": "红色",
			"尺寸": "M",
		},
		Stock:      100,
		SalesCount: 50,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	priceManager.AddProduct(&ProductSku{
		ID:            "SKU002",
		ProductID:     "P001",
		Price:         109.99,
		OriginalPrice: 139.99,
		Attributes: map[string]string{
			"颜色": "蓝色",
			"尺寸": "L",
		},
		Stock:      80,
		SalesCount: 30,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	priceManager.AddProduct(&ProductSku{
		ID:            "SKU003",
		ProductID:     "P002",
		Price:         199.99,
		OriginalPrice: 249.99,
		Attributes: map[string]string{
			"颜色": "黑色",
			"尺寸": "XL",
		},
		Stock:      60,
		SalesCount: 20,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	})

	// 获取商品信息
	sku, _ := priceManager.GetProduct("SKU001")
	fmt.Printf("商品价格: %.2f, 原价: %.2f\n", sku.Price, sku.OriginalPrice)

	// 更新价格
	priceManager.UpdatePrice("SKU001", 89.99)
	updatedSku, _ := priceManager.GetProduct("SKU001")
	fmt.Printf("更新后价格: %.2f, 原价: %.2f\n", updatedSku.Price, updatedSku.OriginalPrice)

	// 按价格区间查询
	rangeProducts, _ := priceManager.GetProductsInPriceRange(80, 150)
	fmt.Println("价格在80-150之间的商品数量:", len(rangeProducts))

	// 按价格排序
	ascendingProducts := priceManager.GetProductsSortedByPrice(true)
	if len(ascendingProducts) > 0 {
		fmt.Printf("最低价商品: %s, 价格: %.2f\n", ascendingProducts[0].ID, ascendingProducts[0].Price)
	}

	descendingProducts := priceManager.GetProductsSortedByPrice(false)
	if len(descendingProducts) > 0 {
		fmt.Printf("最高价商品: %s, 价格: %.2f\n", descendingProducts[0].ID, descendingProducts[0].Price)
	}

	// 批量更新价格
	priceUpdates := map[string]float64{
		"SKU002": 99.99,
		"SKU003": 189.99,
	}
	priceManager.BatchUpdatePrices(priceUpdates)

	// Output:
	// 商品价格: 99.99, 原价: 129.99
	// 更新后价格: 89.99, 原价: 99.99
	// 价格在80-150之间的商品数量: 2
	// 最低价商品: SKU001, 价格: 89.99
	// 最高价商品: SKU003, 价格: 199.99
}

// ExampleSearchEngine 展示搜索引擎的基本用法
func ExampleSearchEngine() {
	// 创建搜索引擎
	searchEngine := NewSearchEngine()

	// 为商品创建索引
	searchEngine.IndexProduct("P001", []string{"苹果", "手机", "iPhone", "智能手机", "5G"})
	searchEngine.IndexProduct("P002", []string{"华为", "手机", "Mate", "智能手机", "5G"})
	searchEngine.IndexProduct("P003", []string{"小米", "手机", "RedMi", "智能手机", "5G"})
	searchEngine.IndexProduct("P004", []string{"苹果", "平板", "iPad", "智能平板"})
	searchEngine.IndexProduct("P005", []string{"华为", "平板", "MatePad", "智能平板"})

	// 搜索
	results := searchEngine.Search("手机", 5)
	fmt.Println("搜索'手机'的结果数:", len(results))
	if len(results) > 0 {
		fmt.Printf("第一个结果: 词=%s, 分数=%d, 商品数=%d\n",
			results[0].Term, results[0].Score, len(results[0].ProductIDs))
	}

	// 自动完成
	completions := searchEngine.AutoComplete("智", 3)
	fmt.Println("以'智'开头的补全结果:", completions)

	// 获取热门搜索词
	topTerms := searchEngine.GetTopSearchTerms(3)
	fmt.Println("热门搜索词:", topTerms)

	// Output:
	// 搜索'手机'的结果数: 1
	// 第一个结果: 词=手机, 分数=3, 商品数=3
	// 以'智'开头的补全结果: [智能平板 智能手机]
	// 热门搜索词: [5G 智能手机 手机]
}

// ExampleECommerceIntegration 展示将库存、价格和搜索整合使用的示例
func ExampleECommerceIntegration() {
	// 创建各个组件
	inventory := NewInventoryManager()
	priceManager := NewPriceManager()
	searchEngine := NewSearchEngine()

	// 初始化商品数据
	now := time.Now()

	// 第一个商品：iPhone 13
	sku1 := &ProductSku{
		ID:            "SKU001",
		ProductID:     "P001",
		Price:         5999.00,
		OriginalPrice: 6999.00,
		Attributes: map[string]string{
			"颜色": "午夜色",
			"内存": "128GB",
		},
		Stock:      100,
		SalesCount: 50,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 第二个商品：iPhone 13 Pro
	sku2 := &ProductSku{
		ID:            "SKU002",
		ProductID:     "P002",
		Price:         7999.00,
		OriginalPrice: 8999.00,
		Attributes: map[string]string{
			"颜色": "远峰蓝",
			"内存": "256GB",
		},
		Stock:      80,
		SalesCount: 30,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 第三个商品：华为 Mate 40 Pro
	sku3 := &ProductSku{
		ID:            "SKU003",
		ProductID:     "P003",
		Price:         6999.00,
		OriginalPrice: 7599.00,
		Attributes: map[string]string{
			"颜色": "釉白色",
			"内存": "256GB",
		},
		Stock:      60,
		SalesCount: 20,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 1. 添加到价格管理器
	priceManager.AddProduct(sku1)
	priceManager.AddProduct(sku2)
	priceManager.AddProduct(sku3)

	// 2. 添加到库存管理器
	inventory.AddSku(sku1.ID, sku1.Stock, 20, "WH-001")
	inventory.AddSku(sku2.ID, sku2.Stock, 15, "WH-001")
	inventory.AddSku(sku3.ID, sku3.Stock, 10, "WH-002")

	// 3. 添加到搜索引擎
	searchEngine.IndexProduct(sku1.ProductID, []string{
		"iPhone", "苹果", "手机", "智能手机", "5G", "128GB", "午夜色",
	})

	searchEngine.IndexProduct(sku2.ProductID, []string{
		"iPhone", "Pro", "苹果", "手机", "智能手机", "5G", "256GB", "远峰蓝",
	})

	searchEngine.IndexProduct(sku3.ProductID, []string{
		"华为", "Mate", "手机", "智能手机", "5G", "256GB", "釉白色",
	})

	// 模拟搜索流程
	fmt.Println("--- 模拟搜索和购买流程 ---")

	// 1. 用户搜索
	searchResults := searchEngine.Search("iPhone", 10)
	fmt.Printf("找到 %d 个相关搜索词\n", len(searchResults))

	// 2. 用户从搜索结果中选择了iPhone 13 (P001)
	// 查询该商品的SKU
	fmt.Printf("选择了商品ID: P001\n")

	// 3. 查询价格
	selectedSku, _ := priceManager.GetProduct("SKU001")
	fmt.Printf("商品价格: %.2f 元\n", selectedSku.Price)

	// 4. 检查库存
	availableStock, _ := inventory.GetStock("SKU001")
	fmt.Printf("可用库存: %d\n", availableStock)

	// 5. 添加到购物车(预留库存)
	quantity := 2
	err := inventory.Reserve("SKU001", quantity)
	if err == nil {
		fmt.Printf("已预留 %d 件商品\n", quantity)
	}

	// 6. 下单并支付(确认库存扣减)
	err = inventory.Commit("SKU001", quantity)
	if err == nil {
		fmt.Printf("已购买 %d 件商品\n", quantity)
	}

	// 7. 查询最终库存
	finalStock, _ := inventory.GetStock("SKU001")
	fmt.Printf("最终库存: %d\n", finalStock)

	// Output:
	// --- 模拟搜索和购买流程 ---
	// 找到 2 个相关搜索词
	// 选择了商品ID: P001
	// 商品价格: 5999.00 元
	// 可用库存: 100
	// 已预留 2 件商品
	// 已购买 2 件商品
	// 最终库存: 98
}
