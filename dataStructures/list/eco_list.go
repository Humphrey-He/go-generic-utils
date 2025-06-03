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
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// 电商特定错误
var (
	ErrStockInsufficient  = errors.New("ggu: 商品库存不足")
	ErrProductNotFound    = errors.New("ggu: 商品未找到")
	ErrCartItemNotFound   = errors.New("ggu: 购物车商品未找到")
	ErrInvalidOrderStatus = errors.New("ggu: 无效的订单状态")
)

// ----------- 商品相关 -----------

// Product 商品信息
type Product struct {
	ID              string            // 商品ID
	Name            string            // 商品名称
	Price           float64           // 商品价格
	Stock           int               // 库存数量
	SKU             string            // SKU编码
	Category        string            // 分类
	Tags            []string          // 标签
	Attributes      map[string]string // 属性
	CreatedAt       time.Time         // 创建时间
	UpdatedAt       time.Time         // 更新时间
	SalesCount      int               // 销量
	IsActive        bool              // 是否上架
	DiscountPercent float64           // 折扣百分比(0-100)
}

// DiscountedPrice 计算折扣后价格
func (p *Product) DiscountedPrice() float64 {
	if p.DiscountPercent <= 0 || p.DiscountPercent >= 100 {
		return p.Price
	}
	return p.Price * (1 - p.DiscountPercent/100)
}

// DeductStock 扣减库存
func (p *Product) DeductStock(count int) error {
	if p.Stock < count {
		return ErrStockInsufficient
	}
	p.Stock -= count
	p.UpdatedAt = time.Now()
	return nil
}

// AddStock 增加库存
func (p *Product) AddStock(count int) {
	p.Stock += count
	p.UpdatedAt = time.Now()
}

// ProductList 商品列表管理
type ProductList struct {
	ConcurrentList[Product]
}

// NewProductList 创建商品列表
func NewProductList(cap int) *ProductList {
	return &ProductList{
		ConcurrentList: ConcurrentList[Product]{
			List: NewArrayList[Product](cap),
		},
	}
}

// FindByID 根据ID查找商品
func (pl *ProductList) FindByID(id string) (Product, error) {
	var zero Product

	var foundProduct Product
	var found bool

	err := pl.Range(func(index int, p Product) error {
		if p.ID == id {
			foundProduct = p
			found = true
			return ErrInvalidArgument // 使用错误中断查找
		}
		return nil
	})

	if err != nil && err != ErrInvalidArgument {
		return zero, err
	}

	if !found {
		return zero, ErrProductNotFound
	}

	return foundProduct, nil
}

// FindByCategory 根据分类查找商品
func (pl *ProductList) FindByCategory(category string) []Product {
	result := make([]Product, 0)

	_ = pl.Range(func(_ int, p Product) error {
		if p.Category == category && p.IsActive {
			result = append(result, p)
		}
		return nil
	})

	return result
}

// FindByTags 根据标签查找商品
func (pl *ProductList) FindByTags(tags []string) []Product {
	result := make([]Product, 0)

	_ = pl.Range(func(_ int, p Product) error {
		if !p.IsActive {
			return nil
		}

		// 检查是否包含任一标签
		for _, tag := range tags {
			for _, productTag := range p.Tags {
				if tag == productTag {
					result = append(result, p)
					return nil // 找到一个匹配标签就添加产品
				}
			}
		}
		return nil
	})

	return result
}

// UpdateStock 更新商品库存
func (pl *ProductList) UpdateStock(productID string, change int) error {
	pl.lock.Lock()
	defer pl.lock.Unlock()

	var productIndex = -1

	err := pl.List.Range(func(i int, p Product) error {
		if p.ID == productID {
			productIndex = i
			return ErrInvalidArgument // 使用错误中断查找
		}
		return nil
	})

	if err != nil && err != ErrInvalidArgument {
		return err
	}

	if productIndex == -1 {
		return ErrProductNotFound
	}

	product, err := pl.List.Get(productIndex)
	if err != nil {
		return err
	}

	if change < 0 {
		if product.Stock < -change {
			return ErrStockInsufficient
		}
	}

	product.Stock += change
	product.UpdatedAt = time.Now()

	return pl.List.Set(productIndex, product)
}

// GetTopSellers 获取销量最高的商品
func (pl *ProductList) GetTopSellers(limit int) []Product {
	if limit <= 0 {
		return []Product{}
	}

	// 获取所有产品的快照
	pl.lock.RLock()
	allProducts := pl.List.AsSlice()
	pl.lock.RUnlock()

	// 只保留上架的商品
	activeProducts := make([]Product, 0, len(allProducts))
	for _, p := range allProducts {
		if p.IsActive {
			activeProducts = append(activeProducts, p)
		}
	}

	// 按销量排序（降序）
	for i := 0; i < len(activeProducts)-1; i++ {
		for j := i + 1; j < len(activeProducts); j++ {
			if activeProducts[j].SalesCount > activeProducts[i].SalesCount {
				activeProducts[i], activeProducts[j] = activeProducts[j], activeProducts[i]
			}
		}
	}

	// 返回前limit个
	if len(activeProducts) <= limit {
		return activeProducts
	}
	return activeProducts[:limit]
}

// GetDiscountedProducts 获取打折商品
func (pl *ProductList) GetDiscountedProducts() []Product {
	result := make([]Product, 0)

	_ = pl.Range(func(_ int, p Product) error {
		if p.IsActive && p.DiscountPercent > 0 && p.DiscountPercent < 100 {
			result = append(result, p)
		}
		return nil
	})

	return result
}

// ----------- 购物车相关 -----------

// CartItem 购物车项目
type CartItem struct {
	ProductID  string    // 商品ID
	Name       string    // 商品名称
	Price      float64   // 加入时的价格
	Quantity   int       // 数量
	Attributes string    // 选择的属性（如颜色、尺寸等）
	AddedAt    time.Time // 添加时间
	IsSelected bool      // 是否选中
	UpdatedAt  time.Time // 更新时间
}

// TotalPrice 计算购物车项目总价
func (ci *CartItem) TotalPrice() float64 {
	return ci.Price * float64(ci.Quantity)
}

// ShoppingCart 购物车
type ShoppingCart struct {
	UserID    string               // 用户ID
	Items     *ArrayList[CartItem] // 购物车项目
	CreatedAt time.Time            // 创建时间
	UpdatedAt time.Time            // 更新时间
	lock      sync.Mutex           // 购物车锁
}

// NewShoppingCart 创建购物车
func NewShoppingCart(userID string) *ShoppingCart {
	now := time.Now()
	return &ShoppingCart{
		UserID:    userID,
		Items:     NewArrayList[CartItem](10),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddItem 添加商品到购物车
func (sc *ShoppingCart) AddItem(product Product, quantity int, attributes string) error {
	if quantity <= 0 {
		return ErrInvalidArgument
	}

	sc.lock.Lock()
	defer sc.lock.Unlock()

	// 检查商品是否已在购物车中
	found := false
	_ = sc.Items.Range(func(i int, item CartItem) error {
		if item.ProductID == product.ID && item.Attributes == attributes {
			// 更新数量
			item.Quantity += quantity
			item.UpdatedAt = time.Now()
			_ = sc.Items.Set(i, item)
			found = true
			return ErrInvalidArgument // 使用错误中断遍历
		}
		return nil
	})

	if !found {
		// 添加新商品
		cartItem := CartItem{
			ProductID:  product.ID,
			Name:       product.Name,
			Price:      product.DiscountedPrice(),
			Quantity:   quantity,
			Attributes: attributes,
			AddedAt:    time.Now(),
			IsSelected: true,
		}
		_ = sc.Items.Append(cartItem)
	}

	sc.UpdatedAt = time.Now()
	return nil
}

// UpdateItemQuantity 更新购物车项目数量
func (sc *ShoppingCart) UpdateItemQuantity(productID string, attributes string, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidArgument
	}

	sc.lock.Lock()
	defer sc.lock.Unlock()

	found := false
	_ = sc.Items.Range(func(i int, item CartItem) error {
		if item.ProductID == productID && item.Attributes == attributes {
			item.Quantity = quantity
			_ = sc.Items.Set(i, item)
			found = true
			return ErrInvalidArgument // 使用错误中断遍历
		}
		return nil
	})

	if !found {
		return ErrCartItemNotFound
	}

	sc.UpdatedAt = time.Now()
	return nil
}

// RemoveItem 从购物车移除商品
func (sc *ShoppingCart) RemoveItem(productID string, attributes string) error {
	sc.lock.Lock()
	defer sc.lock.Unlock()

	items := sc.Items.AsSlice()
	for i, item := range items {
		if item.ProductID == productID && item.Attributes == attributes {
			_, err := sc.Items.Delete(i)
			if err != nil {
				return err
			}
			sc.UpdatedAt = time.Now()
			return nil
		}
	}

	return ErrCartItemNotFound
}

// ClearCart 清空购物车
func (sc *ShoppingCart) ClearCart() {
	sc.lock.Lock()
	defer sc.lock.Unlock()

	sc.Items.Clear()
	sc.UpdatedAt = time.Now()
}

// GetSelectedItems 获取选中的商品
func (sc *ShoppingCart) GetSelectedItems() []CartItem {
	sc.lock.Lock()
	defer sc.lock.Unlock()

	var selectedItems []CartItem
	_ = sc.Items.Range(func(_ int, item CartItem) error {
		if item.IsSelected {
			selectedItems = append(selectedItems, item)
		}
		return nil
	})

	return selectedItems
}

// CalculateTotal 计算购物车总价
func (sc *ShoppingCart) CalculateTotal() float64 {
	selectedItems := sc.GetSelectedItems()
	var total float64
	for _, item := range selectedItems {
		total += item.TotalPrice()
	}
	return total
}

// ----------- 订单相关 -----------

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "待支付"
	OrderStatusPaid       OrderStatus = "已支付"
	OrderStatusProcessing OrderStatus = "处理中"
	OrderStatusShipped    OrderStatus = "已发货"
	OrderStatusDelivered  OrderStatus = "已送达"
	OrderStatusCompleted  OrderStatus = "已完成"
	OrderStatusCancelled  OrderStatus = "已取消"
	OrderStatusRefunded   OrderStatus = "已退款"
)

// OrderItem 订单项目
type OrderItem struct {
	ProductID  string  // 商品ID
	Name       string  // 商品名称
	SKU        string  // SKU
	Price      float64 // 单价
	Quantity   int     // 数量
	Attributes string  // 属性
}

// Order 订单
type Order struct {
	ID           string            // 订单ID
	UserID       string            // 用户ID
	Items        []OrderItem       // 订单项目
	Status       OrderStatus       // 订单状态
	TotalAmount  float64           // 总金额
	PaymentInfo  map[string]string // 支付信息
	ShippingInfo map[string]string // 配送信息
	CreatedAt    time.Time         // 创建时间
	UpdatedAt    time.Time         // 更新时间
	PaidAt       *time.Time        // 支付时间
	ShippedAt    *time.Time        // 发货时间
	DeliveredAt  *time.Time        // 送达时间
	CompletedAt  *time.Time        // 完成时间
	CancelledAt  *time.Time        // 取消时间
	RefundedAt   *time.Time        // 退款时间
	Notes        string            // 备注
}

// UpdateStatus 更新订单状态
func (o *Order) UpdateStatus(status OrderStatus) error {
	now := time.Now()

	switch status {
	case OrderStatusPaid:
		o.PaidAt = &now
	case OrderStatusShipped:
		o.ShippedAt = &now
	case OrderStatusDelivered:
		o.DeliveredAt = &now
	case OrderStatusCompleted:
		o.CompletedAt = &now
	case OrderStatusCancelled:
		o.CancelledAt = &now
	case OrderStatusRefunded:
		o.RefundedAt = &now
	case OrderStatusPending, OrderStatusProcessing:
		// 这些状态不需要记录时间
	default:
		return ErrInvalidOrderStatus
	}

	o.Status = status
	o.UpdatedAt = now
	return nil
}

// OrderList 订单列表
type OrderList struct {
	ConcurrentList[Order]
}

// NewOrderList 创建订单列表
func NewOrderList(cap int) *OrderList {
	return &OrderList{
		ConcurrentList: ConcurrentList[Order]{
			List: NewArrayList[Order](cap),
		},
	}
}

// FindByID 根据ID查找订单
func (ol *OrderList) FindByID(id string) (Order, error) {
	var zero Order

	var foundOrder Order
	var found bool

	err := ol.Range(func(index int, o Order) error {
		if o.ID == id {
			foundOrder = o
			found = true
			return ErrInvalidArgument // 使用错误中断查找
		}
		return nil
	})

	if err != nil && err != ErrInvalidArgument {
		return zero, err
	}

	if !found {
		return zero, fmt.Errorf("订单未找到: %s", id)
	}

	return foundOrder, nil
}

// FindByUserID 查找用户的订单
func (ol *OrderList) FindByUserID(userID string) []Order {
	result := make([]Order, 0)

	_ = ol.Range(func(_ int, o Order) error {
		if o.UserID == userID {
			result = append(result, o)
		}
		return nil
	})

	return result
}

// FindByStatus 根据状态查找订单
func (ol *OrderList) FindByStatus(status OrderStatus) []Order {
	result := make([]Order, 0)

	_ = ol.Range(func(_ int, o Order) error {
		if o.Status == status {
			result = append(result, o)
		}
		return nil
	})

	return result
}

// UpdateOrderStatus 更新订单状态
func (ol *OrderList) UpdateOrderStatus(orderID string, status OrderStatus) error {
	ol.lock.Lock()
	defer ol.lock.Unlock()

	var orderIndex = -1

	err := ol.List.Range(func(i int, o Order) error {
		if o.ID == orderID {
			orderIndex = i
			return ErrInvalidArgument // 使用错误中断查找
		}
		return nil
	})

	if err != nil && err != ErrInvalidArgument {
		return err
	}

	if orderIndex == -1 {
		return fmt.Errorf("订单未找到: %s", orderID)
	}

	order, err := ol.List.Get(orderIndex)
	if err != nil {
		return err
	}

	if err := order.UpdateStatus(status); err != nil {
		return err
	}

	return ol.List.Set(orderIndex, order)
}

// ----------- 商品搜索与排序 -----------

// ProductSearchEngine 商品搜索引擎
type ProductSearchEngine struct {
	ProductList *ProductList
}

// NewProductSearchEngine 创建商品搜索引擎
func NewProductSearchEngine(productList *ProductList) *ProductSearchEngine {
	return &ProductSearchEngine{
		ProductList: productList,
	}
}

// 搜索结果排序方式
const (
	SortByPriceAsc  = "price_asc"  // 价格升序
	SortByPriceDesc = "price_desc" // 价格降序
	SortBySales     = "sales"      // 销量
	SortByNewest    = "newest"     // 最新
)

// SearchResult 搜索结果
type SearchResult struct {
	Products    []Product // 商品列表
	TotalCount  int       // 总数量
	CurrentPage int       // 当前页
	PageSize    int       // 每页大小
	TotalPages  int       // 总页数
}

// Search 搜索商品
func (pse *ProductSearchEngine) Search(
	keyword string,
	category string,
	tags []string,
	priceMin, priceMax float64,
	sortBy string,
	page, pageSize int) SearchResult {

	// 获取所有上架商品
	allProducts := make([]Product, 0)

	_ = pse.ProductList.Range(func(_ int, p Product) error {
		if p.IsActive {
			allProducts = append(allProducts, p)
		}
		return nil
	})

	// 应用筛选条件
	filtered := pse.filterProducts(allProducts, keyword, category, tags, priceMin, priceMax)

	// 排序
	sorted := pse.sortProducts(filtered, sortBy)

	// 分页
	result := SearchResult{
		TotalCount:  len(sorted),
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  int(math.Ceil(float64(len(sorted)) / float64(pageSize))),
	}

	// 计算当前页的商品
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex < 0 {
		startIndex = 0
	}

	if startIndex >= len(sorted) {
		result.Products = []Product{}
		return result
	}

	if endIndex > len(sorted) {
		endIndex = len(sorted)
	}

	result.Products = sorted[startIndex:endIndex]
	return result
}

// filterProducts 过滤商品
func (pse *ProductSearchEngine) filterProducts(
	products []Product,
	keyword string,
	category string,
	tags []string,
	priceMin, priceMax float64) []Product {

	result := make([]Product, 0)

	for _, p := range products {
		// 关键词筛选
		if keyword != "" {
			if !strings.Contains(strings.ToLower(p.Name), strings.ToLower(keyword)) &&
				!strings.Contains(strings.ToLower(p.SKU), strings.ToLower(keyword)) {
				continue
			}
		}

		// 分类筛选
		if category != "" && p.Category != category {
			continue
		}

		// 标签筛选
		if len(tags) > 0 {
			hasTag := false
			for _, tag := range tags {
				for _, pTag := range p.Tags {
					if pTag == tag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// 价格筛选
		price := p.DiscountedPrice()
		if priceMin > 0 && price < priceMin {
			continue
		}
		if priceMax > 0 && price > priceMax {
			continue
		}

		result = append(result, p)
	}

	return result
}

// sortProducts 排序商品
func (pse *ProductSearchEngine) sortProducts(products []Product, sortBy string) []Product {
	result := make([]Product, len(products))
	copy(result, products)

	switch sortBy {
	case SortByPriceAsc:
		sort.Slice(result, func(i, j int) bool {
			return result[i].DiscountedPrice() < result[j].DiscountedPrice()
		})
	case SortByPriceDesc:
		sort.Slice(result, func(i, j int) bool {
			return result[i].DiscountedPrice() > result[j].DiscountedPrice()
		})
	case SortBySales:
		sort.Slice(result, func(i, j int) bool {
			return result[i].SalesCount > result[j].SalesCount
		})
	case SortByNewest:
		sort.Slice(result, func(i, j int) bool {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		})
	default:
		// 默认按ID排序
		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
	}

	return result
}
