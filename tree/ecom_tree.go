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

// Package tree 提供了基于树结构的电商平台组件
package tree

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ----- 电商系统专用错误 -----
var (
	// 商品和SKU相关错误
	ErrProductNotFound = errors.New("ggu: 商品不存在")
	ErrSkuNotFound     = errors.New("ggu: SKU不存在")
	ErrInvalidPrice    = errors.New("ggu: 无效的价格")

	// 库存相关错误
	ErrInvalidQuantity = errors.New("ggu: 无效的数量")
	ErrStockShortage   = errors.New("ggu: 库存不足")
	ErrOutOfStock      = errors.New("ggu: 商品库存不足")

	// 其他错误
	ErrExpired      = errors.New("ggu: 数据已过期")
	ErrNoPermission = errors.New("ggu: 无操作权限")
)

// ----- 商品相关基础类型 -----

// Product 商品信息
type Product struct {
	ID          string            // 商品ID
	Name        string            // 商品名称
	Price       float64           // 价格
	Stock       int               // 库存
	Categories  []string          // 分类路径
	Attributes  map[string]string // 属性
	CreatedAt   time.Time         // 创建时间
	UpdatedAt   time.Time         // 更新时间
	Description string            // 描述
	Status      string            // 状态：active, inactive
}

// ProductSku 商品SKU信息
type ProductSku struct {
	ID              string            // SKU ID
	ProductID       string            // 商品ID
	Price           float64           // 价格
	OriginalPrice   float64           // 原价
	Attributes      map[string]string // 属性组合(如颜色、尺寸等)
	Stock           int               // 库存
	SalesCount      int               // 销量
	Status          string            // 状态(active/inactive)
	CreatedAt       time.Time         // 创建时间
	UpdatedAt       time.Time         // 更新时间
	WeightInGrams   int               // 重量(克)
	VolumeInCubicCm int               // 体积(立方厘米)
}

// InventoryItem 库存项
type InventoryItem struct {
	Sku           string    // 商品SKU
	Stock         int       // 当前库存
	Reserved      int       // 预留库存(已下单未付款)
	SafetyStock   int       // 安全库存
	UpdatedAt     time.Time // 最后更新时间
	WarehouseCode string    // 仓库编码
}

// ----- 其他基础类型 -----

// SearchResult 搜索结果
type SearchResult struct {
	Term       string   // 搜索词
	ProductIDs []string // 相关商品ID
	Score      int      // 相关性分数
}

// TermFreq 词频对
type TermFreq struct {
	Term string // 搜索词
	Freq int    // 频率
}

// ----- 商品目录结构 -----

// CategoryNode 分类树节点
type CategoryNode struct {
	Name          string                   // 分类名称
	Products      map[string]*Product      // 该分类下的商品
	Children      map[string]*CategoryNode // 子分类
	ProductsCount int                      // 该分类下的商品数量(包括子分类)
	Depth         int                      // 分类深度，根节点为0
}

// ProductCatalog 商品目录树
// 使用分类树结构来组织和管理商品
type ProductCatalog struct {
	root  *CategoryNode       // 根分类节点
	mu    sync.RWMutex        // 锁
	count int                 // 商品总数
	index map[string]*Product // 商品ID索引，用于快速查找
}

// NewProductCatalog 创建新的商品目录
func NewProductCatalog() *ProductCatalog {
	return &ProductCatalog{
		root: &CategoryNode{
			Name:     "root",
			Products: make(map[string]*Product),
			Children: make(map[string]*CategoryNode),
			Depth:    0,
		},
		index: make(map[string]*Product),
	}
}

// AddProduct 添加商品到目录
func (pc *ProductCatalog) AddProduct(product Product) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if len(product.Categories) == 0 {
		return errors.New("ggu: 商品必须至少有一个分类")
	}

	// 检查商品是否已存在
	if _, exists := pc.index[product.ID]; exists {
		// 更新现有商品
		return pc.updateProductLocked(product)
	}

	// 添加新商品
	return pc.addProductLocked(product)
}

// 在已持有锁的情况下添加新商品
func (pc *ProductCatalog) addProductLocked(product Product) error {
	// 添加商品到每个分类路径
	for _, path := range product.Categories {
		// 解析分类路径
		categories := strings.Split(path, "/")
		node := pc.root

		// 创建或遍历分类路径
		for depth, category := range categories {
			category = strings.TrimSpace(category)
			if category == "" {
				continue
			}

			// 如果分类不存在，创建它
			if _, exists := node.Children[category]; !exists {
				node.Children[category] = &CategoryNode{
					Name:     category,
					Products: make(map[string]*Product),
					Children: make(map[string]*CategoryNode),
					Depth:    depth + 1,
				}
			}

			// 移动到下一级分类
			node = node.Children[category]
		}

		// 将商品添加到最终分类节点
		productCopy := product // 创建副本
		node.Products[product.ID] = &productCopy
		node.ProductsCount++

		// 更新父分类的商品计数
		pc.updateParentCounts(categories, 1)
	}

	// 添加到索引
	productCopy := product // 创建副本
	pc.index[product.ID] = &productCopy
	pc.count++

	return nil
}

// 在已持有锁的情况下更新商品
func (pc *ProductCatalog) updateProductLocked(product Product) error {
	// 获取现有商品
	existingProduct := pc.index[product.ID]
	oldCategories := existingProduct.Categories

	// 如果分类发生变化，需要重新组织商品
	if !categoriesEqual(oldCategories, product.Categories) {
		// 从旧分类中移除
		for _, path := range oldCategories {
			categories := strings.Split(path, "/")
			node := pc.navigateToCategory(categories)
			if node != nil {
				delete(node.Products, product.ID)
				node.ProductsCount--
				pc.updateParentCounts(categories, -1)
			}
		}

		// 添加到新分类
		for _, path := range product.Categories {
			categories := strings.Split(path, "/")
			node := pc.navigateToCategory(categories)
			if node == nil {
				// 创建不存在的分类路径
				node = pc.createCategoryPath(categories)
			}
			productCopy := product
			node.Products[product.ID] = &productCopy
			node.ProductsCount++
			pc.updateParentCounts(categories, 1)
		}
	} else {
		// 分类没变，只更新商品信息
		for _, path := range product.Categories {
			categories := strings.Split(path, "/")
			node := pc.navigateToCategory(categories)
			if node != nil {
				productCopy := product
				node.Products[product.ID] = &productCopy
			}
		}
	}

	// 更新索引
	productCopy := product
	pc.index[product.ID] = &productCopy

	return nil
}

// RemoveProduct 从目录中移除商品
func (pc *ProductCatalog) RemoveProduct(productID string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	product, exists := pc.index[productID]
	if !exists {
		return ErrProductNotFound
	}

	// 从所有分类中移除
	for _, path := range product.Categories {
		categories := strings.Split(path, "/")
		node := pc.navigateToCategory(categories)
		if node != nil {
			delete(node.Products, productID)
			node.ProductsCount--
			pc.updateParentCounts(categories, -1)
		}
	}

	// 从索引中移除
	delete(pc.index, productID)
	pc.count--

	return nil
}

// GetProduct 获取商品信息
func (pc *ProductCatalog) GetProduct(productID string) (Product, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	product, exists := pc.index[productID]
	if !exists {
		return Product{}, ErrProductNotFound
	}

	return *product, nil
}

// UpdateStock 更新商品库存
func (pc *ProductCatalog) UpdateStock(productID string, delta int) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	product, exists := pc.index[productID]
	if !exists {
		return ErrProductNotFound
	}

	// 检查库存是否足够
	if delta < 0 && product.Stock+delta < 0 {
		return ErrOutOfStock
	}

	// 更新库存
	product.Stock += delta
	product.UpdatedAt = time.Now()

	// 更新所有分类中的商品信息
	for _, path := range product.Categories {
		categories := strings.Split(path, "/")
		node := pc.navigateToCategory(categories)
		if node != nil && node.Products[productID] != nil {
			node.Products[productID].Stock = product.Stock
			node.Products[productID].UpdatedAt = product.UpdatedAt
		}
	}

	return nil
}

// GetProductsByCategory 获取指定分类下的所有商品
func (pc *ProductCatalog) GetProductsByCategory(categoryPath string) []Product {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	categories := strings.Split(categoryPath, "/")
	node := pc.navigateToCategory(categories)

	if node == nil {
		return []Product{}
	}

	// 收集当前分类下的所有商品
	result := make([]Product, 0, len(node.Products))
	for _, product := range node.Products {
		if product.Status == "active" {
			result = append(result, *product)
		}
	}

	return result
}

// GetAllProductsByCategory 获取指定分类及其所有子分类下的商品
func (pc *ProductCatalog) GetAllProductsByCategory(categoryPath string) []Product {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	categories := strings.Split(categoryPath, "/")
	node := pc.navigateToCategory(categories)

	if node == nil {
		return []Product{}
	}

	// 收集当前分类及所有子分类下的商品
	result := make([]Product, 0)
	pc.collectProducts(node, &result)

	return result
}

// 收集节点及其所有子节点中的商品
func (pc *ProductCatalog) collectProducts(node *CategoryNode, result *[]Product) {
	// 添加当前节点的商品
	for _, product := range node.Products {
		if product.Status == "active" {
			*result = append(*result, *product)
		}
	}

	// 递归添加所有子节点的商品
	for _, child := range node.Children {
		pc.collectProducts(child, result)
	}
}

// GetCategories 获取所有分类
func (pc *ProductCatalog) GetCategories() []string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	result := make([]string, 0)
	pc.collectCategories(pc.root, "", &result)

	return result
}

// 收集所有分类路径
func (pc *ProductCatalog) collectCategories(node *CategoryNode, path string, result *[]string) {
	if node != pc.root {
		if path == "" {
			path = node.Name
		} else {
			path = path + "/" + node.Name
		}
		*result = append(*result, path)
	}

	for _, child := range node.Children {
		pc.collectCategories(child, path, result)
	}
}

// GetCategoryInfo 获取分类信息
func (pc *ProductCatalog) GetCategoryInfo(categoryPath string) (string, int, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	categories := strings.Split(categoryPath, "/")
	node := pc.navigateToCategory(categories)

	if node == nil {
		return "", 0, errors.New("ggu: 分类不存在")
	}

	return node.Name, node.ProductsCount, nil
}

// 导航到指定分类
func (pc *ProductCatalog) navigateToCategory(categories []string) *CategoryNode {
	node := pc.root

	for _, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" {
			continue
		}

		child, exists := node.Children[category]
		if !exists {
			return nil
		}
		node = child
	}

	return node
}

// 创建分类路径
func (pc *ProductCatalog) createCategoryPath(categories []string) *CategoryNode {
	node := pc.root

	for depth, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" {
			continue
		}

		if _, exists := node.Children[category]; !exists {
			node.Children[category] = &CategoryNode{
				Name:     category,
				Products: make(map[string]*Product),
				Children: make(map[string]*CategoryNode),
				Depth:    depth + 1,
			}
		}

		node = node.Children[category]
	}

	return node
}

// 更新父分类的商品计数
func (pc *ProductCatalog) updateParentCounts(categories []string, delta int) {
	if len(categories) == 0 {
		return
	}

	// 从根节点开始更新
	node := pc.root
	for i := 0; i < len(categories)-1; i++ {
		category := strings.TrimSpace(categories[i])
		if category == "" {
			continue
		}

		child, exists := node.Children[category]
		if !exists {
			break
		}

		child.ProductsCount += delta
		node = child
	}
}

// Size 返回商品总数
func (pc *ProductCatalog) Size() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.count
}

// ----- 库存管理 -----

// InventoryManager 库存管理器
// 使用AVL树实现高效的库存查询和更新
type InventoryManager struct {
	inventory     *AVLTree[string, *InventoryItem]
	mu            sync.RWMutex
	eventHandlers []func(item *InventoryItem, action string, quantity int)
}

// NewInventoryManager 创建新的库存管理器
func NewInventoryManager() *InventoryManager {
	tree, _ := NewAVLTree[string, *InventoryItem](StringComparator)
	return &InventoryManager{
		inventory:     tree,
		eventHandlers: make([]func(item *InventoryItem, action string, quantity int), 0),
	}
}

// AddSku 添加新的SKU到库存
func (im *InventoryManager) AddSku(sku string, initialStock int, safetyStock int, warehouseCode string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if initialStock < 0 {
		return ErrInvalidQuantity
	}

	// 检查SKU是否已存在
	existing, err := im.inventory.Get(sku)
	if err == nil {
		// 更新现有SKU的库存
		existing.Stock = initialStock
		existing.SafetyStock = safetyStock
		existing.UpdatedAt = time.Now()
		existing.WarehouseCode = warehouseCode

		// 触发事件
		im.triggerEvent(existing, "update", initialStock)
		return nil
	}

	// 创建新的库存项
	item := &InventoryItem{
		Sku:           sku,
		Stock:         initialStock,
		Reserved:      0,
		SafetyStock:   safetyStock,
		UpdatedAt:     time.Now(),
		WarehouseCode: warehouseCode,
	}

	im.inventory.Put(sku, item)

	// 触发事件
	im.triggerEvent(item, "add", initialStock)
	return nil
}

// GetStock 获取商品当前可用库存
func (im *InventoryManager) GetStock(sku string) (int, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	item, err := im.inventory.Get(sku)
	if err != nil {
		return 0, ErrSkuNotFound
	}

	// 可用库存 = 总库存 - 预留库存
	return item.Stock - item.Reserved, nil
}

// Reserve 预留库存(下单未付款)
func (im *InventoryManager) Reserve(sku string, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	item, err := im.inventory.Get(sku)
	if err != nil {
		return ErrSkuNotFound
	}

	// 检查是否有足够库存
	if item.Stock-item.Reserved < quantity {
		return ErrStockShortage
	}

	// 增加预留数量
	item.Reserved += quantity
	item.UpdatedAt = time.Now()

	// 触发事件
	im.triggerEvent(item, "reserve", quantity)
	return nil
}

// Commit 确认库存扣减(订单付款完成)
func (im *InventoryManager) Commit(sku string, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	item, err := im.inventory.Get(sku)
	if err != nil {
		return ErrSkuNotFound
	}

	// 检查预留数量
	if item.Reserved < quantity {
		return fmt.Errorf("预留库存(%d)小于提交数量(%d)", item.Reserved, quantity)
	}

	// 减少总库存和预留库存
	item.Stock -= quantity
	item.Reserved -= quantity
	item.UpdatedAt = time.Now()

	// 触发事件
	im.triggerEvent(item, "commit", quantity)

	// 检查是否低于安全库存
	if item.Stock < item.SafetyStock {
		im.triggerEvent(item, "low_stock", item.Stock)
	}

	return nil
}

// Release 释放预留库存(订单取消)
func (im *InventoryManager) Release(sku string, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	item, err := im.inventory.Get(sku)
	if err != nil {
		return ErrSkuNotFound
	}

	// 检查预留数量
	if item.Reserved < quantity {
		return fmt.Errorf("预留库存(%d)小于释放数量(%d)", item.Reserved, quantity)
	}

	// 减少预留库存
	item.Reserved -= quantity
	item.UpdatedAt = time.Now()

	// 触发事件
	im.triggerEvent(item, "release", quantity)
	return nil
}

// Restock 补充库存
func (im *InventoryManager) Restock(sku string, quantity int) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	item, err := im.inventory.Get(sku)
	if err != nil {
		return ErrSkuNotFound
	}

	// 增加总库存
	item.Stock += quantity
	item.UpdatedAt = time.Now()

	// 触发事件
	im.triggerEvent(item, "restock", quantity)
	return nil
}

// BatchUpdateStock 批量更新库存
func (im *InventoryManager) BatchUpdateStock(updates map[string]int) map[string]error {
	result := make(map[string]error)

	for sku, quantity := range updates {
		if quantity >= 0 {
			err := im.Restock(sku, quantity)
			result[sku] = err
		} else {
			// 负数表示减少库存
			err := im.Commit(sku, -quantity)
			result[sku] = err
		}
	}

	return result
}

// GetLowStockItems 获取低于安全库存的商品
func (im *InventoryManager) GetLowStockItems() []*InventoryItem {
	im.mu.RLock()
	defer im.mu.RUnlock()

	lowStockItems := make([]*InventoryItem, 0)

	im.inventory.ForEach(func(sku string, item *InventoryItem) bool {
		if item.Stock < item.SafetyStock {
			lowStockItems = append(lowStockItems, item)
		}
		return true
	})

	return lowStockItems
}

// GetItemsByWarehouse 获取指定仓库的所有商品库存
func (im *InventoryManager) GetItemsByWarehouse(warehouseCode string) []*InventoryItem {
	im.mu.RLock()
	defer im.mu.RUnlock()

	warehouseItems := make([]*InventoryItem, 0)

	im.inventory.ForEach(func(sku string, item *InventoryItem) bool {
		if item.WarehouseCode == warehouseCode {
			warehouseItems = append(warehouseItems, item)
		}
		return true
	})

	return warehouseItems
}

// AddEventHandler 添加库存变动事件处理器
func (im *InventoryManager) AddEventHandler(handler func(item *InventoryItem, action string, quantity int)) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.eventHandlers = append(im.eventHandlers, handler)
}

// 触发事件处理
func (im *InventoryManager) triggerEvent(item *InventoryItem, action string, quantity int) {
	for _, handler := range im.eventHandlers {
		go handler(item, action, quantity)
	}
}

// ----- 价格管理 -----

// PriceManager 价格管理器
// 使用AVL树实现高效的价格查询和区间搜索
type PriceManager struct {
	products   *AVLTree[string, *ProductSku] // 按SKU ID查询
	priceIndex *AVLTree[float64, []string]   // 按价格索引SKU
	mu         sync.RWMutex
}

// NewPriceManager 创建新的价格管理器
func NewPriceManager() *PriceManager {
	productsTree, _ := NewAVLTree[string, *ProductSku](StringComparator)
	priceTree, _ := NewAVLTree[float64, []string](Float64Comparator)

	return &PriceManager{
		products:   productsTree,
		priceIndex: priceTree,
	}
}

// AddProduct 添加商品SKU及价格
func (pm *PriceManager) AddProduct(sku *ProductSku) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if sku.Price < 0 {
		return ErrInvalidPrice
	}

	// 添加到产品树
	pm.products.Put(sku.ID, sku)

	// 添加到价格索引
	pm.addToPriceIndex(sku.ID, sku.Price)

	return nil
}

// UpdatePrice 更新商品价格
func (pm *PriceManager) UpdatePrice(skuID string, newPrice float64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if newPrice < 0 {
		return ErrInvalidPrice
	}

	sku, err := pm.products.Get(skuID)
	if err != nil {
		return ErrSkuNotFound
	}

	oldPrice := sku.Price

	// 从旧价格索引中移除
	pm.removeFromPriceIndex(skuID, oldPrice)

	// 更新价格
	sku.OriginalPrice = oldPrice
	sku.Price = newPrice
	sku.UpdatedAt = time.Now()

	// 添加到新价格索引
	pm.addToPriceIndex(skuID, newPrice)

	return nil
}

// GetProduct 获取商品信息
func (pm *PriceManager) GetProduct(skuID string) (*ProductSku, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	sku, err := pm.products.Get(skuID)
	if err != nil {
		return nil, ErrSkuNotFound
	}

	return sku, nil
}

// GetProductsInPriceRange 获取指定价格范围内的商品
func (pm *PriceManager) GetProductsInPriceRange(minPrice, maxPrice float64) ([]*ProductSku, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if minPrice < 0 || maxPrice < minPrice {
		return nil, ErrInvalidRange
	}

	// 查找价格范围内的SKU ID
	_, priceValues, err := pm.priceIndex.FindRange(minPrice, maxPrice+0.001) // 增加一点点确保包含上限
	if err != nil {
		return nil, err
	}

	// 收集所有结果
	result := make([]*ProductSku, 0)

	for _, skuIDs := range priceValues {
		for _, skuID := range skuIDs {
			sku, err := pm.products.Get(skuID)
			if err == nil && sku.Status == "active" { // 只返回激活状态的商品
				result = append(result, sku)
			}
		}
	}

	return result, nil
}

// GetProductsSortedByPrice 获取按价格排序的商品
func (pm *PriceManager) GetProductsSortedByPrice(ascending bool) []*ProductSku {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 获取所有价格键和对应的SKU ID
	priceKeys, priceValues := pm.priceIndex.KeyValues()

	// 收集所有结果
	result := make([]*ProductSku, 0)

	// 根据排序方向决定遍历顺序
	if ascending {
		for i := 0; i < len(priceKeys); i++ {
			skuIDs := priceValues[i]
			for _, skuID := range skuIDs {
				sku, err := pm.products.Get(skuID)
				if err == nil && sku.Status == "active" {
					result = append(result, sku)
				}
			}
		}
	} else {
		for i := len(priceKeys) - 1; i >= 0; i-- {
			skuIDs := priceValues[i]
			for _, skuID := range skuIDs {
				sku, err := pm.products.Get(skuID)
				if err == nil && sku.Status == "active" {
					result = append(result, sku)
				}
			}
		}
	}

	return result
}

// BatchUpdatePrices 批量更新价格
func (pm *PriceManager) BatchUpdatePrices(updates map[string]float64) map[string]error {
	result := make(map[string]error)

	for skuID, price := range updates {
		err := pm.UpdatePrice(skuID, price)
		result[skuID] = err
	}

	return result
}

// 添加到价格索引
func (pm *PriceManager) addToPriceIndex(skuID string, price float64) {
	skuIDs, err := pm.priceIndex.Get(price)

	if err != nil {
		// 新价格点
		pm.priceIndex.Put(price, []string{skuID})
	} else {
		// 现有价格点，添加SKU ID
		skuIDs = append(skuIDs, skuID)
		pm.priceIndex.Put(price, skuIDs)
	}
}

// 从价格索引中移除
func (pm *PriceManager) removeFromPriceIndex(skuID string, price float64) {
	skuIDs, err := pm.priceIndex.Get(price)
	if err != nil {
		return
	}

	// 过滤掉要移除的SKU ID
	newSkuIDs := make([]string, 0, len(skuIDs))
	for _, id := range skuIDs {
		if id != skuID {
			newSkuIDs = append(newSkuIDs, id)
		}
	}

	if len(newSkuIDs) > 0 {
		// 更新价格点
		pm.priceIndex.Put(price, newSkuIDs)
	} else {
		// 如果该价格点没有商品了，删除该价格点
		pm.priceIndex.Remove(price)
	}
}

// ----- 辅助函数 -----

// 比较两个分类数组是否相等
func categoriesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// 对搜索结果按分数排序
func sortSearchResults(results []SearchResult) {
	// 简单的冒泡排序实现
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// 对词-频率对按频率排序
func sortTermsByFrequency(pairs []TermFreq) {
	// 简单的冒泡排序实现
	for i := 0; i < len(pairs)-1; i++ {
		for j := 0; j < len(pairs)-i-1; j++ {
			if pairs[j].Freq < pairs[j+1].Freq {
				pairs[j], pairs[j+1] = pairs[j+1], pairs[j]
			}
		}
	}
}
