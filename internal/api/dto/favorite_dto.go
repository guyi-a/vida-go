package dto

import "time"

// FavoriteInfo 点赞记录信息
type FavoriteInfo struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	VideoID   int64     `json:"video_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FavoriteListData 点赞列表数据
type FavoriteListData struct {
	Favorites  []FavoriteInfo `json:"favorites"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int64          `json:"total_pages"`
}

// BatchFavoriteStatusRequest 批量查询点赞状态请求
type BatchFavoriteStatusRequest struct {
	VideoIDs []int64 `json:"video_ids" binding:"required,min=1,max=100"`
}
