// Package paginator 提供了用于处理分页请求和响应的工具函数和结构体。
// 支持基于偏移量的分页和基于游标的分页两种模式。
package paginator

import (
	"math"
)

const (
	// DefaultPageNum 是默认的页码，从 1 开始。
	DefaultPageNum = 1

	// DefaultPageSize 是默认的每页数据条数。
	DefaultPageSize = 10

	// MaxPageSize 是最大允许的每页数据条数，用于防止请求过大的数据量。
	MaxPageSize = 100
)

// Pageable 定义了所有分页请求必须实现的接口。
type Pageable interface {
	// Validate 验证分页参数是否有效。
	// 返回错误信息，如果参数有效则返回 nil。
	Validate() error
}

// PaginatedResponse 表示分页响应的通用结构。
// T 是分页数据项的类型。
type PaginatedResponse[T any] struct {
	// Items 是当前页的数据项列表。
	Items []T `json:"items"`

	// Total 是符合条件的总数据条数（如果可用）。
	// 在某些情况下（如游标分页），可能不提供总数。
	Total int64 `json:"total,omitempty"`

	// HasMore 表示是否还有更多数据可以获取。
	// 主要用于游标分页，但在偏移量分页中也可以使用。
	HasMore bool `json:"hasMore,omitempty"`

	// 以下字段是偏移量分页特有的
	// PageInfo 包含当前分页信息。
	PageInfo *PageInfo `json:"pageInfo,omitempty"`
}

// PageInfo 包含偏移量分页的元数据。
type PageInfo struct {
	// PageNum 是当前页码。
	PageNum int `json:"pageNum"`

	// PageSize 是每页数据条数。
	PageSize int `json:"pageSize"`

	// Pages 是总页数。
	Pages int `json:"pages"`
}

// NewPageInfo 创建一个新的分页信息对象。
func NewPageInfo(pageNum, pageSize int, total int64) *PageInfo {
	// 计算总页数
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages < 1 {
		pages = 1
	}

	return &PageInfo{
		PageNum:  pageNum,
		PageSize: pageSize,
		Pages:    pages,
	}
}

// NewOffsetPaginatedResponse 创建一个基于偏移量分页的响应。
func NewOffsetPaginatedResponse[T any](items []T, pageNum, pageSize int, total int64) *PaginatedResponse[T] {
	pageInfo := NewPageInfo(pageNum, pageSize, total)
	hasMore := pageNum < pageInfo.Pages

	return &PaginatedResponse[T]{
		Items:    items,
		Total:    total,
		HasMore:  hasMore,
		PageInfo: pageInfo,
	}
}

// NewCursorPaginatedResponse 创建一个基于游标分页的响应。
func NewCursorPaginatedResponse[T any](items []T, hasMore bool) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Items:   items,
		HasMore: hasMore,
	}
}
