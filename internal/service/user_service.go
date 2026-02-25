package service

import (
	"errors"

	"vida-go/internal/api/dto"
	"vida-go/internal/model"
	"vida-go/internal/repository"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserByID 获取用户信息
func (s *UserService) GetUserByID(id int64) (*dto.UserFullInfo, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toUserFullInfo(user), nil
}

// UpdateUser 更新用户信息（本人或管理员）
func (s *UserService) UpdateUser(targetID int64, currentUser *dto.UserInfo, req *dto.UserUpdateRequest) (*dto.UserFullInfo, error) {
	if currentUser.ID != targetID && currentUser.UserRole != "admin" {
		return nil, errors.New("没有权限修改该用户信息")
	}

	updates := make(map[string]interface{})
	if req.Username != nil {
		exists, err := s.userRepo.ExistsByUsername(*req.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUsernameExists
		}
		updates["user_name"] = *req.Username
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.BackgroundImage != nil {
		updates["background_image"] = *req.BackgroundImage
	}

	if len(updates) == 0 {
		return s.GetUserByID(targetID)
	}

	user, err := s.userRepo.Update(targetID, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toUserFullInfo(user), nil
}

// SoftDeleteUser 软删除用户（管理员）
func (s *UserService) SoftDeleteUser(userID int64) error {
	_, err := s.userRepo.Update(userID, map[string]interface{}{"is_delete": 1})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// RestoreUser 恢复已删除用户（管理员）
func (s *UserService) RestoreUser(userID int64) error {
	_, err := s.userRepo.Update(userID, map[string]interface{}{"is_delete": 0})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// SetAdminRole 设置管理员角色（管理员）
func (s *UserService) SetAdminRole(userID int64) (*dto.UserFullInfo, error) {
	user, err := s.userRepo.Update(userID, map[string]interface{}{"user_role": "admin"})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return toUserFullInfo(user), nil
}

// ListUsers 获取用户列表（管理员，带筛选和分页）
func (s *UserService) ListUsers(page, pageSize int, username, userRole *string) (*dto.PaginatedData, error) {
	skip := (page - 1) * pageSize
	users, total, err := s.userRepo.ListWithFilters(skip, pageSize, username, userRole)
	if err != nil {
		return nil, err
	}

	items := make([]dto.UserFullInfo, 0, len(users))
	for i := range users {
		items = append(items, *toUserFullInfo(&users[i]))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.PaginatedData{
		Items: items,
		Meta: dto.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func toUserFullInfo(user *model.User) *dto.UserFullInfo {
	return &dto.UserFullInfo{
		ID:              user.ID,
		Username:        user.UserName,
		Avatar:          user.Avatar,
		BackgroundImage: user.BackgroundImage,
		UserRole:        user.UserRole,
		FollowCount:     user.FollowCount,
		FollowerCount:   user.FollowerCount,
		TotalFavorited:  user.TotalFavorited,
		FavoriteCount:   user.FavoriteCount,
	}
}
