package dto

import "time"

// CommentCreateRequest 发表评论请求
type CommentCreateRequest struct {
	Content  string `json:"content" binding:"required,min=1,max=1000"`
	ParentID *int64 `json:"parent_id"`
}

// CommentUpdateRequest 更新评论请求
type CommentUpdateRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// CommentInfo 评论信息
type CommentInfo struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	VideoID      int64     `json:"video_id"`
	Content      string    `json:"content"`
	ParentID     *int64    `json:"parent_id"`
	LikeCount    int64     `json:"like_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Username     *string   `json:"username"`
	Avatar       *string   `json:"avatar"`
	RepliesCount int64     `json:"replies_count"`
	VideoTitle   *string   `json:"video_title,omitempty"`
}

// CommentListData 评论列表数据
type CommentListData struct {
	Comments   []CommentInfo `json:"comments"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}
