package repository

import (
	"vida-go/internal/model"

	"gorm.io/gorm"
)

type RelationRepository struct {
	db *gorm.DB
}

func NewRelationRepository(db *gorm.DB) *RelationRepository {
	return &RelationRepository{db: db}
}

// Create 创建关注关系
func (r *RelationRepository) Create(followerID, followID int64) (*model.Relation, error) {
	relation := &model.Relation{
		FollowerID: followerID,
		FollowID:   followID,
	}
	if err := r.db.Create(relation).Error; err != nil {
		return nil, err
	}
	return relation, nil
}

// Delete 删除关注关系
func (r *RelationRepository) Delete(followerID, followID int64) (bool, error) {
	result := r.db.Where("follower_id = ? AND follow_id = ?", followerID, followID).
		Delete(&model.Relation{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// Exists 检查关注关系是否存在
func (r *RelationRepository) Exists(followerID, followID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Relation{}).
		Where("follower_id = ? AND follow_id = ?", followerID, followID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowingList 获取用户的关注列表（分页）
func (r *RelationRepository) GetFollowingList(userID int64, skip, limit int) ([]int64, error) {
	var followIDs []int64
	err := r.db.Model(&model.Relation{}).
		Where("follower_id = ?", userID).
		Order("created_at DESC").
		Offset(skip).Limit(limit).
		Pluck("follow_id", &followIDs).Error
	return followIDs, err
}

// GetFollowerList 获取用户的粉丝列表（分页）
func (r *RelationRepository) GetFollowerList(userID int64, skip, limit int) ([]int64, error) {
	var followerIDs []int64
	err := r.db.Model(&model.Relation{}).
		Where("follow_id = ?", userID).
		Order("created_at DESC").
		Offset(skip).Limit(limit).
		Pluck("follower_id", &followerIDs).Error
	return followerIDs, err
}

// CountFollowing 统计关注数
func (r *RelationRepository) CountFollowing(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Relation{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

// CountFollowers 统计粉丝数
func (r *RelationRepository) CountFollowers(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Relation{}).Where("follow_id = ?", userID).Count(&count).Error
	return count, err
}

// GetMutualFollowIDs 获取互相关注的用户 ID 列表（分页）
func (r *RelationRepository) GetMutualFollowIDs(userID int64, skip, limit int) ([]int64, error) {
	var mutualIDs []int64
	// 子查询：我关注的人 ∩ 关注我的人
	err := r.db.Raw(`
		SELECT r1.follow_id FROM relations r1
		INNER JOIN relations r2 ON r1.follow_id = r2.follower_id AND r2.follow_id = ?
		WHERE r1.follower_id = ?
		ORDER BY r1.created_at DESC
		OFFSET ? LIMIT ?
	`, userID, userID, skip, limit).Scan(&mutualIDs).Error
	return mutualIDs, err
}

// CountMutualFollows 统计互相关注数
func (r *RelationRepository) CountMutualFollows(userID int64) (int64, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(*) FROM relations r1
		INNER JOIN relations r2 ON r1.follow_id = r2.follower_id AND r2.follow_id = ?
		WHERE r1.follower_id = ?
	`, userID, userID).Scan(&count).Error
	return count, err
}

// BatchCheckFollowing 批量检查关注状态
func (r *RelationRepository) BatchCheckFollowing(followerID int64, followIDs []int64) (map[int64]bool, error) {
	if len(followIDs) == 0 {
		return map[int64]bool{}, nil
	}

	var followedIDs []int64
	err := r.db.Model(&model.Relation{}).
		Where("follower_id = ? AND follow_id IN ?", followerID, followIDs).
		Pluck("follow_id", &followedIDs).Error
	if err != nil {
		return nil, err
	}

	followedSet := make(map[int64]bool, len(followedIDs))
	for _, id := range followedIDs {
		followedSet[id] = true
	}

	result := make(map[int64]bool, len(followIDs))
	for _, id := range followIDs {
		result[id] = followedSet[id]
	}
	return result, nil
}
