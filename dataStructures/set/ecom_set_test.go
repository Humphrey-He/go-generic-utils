package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 购物车测试
func TestShoppingCart(t *testing.T) {
	t.Run("基本购物车操作", func(t *testing.T) {
		cart := NewShoppingCart("user1", 10)
		assert.NotNil(t, cart, "购物车创建失败")
		assert.Equal(t, 0, cart.ItemCount(), "新购物车应为空")

		// 添加商品
		err := cart.AddItem("product1", 2, map[string]string{"color": "red"})
		assert.NoError(t, err, "添加商品不应返回错误")
		assert.Equal(t, 1, cart.ItemCount(), "购物车应有1种商品")
		assert.Equal(t, 2, cart.TotalQuantity(), "购物车总数量应为2")

		// 检查商品是否存在
		assert.True(t, cart.HasItem("product1"), "购物车应包含product1")
		assert.False(t, cart.HasItem("product2"), "购物车不应包含product2")

		// 获取单个商品
		item, exists := cart.GetItem("product1")
		assert.True(t, exists, "商品应存在")
		assert.Equal(t, "product1", string(item.ProductID), "商品ID应匹配")
		assert.Equal(t, 2, item.Quantity, "商品数量应为2")
		assert.Equal(t, "red", item.Attributes["color"], "商品属性应匹配")

		// 更新商品数量
		err = cart.UpdateQuantity("product1", 5)
		assert.NoError(t, err, "更新数量不应返回错误")
		item, _ = cart.GetItem("product1")
		assert.Equal(t, 5, item.Quantity, "商品数量应被更新为5")

		// 更新不存在的商品
		err = cart.UpdateQuantity("product2", 1)
		assert.Error(t, err, "更新不存在的商品应返回错误")

		// 添加新商品
		err = cart.AddItem("product2", 3, nil)
		assert.NoError(t, err, "添加商品不应返回错误")
		assert.Equal(t, 2, cart.ItemCount(), "购物车应有2种商品")
		assert.Equal(t, 8, cart.TotalQuantity(), "购物车总数量应为8")

		// 获取所有商品
		items := cart.GetItems()
		assert.Len(t, items, 2, "应返回2种商品")

		// 通过设置数量为0来删除商品
		err = cart.UpdateQuantity("product1", 0)
		assert.NoError(t, err, "设置数量为0不应返回错误")
		assert.Equal(t, 1, cart.ItemCount(), "购物车应剩余1种商品")
		assert.False(t, cart.HasItem("product1"), "product1应已被删除")

		// 直接删除商品
		cart.RemoveItem("product2")
		assert.Equal(t, 0, cart.ItemCount(), "购物车应为空")

		// 清空购物车
		cart.AddItem("product3", 1, nil)
		assert.Equal(t, 1, cart.ItemCount(), "购物车应有1种商品")
		cart.Clear()
		assert.Equal(t, 0, cart.ItemCount(), "购物车应为空")
	})

	t.Run("购物车容量限制", func(t *testing.T) {
		cart := NewShoppingCart("user1", 2) // 最多容纳2种商品

		// 添加到达上限
		err := cart.AddItem("product1", 1, nil)
		assert.NoError(t, err, "添加第一个商品不应返回错误")
		err = cart.AddItem("product2", 1, nil)
		assert.NoError(t, err, "添加第二个商品不应返回错误")

		// 尝试添加超出上限
		err = cart.AddItem("product3", 1, nil)
		assert.Error(t, err, "超出购物车容量应返回错误")
		assert.Equal(t, 2, cart.ItemCount(), "购物车应仍有2种商品")

		// 更新已有商品不受限制
		err = cart.AddItem("product1", 3, nil) // 增加已有商品的数量
		assert.NoError(t, err, "更新已有商品不应返回错误")
		item, _ := cart.GetItem("product1")
		assert.Equal(t, 4, item.Quantity, "product1数量应为4")
	})

	t.Run("无效数量", func(t *testing.T) {
		cart := NewShoppingCart("user1", 10)

		// 添加数量为0的商品
		err := cart.AddItem("product1", 0, nil)
		assert.Error(t, err, "添加数量为0的商品应返回错误")

		// 添加数量为负的商品
		err = cart.AddItem("product1", -1, nil)
		assert.Error(t, err, "添加数量为负的商品应返回错误")

		// 更新为负数量
		cart.AddItem("product1", 1, nil)
		err = cart.UpdateQuantity("product1", -1)
		assert.Error(t, err, "设置负数量应返回错误")
	})
}

// 标签集测试
func TestTagSet(t *testing.T) {
	t.Run("基本标签操作", func(t *testing.T) {
		tagSet := NewTagSet(10)
		assert.NotNil(t, tagSet, "标签集创建失败")
		assert.Equal(t, 0, tagSet.Count(), "新标签集应为空")

		// 添加标签
		err := tagSet.AddTag("tag1", "标签1")
		assert.NoError(t, err, "添加标签不应返回错误")
		assert.Equal(t, 1, tagSet.Count(), "标签集应有1个标签")

		// 检查标签是否存在
		assert.True(t, tagSet.HasTag("tag1"), "标签集应包含tag1")
		assert.False(t, tagSet.HasTag("tag2"), "标签集不应包含tag2")

		// 获取单个标签
		tag, exists := tagSet.GetTag("tag1")
		assert.True(t, exists, "标签应存在")
		assert.Equal(t, "tag1", tag.ID, "标签ID应匹配")
		assert.Equal(t, "标签1", tag.Name, "标签名称应匹配")

		// 添加更多标签
		tagSet.AddTag("tag2", "标签2")
		tagSet.AddTag("tag3", "标签3")

		// 获取所有标签
		tags := tagSet.GetAllTags()
		assert.Len(t, tags, 3, "应返回3个标签")

		// 删除标签
		tagSet.RemoveTag("tag2")
		assert.Equal(t, 2, tagSet.Count(), "标签集应剩余2个标签")
		assert.False(t, tagSet.HasTag("tag2"), "tag2应已被删除")

		// 清空标签集
		tagSet.Clear()
		assert.Equal(t, 0, tagSet.Count(), "标签集应为空")
	})

	t.Run("标签集容量限制", func(t *testing.T) {
		tagSet := NewTagSet(2) // 最多容纳2个标签

		// 添加到达上限
		err := tagSet.AddTag("tag1", "标签1")
		assert.NoError(t, err, "添加第一个标签不应返回错误")
		err = tagSet.AddTag("tag2", "标签2")
		assert.NoError(t, err, "添加第二个标签不应返回错误")

		// 尝试添加超出上限
		err = tagSet.AddTag("tag3", "标签3")
		assert.Error(t, err, "超出标签集容量应返回错误")
		assert.Equal(t, 2, tagSet.Count(), "标签集应仍有2个标签")

		// 更新已有标签不受限制
		err = tagSet.AddTag("tag1", "标签1更新")
		assert.NoError(t, err, "更新已有标签不应返回错误")
		tag, _ := tagSet.GetTag("tag1")
		assert.Equal(t, "标签1更新", tag.Name, "tag1名称应被更新")
	})
}

// 最近浏览商品测试
func TestRecentlyViewedProducts(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		recent := NewRecentlyViewedProducts("user1", 5)
		assert.NotNil(t, recent, "最近浏览商品集创建失败")

		// 初始状态
		products := recent.GetProducts()
		assert.Empty(t, products, "初始状态应为空")

		// 添加商品
		recent.AddProduct("product1")
		recent.AddProduct("product2")
		recent.AddProduct("product3")

		// 验证顺序 (最新添加的在前面)
		products = recent.GetProducts()
		assert.Equal(t, []ProductID{"product3", "product2", "product1"}, products, "商品顺序应为添加的相反顺序")

		// 重复添加同一商品 (应移到最前面)
		recent.AddProduct("product1")
		products = recent.GetProducts()
		assert.Equal(t, []ProductID{"product1", "product3", "product2"}, products, "product1应移至最前面")

		// 清空
		recent.Clear()
		products = recent.GetProducts()
		assert.Empty(t, products, "清空后应为空")
	})

	t.Run("容量限制", func(t *testing.T) {
		recent := NewRecentlyViewedProducts("user1", 3) // 最多保存3个商品

		// 添加超过容量的商品
		recent.AddProduct("product1")
		recent.AddProduct("product2")
		recent.AddProduct("product3")
		recent.AddProduct("product4") // 这应该导致最早添加的product1被移除

		products := recent.GetProducts()
		assert.Len(t, products, 3, "应只保留3个商品")
		assert.Equal(t, []ProductID{"product4", "product3", "product2"}, products, "应保留最近添加的3个商品")
		assert.NotContains(t, products, ProductID("product1"), "最早添加的product1应被移除")
	})
}

// 心愿单测试
func TestWishList(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		wishlist := NewWishList("user1", 10)
		assert.NotNil(t, wishlist, "心愿单创建失败")
		assert.Equal(t, 0, wishlist.Count(), "新心愿单应为空")

		// 添加商品
		err := wishlist.AddProduct("product1")
		assert.NoError(t, err, "添加商品不应返回错误")
		assert.Equal(t, 1, wishlist.Count(), "心愿单应有1个商品")

		// 检查商品是否存在
		assert.True(t, wishlist.HasProduct("product1"), "心愿单应包含product1")
		assert.False(t, wishlist.HasProduct("product2"), "心愿单不应包含product2")

		// 获取所有商品
		products := wishlist.GetProducts()
		assert.Len(t, products, 1, "应返回1个商品")
		_, exists := products["product1"]
		assert.True(t, exists, "返回的商品中应包含product1")

		// 删除商品
		wishlist.RemoveProduct("product1")
		assert.Equal(t, 0, wishlist.Count(), "心愿单应为空")
		assert.False(t, wishlist.HasProduct("product1"), "product1应已被删除")

		// 清空心愿单
		wishlist.AddProduct("product2")
		assert.Equal(t, 1, wishlist.Count(), "心愿单应有1个商品")
		wishlist.Clear()
		assert.Equal(t, 0, wishlist.Count(), "心愿单应为空")
	})

	t.Run("容量限制", func(t *testing.T) {
		wishlist := NewWishList("user1", 2) // 最多容纳2个商品

		// 添加到达上限
		err := wishlist.AddProduct("product1")
		assert.NoError(t, err, "添加第一个商品不应返回错误")
		err = wishlist.AddProduct("product2")
		assert.NoError(t, err, "添加第二个商品不应返回错误")

		// 尝试添加超出上限
		err = wishlist.AddProduct("product3")
		assert.Error(t, err, "超出心愿单容量应返回错误")
		assert.Equal(t, 2, wishlist.Count(), "心愿单应仍有2个商品")

		// 更新已有商品不受限制
		err = wishlist.AddProduct("product1") // 更新已有商品的时间戳
		assert.NoError(t, err, "更新已有商品不应返回错误")
	})
}

// 商品库存测试
func TestProductInventory(t *testing.T) {
	var changedProductID ProductID
	var oldQty, newQty int

	// 回调函数
	onChangeCallback := func(productID ProductID, oldQuantity, newQuantity int) {
		changedProductID = productID
		oldQty = oldQuantity
		newQty = newQuantity
	}

	t.Run("基本操作", func(t *testing.T) {
		inventory := NewProductInventory(onChangeCallback)
		assert.NotNil(t, inventory, "库存管理创建失败")

		// 设置库存
		inventory.SetStock("product1", 10)
		assert.Equal(t, 10, inventory.GetStock("product1"), "product1库存应为10")
		assert.Equal(t, ProductID("product1"), changedProductID, "回调应提供正确的商品ID")
		assert.Equal(t, 0, oldQty, "原库存应为0")
		assert.Equal(t, 10, newQty, "新库存应为10")

		// 减少库存
		success := inventory.DecreaseStock("product1", 3)
		assert.True(t, success, "减少库存应成功")
		assert.Equal(t, 7, inventory.GetStock("product1"), "减少后product1库存应为7")
		assert.Equal(t, 10, oldQty, "原库存应为10")
		assert.Equal(t, 7, newQty, "新库存应为7")

		// 增加库存
		inventory.IncreaseStock("product1", 5)
		assert.Equal(t, 12, inventory.GetStock("product1"), "增加后product1库存应为12")
		assert.Equal(t, 7, oldQty, "原库存应为7")
		assert.Equal(t, 12, newQty, "新库存应为12")

		// 尝试减少超过现有库存
		success = inventory.DecreaseStock("product1", 20)
		assert.False(t, success, "减少超过库存应失败")
		assert.Equal(t, 12, inventory.GetStock("product1"), "库存应保持不变")
	})

	t.Run("库存状态", func(t *testing.T) {
		inventory := NewProductInventory(nil) // 不需要回调

		inventory.SetStock("product1", 10)
		assert.Equal(t, "in_stock", inventory.GetStockStatus("product1"), "库存充足应返回in_stock")

		inventory.SetLowStockThreshold("product1", 5)
		assert.Equal(t, "in_stock", inventory.GetStockStatus("product1"), "库存大于阈值应返回in_stock")

		inventory.DecreaseStock("product1", 6)
		assert.Equal(t, "low_stock", inventory.GetStockStatus("product1"), "库存低于阈值应返回low_stock")

		inventory.DecreaseStock("product1", 4)
		assert.Equal(t, "out_of_stock", inventory.GetStockStatus("product1"), "库存为0应返回out_of_stock")

		// 低库存商品
		inventory.SetStock("product2", 3)
		inventory.SetLowStockThreshold("product2", 5)
		lowStockProducts := inventory.LowStockProducts()
		assert.Contains(t, lowStockProducts, ProductID("product2"), "低库存商品列表应包含product2")
	})
}

// 分类集合测试
func TestCategorySet(t *testing.T) {
	t.Run("基本操作", func(t *testing.T) {
		categories := NewCategorySet()
		assert.NotNil(t, categories, "分类集合创建失败")

		// 添加分类
		categories.AddCategory("electronics")
		categories.AddCategory("books")
		assert.True(t, categories.HasCategory("electronics"), "分类集合应包含electronics")
		assert.True(t, categories.HasCategory("books"), "分类集合应包含books")

		// 获取所有分类
		allCategories := categories.GetAllCategories()
		assert.Len(t, allCategories, 2, "应返回2个分类")
		assert.Contains(t, allCategories, "electronics", "分类列表应包含electronics")
		assert.Contains(t, allCategories, "books", "分类列表应包含books")

		// 分配分类到商品
		categories.AssignCategoryToProduct("product1", "electronics")
		categories.AssignCategoryToProduct("product1", "books")
		categories.AssignCategoryToProduct("product2", "electronics")

		// 获取商品分类
		productCategories := categories.GetProductCategories("product1")
		assert.Len(t, productCategories, 2, "product1应有2个分类")
		assert.Contains(t, productCategories, "electronics", "product1分类应包含electronics")
		assert.Contains(t, productCategories, "books", "product1分类应包含books")

		// 获取分类下的商品
		electronicsProducts := categories.GetProductsByCategory("electronics")
		assert.Len(t, electronicsProducts, 2, "electronics分类下应有2个商品")
		assert.Contains(t, electronicsProducts, ProductID("product1"), "electronics分类下应包含product1")
		assert.Contains(t, electronicsProducts, ProductID("product2"), "electronics分类下应包含product2")

		// 从商品移除分类
		categories.RemoveCategoryFromProduct("product1", "books")
		productCategories = categories.GetProductCategories("product1")
		assert.Len(t, productCategories, 1, "移除后product1应只有1个分类")
		assert.Contains(t, productCategories, "electronics", "product1仍应包含electronics分类")
		assert.NotContains(t, productCategories, "books", "product1不应再包含books分类")

		// 删除分类
		categories.RemoveCategory("electronics")
		assert.False(t, categories.HasCategory("electronics"), "electronics分类应已被删除")
		// 删除分类后，商品与该分类的关联也应被删除
		productCategories = categories.GetProductCategories("product1")
		assert.Empty(t, productCategories, "删除分类后，product1不应有任何分类")
	})
}

// 商品过滤器测试
func TestProductFilter(t *testing.T) {
	// 准备测试数据
	categories := NewCategorySet()
	categories.AddCategory("electronics")
	categories.AddCategory("books")
	categories.AddCategory("clothing")

	// 分配分类
	categories.AssignCategoryToProduct("product1", "electronics")
	categories.AssignCategoryToProduct("product1", "books")
	categories.AssignCategoryToProduct("product2", "electronics")
	categories.AssignCategoryToProduct("product3", "books")
	categories.AssignCategoryToProduct("product4", "clothing")

	// 准备库存
	inventory := NewProductInventory(nil)
	inventory.SetStock("product1", 10)
	inventory.SetStock("product2", 0)
	inventory.SetStock("product3", 5)
	inventory.SetStock("product4", 3)
	inventory.SetLowStockThreshold("product3", 10) // 设置为低库存

	filter := NewProductFilter(categories)

	t.Run("按分类过滤", func(t *testing.T) {
		allProducts := []ProductID{"product1", "product2", "product3", "product4"}

		// 单分类过滤
		electronicsProducts := filter.FilterByCategory(allProducts, "electronics")
		assert.Len(t, electronicsProducts, 2, "electronics分类下应有2个商品")
		assert.Contains(t, electronicsProducts, ProductID("product1"), "结果应包含product1")
		assert.Contains(t, electronicsProducts, ProductID("product2"), "结果应包含product2")

		// 多分类过滤 (OR 关系)
		booksOrElectronics := filter.FilterByCategories(allProducts, []string{"books", "electronics"}, false)
		assert.Len(t, booksOrElectronics, 3, "books或electronics分类下应有3个商品")

		// 多分类过滤 (AND 关系)
		booksAndElectronics := filter.FilterByCategories(allProducts, []string{"books", "electronics"}, true)
		assert.Len(t, booksAndElectronics, 1, "同时属于books和electronics分类的应只有1个商品")
		assert.Contains(t, booksAndElectronics, ProductID("product1"), "只有product1同时属于两个分类")
	})

	t.Run("按库存状态过滤", func(t *testing.T) {
		allProducts := []ProductID{"product1", "product2", "product3", "product4"}

		// 按库存状态过滤
		inStockProducts := filter.FilterByStock(allProducts, inventory, "in_stock")
		assert.Len(t, inStockProducts, 2, "有库存的商品应有2个")
		assert.Contains(t, inStockProducts, ProductID("product1"), "结果应包含product1")
		assert.Contains(t, inStockProducts, ProductID("product4"), "结果应包含product4")

		// 低库存商品
		lowStockProducts := filter.FilterByStock(allProducts, inventory, "low_stock")
		assert.Len(t, lowStockProducts, 1, "低库存商品应有1个")
		assert.Contains(t, lowStockProducts, ProductID("product3"), "结果应包含product3")

		// 无库存商品
		outOfStockProducts := filter.FilterByStock(allProducts, inventory, "out_of_stock")
		assert.Len(t, outOfStockProducts, 1, "无库存商品应有1个")
		assert.Contains(t, outOfStockProducts, ProductID("product2"), "结果应包含product2")

		// 使用便捷方法过滤有库存商品
		inStockProducts2 := filter.FilterByInStock(allProducts, inventory)
		assert.Len(t, inStockProducts2, 3, "有库存和低库存商品总共应有3个")
		assert.Contains(t, inStockProducts2, ProductID("product1"), "结果应包含product1")
		assert.Contains(t, inStockProducts2, ProductID("product3"), "结果应包含product3")
		assert.Contains(t, inStockProducts2, ProductID("product4"), "结果应包含product4")
	})
}
