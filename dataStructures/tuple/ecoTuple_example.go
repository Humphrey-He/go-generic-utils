package tuple

import (
	"encoding/json"
	"fmt"
	"time"
)

// 以下代码展示了如何在电商平台后端开发中使用tuple包

// 1. 商品价格管理示例
func ExampleProductPrice() {
	// 创建商品价格列表
	prices := ProductPriceList{
		NewProductPrice("P001", 199.00),
		NewProductPrice("P002", 299.50),
		NewProductPrice("P003", 99.99),
		NewProductPrice("P004", 499.00),
	}

	// 按价格排序
	prices.SortByPrice()
	fmt.Println("按价格升序排序:")
	for _, p := range prices {
		fmt.Println(p)
	}

	// 过滤特定价格区间的商品
	filtered := prices.FilterByPriceRange(100, 300)
	fmt.Println("\n价格区间 100-300 的商品:")
	for _, p := range filtered {
		fmt.Println(p)
	}

	// 计算总价
	total := prices.TotalPrice()
	fmt.Printf("\n所有商品总价: $%.2f\n", total)

	// 转换为通用Pair
	pairs := make([]Pair[string, float64], len(prices))
	for i, p := range prices {
		pairs[i] = p.AsPair()
	}

	// 从Pair创建map
	priceMap := MapFromPairs(pairs)
	fmt.Println("\n价格映射:")
	for id, price := range priceMap {
		fmt.Printf("商品 %s: $%.2f\n", id, price)
	}

	// Output:
	// 按价格升序排序:
	// 商品 P003: $99.99
	// 商品 P001: $199.00
	// 商品 P002: $299.50
	// 商品 P004: $499.00
	//
	// 价格区间 100-300 的商品:
	// 商品 P001: $199.00
	// 商品 P002: $299.50
	//
	// 所有商品总价: $1097.49
	//
	// 价格映射:
	// 商品 P001: $199.00
	// 商品 P002: $299.50
	// 商品 P003: $99.99
	// 商品 P004: $499.00
}

// 2. 库存管理示例
func ExampleProductStock() {
	// 创建商品库存列表
	stocks := ProductStockList{
		NewProductStock("P001", 100),
		NewProductStock("P002", 5),
		NewProductStock("P003", 0),
		NewProductStock("P004", 50),
	}

	// 设置库存警告阈值
	stocks[1].LowWarning = 10
	stocks[3].LowWarning = 20
	stocks[3].HighWarning = 100

	// 查找低库存商品
	lowStocks := stocks.FindLowStock()
	fmt.Println("低库存商品:")
	for _, s := range lowStocks {
		fmt.Println(s)
	}

	// 查找缺货商品
	outOfStocks := stocks.FindOutOfStock()
	fmt.Println("\n缺货商品:")
	for _, s := range outOfStocks {
		fmt.Println(s)
	}

	// 按库存排序
	stocks.SortByStock()
	fmt.Println("\n按库存排序:")
	for _, s := range stocks {
		fmt.Println(s)
	}

	// 计算总库存
	totalStock := stocks.TotalStock()
	fmt.Printf("\n总库存: %d 件\n", totalStock)

	// Output:
	// 低库存商品:
	// 商品 P002: 库存 5
	// 商品 P004: 库存 50
	//
	// 缺货商品:
	// 商品 P003: 库存 0
	//
	// 按库存排序:
	// 商品 P003: 库存 0
	// 商品 P002: 库存 5
	// 商品 P004: 库存 50
	// 商品 P001: 库存 100
	//
	// 总库存: 155 件
}

// 3. 订单管理示例
func ExampleUserOrder() {
	// 创建用户订单列表
	orders := UserOrderList{
		NewUserOrder("U001", "ORD001", 199.00),
		NewUserOrder("U002", "ORD002", 599.50),
		NewUserOrder("U001", "ORD003", 299.99),
		NewUserOrder("U003", "ORD004", 899.00),
	}

	// 模拟订单状态变化
	orders[0].Status = "已支付"
	orders[1].Status = "已发货"
	orders[2].Status = "已支付"

	// 查询用户订单
	user1Orders := orders.FilterByUser("U001")
	fmt.Println("用户U001的订单:")
	for _, o := range user1Orders {
		fmt.Println(o)
	}

	// 按状态过滤
	paidOrders := orders.FilterByStatus("已支付")
	fmt.Println("\n已支付订单:")
	for _, o := range paidOrders {
		fmt.Println(o)
	}

	// 按金额排序
	orders.SortByAmount()
	fmt.Println("\n按金额排序:")
	for _, o := range orders {
		fmt.Println(o)
	}

	// 计算总金额
	totalAmount := orders.SumAmount()
	fmt.Printf("\n所有订单总金额: $%.2f\n", totalAmount)

	// 转换为三元组
	triples := make([]Triple[string, string, float64], len(orders))
	for i, o := range orders {
		triples[i] = o.AsTriple()
	}

	// 从三元组中提取信息
	fmt.Println("\n从三元组中提取的订单信息:")
	for _, t := range triples {
		userID, orderID, amount := t.Split()
		fmt.Printf("用户 %s, 订单 %s, 金额 $%.2f\n", userID, orderID, amount)
	}

	// Output:
	// 用户U001的订单:
	// 用户 U001 的订单 ORD001: $199.00 (已支付)
	// 用户 U001 的订单 ORD003: $299.99 (已支付)
	//
	// 已支付订单:
	// 用户 U001 的订单 ORD001: $199.00 (已支付)
	// 用户 U001 的订单 ORD003: $299.99 (已支付)
	//
	// 按金额排序:
	// 用户 U001 的订单 ORD001: $199.00 (已支付)
	// 用户 U001 的订单 ORD003: $299.99 (已支付)
	// 用户 U002 的订单 ORD002: $599.50 (已发货)
	// 用户 U003 的订单 ORD004: $899.00 (待支付)
	//
	// 所有订单总金额: $1997.49
	//
	// 从三元组中提取的订单信息:
	// 用户 U001, 订单 ORD001, 金额 $199.00
	// 用户 U002, 订单 ORD002, 金额 $599.50
	// 用户 U001, 订单 ORD003, 金额 $299.99
	// 用户 U003, 订单 ORD004, 金额 $899.00
}

// 4. 商品分类管理示例
func ExampleProductCategory() {
	// 创建商品分类列表
	categories := ProductCategoryList{
		NewProductCategory("P001", "电子产品", "手机"),
		NewProductCategory("P002", "电子产品", "电脑"),
		NewProductCategory("P003", "服装", "男装"),
		NewProductCategory("P004", "服装", "女装"),
		NewProductCategory("P005", "电子产品", "耳机"),
	}

	// 按分类过滤
	electronics := categories.FilterByCategory("电子产品")
	fmt.Println("电子产品分类下的商品:")
	for _, c := range electronics {
		fmt.Println(c)
	}

	// 按子分类过滤
	computers := categories.FilterBySubCategory("电脑")
	fmt.Println("\n电脑子分类下的商品:")
	for _, c := range computers {
		fmt.Println(c)
	}

	// 统计各分类商品数量
	categoryCounts := categories.CountByCategory()
	fmt.Println("\n各分类商品数量:")
	for category, count := range categoryCounts {
		fmt.Printf("%s: %d件\n", category, count)
	}

	// 转换为三元组并操作
	triples := make([]Triple[string, string, string], len(categories))
	for i, c := range categories {
		triples[i] = c.AsTriple()
	}

	// 从三元组中过滤电子产品-手机
	fmt.Println("\n电子产品-手机类商品:")
	for _, t := range triples {
		productID, category, subCategory := t.Split()
		if category == "电子产品" && subCategory == "手机" {
			fmt.Printf("商品 %s: %s > %s\n", productID, category, subCategory)
		}
	}

	// Output:
	// 电子产品分类下的商品:
	// 商品 P001: 电子产品 > 手机
	// 商品 P002: 电子产品 > 电脑
	// 商品 P005: 电子产品 > 耳机
	//
	// 电脑子分类下的商品:
	// 商品 P002: 电子产品 > 电脑
	//
	// 各分类商品数量:
	// 电子产品: 3件
	// 服装: 2件
	//
	// 电子产品-手机类商品:
	// 商品 P001: 电子产品 > 手机
}

// 5. 商品评分示例
func ExampleProductRating() {
	// 创建商品评分列表
	ratings := ProductRatingList{
		NewProductRating("P001", "U001", 4.5, "质量不错"),
		NewProductRating("P001", "U002", 3.0, "一般般"),
		NewProductRating("P002", "U001", 5.0, "非常好"),
		NewProductRating("P002", "U003", 4.0, "物美价廉"),
		NewProductRating("P001", "U004", 2.0, "有点失望"),
	}

	// 按商品分组
	groupedRatings := ratings.GroupByProduct()
	for productID, productRatings := range groupedRatings {
		avgRating := productRatings.AverageRating()
		fmt.Printf("商品 %s 的平均评分: %.1f (基于 %d 条评价)\n",
			productID, avgRating, len(productRatings))
	}

	// 筛选高评分
	highRatings := ratings.FilterByMinRating(4.0)
	fmt.Println("\n4分以上的评价:")
	for _, r := range highRatings {
		fmt.Println(r)
	}

	// 转换为键值对
	pairs := make([]Pair[string, float64], len(ratings))
	for i, r := range ratings {
		pairs[i] = r.AsPair()
	}

	// 使用通用Filter函数过滤
	lowRatingPairs := Filter(pairs, func(productID string, rating float64) bool {
		return rating < 3.0
	})

	fmt.Println("\n低于3分的评价:")
	for _, p := range lowRatingPairs {
		fmt.Printf("商品 %s: %.1f分\n", p.Key, p.Value)
	}

	// Output:
	// 商品 P001 的平均评分: 3.2 (基于 3 条评价)
	// 商品 P002 的平均评分: 4.5 (基于 2 条评价)
	//
	// 4分以上的评价:
	// 商品 P001: 4.5分 (U001)
	// 商品 P002: 5.0分 (U001)
	// 商品 P002: 4.0分 (U003)
	//
	// 低于3分的评价:
	// 商品 P001: 2.0分
}

// 6. 购物车管理示例
func ExampleCartItem() {
	// 创建购物车
	cart := CartItemList{
		NewCartItem("P001", 2, 199.00),
		NewCartItem("P002", 1, 299.50),
		NewCartItem("P003", 3, 99.99),
	}

	// 模拟取消选中一个商品
	cart[2].Selected = false

	// 计算选中商品的总数量和金额
	totalQty := cart.TotalQuantity()
	totalAmount := cart.TotalAmount()
	fmt.Printf("购物车: 选中 %d 件商品, 总金额 $%.2f\n", totalQty, totalAmount)

	// 获取选中的商品
	selectedItems := cart.FilterSelected()
	fmt.Println("\n选中的商品:")
	for _, item := range selectedItems {
		fmt.Printf("%s, 小计: $%.2f\n", item, item.TotalPrice())
	}

	// 更新商品数量
	updated := cart.UpdateQuantity("P001", 5)
	if updated {
		fmt.Println("\n更新商品P001数量后:")
		for _, item := range cart {
			if item.ProductID == "P001" {
				fmt.Printf("%s, 小计: $%.2f\n", item, item.TotalPrice())
			}
		}
	}

	// 转换为通用Pair
	pairs := make([]Pair[string, int], len(cart))
	for i, item := range cart {
		pairs[i] = item.AsPair()
	}

	// 使用Map函数转换为商品ID到总价的映射
	totalPrices := Map(pairs, func(productID string, quantity int) (string, float64) {
		for _, item := range cart {
			if item.ProductID == productID {
				return productID, item.TotalPrice()
			}
		}
		return productID, 0
	})

	fmt.Println("\n各商品总价:")
	for _, p := range totalPrices {
		fmt.Printf("商品 %s: $%.2f\n", p.Key, p.Value)
	}

	// Output:
	// 购物车: 选中 3 件商品, 总金额 $697.50
	//
	// 选中的商品:
	// 商品 P001: 2件 x $199.00, 小计: $398.00
	// 商品 P002: 1件 x $299.50, 小计: $299.50
	//
	// 更新商品P001数量后:
	// 商品 P001: 5件 x $199.00, 小计: $995.00
	//
	// 各商品总价:
	// 商品 P001: $995.00
	// 商品 P002: $299.50
	// 商品 P003: $299.97
}

// 7. 销售趋势分析示例
func ExampleTimeValuePair() {
	// 创建一周销售数据
	now := time.Now()
	salesData := TimeValueList{
		NewTimeValuePair(now.AddDate(0, 0, -6), 2345.67),
		NewTimeValuePair(now.AddDate(0, 0, -5), 3456.78),
		NewTimeValuePair(now.AddDate(0, 0, -4), 2345.67),
		NewTimeValuePair(now.AddDate(0, 0, -3), 4567.89),
		NewTimeValuePair(now.AddDate(0, 0, -2), 5678.90),
		NewTimeValuePair(now.AddDate(0, 0, -1), 6789.01),
		NewTimeValuePair(now, 7890.12),
	}

	// 计算总销售额和平均销售额
	totalSales := salesData.SumValues()
	avgSales := salesData.AverageValue()
	fmt.Printf("总销售额: $%.2f, 日均销售额: $%.2f\n", totalSales, avgSales)

	// 过滤最近3天的销售数据
	recentSales := salesData.FilterByTimeRange(now.AddDate(0, 0, -2), time.Time{})
	fmt.Println("\n最近3天销售数据:")
	for _, sale := range recentSales {
		fmt.Printf("%s: $%.2f\n",
			sale.Time.Format("2006-01-02"), sale.Value)
	}

	// 按天分组并计算每日总额
	dailyTotals := salesData.DailyTotal()
	fmt.Println("\n每日销售总额:")
	for _, pair := range dailyTotals {
		fmt.Printf("%s: $%.2f\n", pair.Key, pair.Value)
	}

	// 转换为JSON格式
	pair := salesData[0].AsPair()
	jsonData, _ := json.MarshalIndent(pair, "", "  ")
	fmt.Println("\n销售数据JSON格式示例:")
	fmt.Println(string(jsonData))

	// Output:
	// 总销售额: $33074.04, 日均销售额: $4725.15
	//
	// 最近3天销售数据:
	// 2024-11-14: $5678.90
	// 2024-11-15: $6789.01
	// 2024-11-16: $7890.12
	//
	// 每日销售总额:
	// 2024-11-10: $2345.67
	// 2024-11-11: $3456.78
	// 2024-11-12: $2345.67
	// 2024-11-13: $4567.89
	// 2024-11-14: $5678.90
	// 2024-11-15: $6789.01
	// 2024-11-16: $7890.12
	//
	// 销售数据JSON格式示例:
	// {
	//   "key": "2024-11-10T00:00:00Z",
	//   "value": 2345.67
	// }
}

// 8. 基本元组操作示例
func ExamplePair() {
	// 创建键值对
	p := NewPair("商品ID", "P001")
	fmt.Println("键值对:", p)

	// 分解键值对
	key, value := p.Split()
	fmt.Printf("键: %v, 值: %v\n", key, value)

	// 创建多个键值对
	keys := []string{"名称", "价格", "库存"}
	values := []interface{}{"高端机械键盘", 499.00, 100}
	pairs, _ := NewPairs(keys, values)

	fmt.Println("\n商品属性:")
	for _, pair := range pairs {
		fmt.Printf("%v\n", pair)
	}

	// 展平键值对数组
	flat := FlattenPairs(pairs)
	fmt.Println("\n展平后的数组:", flat)

	// 创建三元组
	t := NewTriple("P001", "高端机械键盘", 499.00)
	fmt.Println("\n三元组:", t)

	// 分解三元组
	id, name, price := t.Split()
	fmt.Printf("商品ID: %v, 名称: %v, 价格: %v\n", id, name, price)

	// Output:
	// 键值对: <商品ID, P001>
	// 键: 商品ID, 值: P001
	//
	// 商品属性:
	// <名称, 高端机械键盘>
	// <价格, 499>
	// <库存, 100>
	//
	// 展平后的数组: [名称 高端机械键盘 价格 499 库存 100]
	//
	// 三元组: <P001, 高端机械键盘, 499>
	// 商品ID: P001, 名称: 高端机械键盘, 价格: 499
}
