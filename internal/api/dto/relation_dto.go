package dto

// RelationUserInfo 关注关系中的用户简要信息
type RelationUserInfo struct {
	ID            int64   `json:"id"`
	Username      string  `json:"user_name"`
	Avatar        *string `json:"avatar"`
	FollowCount   int64   `json:"follow_count"`
	FollowerCount int64   `json:"follower_count"`
}

// FollowResult 关注/取关操作结果
type FollowResult struct {
	FollowerID    int64 `json:"follower_id"`
	FollowID      int64 `json:"follow_id"`
	FollowCount   int64 `json:"follow_count"`
	FollowerCount int64 `json:"follower_count"`
}

// RelationListData 关注/粉丝列表数据
type RelationListData struct {
	Users      []RelationUserInfo `json:"users"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int64              `json:"total_pages"`
}

// BatchFollowStatusRequest 批量查询关注状态请求
type BatchFollowStatusRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required,min=1,max=100"`
}
