package dto

import "time"

// VideoUploadRequest 视频上传请求（multipart/form-data）
type VideoUploadRequest struct {
	Title       string `form:"title" binding:"required,min=1,max=200"`
	Description string `form:"description" binding:"omitempty"`
}

// VideoUpdateRequest 视频更新请求
type VideoUpdateRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description"`
	Status      *string `json:"status" binding:"omitempty,oneof=pending processing published failed deleted"`
}

// AuthorBrief 视频中嵌套的作者简要信息
type AuthorBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
}

// VideoInfo 视频详情
type VideoInfo struct {
	ID            int64        `json:"id"`
	AuthorID      int64        `json:"author_id"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	PlayURL       string       `json:"play_url"`
	CoverURL      string       `json:"cover_url"`
	Duration      int          `json:"duration"`
	FileSize      int64        `json:"file_size"`
	FileFormat    string       `json:"file_format"`
	Width         int          `json:"width"`
	Height        int          `json:"height"`
	Status        string       `json:"status"`
	ViewCount     int64        `json:"view_count"`
	FavoriteCount int64        `json:"favorite_count"`
	CommentCount  int64        `json:"comment_count"`
	PublishTime   *int64       `json:"publish_time"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Author        *AuthorBrief `json:"author,omitempty"`
}

// VideoListData 视频列表响应数据
type VideoListData struct {
	Videos     []VideoInfo `json:"videos"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int64       `json:"total_pages"`
}
