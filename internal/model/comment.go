package model

import "time"

// Comment 评论模型
type Comment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;comment:评论ID" json:"id"`
	UserID    int64     `gorm:"not null;index:idx_comments_user_id;comment:评论用户ID" json:"user_id"`
	VideoID   int64     `gorm:"not null;index:idx_comments_video_id;index:idx_composite_video_created,priority:1;comment:被评论视频ID" json:"video_id"`
	Content   string    `gorm:"type:text;not null;comment:评论内容" json:"content"`
	ParentID  *int64    `gorm:"index:idx_comments_parent_id;comment:父评论ID" json:"parent_id"`
	LikeCount int64     `gorm:"default:0;comment:评论点赞数" json:"like_count"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_comments_created_at;index:idx_composite_video_created,priority:2;comment:评论时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`

	// 关联关系
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Video   Video     `gorm:"foreignKey:VideoID" json:"video,omitempty"`
	Parent  *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}
