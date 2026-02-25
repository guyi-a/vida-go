package dto

// UserUpdateRequest 用户信息更新请求
type UserUpdateRequest struct {
	Username        *string `json:"username" binding:"omitempty,min=1,max=255"`
	Avatar          *string `json:"avatar" binding:"omitempty,max=500"`
	BackgroundImage *string `json:"background_image" binding:"omitempty,max=500"`
}

// UserFullInfo 用户完整公开信息（含收藏统计）
type UserFullInfo struct {
	ID              int64   `json:"id"`
	Username        string  `json:"user_name"`
	Avatar          *string `json:"avatar"`
	BackgroundImage *string `json:"background_image"`
	UserRole        string  `json:"user_role"`
	FollowCount     int64   `json:"follow_count"`
	FollowerCount   int64   `json:"follower_count"`
	TotalFavorited  int64   `json:"total_favorited"`
	FavoriteCount   int64   `json:"favorite_count"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// PaginatedData 带分页的数据
type PaginatedData struct {
	Items interface{}    `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}
