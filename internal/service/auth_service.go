package service

import (
	"errors"

	"vida-go/internal/api/dto"
	"vida-go/internal/config"
	"vida-go/internal/model"
	"vida-go/internal/repository"
	"vida-go/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUsernameExists    = errors.New("用户名已存在")
	ErrInvalidCredential = errors.New("用户名或密码错误")
	ErrUserDeleted       = errors.New("该用户已被删除")
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Register 用户注册
func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.UserInfo, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	role := req.UserRole
	if role == "" {
		role = "user"
	}

	user := &model.User{
		UserName:        req.Username,
		Password:        hashedPassword,
		Avatar:          req.Avatar,
		BackgroundImage: req.BackgroundImage,
		UserRole:        role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return toUserInfo(user), nil
}

// Login 用户登录，返回 token 数据
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.TokenData, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredential
		}
		return nil, err
	}

	if user.IsDelete != 0 {
		return nil, ErrUserDeleted
	}

	if !utils.VerifyPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredential
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	expireSeconds := int(config.GetJWT().ExpireHours) * 3600

	return &dto.TokenData{
		Token:     token,
		TokenType: "bearer",
		ExpiresIn: expireSeconds,
		User:      *toUserInfo(user),
	}, nil
}

// GetCurrentUser 根据用户 ID 获取用户信息
func (s *AuthService) GetCurrentUser(userID int64) (*dto.UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.IsDelete != 0 {
		return nil, ErrUserDeleted
	}

	return toUserInfo(user), nil
}

func toUserInfo(user *model.User) *dto.UserInfo {
	return &dto.UserInfo{
		ID:              user.ID,
		Username:        user.UserName,
		Avatar:          user.Avatar,
		BackgroundImage: user.BackgroundImage,
		UserRole:        user.UserRole,
		FollowCount:     user.FollowCount,
		FollowerCount:   user.FollowerCount,
		TotalFavorited:  user.TotalFavorited,
	}
}
