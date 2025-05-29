package paginator

import (
	"errors"
	"strconv"

	"github.com/noobtrump/go-generic-utils/ginutil/binding"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	// ErrInvalidLimit 表示无效的限制数错误。
	ErrInvalidLimit = errors.New("限制数必须是大于0且不超过最大限制的整数")

	// ErrInvalidCursor 表示无效的游标错误。
	ErrInvalidCursor = errors.New("无效的游标")
)

// CursorType 定义了游标的类型。
type CursorType string

const (
	// CursorTypeNext 表示获取下一页数据的游标。
	CursorTypeNext CursorType = "after"

	// CursorTypePrev 表示获取上一页数据的游标。
	CursorTypePrev CursorType = "before"
)

// CursorPageRequest 保存基于游标的分页参数。
type CursorPageRequest[C any] struct {
	// Cursor 是当前游标值，可以是任意类型。
	Cursor C `json:"cursor,omitempty"`

	// CursorType 表示游标类型，可以是 "after" 或 "before"。
	CursorType CursorType `json:"cursorType,omitempty"`

	// Limit 是要获取的数据条数。
	Limit int `form:"limit" json:"limit" binding:"omitempty,gte=1,lte=100"`

	// 解析函数，用于将字符串转换为游标类型。
	parseCursorFunc func(string) (C, error)
}

// GetLimit 返回要获取的数据条数，确保其在有效范围内。
func (r *CursorPageRequest[C]) GetLimit() int {
	if r.Limit <= 0 {
		r.Limit = DefaultPageSize
	}
	if r.Limit > MaxPageSize {
		return MaxPageSize
	}
	return r.Limit
}

// Validate 实现 Pageable 接口，验证分页参数是否有效。
func (r *CursorPageRequest[C]) Validate() error {
	if r.Limit <= 0 || r.Limit > MaxPageSize {
		return ErrInvalidLimit
	}
	return nil
}

// HasCursor 返回是否提供了游标。
func (r *CursorPageRequest[C]) HasCursor() bool {
	// 对于字符串类型的游标，检查是否为空字符串
	if cursorStr, ok := any(r.Cursor).(string); ok {
		return cursorStr != ""
	}

	// 对于整数类型的游标，检查是否为0（假设0是无效游标）
	if cursorInt, ok := any(r.Cursor).(int); ok {
		return cursorInt != 0
	}

	// 对于其他类型，假设非零值表示有游标
	return any(r.Cursor) != nil
}

// IsForward 返回是否是向前查询（使用 after 游标）。
func (r *CursorPageRequest[C]) IsForward() bool {
	return r.CursorType == CursorTypeNext || r.CursorType == ""
}

// IsBackward 返回是否是向后查询（使用 before 游标）。
func (r *CursorPageRequest[C]) IsBackward() bool {
	return r.CursorType == CursorTypePrev
}

// ParseCursorParams 从 gin.Context 解析游标分页参数。
// 参数：
//   - c: gin.Context
//   - parseCursor: 函数，用于将字符串转换为游标类型
//
// 返回值：
//   - *CursorPageRequest[C]: 游标分页请求
//   - error: 解析错误
func ParseCursorParams[C any](c *gin.Context, parseCursor func(string) (C, error)) (*CursorPageRequest[C], error) {
	if c == nil {
		return DefaultCursorPageRequest[C](parseCursor), nil
	}

	// 解析 limit
	limitStr := c.DefaultQuery("limit", strconv.Itoa(DefaultPageSize))
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
		if err != nil {
			return nil, errors.New("无效的限制数: " + err.Error())
		}
		if limit <= 0 || limit > MaxPageSize {
			return nil, ErrInvalidLimit
		}
	}

	// 创建请求对象
	req := &CursorPageRequest[C]{
		Limit:           limit,
		parseCursorFunc: parseCursor,
	}

	// 解析 after 游标
	afterCursor := c.Query("after")
	if afterCursor != "" {
		cursor, err := parseCursor(afterCursor)
		if err != nil {
			return nil, errors.New("无效的 after 游标: " + err.Error())
		}
		req.Cursor = cursor
		req.CursorType = CursorTypeNext
		return req, nil
	}

	// 解析 before 游标
	beforeCursor := c.Query("before")
	if beforeCursor != "" {
		cursor, err := parseCursor(beforeCursor)
		if err != nil {
			return nil, errors.New("无效的 before 游标: " + err.Error())
		}
		req.Cursor = cursor
		req.CursorType = CursorTypePrev
		return req, nil
	}

	// 没有游标，默认为向前查询
	return req, nil
}

// ParseCursorParamsWithBinding 使用 gin 的绑定功能解析游标分页参数。
// 返回详细的字段验证错误。
func ParseCursorParamsWithBinding[C any](c *gin.Context, parseCursor func(string) (C, error)) (*CursorPageRequest[C], binding.FieldErrors) {
	if c == nil {
		return DefaultCursorPageRequest[C](parseCursor), nil
	}

	// 解析 limit
	var limitReq struct {
		Limit int `form:"limit" binding:"omitempty,gte=1,lte=100"`
	}

	if err := c.ShouldBindQuery(&limitReq); err != nil {
		var fieldErrors binding.FieldErrors

		// 处理验证错误
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrs {
				var message string
				switch fieldErr.Tag() {
				case "gte":
					message = "必须大于或等于 1"
				case "lte":
					message = "必须小于或等于 " + strconv.Itoa(MaxPageSize)
				default:
					message = "无效的参数值"
				}

				fieldErrors.Add(fieldErr.Field(), fieldErr.Tag(), message, fieldErr.Value())
			}
			return nil, fieldErrors
		}

		// 其他绑定错误
		fieldErrors.Add("pagination", "binding", "分页参数绑定失败", nil)
		return nil, fieldErrors
	}

	// 设置默认值
	limit := limitReq.Limit
	if limit <= 0 {
		limit = DefaultPageSize
	} else if limit > MaxPageSize {
		limit = MaxPageSize
	}

	// 创建请求对象
	req := &CursorPageRequest[C]{
		Limit:           limit,
		parseCursorFunc: parseCursor,
	}

	// 解析 after 游标
	afterCursor := c.Query("after")
	if afterCursor != "" {
		cursor, err := parseCursor(afterCursor)
		if err != nil {
			var fieldErrors binding.FieldErrors
			fieldErrors.Add("after", "invalid", "无效的 after 游标: "+err.Error(), afterCursor)
			return nil, fieldErrors
		}
		req.Cursor = cursor
		req.CursorType = CursorTypeNext
		return req, nil
	}

	// 解析 before 游标
	beforeCursor := c.Query("before")
	if beforeCursor != "" {
		cursor, err := parseCursor(beforeCursor)
		if err != nil {
			var fieldErrors binding.FieldErrors
			fieldErrors.Add("before", "invalid", "无效的 before 游标: "+err.Error(), beforeCursor)
			return nil, fieldErrors
		}
		req.Cursor = cursor
		req.CursorType = CursorTypePrev
		return req, nil
	}

	// 没有游标，默认为向前查询
	return req, nil
}

// DefaultCursorPageRequest 返回默认的游标分页请求。
func DefaultCursorPageRequest[C any](parseCursor func(string) (C, error)) *CursorPageRequest[C] {
	return &CursorPageRequest[C]{
		Limit:           DefaultPageSize,
		CursorType:      CursorTypeNext,
		parseCursorFunc: parseCursor,
	}
}

// CursorPaginationResult 表示游标分页的结果。
type CursorPaginationResult[T any, C any] struct {
	// Items 是当前页的数据项列表。
	Items []T `json:"items"`

	// HasPrevPage 表示是否有上一页。
	HasPrevPage bool `json:"hasPrevPage"`

	// HasNextPage 表示是否有下一页。
	HasNextPage bool `json:"hasNextPage"`

	// StartCursor 是第一条数据的游标。
	StartCursor string `json:"startCursor,omitempty"`

	// EndCursor 是最后一条数据的游标。
	EndCursor string `json:"endCursor,omitempty"`
}

// NewCursorPaginationResult 创建一个新的游标分页结果。
// 参数：
//   - items: 当前页的数据项列表
//   - hasPrevPage: 是否有上一页
//   - hasNextPage: 是否有下一页
//   - getItemCursor: 函数，用于从数据项获取游标
//   - cursorToString: 函数，用于将游标转换为字符串
//
// 返回值：
//   - *CursorPaginationResult[T, C]: 游标分页结果
func NewCursorPaginationResult[T any, C any](
	items []T,
	hasPrevPage bool,
	hasNextPage bool,
	getItemCursor func(T) C,
	cursorToString func(C) string,
) *CursorPaginationResult[T, C] {
	result := &CursorPaginationResult[T, C]{
		Items:       items,
		HasPrevPage: hasPrevPage,
		HasNextPage: hasNextPage,
	}

	// 设置开始和结束游标
	if len(items) > 0 {
		firstItem := items[0]
		lastItem := items[len(items)-1]

		firstCursor := getItemCursor(firstItem)
		lastCursor := getItemCursor(lastItem)

		result.StartCursor = cursorToString(firstCursor)
		result.EndCursor = cursorToString(lastCursor)
	}

	return result
}

// ToPaginatedResponse 将游标分页结果转换为通用分页响应。
func (r *CursorPaginationResult[T, C]) ToPaginatedResponse() *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Items:   r.Items,
		HasMore: r.HasNextPage,
	}
}
