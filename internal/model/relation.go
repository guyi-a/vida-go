package model

import "time"

// Relation 用户关注关系模型
type Relation struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:用户关系id" json:"id"`
	FollowID   int64     `gorm:"not null;uniqueIndex:idx_unique_follow_relation;index:idx_follow_id;comment:关注的用户id" json:"follow_id"`
	FollowerID int64     `gorm:"not null;uniqueIndex:idx_unique_follow_relation;index:idx_follower_id;comment:粉丝用户id" json:"follower_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index:idx_relations_created_at;comment:关注时间" json:"created_at"`
}

func (Relation) TableName() string {
	return "relations"
}
