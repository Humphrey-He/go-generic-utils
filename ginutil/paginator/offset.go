package paginator

import (
	"errors"
	"fmt"
	"ggu/ginutil/binding"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	// ErrInvalidPageNum 表示无效的页码错误。
	ErrInvalidPageNum = errors.New("页码必须是大于0的整数")

	// ErrInvalidPageSize 表示无效的每页数据条数错误。
	ErrInvalidPageSize = errors.New("每页数据条数必须是大于0且不超过最大限制的整数")
)

// OffsetPageRequest 保存基于偏移量的分页参数。
type OffsetPageRequest struct {
	// PageNum 是当前页码，从 1 开始。
	PageNum int `form:"pageNum" json:"pageNum" binding:"omitempty,gte=1"`

	// PageSize 是每页数据条数。
	PageSize int `form:"pageSize" json:"pageSize" binding:"omitempty,gte=1,lte=100"`

	// Sort 是可选的排序字段，例如 "created_at desc, name asc"。
	Sort string `form:"sort" json:"sort,omitempty"`
}

// GetOffset 计算数据库查询的偏移量。
func (r *OffsetPageRequest) GetOffset() int {
	if r.PageNum <= 0 {
		r.PageNum = DefaultPageNum
	}
	if r.PageSize <= 0 {
		r.PageSize = DefaultPageSize
	}
	return (r.PageNum - 1) * r.PageSize
}

// GetLimit 返回每页数据条数，确保其在有效范围内。
func (r *OffsetPageRequest) GetLimit() int {
	if r.PageSize <= 0 {
		r.PageSize = DefaultPageSize
	}
	if r.PageSize > MaxPageSize {
		return MaxPageSize
	}
	return r.PageSize
}

// Validate 实现 Pageable 接口，验证分页参数是否有效。
func (r *OffsetPageRequest) Validate() error {
	if r.PageNum <= 0 {
		return ErrInvalidPageNum
	}
	if r.PageSize <= 0 || r.PageSize > MaxPageSize {
		return ErrInvalidPageSize
	}
	return nil
}

// ParseOffsetParams 从 gin.Context 解析偏移量分页参数。
// 如果参数无效，将使用默认值，并返回可能的错误。
func ParseOffsetParams(c *gin.Context) (*OffsetPageRequest, error) {
	if c == nil {
		return DefaultOffsetPageRequest(), nil
	}

	// 解析 pageNum
	pageNumStr := c.DefaultQuery("pageNum", strconv.Itoa(DefaultPageNum))
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum <= 0 {
		pageNum = DefaultPageNum
		if err != nil {
			return nil, fmt.Errorf("无效的页码: %w", err)
		}
		if pageNum <= 0 {
			return nil, ErrInvalidPageNum
		}
	}

	// 解析 pageSize
	pageSizeStr := c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize))
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > MaxPageSize {
		pageSize = DefaultPageSize
		if err != nil {
			return nil, fmt.Errorf("无效的每页数据条数: %w", err)
		}
		if pageSize <= 0 || pageSize > MaxPageSize {
			return nil, ErrInvalidPageSize
		}
	}

	// 解析排序参数
	sort := c.Query("sort")

	return &OffsetPageRequest{
		PageNum:  pageNum,
		PageSize: pageSize,
		Sort:     sort,
	}, nil
}

// ParseOffsetParamsWithBinding 使用 gin 的绑定功能解析偏移量分页参数。
// 返回详细的字段验证错误。
func ParseOffsetParamsWithBinding(c *gin.Context) (*OffsetPageRequest, binding.FieldErrors) {
	if c == nil {
		return DefaultOffsetPageRequest(), nil
	}

	var req OffsetPageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
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
	if req.PageNum <= 0 {
		req.PageNum = DefaultPageNum
	}

	if req.PageSize <= 0 {
		req.PageSize = DefaultPageSize
	} else if req.PageSize > MaxPageSize {
		req.PageSize = MaxPageSize
	}

	return &req, nil
}

// DefaultOffsetPageRequest 返回默认的偏移量分页请求。
func DefaultOffsetPageRequest() *OffsetPageRequest {
	return &OffsetPageRequest{
		PageNum:  DefaultPageNum,
		PageSize: DefaultPageSize,
	}
}

// GetLimitOffset 是一个便捷方法，从 gin.Context 解析分页参数并返回 limit 和 offset。
// 返回值：
//   - limit：每页数据条数
//   - offset：偏移量
//   - error：解析错误
func GetLimitOffset(c *gin.Context) (limit, offset int, err error) {
	req, err := ParseOffsetParams(c)
	if err != nil {
		return DefaultPageSize, 0, err
	}

	return req.GetLimit(), req.GetOffset(), nil
}

// GetLimitOffsetWithBinding 是一个便捷方法，使用绑定功能解析分页参数并返回 limit 和 offset。
// 返回值：
//   - limit：每页数据条数
//   - offset：偏移量
//   - errors：字段验证错误
func GetLimitOffsetWithBinding(c *gin.Context) (limit, offset int, errors binding.FieldErrors) {
	req, errs := ParseOffsetParamsWithBinding(c)
	if errs != nil && errs.HasErrors() {
		return DefaultPageSize, 0, errs
	}

	return req.GetLimit(), req.GetOffset(), nil
}
