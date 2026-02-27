package repository

import (
	"vida-go/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID 根据 ID 查询用户（排除已删除）
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ? AND is_delete = 0", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByIDIncludeDeleted 根据 ID 查询用户（包含已删除，管理员用）
func (r *UserRepository) GetByIDIncludeDeleted(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名查询用户（排除已删除）
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("user_name = ? AND is_delete = 0", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户字段（传入 map，只更新非零值字段）
func (r *UserRepository) Update(id int64, updates map[string]interface{}) (*model.User, error) {
	result := r.db.Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetByIDIncludeDeleted(id)
}

// ExistsByUsername 检查用户名是否已存在
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("user_name = ? AND is_delete = 0", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListWithFilters 带筛选条件的分页查询
func (r *UserRepository) ListWithFilters(skip, limit int, username, userRole *string) ([]model.User, int64, error) {
	query := r.db.Model(&model.User{}).Where("is_delete = 0")

	if username != nil && *username != "" {
		query = query.Where("user_name ILIKE ?", "%"+*username+"%")
	}
	if userRole != nil && *userRole != "" {
		query = query.Where("user_role = ?", *userRole)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	if err := query.Offset(skip).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetByIDs 批量查询用户
func (r *UserRepository) GetByIDs(ids []int64) ([]model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []model.User
	err := r.db.Where("id IN ? AND is_delete = 0", ids).Find(&users).Error
	return users, err
}

// IncrementFollowCount 关注数 +1
func (r *UserRepository) IncrementFollowCount(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("follow_count", gorm.Expr("follow_count + 1")).Error
}

// DecrementFollowCount 关注数 -1（不低于 0）
func (r *UserRepository) DecrementFollowCount(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ? AND follow_count > 0", id).
		UpdateColumn("follow_count", gorm.Expr("follow_count - 1")).Error
}

// IncrementFollowerCount 粉丝数 +1
func (r *UserRepository) IncrementFollowerCount(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
}

// DecrementFollowerCount 粉丝数 -1（不低于 0）
func (r *UserRepository) DecrementFollowerCount(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ? AND follower_count > 0", id).
		UpdateColumn("follower_count", gorm.Expr("follower_count - 1")).Error
}

// IncrementTotalFavorited 获赞数 +1（视频作者被点赞总数）
func (r *UserRepository) IncrementTotalFavorited(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		UpdateColumn("total_favorited", gorm.Expr("total_favorited + 1")).Error
}

// DecrementTotalFavorited 获赞数 -1（不低于 0）
func (r *UserRepository) DecrementTotalFavorited(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ? AND total_favorited > 0", id).
		UpdateColumn("total_favorited", gorm.Expr("total_favorited - 1")).Error
}
