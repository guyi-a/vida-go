package repository

import (
	"vida-go/internal/model"

	"gorm.io/gorm"
)

type VideoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
	return &VideoRepository{db: db}
}

// GetByID 根据 ID 获取视频
func (r *VideoRepository) GetByID(id int64) (*model.Video, error) {
	var video model.Video
	err := r.db.Where("id = ? AND status != 'deleted'", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// GetByIDWithAuthor 根据 ID 获取视频（含作者信息）
func (r *VideoRepository) GetByIDWithAuthor(id int64) (*model.Video, error) {
	var video model.Video
	err := r.db.Preload("Author").Where("id = ? AND status != 'deleted'", id).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// GetByIDAndAuthor 根据视频 ID + 作者 ID 查询（权限校验用）
func (r *VideoRepository) GetByIDAndAuthor(videoID, authorID int64) (*model.Video, error) {
	var video model.Video
	err := r.db.Where("id = ? AND author_id = ? AND status != 'deleted'", videoID, authorID).First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// Create 创建视频记录
func (r *VideoRepository) Create(video *model.Video) error {
	return r.db.Create(video).Error
}

// Update 更新视频字段
func (r *VideoRepository) Update(id int64, updates map[string]interface{}) (*model.Video, error) {
	result := r.db.Model(&model.Video{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetByID(id)
}

// SoftDelete 软删除（设置 status = 'deleted'）
func (r *VideoRepository) SoftDelete(id int64) error {
	result := r.db.Model(&model.Video{}).Where("id = ? AND status != 'deleted'", id).
		Update("status", "deleted")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListVideos 视频列表查询（分页、筛选、排序）
func (r *VideoRepository) ListVideos(skip, limit int, authorID *int64, status *string, search *string, withAuthor bool) ([]model.Video, int64, error) {
	query := r.db.Model(&model.Video{}).Where("status != 'deleted'")

	if authorID != nil {
		query = query.Where("author_id = ?", *authorID)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
		if *status == "published" {
			query = query.Where("play_url IS NOT NULL AND play_url != ''")
		}
	}
	if search != nil && *search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+*search+"%", "%"+*search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	findQuery := query.Order("created_at DESC").Offset(skip).Limit(limit)
	if withAuthor {
		findQuery = findQuery.Preload("Author")
	}

	var videos []model.Video
	if err := findQuery.Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

// IncrementViewCount 观看数 +1
func (r *VideoRepository) IncrementViewCount(id int64) error {
	return r.db.Model(&model.Video{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// IncrementCommentCount 评论数 +1
func (r *VideoRepository) IncrementCommentCount(id int64) error {
	return r.db.Model(&model.Video{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// DecrementCommentCount 评论数 -1
func (r *VideoRepository) DecrementCommentCount(id int64) error {
	return r.db.Model(&model.Video{}).Where("id = ? AND comment_count > 0", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
}

// IncrementFavoriteCount 点赞数 +1
func (r *VideoRepository) IncrementFavoriteCount(id int64) error {
	return r.db.Model(&model.Video{}).Where("id = ?", id).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error
}

// DecrementFavoriteCount 点赞数 -1
func (r *VideoRepository) DecrementFavoriteCount(id int64) error {
	return r.db.Model(&model.Video{}).Where("id = ? AND favorite_count > 0", id).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count - 1")).Error
}
