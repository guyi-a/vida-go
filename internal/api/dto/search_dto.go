package dto

// SearchVideoRequest 搜索请求参数
type SearchVideoRequest struct {
	Q         string `form:"q"`
	AuthorID  *int64 `form:"author_id"`
	VideoID   *int64 `form:"video_id"`
	Sort      string `form:"sort"` // relevance, time, hot
	StartTime *int64 `form:"start_time"`
	EndTime   *int64 `form:"end_time"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

// SearchVideoInfo 搜索结果中的视频信息
type SearchVideoInfo struct {
	ID            int64             `json:"id"`
	AuthorID      int64             `json:"author_id"`
	AuthorName    string             `json:"author_name"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	PlayURL       string             `json:"play_url"`
	CoverURL      string             `json:"cover_url"`
	ViewCount     int64              `json:"view_count"`
	FavoriteCount int64              `json:"favorite_count"`
	CommentCount  int64              `json:"comment_count"`
	PublishTime   *int64             `json:"publish_time"`
	Highlight     map[string][]string `json:"highlight,omitempty"`
}

// SearchVideoData 搜索结果
type SearchVideoData struct {
	Videos     []SearchVideoInfo `json:"videos"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int64             `json:"total_pages"`
}
