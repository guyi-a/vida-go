package model

import "time"

// Video 视频模型
type Video struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;comment:视频标识" json:"id"`
	AuthorID      int64      `gorm:"not null;index:idx_author_id;index:idx_composite_author_status;comment:视频作者ID" json:"author_id"`
	Title         string     `gorm:"size:200;not null;comment:视频标题" json:"title"`
	Description   string     `gorm:"type:text;comment:视频描述" json:"description"`
	PlayURL       string     `gorm:"size:500;comment:视频播放地址" json:"play_url"`
	CoverURL      string     `gorm:"size:500;comment:视频封面地址" json:"cover_url"`
	Duration      int        `gorm:"default:0;comment:视频时长（秒）" json:"duration"`
	FileSize      int64      `gorm:"default:0;comment:文件大小（字节）" json:"file_size"`
	FileFormat    string     `gorm:"size:20;comment:文件格式" json:"file_format"`
	Width         int        `gorm:"comment:视频宽度" json:"width"`
	Height        int        `gorm:"comment:视频高度" json:"height"`
	Status        string     `gorm:"size:20;default:'pending';index:idx_status;index:idx_composite_author_status;comment:视频状态" json:"status"`
	ViewCount     int64      `gorm:"default:0;comment:播放量" json:"view_count"`
	FavoriteCount int64      `gorm:"default:0;comment:点赞数" json:"favorite_count"`
	CommentCount  int64      `gorm:"default:0;comment:评论数" json:"comment_count"`
	PublishTime   *int64     `gorm:"index:idx_publish_time;comment:发布时间" json:"publish_time"`
	CreatedAt     time.Time  `gorm:"autoCreateTime;index:idx_videos_created_at;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`

	// 关联关系
	Author    User       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Favorites []Favorite `gorm:"foreignKey:VideoID" json:"favorites,omitempty"`
	Comments  []Comment  `gorm:"foreignKey:VideoID" json:"comments,omitempty"`
}

func (Video) TableName() string {
	return "videos"
}
