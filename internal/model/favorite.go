package model

import "time"

// Favorite 点赞/收藏模型
type Favorite struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:点赞记录ID" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:uq_user_video_favorite;index:idx_favorites_user_id;comment:点赞用户ID" json:"user_id"`
	VideoID   int64     `gorm:"not null;uniqueIndex:uq_user_video_favorite;index:idx_favorites_video_id;comment:被点赞视频ID" json:"video_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_favorites_created_at;comment:点赞时间" json:"created_at"`

	// 关联关系
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Video Video `gorm:"foreignKey:VideoID" json:"video,omitempty"`
}

func (Favorite) TableName() string {
	return "favorites"
}
