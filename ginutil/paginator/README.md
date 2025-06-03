# Paginator 分页工具包

`paginator` 包提供了在 Gin 框架中处理分页请求和响应的工具函数和结构体，支持基于偏移量的分页和基于游标的分页两种模式。

## 主要特性

* **基于偏移量的分页**：支持传统的页码/每页数量 (pageNum, pageSize) 或偏移量/限制数量 (offset, limit) 的分页方式。
* **基于游标的分页**：支持使用游标 (cursor) 进行分页，适用于大数据集和实时数据流。
* **泛型支持**：使用 Go 1.18+ 泛型特性，提供类型安全的分页操作。
* **参数验证**：提供参数验证和错误处理，确保分页参数的有效性。
* **灵活的响应格式**：支持自定义分页响应格式，满足不同的前端需求。

## 安装

该包是 `ggu/ginutil` 项目的一部分，无需单独安装。

## 使用方法

### 基于偏移量的分页

```go
import (
    "ggu/ginutil/paginator"
    "github.com/gin-gonic/gin"
    "net/http"
)

// 处理基于偏移量的分页请求
func GetUserList(c *gin.Context) {
    // 方法一：简单解析分页参数
    limit, offset, err := paginator.GetLimitOffset(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 方法二：使用绑定功能，获取详细的验证错误
    limit, offset, errs := paginator.GetLimitOffsetWithBinding(c)
    if errs != nil && errs.HasErrors() {
        c.JSON(http.StatusBadRequest, gin.H{"errors": errs})
        return
    }
    
    // 查询数据库
    users, total := findUsers(limit, offset)
    
    // 构建分页响应
    pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
    response := paginator.NewOffsetPaginatedResponse(users, pageNum, limit, total)
    
    c.JSON(http.StatusOK, response)
}
```

### 基于游标的分页

```go
import (
    "ggu/ginutil/paginator"
    "github.com/gin-gonic/gin"
    "net/http"
    "encoding/base64"
    "strconv"
)

// 处理基于游标的分页请求
func GetUserListWithCursor(c *gin.Context) {
    // 定义游标解析函数
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
    
    // 查询数据库
    users, hasPrevPage, hasNextPage := findUsersWithCursor(req.GetLimit(), afterID, beforeID)
    
    // 定义游标转字符串函数
    cursorToString := func(id int) string {
        if id == 0 {
            return ""
        }
        return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(id)))
    }
    
    // 定义从数据项获取游标函数
    getItemCursor := func(user User) int {
        return user.ID
    }
    
    // 构建游标分页结果
    result := paginator.NewCursorPaginationResult(
        users,
        hasPrevPage,
        hasNextPage,
        getItemCursor,
        cursorToString,
    )
    
    c.JSON(http.StatusOK, result)
}
```

## 分页响应格式

### 基于偏移量的分页响应

```json
{
  "items": [
    {
      "id": 11,
      "name": "User 11",
      "email": "user11@example.com",
      "createdAt": "2024-06-01T10:00:00Z"
    },
    // ...更多数据项
  ],
  "total": 100,
  "hasMore": true,
  "pageInfo": {
    "pageNum": 2,
    "pageSize": 10,
    "pages": 10
  }
}
```

### 基于游标的分页响应

```json
{
  "items": [
    {
      "id": 11,
      "name": "User 11",
      "email": "user11@example.com",
      "createdAt": "2024-06-01T10:00:00Z"
    },
    // ...更多数据项
  ],
  "hasPrevPage": true,
  "hasNextPage": true,
  "startCursor": "MTE=",
  "endCursor": "MjA="
}
```

## 最佳实践

1. **选择合适的分页方式**：
   - 对于小型数据集和需要显示总页数的场景，使用基于偏移量的分页。
   - 对于大型数据集、实时数据流或需要高性能的场景，使用基于游标的分页。

2. **游标设计**：
   - 游标应该是唯一且稳定的，通常使用主键或创建时间戳。
   - 对于复合排序条件，可以将多个字段组合成一个游标字符串。
   - 游标通常应该进行 Base64 编码，以便在 URL 中安全传输。

3. **错误处理**：
   - 使用 `ParseOffsetParamsWithBinding` 和 `ParseCursorParamsWithBinding` 获取详细的验证错误，便于前端展示。
   - 对于无效的游标，返回明确的错误信息。

4. **安全性**：
   - 始终验证和限制 `pageSize` 或 `limit` 参数，防止请求过大的数据量。
   - 对于敏感数据，考虑对游标进行加密或签名。

5. **与数据库查询的集成**：
   - 基于偏移量的分页：使用 `LIMIT offset, limit` 子句。
   - 基于游标的分页：使用 `WHERE id > cursor ORDER BY id ASC LIMIT limit` 或 `WHERE id < cursor ORDER BY id DESC LIMIT limit` 子句。

## 示例

完整的示例代码可以在 `example_test.go` 文件中找到，包括：

- 基于偏移量的分页示例
- 基于游标的分页示例
- 组合使用两种分页方式的示例 