package repository

import (
	"vida-go/internal/model"

	"gorm.io/gorm"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Create(userID, videoID int64) (*model.Favorite, error) {
	fav := &model.Favorite{UserID: userID, VideoID: videoID}
	if err := r.db.Create(fav).Error; err != nil {
		return nil, err
	}
	return fav, nil
}

func (r *FavoriteRepository) Delete(userID, videoID int64) (bool, error) {
	result := r.db.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&model.Favorite{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *FavoriteRepository) Exists(userID, videoID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).Count(&count).Error
	return count > 0, err
}

// ListByUser 获取用户的点赞列表
func (r *FavoriteRepository) ListByUser(userID int64, skip, limit int) ([]model.Favorite, int64, error) {
	query := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var favorites []model.Favorite
	err := query.Order("created_at DESC").Offset(skip).Limit(limit).Find(&favorites).Error
	if err != nil {
		return nil, 0, err
	}
	return favorites, total, nil
}

// ListByVideo 获取视频的点赞列表
func (r *FavoriteRepository) ListByVideo(videoID int64, skip, limit int) ([]model.Favorite, int64, error) {
	query := r.db.Model(&model.Favorite{}).Where("video_id = ?", videoID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var favorites []model.Favorite
	err := query.Order("created_at DESC").Offset(skip).Limit(limit).Find(&favorites).Error
	if err != nil {
		return nil, 0, err
	}
	return favorites, total, nil
}

// CountByVideo 统计视频的点赞数
func (r *FavoriteRepository) CountByVideo(videoID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("video_id = ?", videoID).Count(&count).Error
	return count, err
}

// BatchCheckFavorited 批量查询点赞状态
func (r *FavoriteRepository) BatchCheckFavorited(userID int64, videoIDs []int64) (map[int64]bool, error) {
	if len(videoIDs) == 0 {
		return map[int64]bool{}, nil
	}

	var favVideoIDs []int64
	err := r.db.Model(&model.Favorite{}).
		Where("user_id = ? AND video_id IN ?", userID, videoIDs).
		Pluck("video_id", &favVideoIDs).Error
	if err != nil {
		return nil, err
	}

	favSet := make(map[int64]bool, len(favVideoIDs))
	for _, id := range favVideoIDs {
		favSet[id] = true
	}

	result := make(map[int64]bool, len(videoIDs))
	for _, id := range videoIDs {
		result[id] = favSet[id]
	}
	return result, nil
}

// GetFavoritedVideoIDs 获取用户点赞的视频 ID 列表
func (r *FavoriteRepository) GetFavoritedVideoIDs(userID int64, skip, limit int) ([]int64, int64, error) {
	query := r.db.Model(&model.Favorite{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ids []int64
	err := query.Order("created_at DESC").Offset(skip).Limit(limit).Pluck("video_id", &ids).Error
	return ids, total, err
}
