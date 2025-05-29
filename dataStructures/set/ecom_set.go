package set

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ProductID 商品ID类型
type ProductID string

// UserID 用户ID类型
type UserID string

// CartItem 购物车项
type CartItem struct {
	ProductID  ProductID         `json:"product_id"`
	Quantity   int               `json:"quantity"`
	AddedAt    time.Time         `json:"added_at"`
	Attributes map[string]string `json:"attributes"` // 商品属性（如尺寸、颜色等）
}

// ShoppingCart 购物车实现
type ShoppingCart struct {
	userID   UserID
	items    map[ProductID]*CartItem
	lock     sync.RWMutex
	maxItems int // 购物车最大商品数
}

// NewShoppingCart 创建新的购物车
func NewShoppingCart(userID UserID, maxItems int) *ShoppingCart {
	if maxItems <= 0 {
		maxItems = 100 // 默认最大100件商品
	}

	return &ShoppingCart{
		userID:   userID,
		items:    make(map[ProductID]*CartItem),
		maxItems: maxItems,
	}
}

// AddItem 添加商品到购物车
func (c *ShoppingCart) AddItem(productID ProductID, quantity int, attrs map[string]string) error {
	if quantity <= 0 {
		return errors.New("商品数量必须大于0")
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// 检查购物车是否已满
	if len(c.items) >= c.maxItems && c.items[productID] == nil {
		return fmt.Errorf("购物车已满，最多支持%d件商品", c.maxItems)
	}

	// 检查商品是否已在购物车中
	if item, exists := c.items[productID]; exists {
		// 更新数量和属性
		item.Quantity += quantity

		// 合并属性
		if attrs != nil {
			if item.Attributes == nil {
				item.Attributes = make(map[string]string)
			}
			for k, v := range attrs {
				item.Attributes[k] = v
			}
		}
	} else {
		// 添加新商品
		c.items[productID] = &CartItem{
			ProductID:  productID,
			Quantity:   quantity,
			AddedAt:    time.Now(),
			Attributes: attrs,
		}
	}

	return nil
}

// UpdateQuantity 更新购物车中商品数量
func (c *ShoppingCart) UpdateQuantity(productID ProductID, quantity int) error {
	if quantity < 0 {
		return errors.New("商品数量不能为负数")
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	item, exists := c.items[productID]
	if !exists {
		return fmt.Errorf("商品 %s 不在购物车中", productID)
	}

	if quantity == 0 {
		// 删除商品
		delete(c.items, productID)
		return nil
	}

	// 更新数量
	item.Quantity = quantity
	return nil
}

// RemoveItem 从购物车中移除商品
func (c *ShoppingCart) RemoveItem(productID ProductID) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.items, productID)
}

// GetItem 获取购物车中的商品
func (c *ShoppingCart) GetItem(productID ProductID) (*CartItem, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, exists := c.items[productID]
	if !exists {
		return nil, false
	}

	// 返回副本，避免外部修改
	copiedItem := *item
	return &copiedItem, true
}

// GetItems 获取购物车中的所有商品
func (c *ShoppingCart) GetItems() []*CartItem {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]*CartItem, 0, len(c.items))
	for _, item := range c.items {
		// 复制一份返回，避免外部修改
		copiedItem := *item
		result = append(result, &copiedItem)
	}

	return result
}

// Clear 清空购物车
func (c *ShoppingCart) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items = make(map[ProductID]*CartItem)
}

// ItemCount 获取购物车中的商品种类数
func (c *ShoppingCart) ItemCount() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return len(c.items)
}

// TotalQuantity 获取购物车中所有商品的总数量
func (c *ShoppingCart) TotalQuantity() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	total := 0
	for _, item := range c.items {
		total += item.Quantity
	}

	return total
}

// HasItem 检查商品是否在购物车中
func (c *ShoppingCart) HasItem(productID ProductID) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, exists := c.items[productID]
	return exists
}

// ProductTag 商品标签结构
type ProductTag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// TagSet 标签集合，用于管理商品或用户标签
type TagSet struct {
	tags    map[string]*ProductTag
	lock    sync.RWMutex
	maxTags int // 最大标签数
}

// NewTagSet 创建标签集合
func NewTagSet(maxTags int) *TagSet {
	if maxTags <= 0 {
		maxTags = 100 // 默认最多100个标签
	}

	return &TagSet{
		tags:    make(map[string]*ProductTag),
		maxTags: maxTags,
	}
}

// AddTag 添加标签
func (s *TagSet) AddTag(id, name string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 检查是否已达到最大标签数
	if len(s.tags) >= s.maxTags && s.tags[id] == nil {
		return fmt.Errorf("已达最大标签数: %d", s.maxTags)
	}

	s.tags[id] = &ProductTag{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}

	return nil
}

// RemoveTag 删除标签
func (s *TagSet) RemoveTag(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.tags, id)
}

// GetTag 获取标签
func (s *TagSet) GetTag(id string) (*ProductTag, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tag, exists := s.tags[id]
	if !exists {
		return nil, false
	}

	// 返回副本
	copiedTag := *tag
	return &copiedTag, true
}

// GetAllTags 获取所有标签
func (s *TagSet) GetAllTags() []*ProductTag {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make([]*ProductTag, 0, len(s.tags))
	for _, tag := range s.tags {
		copiedTag := *tag
		result = append(result, &copiedTag)
	}

	return result
}

// HasTag 检查标签是否存在
func (s *TagSet) HasTag(id string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, exists := s.tags[id]
	return exists
}

// Count 获取标签数量
func (s *TagSet) Count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.tags)
}

// Clear 清空所有标签
func (s *TagSet) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.tags = make(map[string]*ProductTag)
}

// RecentlyViewedProducts 最近浏览的商品集合
type RecentlyViewedProducts struct {
	userID      UserID
	products    []ProductID
	lock        sync.RWMutex
	maxProducts int // 最大记录数
}

// NewRecentlyViewedProducts 创建最近浏览商品集合
func NewRecentlyViewedProducts(userID UserID, maxProducts int) *RecentlyViewedProducts {
	if maxProducts <= 0 {
		maxProducts = 50 // 默认记录最近50个商品
	}

	return &RecentlyViewedProducts{
		userID:      userID,
		products:    make([]ProductID, 0, maxProducts),
		maxProducts: maxProducts,
	}
}

// AddProduct 添加最近浏览的商品
func (r *RecentlyViewedProducts) AddProduct(productID ProductID) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 检查商品是否已在列表中
	for i, id := range r.products {
		if id == productID {
			// 找到了，将其移到列表最前面
			r.products = append(r.products[:i], r.products[i+1:]...)
			r.products = append([]ProductID{productID}, r.products...)
			return
		}
	}

	// 商品不在列表中，添加到最前面
	r.products = append([]ProductID{productID}, r.products...)

	// 如果超出最大长度，删除最旧的
	if len(r.products) > r.maxProducts {
		r.products = r.products[:r.maxProducts]
	}
}

// GetProducts 获取最近浏览的商品列表
func (r *RecentlyViewedProducts) GetProducts() []ProductID {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// 返回副本
	result := make([]ProductID, len(r.products))
	copy(result, r.products)
	return result
}

// Clear 清空最近浏览的商品列表
func (r *RecentlyViewedProducts) Clear() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.products = make([]ProductID, 0, r.maxProducts)
}

// WishList 心愿单实现
type WishList struct {
	userID   UserID
	products map[ProductID]time.Time // 记录添加时间
	lock     sync.RWMutex
	maxItems int
}

// NewWishList 创建心愿单
func NewWishList(userID UserID, maxItems int) *WishList {
	if maxItems <= 0 {
		maxItems = 100 // 默认最多100个商品
	}

	return &WishList{
		userID:   userID,
		products: make(map[ProductID]time.Time),
		maxItems: maxItems,
	}
}

// AddProduct 添加商品到心愿单
func (w *WishList) AddProduct(productID ProductID) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	// 检查是否已达最大商品数
	if len(w.products) >= w.maxItems && w.products[productID] == (time.Time{}) {
		return fmt.Errorf("心愿单已满，最多支持%d件商品", w.maxItems)
	}

	w.products[productID] = time.Now()
	return nil
}

// RemoveProduct 从心愿单移除商品
func (w *WishList) RemoveProduct(productID ProductID) {
	w.lock.Lock()
	defer w.lock.Unlock()

	delete(w.products, productID)
}

// HasProduct 检查商品是否在心愿单中
func (w *WishList) HasProduct(productID ProductID) bool {
	w.lock.RLock()
	defer w.lock.RUnlock()

	_, exists := w.products[productID]
	return exists
}

// GetProducts 获取心愿单中的所有商品
func (w *WishList) GetProducts() map[ProductID]time.Time {
	w.lock.RLock()
	defer w.lock.RUnlock()

	// 返回副本
	result := make(map[ProductID]time.Time, len(w.products))
	for id, t := range w.products {
		result[id] = t
	}

	return result
}

// Count 获取心愿单商品数量
func (w *WishList) Count() int {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return len(w.products)
}

// Clear 清空心愿单
func (w *WishList) Clear() {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.products = make(map[ProductID]time.Time)
}

// 商品库存状态
const (
	StockStatusInStock    = "in_stock"     // 有库存
	StockStatusLowStock   = "low_stock"    // 库存不足
	StockStatusOutOfStock = "out_of_stock" // 无库存
	StockStatusPreOrder   = "pre_order"    // 预购
)

// ProductInventory 商品库存追踪器
type ProductInventory struct {
	products map[ProductID]int // 商品ID到库存数量
	lock     sync.RWMutex
	// 库存变化通知回调
	onChange func(productID ProductID, oldQty, newQty int)
	// 库存阈值（低于此值视为库存不足）
	lowStockThreshold map[ProductID]int
}

// NewProductInventory 创建商品库存追踪器
func NewProductInventory(onChange func(productID ProductID, oldQty, newQty int)) *ProductInventory {
	return &ProductInventory{
		products:          make(map[ProductID]int),
		lowStockThreshold: make(map[ProductID]int),
		onChange:          onChange,
	}
}

// SetStock 设置商品库存
func (p *ProductInventory) SetStock(productID ProductID, quantity int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	oldQty := p.products[productID]
	p.products[productID] = quantity

	// 调用变化回调
	if p.onChange != nil && oldQty != quantity {
		p.onChange(productID, oldQty, quantity)
	}
}

// GetStock 获取商品库存
func (p *ProductInventory) GetStock(productID ProductID) int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.products[productID]
}

// DecreaseStock 减少商品库存
func (p *ProductInventory) DecreaseStock(productID ProductID, quantity int) bool {
	if quantity <= 0 {
		return false
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	currentQty := p.products[productID]
	if currentQty < quantity {
		return false // 库存不足
	}

	newQty := currentQty - quantity
	p.products[productID] = newQty

	// 调用变化回调
	if p.onChange != nil {
		p.onChange(productID, currentQty, newQty)
	}

	return true
}

// IncreaseStock 增加商品库存
func (p *ProductInventory) IncreaseStock(productID ProductID, quantity int) {
	if quantity <= 0 {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	currentQty := p.products[productID]
	newQty := currentQty + quantity
	p.products[productID] = newQty

	// 调用变化回调
	if p.onChange != nil {
		p.onChange(productID, currentQty, newQty)
	}
}

// SetLowStockThreshold 设置库存不足阈值
func (p *ProductInventory) SetLowStockThreshold(productID ProductID, threshold int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.lowStockThreshold[productID] = threshold
}

// GetStockStatus 获取商品库存状态
func (p *ProductInventory) GetStockStatus(productID ProductID) string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	quantity := p.products[productID]
	if quantity <= 0 {
		return StockStatusOutOfStock
	}

	threshold, exists := p.lowStockThreshold[productID]
	if exists && quantity <= threshold {
		return StockStatusLowStock
	}

	return StockStatusInStock
}

// LowStockProducts 获取所有库存不足的商品
func (p *ProductInventory) LowStockProducts() []ProductID {
	p.lock.RLock()
	defer p.lock.RUnlock()

	var result []ProductID

	for productID, quantity := range p.products {
		threshold, exists := p.lowStockThreshold[productID]
		if (exists && quantity <= threshold && quantity > 0) || quantity == 0 {
			result = append(result, productID)
		}
	}

	return result
}

// CategorySet 商品分类集合
type CategorySet struct {
	categories      map[string]bool
	productCategory map[ProductID]map[string]bool // 商品ID到分类的映射
	lock            sync.RWMutex
}

// NewCategorySet 创建商品分类集合
func NewCategorySet() *CategorySet {
	return &CategorySet{
		categories:      make(map[string]bool),
		productCategory: make(map[ProductID]map[string]bool),
	}
}

// AddCategory 添加分类
func (c *CategorySet) AddCategory(category string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.categories[category] = true
}

// RemoveCategory 删除分类
func (c *CategorySet) RemoveCategory(category string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.categories, category)

	// 同时从所有商品中移除该分类
	for productID, categories := range c.productCategory {
		if categories[category] {
			delete(categories, category)

			// 如果商品没有分类了，删除该商品的记录
			if len(categories) == 0 {
				delete(c.productCategory, productID)
			}
		}
	}
}

// GetAllCategories 获取所有分类
func (c *CategorySet) GetAllCategories() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]string, 0, len(c.categories))
	for category := range c.categories {
		result = append(result, category)
	}

	return result
}

// AssignCategoryToProduct 将分类分配给商品
func (c *CategorySet) AssignCategoryToProduct(productID ProductID, category string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// 确保分类存在
	c.categories[category] = true

	// 为商品添加分类
	if c.productCategory[productID] == nil {
		c.productCategory[productID] = make(map[string]bool)
	}

	c.productCategory[productID][category] = true
}

// RemoveCategoryFromProduct 从商品中移除分类
func (c *CategorySet) RemoveCategoryFromProduct(productID ProductID, category string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	categories := c.productCategory[productID]
	if categories != nil {
		delete(categories, category)

		// 如果商品没有分类了，删除该商品的记录
		if len(categories) == 0 {
			delete(c.productCategory, productID)
		}
	}
}

// GetProductCategories 获取商品的所有分类
func (c *CategorySet) GetProductCategories(productID ProductID) []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	categories := c.productCategory[productID]
	if categories == nil {
		return []string{}
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}

	return result
}

// GetProductsByCategory 获取分类下的所有商品
func (c *CategorySet) GetProductsByCategory(category string) []ProductID {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var result []ProductID

	for productID, categories := range c.productCategory {
		if categories[category] {
			result = append(result, productID)
		}
	}

	return result
}

// HasCategory 检查分类是否存在
func (c *CategorySet) HasCategory(category string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.categories[category]
}

// ProductFilter 商品过滤器
type ProductFilter struct {
	categorySet *CategorySet
}

// NewProductFilter 创建商品过滤器
func NewProductFilter(categorySet *CategorySet) *ProductFilter {
	return &ProductFilter{
		categorySet: categorySet,
	}
}

// FilterByCategory 按分类过滤商品
func (f *ProductFilter) FilterByCategory(products []ProductID, category string) []ProductID {
	if !f.categorySet.HasCategory(category) {
		return []ProductID{}
	}

	result := make([]ProductID, 0)
	for _, productID := range products {
		categories := f.categorySet.GetProductCategories(productID)
		for _, cat := range categories {
			if cat == category {
				result = append(result, productID)
				break
			}
		}
	}

	return result
}

// FilterByCategories 按多个分类过滤商品
func (f *ProductFilter) FilterByCategories(products []ProductID, categories []string, requireAll bool) []ProductID {
	if len(categories) == 0 {
		return products
	}

	result := make([]ProductID, 0)
	for _, productID := range products {
		productCategories := f.categorySet.GetProductCategories(productID)
		if requireAll {
			// 必须满足所有分类
			allMatch := true
			for _, category := range categories {
				found := false
				for _, prodCat := range productCategories {
					if prodCat == category {
						found = true
						break
					}
				}
				if !found {
					allMatch = false
					break
				}
			}
			if allMatch {
				result = append(result, productID)
			}
		} else {
			// 满足任一分类即可
			for _, prodCat := range productCategories {
				for _, category := range categories {
					if prodCat == category {
						result = append(result, productID)
						goto nextProduct // 找到一个匹配的分类，添加商品并跳到下一个商品
					}
				}
			}
		nextProduct:
		}
	}

	return result
}

// FilterByStock 按库存状态过滤商品
func (f *ProductFilter) FilterByStock(products []ProductID, inventory *ProductInventory, status string) []ProductID {
	result := make([]ProductID, 0)
	for _, productID := range products {
		if inventory.GetStockStatus(productID) == status {
			result = append(result, productID)
		}
	}

	return result
}

// FilterByInStock 过滤出有库存的商品
func (f *ProductFilter) FilterByInStock(products []ProductID, inventory *ProductInventory) []ProductID {
	result := make([]ProductID, 0)
	for _, productID := range products {
		status := inventory.GetStockStatus(productID)
		if status == StockStatusInStock || status == StockStatusLowStock {
			result = append(result, productID)
		}
	}

	return result
}
