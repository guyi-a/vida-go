package repository

import (
	"vida-go/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) GetByID(id int64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) GetByIDWithUser(id int64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Preload("User").First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// Update 更新评论（仅作者本人）
func (r *CommentRepository) Update(commentID, userID int64, content string) error {
	result := r.db.Model(&model.Comment{}).
		Where("id = ? AND user_id = ?", commentID, userID).
		Update("content", content)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除评论（仅作者本人）
func (r *CommentRepository) Delete(commentID, userID int64) (bool, error) {
	result := r.db.Where("id = ? AND user_id = ?", commentID, userID).Delete(&model.Comment{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ListByVideo 获取视频的评论列表（支持父评论筛选）
func (r *CommentRepository) ListByVideo(videoID int64, parentID *int64, skip, limit int) ([]model.Comment, int64, error) {
	query := r.db.Model(&model.Comment{}).Where("video_id = ?", videoID)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []model.Comment
	err := query.Preload("User").Order("created_at DESC").
		Offset(skip).Limit(limit).Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// ListReplies 获取某条评论的回复
func (r *CommentRepository) ListReplies(parentID int64, skip, limit int) ([]model.Comment, int64, error) {
	query := r.db.Model(&model.Comment{}).Where("parent_id = ?", parentID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []model.Comment
	err := query.Preload("User").Order("created_at ASC").
		Offset(skip).Limit(limit).Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// ListByUser 获取用户的评论列表
func (r *CommentRepository) ListByUser(userID int64, skip, limit int) ([]model.Comment, int64, error) {
	query := r.db.Model(&model.Comment{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []model.Comment
	err := query.Preload("Video").Order("created_at DESC").
		Offset(skip).Limit(limit).Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// CountReplies 统计某条评论的回复数
func (r *CommentRepository) CountReplies(commentID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Where("parent_id = ?", commentID).Count(&count).Error
	return count, err
}
