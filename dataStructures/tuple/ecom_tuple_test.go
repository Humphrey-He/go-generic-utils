package tuple

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ProductPrice类型测试
func TestProductPrice(t *testing.T) {
	t.Run("创建和基本属性", func(t *testing.T) {
		pp := NewProductPrice("P001", 299.99)
		assert.Equal(t, "P001", pp.ProductID, "ProductID应匹配")
		assert.Equal(t, 299.99, pp.Price, "Price应匹配")
	})

	t.Run("字符串表示", func(t *testing.T) {
		pp := NewProductPrice("P001", 299.99)
		expected := "商品 P001: $299.99"
		assert.Equal(t, expected, pp.String(), "String()应返回格式化的字符串")
	})

	t.Run("转换为Pair", func(t *testing.T) {
		pp := NewProductPrice("P001", 299.99)
		pair := pp.AsPair()
		assert.Equal(t, "P001", pair.Key, "Key应匹配ProductID")
		assert.Equal(t, 299.99, pair.Value, "Value应匹配Price")
	})
}

// ProductPriceList类型测试
func TestProductPriceList(t *testing.T) {
	prices := ProductPriceList{
		NewProductPrice("P001", 199.00),
		NewProductPrice("P002", 299.50),
		NewProductPrice("P003", 99.99),
		NewProductPrice("P004", 499.00),
	}

	t.Run("按价格排序", func(t *testing.T) {
		// 复制一份避免影响其他测试
		priceCopy := make(ProductPriceList, len(prices))
		copy(priceCopy, prices)

		// 测试升序排序
		priceCopy.SortByPrice()
		assert.Equal(t, "P003", priceCopy[0].ProductID, "最低价格商品应在首位")
		assert.Equal(t, "P004", priceCopy[3].ProductID, "最高价格商品应在末位")

		// 测试降序排序
		priceCopy.SortByPriceDesc()
		assert.Equal(t, "P004", priceCopy[0].ProductID, "最高价格商品应在首位")
		assert.Equal(t, "P003", priceCopy[3].ProductID, "最低价格商品应在末位")
	})

	t.Run("按价格区间过滤", func(t *testing.T) {
		// 测试在特定价格区间内的商品
		filtered := prices.FilterByPriceRange(100, 300)
		assert.Len(t, filtered, 2, "价格在100-300之间的商品应有2个")
		if len(filtered) == 2 {
			productIds := []string{filtered[0].ProductID, filtered[1].ProductID}
			assert.Contains(t, productIds, "P001", "应包含P001")
			assert.Contains(t, productIds, "P002", "应包含P002")
		}

		// 测试最小价格过滤
		minFiltered := prices.FilterByPriceRange(300, 0)
		assert.Len(t, minFiltered, 1, "价格>=300的商品应有1个")
		if len(minFiltered) > 0 {
			assert.Equal(t, "P004", minFiltered[0].ProductID, "应包含P004")
		}
	})

	t.Run("计算总价", func(t *testing.T) {
		total := prices.TotalPrice()
		expected := 199.00 + 299.50 + 99.99 + 499.00
		assert.InDelta(t, expected, total, 0.01, "总价计算应正确")
	})
}

// ProductStock类型测试
func TestProductStock(t *testing.T) {
	t.Run("创建和基本属性", func(t *testing.T) {
		ps := NewProductStock("P001", 100)
		assert.Equal(t, "P001", ps.ProductID, "ProductID应匹配")
		assert.Equal(t, 100, ps.Stock, "Stock应匹配")
		assert.False(t, ps.UpdatedAt.IsZero(), "UpdatedAt应被设置")
	})

	t.Run("字符串表示", func(t *testing.T) {
		ps := NewProductStock("P001", 100)
		expected := "商品 P001: 库存 100"
		assert.Equal(t, expected, ps.String(), "String()应返回格式化的字符串")
	})

	t.Run("库存状态检查", func(t *testing.T) {
		ps := NewProductStock("P001", 8)

		// 初始状态
		assert.False(t, ps.IsLowStock(), "未设置警告阈值时不应为低库存")
		assert.False(t, ps.IsHighStock(), "未设置警告阈值时不应为高库存")

		// 设置警告阈值
		ps.LowWarning = 10
		ps.HighWarning = 100

		// 检查低库存
		assert.True(t, ps.IsLowStock(), "库存<LowWarning应判定为低库存")

		// 检查高库存
		ps.Stock = 150
		assert.True(t, ps.IsHighStock(), "库存>HighWarning应判定为高库存")
	})

	t.Run("转换为Pair", func(t *testing.T) {
		ps := NewProductStock("P001", 100)
		pair := ps.AsPair()
		assert.Equal(t, "P001", pair.Key, "Key应匹配ProductID")
		assert.Equal(t, 100, pair.Value, "Value应匹配Stock")
	})
}

// ProductStockList类型测试
func TestProductStockList(t *testing.T) {
	stocks := ProductStockList{
		NewProductStock("P001", 100),
		NewProductStock("P002", 5),
		NewProductStock("P003", 0),
		NewProductStock("P004", 50),
	}

	// 设置库存警告阈值
	stocks[1].LowWarning = 10
	stocks[3].LowWarning = 60

	t.Run("按库存排序", func(t *testing.T) {
		// 复制一份避免影响其他测试
		stockCopy := make(ProductStockList, len(stocks))
		copy(stockCopy, stocks)

		stockCopy.SortByStock()
		assert.Equal(t, "P003", stockCopy[0].ProductID, "库存最少的商品应在首位")
		assert.Equal(t, "P001", stockCopy[3].ProductID, "库存最多的商品应在末位")
	})

	t.Run("查找低库存商品", func(t *testing.T) {
		lowStocks := stocks.FindLowStock()
		assert.Len(t, lowStocks, 2, "应有2个低库存商品")
		assert.Equal(t, "P002", lowStocks[0].ProductID, "P002应为低库存")
		assert.Equal(t, "P004", lowStocks[1].ProductID, "P004应为低库存")
	})

	t.Run("查找缺货商品", func(t *testing.T) {
		outOfStock := stocks.FindOutOfStock()
		assert.Len(t, outOfStock, 1, "应有1个缺货商品")
		assert.Equal(t, "P003", outOfStock[0].ProductID, "P003应为缺货")
	})

	t.Run("计算总库存", func(t *testing.T) {
		total := stocks.TotalStock()
		expected := 100 + 5 + 0 + 50
		assert.Equal(t, expected, total, "总库存计算应正确")
	})
}

// UserOrder类型测试
func TestUserOrder(t *testing.T) {
	t.Run("创建和基本属性", func(t *testing.T) {
		uo := NewUserOrder("U001", "ORD001", 299.99)
		assert.Equal(t, "U001", uo.UserID, "UserID应匹配")
		assert.Equal(t, "ORD001", uo.OrderID, "OrderID应匹配")
		assert.Equal(t, 299.99, uo.Amount, "Amount应匹配")
		assert.Equal(t, "待支付", uo.Status, "默认状态应为待支付")
		assert.False(t, uo.OrderTime.IsZero(), "OrderTime应被设置")
	})

	t.Run("字符串表示", func(t *testing.T) {
		uo := NewUserOrder("U001", "ORD001", 299.99)
		expected := "用户 U001 的订单 ORD001: $299.99 (待支付)"
		assert.Equal(t, expected, uo.String(), "String()应返回格式化的字符串")
	})

	t.Run("转换为Triple", func(t *testing.T) {
		uo := NewUserOrder("U001", "ORD001", 299.99)
		triple := uo.AsTriple()
		assert.Equal(t, "U001", triple.First, "First应匹配UserID")
		assert.Equal(t, "ORD001", triple.Second, "Second应匹配OrderID")
		assert.Equal(t, 299.99, triple.Third, "Third应匹配Amount")
	})
}

// UserOrderList类型测试
func TestUserOrderList(t *testing.T) {
	orders := UserOrderList{
		NewUserOrder("U001", "ORD001", 199.00),
		NewUserOrder("U002", "ORD002", 599.50),
		NewUserOrder("U001", "ORD003", 299.99),
		NewUserOrder("U003", "ORD004", 899.00),
	}

	// 设置订单状态
	orders[0].Status = "已支付"
	orders[1].Status = "已发货"
	orders[2].Status = "已支付"

	t.Run("按用户过滤", func(t *testing.T) {
		u1Orders := orders.FilterByUser("U001")
		assert.Len(t, u1Orders, 2, "用户U001应有2个订单")
		assert.Equal(t, "ORD001", u1Orders[0].OrderID, "第一个订单应为ORD001")
		assert.Equal(t, "ORD003", u1Orders[1].OrderID, "第二个订单应为ORD003")
	})

	t.Run("按状态过滤", func(t *testing.T) {
		paidOrders := orders.FilterByStatus("已支付")
		assert.Len(t, paidOrders, 2, "应有2个已支付订单")
		assert.Equal(t, "U001", paidOrders[0].UserID, "第一个已支付订单应属于U001")
		assert.Equal(t, "U001", paidOrders[1].UserID, "第二个已支付订单应属于U001")
	})

	t.Run("按时间排序", func(t *testing.T) {
		// 为了测试排序，手动设置不同的订单时间
		now := time.Now()
		ordersForSort := UserOrderList{
			{UserID: "U001", OrderID: "ORD001", Amount: 100, OrderTime: now.Add(-2 * time.Hour)},
			{UserID: "U002", OrderID: "ORD002", Amount: 200, OrderTime: now.Add(-1 * time.Hour)},
			{UserID: "U003", OrderID: "ORD003", Amount: 300, OrderTime: now},
		}

		ordersForSort.SortByTime()
		assert.Equal(t, "ORD003", ordersForSort[0].OrderID, "最新订单应在首位")
		assert.Equal(t, "ORD001", ordersForSort[2].OrderID, "最早订单应在末位")
	})

	t.Run("按金额排序", func(t *testing.T) {
		// 复制一份避免影响其他测试
		ordersCopy := make(UserOrderList, len(orders))
		copy(ordersCopy, orders)

		ordersCopy.SortByAmount()
		assert.Equal(t, "ORD001", ordersCopy[0].OrderID, "金额最低的订单应在首位")
		assert.Equal(t, "ORD004", ordersCopy[3].OrderID, "金额最高的订单应在末位")
	})

	t.Run("计算总金额", func(t *testing.T) {
		total := orders.SumAmount()
		expected := 199.00 + 599.50 + 299.99 + 899.00
		assert.InDelta(t, expected, total, 0.01, "总金额计算应正确")
	})
}

// 测试CartItem类型
func TestCartItem(t *testing.T) {
	t.Run("创建和基本属性", func(t *testing.T) {
		ci := NewCartItem("P001", 2, 199.99)
		assert.Equal(t, "P001", ci.ProductID, "ProductID应匹配")
		assert.Equal(t, 2, ci.Quantity, "Quantity应匹配")
		assert.Equal(t, 199.99, ci.Price, "Price应匹配")
		assert.True(t, ci.Selected, "默认应被选中")
	})

	t.Run("字符串表示", func(t *testing.T) {
		ci := NewCartItem("P001", 2, 199.99)
		expected := "商品 P001: 2件, 单价 $199.99"
		assert.Equal(t, expected, ci.String(), "String()应返回格式化的字符串")
	})

	t.Run("计算总价", func(t *testing.T) {
		ci := NewCartItem("P001", 2, 199.99)
		expected := 2 * 199.99
		assert.InDelta(t, expected, ci.TotalPrice(), 0.01, "总价应为数量*单价")
	})

	t.Run("转换为Pair", func(t *testing.T) {
		ci := NewCartItem("P001", 2, 199.99)
		pair := ci.AsPair()
		assert.Equal(t, "P001", pair.Key, "Key应匹配ProductID")
		assert.Equal(t, 2, pair.Value, "Value应匹配Quantity")
	})
}

// 测试CartItemList类型
func TestCartItemList(t *testing.T) {
	items := CartItemList{
		NewCartItem("P001", 2, 199.99),
		NewCartItem("P002", 1, 299.50),
		NewCartItem("P003", 3, 99.99),
	}

	// 设置选中状态
	items[1].Selected = false

	t.Run("计算总数量", func(t *testing.T) {
		total := items.TotalQuantity()
		expected := 2 + 1 + 3
		assert.Equal(t, expected, total, "总数量计算应正确")
	})

	t.Run("计算总金额", func(t *testing.T) {
		total := items.TotalAmount()
		expected := (2 * 199.99) + (1 * 299.50) + (3 * 99.99)
		assert.InDelta(t, expected, total, 0.01, "总金额计算应正确")
	})

	t.Run("过滤已选商品", func(t *testing.T) {
		selected := items.FilterSelected()
		assert.Len(t, selected, 2, "应有2个选中的商品")
		assert.Equal(t, "P001", selected[0].ProductID, "选中商品应包含P001")
		assert.Equal(t, "P003", selected[1].ProductID, "选中商品应包含P003")
	})

	t.Run("更新商品数量", func(t *testing.T) {
		// 复制一份避免影响其他测试
		itemsCopy := make(CartItemList, len(items))
		copy(itemsCopy, items)

		// 更新已有商品
		success := itemsCopy.UpdateQuantity("P001", 5)
		assert.True(t, success, "更新已有商品应成功")
		assert.Equal(t, 5, itemsCopy[0].Quantity, "数量应被更新为5")

		// 更新不存在的商品
		success = itemsCopy.UpdateQuantity("P999", 1)
		assert.False(t, success, "更新不存在的商品应失败")
	})
}

// 测试TimeValuePair类型
func TestTimeValuePair(t *testing.T) {
	now := time.Now()

	t.Run("创建和基本属性", func(t *testing.T) {
		tvp := NewTimeValuePair(now, 123.45)
		assert.Equal(t, now, tvp.Time, "Time应匹配")
		assert.Equal(t, 123.45, tvp.Value, "Value应匹配")
	})

	t.Run("字符串表示", func(t *testing.T) {
		tvp := NewTimeValuePair(now, 123.45)
		expected := now.Format("2006-01-02 15:04:05") + ": 123.45"
		assert.Equal(t, expected, tvp.String(), "String()应返回格式化的字符串")
	})

	t.Run("转换为Pair", func(t *testing.T) {
		tvp := NewTimeValuePair(now, 123.45)
		pair := tvp.AsPair()
		assert.Equal(t, now, pair.Key, "Key应匹配Time")
		assert.Equal(t, 123.45, pair.Value, "Value应匹配Value")
	})
}

// 测试TimeValueList类型
func TestTimeValueList(t *testing.T) {
	now := time.Now()
	values := TimeValueList{
		NewTimeValuePair(now.Add(-2*time.Hour), 100),
		NewTimeValuePair(now.Add(-1*time.Hour), 200),
		NewTimeValuePair(now, 300),
	}

	t.Run("按时间排序", func(t *testing.T) {
		// 复制一份避免影响其他测试
		valuesCopy := make(TimeValueList, len(values))
		copy(valuesCopy, values)

		valuesCopy.SortByTime()
		assert.InDelta(t, 300, valuesCopy[0].Value, 0.01, "最新时间点的值应在首位")
		assert.InDelta(t, 100, valuesCopy[2].Value, 0.01, "最早时间点的值应在末位")
	})

	t.Run("求和", func(t *testing.T) {
		sum := values.SumValues()
		expected := 100 + 200 + 300
		assert.InDelta(t, expected, sum, 0.01, "总和计算应正确")
	})

	t.Run("求平均值", func(t *testing.T) {
		avg := values.AverageValue()
		expected := (100 + 200 + 300) / 3.0
		assert.InDelta(t, expected, avg, 0.01, "平均值计算应正确")
	})

	t.Run("按时间范围过滤", func(t *testing.T) {
		filtered := values.FilterByTimeRange(now.Add(-90*time.Minute), now.Add(-30*time.Minute))
		assert.Len(t, filtered, 1, "时间范围内应有1个值")
		assert.InDelta(t, 200, filtered[0].Value, 0.01, "过滤后应只包含中间时间点的值")
	})

	t.Run("按天分组", func(t *testing.T) {
		// 创建跨天的数据
		today := time.Now()
		yesterday := today.AddDate(0, 0, -1)

		crossDayValues := TimeValueList{
			NewTimeValuePair(yesterday, 100),
			NewTimeValuePair(yesterday.Add(1*time.Hour), 200),
			NewTimeValuePair(today, 300),
			NewTimeValuePair(today.Add(1*time.Hour), 400),
		}

		groups := crossDayValues.GroupByDay()
		assert.Len(t, groups, 2, "应分为2组(今天和昨天)")

		// 检查每组数据
		todayKey := today.Format("2006-01-02")
		yesterdayKey := yesterday.Format("2006-01-02")

		assert.Len(t, groups[todayKey], 2, "今天应有2个数据点")
		assert.Len(t, groups[yesterdayKey], 2, "昨天应有2个数据点")
	})

	t.Run("每日汇总", func(t *testing.T) {
		// 创建跨天的数据
		today := time.Now()
		yesterday := today.AddDate(0, 0, -1)

		crossDayValues := TimeValueList{
			NewTimeValuePair(yesterday, 100),
			NewTimeValuePair(yesterday.Add(1*time.Hour), 200),
			NewTimeValuePair(today, 300),
			NewTimeValuePair(today.Add(1*time.Hour), 400),
		}

		dailyTotals := crossDayValues.DailyTotal()
		assert.Len(t, dailyTotals, 2, "应有2天的汇总")

		// 检查汇总结果
		todayKey := today.Format("2006-01-02")
		yesterdayKey := yesterday.Format("2006-01-02")

		// 由于map迭代顺序不确定，需要查找指定key的结果
		var todayTotal, yesterdayTotal float64
		for _, p := range dailyTotals {
			if p.Key == todayKey {
				todayTotal = p.Value
			} else if p.Key == yesterdayKey {
				yesterdayTotal = p.Value
			}
		}

		assert.InDelta(t, 700, todayTotal, 0.01, "今天的总和应为300+400=700")
		assert.InDelta(t, 300, yesterdayTotal, 0.01, "昨天的总和应为100+200=300")
	})

	// 测试大数据量的性能表现
	t.Run("大数据量每日汇总", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过大数据量测试")
		}

		// 创建1000个跨越30天的数据点
		baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local) // 使用固定日期避免跨月问题
		largeDataset := make(TimeValueList, 1000)
		for i := 0; i < 1000; i++ {
			// 在过去30天内的固定天数
			dayOffset := i % 30 // 0-29之间的值，共30天
			hourOffset := i % 24
			timePoint := baseTime.AddDate(0, 0, -dayOffset).Add(time.Duration(-hourOffset) * time.Hour)
			largeDataset[i] = NewTimeValuePair(timePoint, float64(i%100))
		}

		// 性能测试 - 开始计时
		start := time.Now()
		dailyTotals := largeDataset.DailyTotal()
		duration := time.Since(start)

		// 验证结果
		assert.LessOrEqual(t, duration.Milliseconds(), int64(200), "处理1000个数据点应在200ms内完成")
		// 不检查确切的天数，因为日期计算可能因时区等因素而有所不同
		assert.GreaterOrEqual(t, len(dailyTotals), 29, "应该至少有29天的汇总数据")
		assert.LessOrEqual(t, len(dailyTotals), 31, "应该最多有31天的汇总数据")
	})
}

// 测试ProductCategory类型
func TestProductCategory(t *testing.T) {
	t.Run("创建和基本属性", func(t *testing.T) {
		pc := NewProductCategory("P001", "电子产品", "手机")
		assert.Equal(t, "P001", pc.ProductID, "ProductID应匹配")
		assert.Equal(t, "电子产品", pc.Category, "Category应匹配")
		assert.Equal(t, "手机", pc.SubCategory, "SubCategory应匹配")
	})

	t.Run("字符串表示", func(t *testing.T) {
		// 测试有子分类
		pc1 := NewProductCategory("P001", "电子产品", "手机")
		expected1 := "商品 P001: 电子产品 > 手机"
		assert.Equal(t, expected1, pc1.String(), "带子分类的字符串格式应正确")

		// 测试无子分类
		pc2 := NewProductCategory("P002", "食品", "")
		expected2 := "商品 P002: 食品"
		assert.Equal(t, expected2, pc2.String(), "无子分类的字符串格式应正确")
	})

	t.Run("转换为Pair和Triple", func(t *testing.T) {
		pc := NewProductCategory("P001", "电子产品", "手机")

		// 测试转换为Pair
		pair := pc.AsPair()
		assert.Equal(t, "P001", pair.Key, "Pair.Key应匹配ProductID")
		assert.Equal(t, "电子产品", pair.Value, "Pair.Value应匹配Category")

		// 测试转换为Triple
		triple := pc.AsTriple()
		assert.Equal(t, "P001", triple.First, "Triple.First应匹配ProductID")
		assert.Equal(t, "电子产品", triple.Second, "Triple.Second应匹配Category")
		assert.Equal(t, "手机", triple.Third, "Triple.Third应匹配SubCategory")
	})
}

// 测试ProductCategoryList类型
func TestProductCategoryList(t *testing.T) {
	categories := ProductCategoryList{
		NewProductCategory("P001", "电子产品", "手机"),
		NewProductCategory("P002", "电子产品", "电脑"),
		NewProductCategory("P003", "服装", "上衣"),
		NewProductCategory("P004", "服装", "裤子"),
		NewProductCategory("P005", "电子产品", "相机"),
	}

	t.Run("按分类过滤", func(t *testing.T) {
		filtered := categories.FilterByCategory("电子产品")
		assert.Len(t, filtered, 3, "电子产品分类应有3个商品")

		// 检查过滤结果包含预期的商品
		productIDs := []string{filtered[0].ProductID, filtered[1].ProductID, filtered[2].ProductID}
		assert.Contains(t, productIDs, "P001", "应包含P001")
		assert.Contains(t, productIDs, "P002", "应包含P002")
		assert.Contains(t, productIDs, "P005", "应包含P005")
	})

	t.Run("按子分类过滤", func(t *testing.T) {
		filtered := categories.FilterBySubCategory("手机")
		assert.Len(t, filtered, 1, "手机子分类应有1个商品")
		assert.Equal(t, "P001", filtered[0].ProductID, "应为P001")
	})

	t.Run("分类计数", func(t *testing.T) {
		counts := categories.CountByCategory()
		assert.Equal(t, 3, counts["电子产品"], "电子产品应有3个")
		assert.Equal(t, 2, counts["服装"], "服装应有2个")
	})
}

// 测试ProductRating类型
func TestProductRating(t *testing.T) {
	t.Run("创建和基本属性", func(t *testing.T) {
		pr := NewProductRating("P001", "U001", 4.5, "很好用")
		assert.Equal(t, "P001", pr.ProductID, "ProductID应匹配")
		assert.Equal(t, "U001", pr.UserID, "UserID应匹配")
		assert.Equal(t, 4.5, pr.Rating, "Rating应匹配")
		assert.Equal(t, "很好用", pr.Comment, "Comment应匹配")
		assert.False(t, pr.Time.IsZero(), "Time应被设置")
	})

	t.Run("字符串表示", func(t *testing.T) {
		pr := NewProductRating("P001", "U001", 4.5, "很好用")
		expected := "商品 P001: 4.5分 (U001)"
		assert.Equal(t, expected, pr.String(), "String()应返回格式化的字符串")
	})

	t.Run("转换为Pair", func(t *testing.T) {
		pr := NewProductRating("P001", "U001", 4.5, "很好用")
		pair := pr.AsPair()
		assert.Equal(t, "P001", pair.Key, "Key应匹配ProductID")
		assert.Equal(t, 4.5, pair.Value, "Value应匹配Rating")
	})
}

// 测试ProductRatingList类型
func TestProductRatingList(t *testing.T) {
	ratings := ProductRatingList{
		NewProductRating("P001", "U001", 5.0, "非常好"),
		NewProductRating("P001", "U002", 4.0, "还不错"),
		NewProductRating("P002", "U001", 3.0, "一般"),
		NewProductRating("P002", "U003", 2.0, "不太好"),
		NewProductRating("P003", "U002", 1.0, "很差"),
	}

	t.Run("计算平均评分", func(t *testing.T) {
		// 测试全部评分
		avgAll := ratings.AverageRating()
		expectedAll := (5.0 + 4.0 + 3.0 + 2.0 + 1.0) / 5.0
		assert.InDelta(t, expectedAll, avgAll, 0.01, "所有评分的平均值计算应正确")

		// 测试空列表
		emptyList := ProductRatingList{}
		assert.Equal(t, 0.0, emptyList.AverageRating(), "空列表平均值应为0")
	})

	t.Run("按最低评分过滤", func(t *testing.T) {
		filtered := ratings.FilterByMinRating(4.0)
		assert.Len(t, filtered, 2, "评分>=4.0的应有2个")

		// 检查过滤结果
		ratingSum := 0.0
		for _, r := range filtered {
			ratingSum += r.Rating
			assert.GreaterOrEqual(t, r.Rating, 4.0, "评分应>=4.0")
		}
		assert.InDelta(t, 9.0, ratingSum, 0.01, "总评分应为9.0")
	})

	t.Run("按商品分组", func(t *testing.T) {
		groups := ratings.GroupByProduct()

		assert.Len(t, groups, 3, "应有3个商品组")
		assert.Len(t, groups["P001"], 2, "P001应有2个评分")
		assert.Len(t, groups["P002"], 2, "P002应有2个评分")
		assert.Len(t, groups["P003"], 1, "P003应有1个评分")

		// 检查分组内容
		assert.InDelta(t, 4.5, groups["P001"].AverageRating(), 0.01, "P001平均分应为4.5")
		assert.InDelta(t, 2.5, groups["P002"].AverageRating(), 0.01, "P002平均分应为2.5")
	})
}
