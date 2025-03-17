package model

// PageResult 分页结果
type PageResult struct {
	List     interface{} `json:"list"`      // 数据列表
	Total    int64       `json:"total"`     // 总记录数
	Page     int         `json:"page"`      // 当前页码
	PageSize int         `json:"page_size"` // 每页记录数
	Pages    int         `json:"pages"`     // 总页数
}

// NewPageResult 创建分页结果
func NewPageResult(list interface{}, total int64, page, pageSize int) *PageResult {
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}
	return &PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}
}

// Option 选项结构
type Option struct {
	Label string      `json:"label"` // 选项标签
	Value interface{} `json:"value"` // 选项值
}

// TreeNode 树节点结构
type TreeNode struct {
	ID       interface{} `json:"id"`       // 节点ID
	Label    string      `json:"label"`    // 节点标签
	Value    interface{} `json:"value"`    // 节点值
	Children []*TreeNode `json:"children"` // 子节点
}

// EnumItem 枚举项结构
type EnumItem struct {
	Value int    `json:"value"` // 枚举值
	Label string `json:"label"` // 枚举标签
}

// LoginResult 登录结果
type LoginResult struct {
	Token        string `json:"token"`         // 访问令牌
	RefreshToken string `json:"refresh_token"` // 刷新令牌
	ExpiresIn    int    `json:"expires_in"`    // 过期时间(秒)
	TokenType    string `json:"token_type"`    // 令牌类型
}

// RefreshTokenResult 刷新令牌结果
type RefreshTokenResult struct {
	Token     string `json:"token"`      // 新的访问令牌
	ExpiresIn int    `json:"expires_in"` // 过期时间(秒)
}

// UploadResult 上传结果
type UploadResult struct {
	FileID       int64  `json:"file_id"`       // 文件ID
	OriginalName string `json:"original_name"` // 原始文件名
	FileName     string `json:"file_name"`     // 存储文件名
	FileExt      string `json:"file_ext"`      // 文件扩展名
	FileSize     int64  `json:"file_size"`     // 文件大小(字节)
	URL          string `json:"url"`           // 文件URL
}

// IDRequest ID请求
type IDRequest struct {
	ID int `json:"id" binding:"required" example:"1"` // ID
}

// IDsRequest IDs请求
type IDsRequest struct {
	IDs []int `json:"ids" binding:"required" example:"[1,2,3]"` // ID列表
}

// StatusRequest 状态请求
type StatusRequest struct {
	Status int8 `json:"status" binding:"required" example:"1"` // 状态值
}
