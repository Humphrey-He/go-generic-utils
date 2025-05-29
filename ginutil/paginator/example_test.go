package paginator_test

import (
	"encoding/base64"
	"fmt"
	"ggu/ginutil/paginator"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 示例数据结构
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// 模拟数据库查询
func findUsers(limit, offset int) ([]User, int64) {
	// 模拟总数据量
	total := int64(100)

	// 模拟数据
	users := make([]User, 0, limit)
	for i := 0; i < limit && offset+i < int(total); i++ {
		users = append(users, User{
			ID:        offset + i + 1,
			Name:      fmt.Sprintf("User %d", offset+i+1),
			Email:     fmt.Sprintf("user%d@example.com", offset+i+1),
			CreatedAt: time.Now().Add(-time.Duration(offset+i) * time.Hour),
		})
	}

	return users, total
}

// 模拟基于游标的数据库查询
func findUsersWithCursor(limit int, afterID, beforeID int) ([]User, bool, bool) {
	// 模拟总数据量
	total := 100

	// 确定查询范围
	var start, end int
	var users []User

	if afterID > 0 {
		// 向后查询
		start = afterID
		end = afterID + limit
		if end > total {
			end = total
		}
	} else if beforeID > 0 {
		// 向前查询
		end = beforeID
		start = beforeID - limit
		if start < 1 {
			start = 1
		}
	} else {
		// 第一页
		start = 1
		end = limit + 1
		if end > total {
			end = total
		}
	}

	// 模拟数据
	for i := start; i < end; i++ {
		users = append(users, User{
			ID:        i,
			Name:      fmt.Sprintf("User %d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}

	// 判断是否有上一页和下一页
	hasPrevPage := start > 1
	hasNextPage := end < total

	return users, hasPrevPage, hasNextPage
}

// 示例：使用偏移量分页
func ExampleOffsetPagination() {
	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 定义获取用户列表的路由
	r.GET("/users", func(c *gin.Context) {
		// 解析分页参数
		limit, offset, err := paginator.GetLimitOffset(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查询数据
		users, total := findUsers(limit, offset)

		// 构建分页响应
		pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
		response := paginator.NewOffsetPaginatedResponse(users, pageNum, limit, total)

		// 返回响应
		c.JSON(http.StatusOK, response)
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/users?pageNum=2&pageSize=10", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 输出响应状态码
	fmt.Println("Status Code:", w.Code)

	// Output:
	// Status Code: 200
}

// 示例：使用游标分页
func ExampleCursorPagination() {
	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 解析游标函数
	parseCursor := func(cursor string) (int, error) {
		if cursor == "" {
			return 0, nil
		}

		// 解码 Base64
		decoded, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return 0, err
		}

		// 转换为整数
		id, err := strconv.Atoi(string(decoded))
		if err != nil {
			return 0, err
		}

		return id, nil
	}

	// 游标转字符串函数
	cursorToString := func(id int) string {
		if id == 0 {
			return ""
		}
		return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(id)))
	}

	// 从用户获取游标函数
	getItemCursor := func(user User) int {
		return user.ID
	}

	// 定义获取用户列表的路由
	r.GET("/users/cursor", func(c *gin.Context) {
		// 解析游标分页参数
		req, err := paginator.ParseCursorParams(c, parseCursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取游标值
		var afterID, beforeID int
		if req.HasCursor() {
			if req.IsForward() {
				afterID = req.Cursor
			} else {
				beforeID = req.Cursor
			}
		}

		// 查询数据
		users, hasPrevPage, hasNextPage := findUsersWithCursor(req.GetLimit(), afterID, beforeID)

		// 构建游标分页结果
		result := paginator.NewCursorPaginationResult(
			users,
			hasPrevPage,
			hasNextPage,
			getItemCursor,
			cursorToString,
		)

		// 返回响应
		c.JSON(http.StatusOK, result)
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/users/cursor?limit=10", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 输出响应状态码
	fmt.Println("Status Code:", w.Code)

	// Output:
	// Status Code: 200
}

// 示例：组合使用偏移量分页和游标分页
func ExampleCombinedPagination() {
	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 定义获取用户列表的路由（支持两种分页方式）
	r.GET("/users/combined", func(c *gin.Context) {
		// 检查是否有游标参数
		if c.Query("after") != "" || c.Query("before") != "" {
			// 使用游标分页
			parseCursor := func(cursor string) (int, error) {
				if cursor == "" {
					return 0, nil
				}
				decoded, err := base64.StdEncoding.DecodeString(cursor)
				if err != nil {
					return 0, err
				}
				id, err := strconv.Atoi(string(decoded))
				if err != nil {
					return 0, err
				}
				return id, nil
			}

			req, err := paginator.ParseCursorParams(c, parseCursor)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			var afterID, beforeID int
			if req.HasCursor() {
				if req.IsForward() {
					afterID = req.Cursor
				} else {
					beforeID = req.Cursor
				}
			}

			users, hasPrevPage, hasNextPage := findUsersWithCursor(req.GetLimit(), afterID, beforeID)

			cursorToString := func(id int) string {
				if id == 0 {
					return ""
				}
				return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(id)))
			}

			getItemCursor := func(user User) int {
				return user.ID
			}

			result := paginator.NewCursorPaginationResult(
				users,
				hasPrevPage,
				hasNextPage,
				getItemCursor,
				cursorToString,
			)

			c.JSON(http.StatusOK, result)
			return
		}

		// 使用偏移量分页
		limit, offset, err := paginator.GetLimitOffset(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		users, total := findUsers(limit, offset)

		pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
		response := paginator.NewOffsetPaginatedResponse(users, pageNum, limit, total)

		c.JSON(http.StatusOK, response)
	})

	// 创建请求
	req := httptest.NewRequest("GET", "/users/combined?pageNum=2&pageSize=10", nil)
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// 输出响应状态码
	fmt.Println("Status Code:", w.Code)

	// Output:
	// Status Code: 200
}
