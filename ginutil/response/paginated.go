package response

import (
	"ggu/ginutil/ecode"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PaginatedData 包含项目列表和分页元数据。
type PaginatedData[T any] struct {
	List       []T   `json:"list"`       // 当前页的数据列表
	Total      int64 `json:"total"`      // 总记录数
	PageNum    int   `json:"pageNum"`    // 当前页码
	PageSize   int   `json:"pageSize"`   // 每页大小
	TotalPages int   `json:"totalPages"` // 总页数
}

// NewPaginatedData 创建一个 PaginatedData 实例。
func NewPaginatedData[T any](list []T, totalCount int64, pageNum, pageSize int) PaginatedData[T] {
	// 参数校验和默认值设置
	if pageSize <= 0 {
		pageSize = 10 // 默认每页10条记录
	}

	// 计算总页数
	totalPages := int(totalCount / int64(pageSize))
	if totalCount%int64(pageSize) != 0 {
		totalPages++
	}

	// 特殊情况处理
	if totalPages == 0 && totalCount > 0 { // 例如：totalCount=5, pageSize=10
		totalPages = 1
	}

	// 确保页码至少为1
	if pageNum <= 0 {
		pageNum = 1
	}

	return PaginatedData[T]{
		List:       list,
		Total:      totalCount,
		PageNum:    pageNum,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// RespondPaginated 发送包含分页数据的成功响应。
func RespondPaginated[T any](c *gin.Context, list []T, totalCount int64, pageNum, pageSize int, customMessage ...string) {
	// 创建分页数据
	paginatedData := NewPaginatedData(list, totalCount, pageNum, pageSize)

	// 确定消息内容
	msg := ecode.DefaultSuccessMessage
	if len(customMessage) > 0 && customMessage[0] != "" {
		msg = customMessage[0]
	}

	// 构建并发送响应
	resp := StandardResponse[PaginatedData[T]]{
		Code:       ecode.OK,
		Message:    msg,
		Data:       paginatedData,
		TraceID:    c.GetString(GinTraceIDKey),
		ServerTime: getServerTime(),
	}
	sendJSON(c, http.StatusOK, resp)
}

// PageInfo 表示分页请求的基本信息
type PageInfo struct {
	PageNum  int `form:"pageNum" json:"pageNum"`   // 页码，从1开始
	PageSize int `form:"pageSize" json:"pageSize"` // 每页记录数
}

// GetPageInfo 从请求中提取分页信息，并应用默认值
func GetPageInfo(c *gin.Context) PageInfo {
	var pageInfo PageInfo

	// 尝试从查询参数绑定
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		// 使用默认值
		pageInfo.PageNum = 1
		pageInfo.PageSize = 10
	}

	// 确保有效的分页参数
	if pageInfo.PageNum <= 0 {
		pageInfo.PageNum = 1
	}

	if pageInfo.PageSize <= 0 {
		pageInfo.PageSize = 10
	} else if pageInfo.PageSize > 100 {
		// 限制最大页大小，防止请求过大数据量
		pageInfo.PageSize = 100
	}

	return pageInfo
}

// GetOffset 根据分页信息计算数据库查询的偏移量
func (p PageInfo) GetOffset() int {
	return (p.PageNum - 1) * p.PageSize
}

// GetLimit 获取查询限制（即每页大小）
func (p PageInfo) GetLimit() int {
	return p.PageSize
}
