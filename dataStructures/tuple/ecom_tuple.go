package tuple

import (
	"fmt"
	"sort"
	"time"
)

// ----------- 商品与价格相关元组 -----------

// ProductPrice 商品与价格的键值对
// 适用于商品ID与价格的映射、批量价格更新等场景
type ProductPrice struct {
	ProductID string  // 商品ID
	Price     float64 // 价格
}

// NewProductPrice 创建商品-价格对
func NewProductPrice(productID string, price float64) ProductPrice {
	return ProductPrice{
		ProductID: productID,
		Price:     price,
	}
}

// String 返回ProductPrice的字符串表示
func (pp ProductPrice) String() string {
	return fmt.Sprintf("商品 %s: $%.2f", pp.ProductID, pp.Price)
}

// AsPair 转换为通用Pair
func (pp ProductPrice) AsPair() Pair[string, float64] {
	return NewPair(pp.ProductID, pp.Price)
}

// ProductPriceList 商品价格列表
type ProductPriceList []ProductPrice

// SortByPrice 按价格排序（升序）
func (ppl ProductPriceList) SortByPrice() {
	sort.Slice(ppl, func(i, j int) bool {
		return ppl[i].Price < ppl[j].Price
	})
}

// SortByPriceDesc 按价格排序（降序）
func (ppl ProductPriceList) SortByPriceDesc() {
	sort.Slice(ppl, func(i, j int) bool {
		return ppl[i].Price > ppl[j].Price
	})
}

// FilterByPriceRange 按价格范围过滤
func (ppl ProductPriceList) FilterByPriceRange(min, max float64) ProductPriceList {
	result := make(ProductPriceList, 0)
	for _, pp := range ppl {
		if pp.Price >= min && (max <= 0 || pp.Price <= max) {
			result = append(result, pp)
		}
	}
	return result
}

// TotalPrice 计算总价
func (ppl ProductPriceList) TotalPrice() float64 {
	var total float64
	for _, pp := range ppl {
		total += pp.Price
	}
	return total
}

// ----------- 商品与库存相关元组 -----------

// ProductStock 商品与库存的键值对
// 适用于库存管理、库存警告等场景
type ProductStock struct {
	ProductID   string    // 商品ID
	Stock       int       // 库存数量
	UpdatedAt   time.Time // 更新时间
	LowWarning  int       // 低库存警告阈值
	HighWarning int       // 高库存警告阈值
}

// NewProductStock 创建商品-库存对
func NewProductStock(productID string, stock int) ProductStock {
	return ProductStock{
		ProductID: productID,
		Stock:     stock,
		UpdatedAt: time.Now(),
	}
}

// String 返回ProductStock的字符串表示
func (ps ProductStock) String() string {
	return fmt.Sprintf("商品 %s: 库存 %d", ps.ProductID, ps.Stock)
}

// IsLowStock 判断是否低库存
func (ps ProductStock) IsLowStock() bool {
	return ps.LowWarning > 0 && ps.Stock <= ps.LowWarning
}

// IsHighStock 判断是否高库存
func (ps ProductStock) IsHighStock() bool {
	return ps.HighWarning > 0 && ps.Stock >= ps.HighWarning
}

// AsPair 转换为通用Pair
func (ps ProductStock) AsPair() Pair[string, int] {
	return NewPair(ps.ProductID, ps.Stock)
}

// ProductStockList 商品库存列表
type ProductStockList []ProductStock

// SortByStock 按库存排序（升序）
func (psl ProductStockList) SortByStock() {
	sort.Slice(psl, func(i, j int) bool {
		return psl[i].Stock < psl[j].Stock
	})
}

// FindLowStock 查找低库存商品
func (psl ProductStockList) FindLowStock() ProductStockList {
	result := make(ProductStockList, 0)
	for _, ps := range psl {
		if ps.IsLowStock() {
			result = append(result, ps)
		}
	}
	return result
}

// FindOutOfStock 查找缺货商品
func (psl ProductStockList) FindOutOfStock() ProductStockList {
	result := make(ProductStockList, 0)
	for _, ps := range psl {
		if ps.Stock <= 0 {
			result = append(result, ps)
		}
	}
	return result
}

// TotalStock 计算总库存
func (psl ProductStockList) TotalStock() int {
	var total int
	for _, ps := range psl {
		total += ps.Stock
	}
	return total
}

// ----------- 用户与订单相关元组 -----------

// UserOrder 用户与订单的关联
// 适用于用户订单查询、订单分析等场景
type UserOrder struct {
	UserID    string    // 用户ID
	OrderID   string    // 订单ID
	OrderTime time.Time // 订单时间
	Status    string    // 订单状态
	Amount    float64   // 订单金额
}

// NewUserOrder 创建用户-订单关联
func NewUserOrder(userID, orderID string, amount float64) UserOrder {
	return UserOrder{
		UserID:    userID,
		OrderID:   orderID,
		OrderTime: time.Now(),
		Status:    "待支付",
		Amount:    amount,
	}
}

// String 返回UserOrder的字符串表示
func (uo UserOrder) String() string {
	return fmt.Sprintf("用户 %s 的订单 %s: $%.2f (%s)", uo.UserID, uo.OrderID, uo.Amount, uo.Status)
}

// AsTriple 转换为三元组
func (uo UserOrder) AsTriple() Triple[string, string, float64] {
	return NewTriple(uo.UserID, uo.OrderID, uo.Amount)
}

// UserOrderList 用户订单列表
type UserOrderList []UserOrder

// FilterByUser 按用户过滤
func (uol UserOrderList) FilterByUser(userID string) UserOrderList {
	result := make(UserOrderList, 0)
	for _, uo := range uol {
		if uo.UserID == userID {
			result = append(result, uo)
		}
	}
	return result
}

// FilterByStatus 按状态过滤
func (uol UserOrderList) FilterByStatus(status string) UserOrderList {
	result := make(UserOrderList, 0)
	for _, uo := range uol {
		if uo.Status == status {
			result = append(result, uo)
		}
	}
	return result
}

// SortByTime 按时间排序
func (uol UserOrderList) SortByTime() {
	sort.Slice(uol, func(i, j int) bool {
		return uol[i].OrderTime.Before(uol[j].OrderTime)
	})
}

// SortByAmount 按金额排序
func (uol UserOrderList) SortByAmount() {
	sort.Slice(uol, func(i, j int) bool {
		return uol[i].Amount < uol[j].Amount
	})
}

// SumAmount 计算总金额
func (uol UserOrderList) SumAmount() float64 {
	var total float64
	for _, uo := range uol {
		total += uo.Amount
	}
	return total
}

// ----------- 商品与分类相关元组 -----------

// ProductCategory 商品与分类的键值对
// 适用于商品分类管理、分类统计等场景
type ProductCategory struct {
	ProductID   string // 商品ID
	Category    string // 分类
	SubCategory string // 子分类
}

// NewProductCategory 创建商品-分类对
func NewProductCategory(productID, category, subCategory string) ProductCategory {
	return ProductCategory{
		ProductID:   productID,
		Category:    category,
		SubCategory: subCategory,
	}
}

// String 返回ProductCategory的字符串表示
func (pc ProductCategory) String() string {
	if pc.SubCategory != "" {
		return fmt.Sprintf("商品 %s: %s > %s", pc.ProductID, pc.Category, pc.SubCategory)
	}
	return fmt.Sprintf("商品 %s: %s", pc.ProductID, pc.Category)
}

// AsPair 转换为通用Pair
func (pc ProductCategory) AsPair() Pair[string, string] {
	return NewPair(pc.ProductID, pc.Category)
}

// AsTriple 转换为三元组
func (pc ProductCategory) AsTriple() Triple[string, string, string] {
	return NewTriple(pc.ProductID, pc.Category, pc.SubCategory)
}

// ProductCategoryList 商品分类列表
type ProductCategoryList []ProductCategory

// FilterByCategory 按分类过滤
func (pcl ProductCategoryList) FilterByCategory(category string) ProductCategoryList {
	result := make(ProductCategoryList, 0)
	for _, pc := range pcl {
		if pc.Category == category {
			result = append(result, pc)
		}
	}
	return result
}

// FilterBySubCategory 按子分类过滤
func (pcl ProductCategoryList) FilterBySubCategory(subCategory string) ProductCategoryList {
	result := make(ProductCategoryList, 0)
	for _, pc := range pcl {
		if pc.SubCategory == subCategory {
			result = append(result, pc)
		}
	}
	return result
}

// CountByCategory 按分类统计数量
func (pcl ProductCategoryList) CountByCategory() map[string]int {
	result := make(map[string]int)
	for _, pc := range pcl {
		result[pc.Category]++
	}
	return result
}

// ----------- 商品与评分相关元组 -----------

// ProductRating 商品与评分的键值对
// 适用于商品评分、评价统计等场景
type ProductRating struct {
	ProductID string    // 商品ID
	UserID    string    // 用户ID
	Rating    float64   // 评分(1-5)
	Comment   string    // 评价内容
	Time      time.Time // 评价时间
}

// NewProductRating 创建商品-评分对
func NewProductRating(productID, userID string, rating float64, comment string) ProductRating {
	return ProductRating{
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
		Time:      time.Now(),
	}
}

// String 返回ProductRating的字符串表示
func (pr ProductRating) String() string {
	return fmt.Sprintf("商品 %s: %.1f分 (%s)", pr.ProductID, pr.Rating, pr.UserID)
}

// AsPair 转换为通用Pair
func (pr ProductRating) AsPair() Pair[string, float64] {
	return NewPair(pr.ProductID, pr.Rating)
}

// ProductRatingList 商品评分列表
type ProductRatingList []ProductRating

// AverageRating 计算平均评分
func (prl ProductRatingList) AverageRating() float64 {
	if len(prl) == 0 {
		return 0
	}

	var total float64
	for _, pr := range prl {
		total += pr.Rating
	}
	return total / float64(len(prl))
}

// FilterByMinRating 按最低评分过滤
func (prl ProductRatingList) FilterByMinRating(minRating float64) ProductRatingList {
	result := make(ProductRatingList, 0)
	for _, pr := range prl {
		if pr.Rating >= minRating {
			result = append(result, pr)
		}
	}
	return result
}

// GroupByProduct 按商品分组
func (prl ProductRatingList) GroupByProduct() map[string]ProductRatingList {
	result := make(map[string]ProductRatingList)
	for _, pr := range prl {
		result[pr.ProductID] = append(result[pr.ProductID], pr)
	}
	return result
}

// ----------- 购物车商品与数量相关元组 -----------

// CartItem 购物车项目
// 适用于购物车管理、结算等场景
type CartItem struct {
	ProductID string  // 商品ID
	Quantity  int     // 数量
	Price     float64 // 单价
	Selected  bool    // 是否选中
}

// NewCartItem 创建购物车项目
func NewCartItem(productID string, quantity int, price float64) CartItem {
	return CartItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
		Selected:  true,
	}
}

// String 返回CartItem的字符串表示
func (ci CartItem) String() string {
	return fmt.Sprintf("商品 %s: %d件 x $%.2f", ci.ProductID, ci.Quantity, ci.Price)
}

// TotalPrice 计算总价
func (ci CartItem) TotalPrice() float64 {
	return ci.Price * float64(ci.Quantity)
}

// AsPair 转换为通用Pair
func (ci CartItem) AsPair() Pair[string, int] {
	return NewPair(ci.ProductID, ci.Quantity)
}

// CartItemList 购物车项目列表
type CartItemList []CartItem

// TotalQuantity 计算总数量
func (cil CartItemList) TotalQuantity() int {
	var total int
	for _, ci := range cil {
		if ci.Selected {
			total += ci.Quantity
		}
	}
	return total
}

// TotalAmount 计算总金额
func (cil CartItemList) TotalAmount() float64 {
	var total float64
	for _, ci := range cil {
		if ci.Selected {
			total += ci.TotalPrice()
		}
	}
	return total
}

// FilterSelected 过滤出选中的项目
func (cil CartItemList) FilterSelected() CartItemList {
	result := make(CartItemList, 0)
	for _, ci := range cil {
		if ci.Selected {
			result = append(result, ci)
		}
	}
	return result
}

// UpdateQuantity 更新商品数量
func (cil CartItemList) UpdateQuantity(productID string, quantity int) bool {
	for i, ci := range cil {
		if ci.ProductID == productID {
			cil[i].Quantity = quantity
			return true
		}
	}
	return false
}

// ----------- 运营常用的分析元组 -----------

// TimeValuePair 时间-数值对
// 适用于销售趋势、访问量统计等场景
type TimeValuePair struct {
	Time  time.Time // 时间点
	Value float64   // 数值
}

// NewTimeValuePair 创建时间-数值对
func NewTimeValuePair(t time.Time, value float64) TimeValuePair {
	return TimeValuePair{
		Time:  t,
		Value: value,
	}
}

// String 返回TimeValuePair的字符串表示
func (tvp TimeValuePair) String() string {
	return fmt.Sprintf("%s: %.2f", tvp.Time.Format("2006-01-02 15:04:05"), tvp.Value)
}

// AsPair 转换为通用Pair
func (tvp TimeValuePair) AsPair() Pair[time.Time, float64] {
	return NewPair(tvp.Time, tvp.Value)
}

// TimeValueList 时间-数值列表
type TimeValueList []TimeValuePair

// SortByTime 按时间排序
func (tvl TimeValueList) SortByTime() {
	sort.Slice(tvl, func(i, j int) bool {
		return tvl[i].Time.Before(tvl[j].Time)
	})
}

// SumValues 计算数值总和
func (tvl TimeValueList) SumValues() float64 {
	var total float64
	for _, tv := range tvl {
		total += tv.Value
	}
	return total
}

// AverageValue 计算平均值
func (tvl TimeValueList) AverageValue() float64 {
	if len(tvl) == 0 {
		return 0
	}
	return tvl.SumValues() / float64(len(tvl))
}

// FilterByTimeRange 按时间范围过滤
func (tvl TimeValueList) FilterByTimeRange(start, end time.Time) TimeValueList {
	result := make(TimeValueList, 0)
	for _, tv := range tvl {
		if (start.IsZero() || !tv.Time.Before(start)) && (end.IsZero() || !tv.Time.After(end)) {
			result = append(result, tv)
		}
	}
	return result
}

// GroupByDay 按天分组
func (tvl TimeValueList) GroupByDay() map[string]TimeValueList {
	result := make(map[string]TimeValueList)
	for _, tv := range tvl {
		day := tv.Time.Format("2006-01-02")
		result[day] = append(result[day], tv)
	}
	return result
}

// DailyTotal 计算每日总量
func (tvl TimeValueList) DailyTotal() []Pair[string, float64] {
	groups := tvl.GroupByDay()
	result := make([]Pair[string, float64], 0, len(groups))

	for day, values := range groups {
		total := 0.0
		for _, v := range values {
			total += v.Value
		}
		result = append(result, NewPair(day, total))
	}

	// 按日期排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result
}
