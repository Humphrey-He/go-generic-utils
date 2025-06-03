// Copyright 2024 Humphrey-He
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package list

import (
	"fmt"
	"time"
)

// 以下代码演示如何在电商平台后端开发中使用list包

// 1. 商品管理示例
func ExampleProductList() {
	// 创建商品列表
	productList := NewProductList(100)

	// 创建几个示例商品
	product1 := Product{
		ID:              "P001",
		Name:            "高端机械键盘",
		Price:           499.00,
		Stock:           100,
		SKU:             "KB-MEC-001",
		Category:        "电脑配件",
		Tags:            []string{"键盘", "机械键盘", "游戏外设"},
		Attributes:      map[string]string{"颜色": "黑色", "轴体": "青轴"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		SalesCount:      0,
		IsActive:        true,
		DiscountPercent: 0,
	}

	product2 := Product{
		ID:              "P002",
		Name:            "无线蓝牙耳机",
		Price:           299.00,
		Stock:           200,
		SKU:             "HP-BT-002",
		Category:        "音频设备",
		Tags:            []string{"耳机", "蓝牙", "无线"},
		Attributes:      map[string]string{"颜色": "白色", "连接": "蓝牙5.0"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		SalesCount:      0,
		IsActive:        true,
		DiscountPercent: 10, // 10%折扣
	}

	// 添加商品到列表
	_ = productList.Append(product1)
	_ = productList.Append(product2)

	// 查找商品
	if foundProduct, err := productList.FindByID("P001"); err == nil {
		fmt.Printf("找到商品：%s，价格：%.2f元\n", foundProduct.Name, foundProduct.Price)
	}

	// 更新库存
	_ = productList.UpdateStock("P001", -5) // 减少5个库存
	_ = productList.UpdateStock("P002", 10) // 增加10个库存

	// 查找特定分类的商品
	audioProducts := productList.FindByCategory("音频设备")
	fmt.Printf("音频设备类别商品数量：%d\n", len(audioProducts))

	// 查找特定标签的商品
	wirelessProducts := productList.FindByTags([]string{"无线"})
	fmt.Printf("无线标签商品数量：%d\n", len(wirelessProducts))

	// 获取折扣商品
	discountedProducts := productList.GetDiscountedProducts()
	for _, p := range discountedProducts {
		fmt.Printf("折扣商品：%s, 原价：%.2f, 折扣后：%.2f\n",
			p.Name, p.Price, p.DiscountedPrice())
	}

	// Output:
	// 找到商品：高端机械键盘，价格：499.00元
	// 音频设备类别商品数量：1
	// 无线标签商品数量：1
	// 折扣商品：无线蓝牙耳机, 原价：299.00, 折扣后：269.10
}

// 2. 购物车示例
func ExampleShoppingCart() {
	// 创建购物车
	cart := NewShoppingCart("user123")

	// 创建示例商品
	keyboard := Product{
		ID:              "P001",
		Name:            "高端机械键盘",
		Price:           499.00,
		Stock:           100,
		SKU:             "KB-MEC-001",
		Category:        "电脑配件",
		Tags:            []string{"键盘", "机械键盘", "游戏外设"},
		Attributes:      map[string]string{"颜色": "黑色", "轴体": "青轴"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsActive:        true,
		DiscountPercent: 0,
	}

	headphone := Product{
		ID:              "P002",
		Name:            "无线蓝牙耳机",
		Price:           299.00,
		Stock:           200,
		SKU:             "HP-BT-002",
		Category:        "音频设备",
		Tags:            []string{"耳机", "蓝牙", "无线"},
		Attributes:      map[string]string{"颜色": "白色", "连接": "蓝牙5.0"},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsActive:        true,
		DiscountPercent: 10, // 10%折扣
	}

	// 添加商品到购物车
	_ = cart.AddItem(keyboard, 1, "黑色-青轴")
	_ = cart.AddItem(headphone, 2, "白色")

	// 更新商品数量
	_ = cart.UpdateItemQuantity("P002", "白色", 3)

	// 计算总价
	total := cart.CalculateTotal()
	fmt.Printf("购物车总价：%.2f元\n", total)

	// 获取选中的商品
	selectedItems := cart.GetSelectedItems()
	fmt.Printf("选中的商品数量：%d\n", len(selectedItems))

	// 移除商品
	_ = cart.RemoveItem("P001", "黑色-青轴")

	// 再次计算总价
	total = cart.CalculateTotal()
	fmt.Printf("移除商品后的购物车总价：%.2f元\n", total)

	// Output:
	// 购物车总价：1307.30元
	// 选中的商品数量：2
	// 移除商品后的购物车总价：807.30元
}

// 3. 订单管理示例
func ExampleOrderList() {
	// 创建订单列表
	orderList := NewOrderList(100)

	// 创建一个示例订单
	order1 := Order{
		ID:     "ORD20240001",
		UserID: "user123",
		Items: []OrderItem{
			{
				ProductID:  "P001",
				Name:       "高端机械键盘",
				SKU:        "KB-MEC-001",
				Price:      499.00,
				Quantity:   1,
				Attributes: "黑色-青轴",
			},
			{
				ProductID:  "P002",
				Name:       "无线蓝牙耳机",
				SKU:        "HP-BT-002",
				Price:      269.10, // 折扣后价格
				Quantity:   2,
				Attributes: "白色",
			},
		},
		Status:       OrderStatusPending,
		TotalAmount:  1037.20,
		PaymentInfo:  map[string]string{"方式": "支付宝", "交易号": "12345678"},
		ShippingInfo: map[string]string{"地址": "北京市海淀区", "收件人": "张三", "电话": "13800138000"},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Notes:        "请尽快发货",
	}

	// 添加订单到列表
	_ = orderList.Append(order1)

	// 查找订单
	if foundOrder, err := orderList.FindByID("ORD20240001"); err == nil {
		fmt.Printf("找到订单：%s，总金额：%.2f元\n", foundOrder.ID, foundOrder.TotalAmount)
	}

	// 更新订单状态为已支付
	_ = orderList.UpdateOrderStatus("ORD20240001", OrderStatusPaid)

	// 查找特定状态的订单
	paidOrders := orderList.FindByStatus(OrderStatusPaid)
	fmt.Printf("已支付订单数量：%d\n", len(paidOrders))

	// 查找特定用户的订单
	userOrders := orderList.FindByUserID("user123")
	fmt.Printf("用户user123的订单数量：%d\n", len(userOrders))

	// Output:
	// 找到订单：ORD20240001，总金额：1037.20元
	// 已支付订单数量：1
	// 用户user123的订单数量：1
}

// 4. 商品搜索示例
func ExampleProductSearchEngine() {
	// 创建商品列表
	productList := NewProductList(100)

	// 添加示例商品
	products := []Product{
		{
			ID:              "P001",
			Name:            "高端机械键盘",
			Price:           499.00,
			Stock:           100,
			SKU:             "KB-MEC-001",
			Category:        "电脑配件",
			Tags:            []string{"键盘", "机械键盘", "游戏外设"},
			Attributes:      map[string]string{"颜色": "黑色", "轴体": "青轴"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			SalesCount:      50,
			IsActive:        true,
			DiscountPercent: 0,
		},
		{
			ID:              "P002",
			Name:            "无线蓝牙耳机",
			Price:           299.00,
			Stock:           200,
			SKU:             "HP-BT-002",
			Category:        "音频设备",
			Tags:            []string{"耳机", "蓝牙", "无线"},
			Attributes:      map[string]string{"颜色": "白色", "连接": "蓝牙5.0"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			SalesCount:      80,
			IsActive:        true,
			DiscountPercent: 10,
		},
		{
			ID:              "P003",
			Name:            "游戏鼠标",
			Price:           199.00,
			Stock:           150,
			SKU:             "MS-GAME-003",
			Category:        "电脑配件",
			Tags:            []string{"鼠标", "游戏外设", "RGB"},
			Attributes:      map[string]string{"颜色": "黑色", "DPI": "16000"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			SalesCount:      100,
			IsActive:        true,
			DiscountPercent: 5,
		},
		{
			ID:              "P004",
			Name:            "高清摄像头",
			Price:           399.00,
			Stock:           80,
			SKU:             "CAM-HD-004",
			Category:        "电脑配件",
			Tags:            []string{"摄像头", "高清", "直播"},
			Attributes:      map[string]string{"分辨率": "1080P", "接口": "USB"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			SalesCount:      30,
			IsActive:        true,
			DiscountPercent: 0,
		},
	}

	for _, p := range products {
		_ = productList.Append(p)
	}

	// 创建搜索引擎
	searchEngine := NewProductSearchEngine(productList)

	// 搜索示例：按关键词搜索
	result := searchEngine.Search("游戏", "", nil, 0, 0, SortBySales, 1, 10)
	fmt.Printf("搜索 '游戏' 找到商品数量：%d\n", result.TotalCount)
	for _, p := range result.Products {
		fmt.Printf("- %s (销量：%d)\n", p.Name, p.SalesCount)
	}

	// 搜索示例：按分类搜索
	result = searchEngine.Search("", "电脑配件", nil, 0, 0, SortByPriceDesc, 1, 10)
	fmt.Printf("\n分类 '电脑配件' 找到商品数量：%d\n", result.TotalCount)
	for _, p := range result.Products {
		fmt.Printf("- %s (价格：%.2f)\n", p.Name, p.Price)
	}

	// 搜索示例：按价格范围搜索
	result = searchEngine.Search("", "", nil, 200, 400, SortByPriceAsc, 1, 10)
	fmt.Printf("\n价格区间 200-400 找到商品数量：%d\n", result.TotalCount)
	for _, p := range result.Products {
		fmt.Printf("- %s (价格：%.2f)\n", p.Name, p.Price)
	}

	// Output:
	// 搜索 '游戏' 找到商品数量：2
	// - 游戏鼠标 (销量：100)
	// - 高端机械键盘 (销量：50)
	//
	// 分类 '电脑配件' 找到商品数量：3
	// - 高清摄像头 (价格：399.00)
	// - 高端机械键盘 (价格：499.00)
	// - 游戏鼠标 (价格：199.00)
	//
	// 价格区间 200-400 找到商品数量：2
	// - 无线蓝牙耳机 (价格：299.00)
	// - 游戏鼠标 (价格：199.00)
}

// 5. 并发安全库存管理示例
func ExampleConcurrentInventory() {
	// 创建并发安全的库存管理系统
	inventory := NewConcurrentInventory[Product](100)

	// 添加商品
	keyboard := Product{
		ID:        "P001",
		Name:      "高端机械键盘",
		Price:     499.00,
		Stock:     100,
		SKU:       "KB-MEC-001",
		Category:  "电脑配件",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	headphone := Product{
		ID:        "P002",
		Name:      "无线蓝牙耳机",
		Price:     299.00,
		Stock:     200,
		SKU:       "HP-BT-002",
		Category:  "音频设备",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	_ = inventory.Append(keyboard)
	_ = inventory.Append(headphone)

	// 安全地更新库存
	err := inventory.UpdateInventory("P001",
		func(p Product) (Product, error) {
			// 模拟库存扣减
			if p.Stock < 5 {
				return p, ErrStockInsufficient
			}
			p.Stock -= 5
			p.UpdatedAt = time.Now()
			return p, nil
		},
		func(p Product) bool {
			return p.ID == "P001"
		})

	if err == nil {
		fmt.Println("成功更新键盘库存")
	} else {
		fmt.Printf("更新库存失败: %v\n", err)
	}

	// 批量更新库存（假设有一个订单需要同时更新多个商品的库存）
	updates := map[string]func(Product) (Product, bool){
		"P001": func(p Product) (Product, bool) {
			if p.Stock >= 2 {
				p.Stock -= 2
				p.UpdatedAt = time.Now()
				return p, true
			}
			return p, false
		},
		"P002": func(p Product) (Product, bool) {
			if p.Stock >= 3 {
				p.Stock -= 3
				p.UpdatedAt = time.Now()
				return p, true
			}
			return p, false
		},
	}

	err = inventory.BatchUpdate(updates, func(p Product) string {
		return p.ID
	})

	if err == nil {
		fmt.Println("成功批量更新库存")
	} else {
		fmt.Printf("批量更新库存失败: %v\n", err)
	}

	// 查看更新后的库存
	_ = inventory.Range(func(_ int, p Product) error {
		fmt.Printf("%s 剩余库存: %d\n", p.Name, p.Stock)
		return nil
	})

	// Output:
	// 成功更新键盘库存
	// 成功批量更新库存
	// 高端机械键盘 剩余库存: 93
	// 无线蓝牙耳机 剩余库存: 197
}

// 6. 优先级订单队列示例
func ExamplePriorityLinkedList() {
	// 创建优先级队列
	orderQueue := NewPriorityLinkedList[string]()

	// 添加不同优先级的订单
	_ = orderQueue.AddWithPriority("普通订单1", 1, string(OrderStatusPaid), time.Now().UnixNano())
	_ = orderQueue.AddWithPriority("加急订单1", 2, string(OrderStatusPaid), time.Now().UnixNano())
	time.Sleep(10 * time.Millisecond) // 等待一段时间再添加
	_ = orderQueue.AddWithPriority("普通订单2", 1, string(OrderStatusPaid), time.Now().UnixNano())
	_ = orderQueue.AddWithPriority("特急订单1", 3, string(OrderStatusPaid), time.Now().UnixNano())

	// 按优先级处理订单
	for orderQueue.LinkedList.Len() > 0 {
		order, _ := orderQueue.PopHighestPriority()
		fmt.Printf("处理订单: %s\n", order)
	}

	// 再次添加一些不同状态的订单
	_ = orderQueue.AddWithPriority("待支付订单1", 1, string(OrderStatusPending), time.Now().UnixNano())
	_ = orderQueue.AddWithPriority("已支付订单1", 1, string(OrderStatusPaid), time.Now().UnixNano())
	_ = orderQueue.AddWithPriority("已支付订单2", 2, string(OrderStatusPaid), time.Now().UnixNano())
	_ = orderQueue.AddWithPriority("待支付订单2", 2, string(OrderStatusPending), time.Now().UnixNano())

	// 获取特定状态的订单
	paidOrders := orderQueue.FilterByStatus(string(OrderStatusPaid))
	fmt.Printf("已支付订单数量: %d\n", len(paidOrders))

	pendingOrders := orderQueue.FilterByStatus(string(OrderStatusPending))
	fmt.Printf("待支付订单数量: %d\n", len(pendingOrders))

	// Output:
	// 处理订单: 特急订单1
	// 处理订单: 加急订单1
	// 处理订单: 普通订单1
	// 处理订单: 普通订单2
	// 已支付订单数量: 2
	// 待支付订单数量: 2
}
