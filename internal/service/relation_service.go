package service

import (
	"errors"

	"vida-go/internal/api/dto"
	"vida-go/internal/model"
	"vida-go/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrCannotFollowSelf = errors.New("不能关注自己")
	ErrAlreadyFollowed  = errors.New("您已经关注过该用户了")
	ErrNotFollowed      = errors.New("您尚未关注该用户")
)

type RelationService struct {
	relationRepo *repository.RelationRepository
	userRepo     *repository.UserRepository
}

func NewRelationService(relationRepo *repository.RelationRepository, userRepo *repository.UserRepository) *RelationService {
	return &RelationService{
		relationRepo: relationRepo,
		userRepo:     userRepo,
	}
}

// Follow 关注用户
func (s *RelationService) Follow(currentUserID, targetUserID int64) (*dto.FollowResult, error) {
	if currentUserID == targetUserID {
		return nil, ErrCannotFollowSelf
	}

	// 检查目标用户是否存在
	if _, err := s.userRepo.GetByID(targetUserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 检查是否已关注
	exists, err := s.relationRepo.Exists(currentUserID, targetUserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyFollowed
	}

	// 创建关注关系
	if _, err := s.relationRepo.Create(currentUserID, targetUserID); err != nil {
		return nil, err
	}

	// 更新计数
	_ = s.userRepo.IncrementFollowCount(currentUserID)
	_ = s.userRepo.IncrementFollowerCount(targetUserID)

	// 获取更新后的计数
	follower, _ := s.userRepo.GetByID(currentUserID)
	target, _ := s.userRepo.GetByID(targetUserID)

	result := &dto.FollowResult{
		FollowerID: currentUserID,
		FollowID:   targetUserID,
	}
	if follower != nil {
		result.FollowCount = follower.FollowCount
	}
	if target != nil {
		result.FollowerCount = target.FollowerCount
	}

	return result, nil
}

// Unfollow 取消关注
func (s *RelationService) Unfollow(currentUserID, targetUserID int64) (*dto.FollowResult, error) {
	deleted, err := s.relationRepo.Delete(currentUserID, targetUserID)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, ErrNotFollowed
	}

	// 更新计数
	_ = s.userRepo.DecrementFollowCount(currentUserID)
	_ = s.userRepo.DecrementFollowerCount(targetUserID)

	follower, _ := s.userRepo.GetByID(currentUserID)
	target, _ := s.userRepo.GetByID(targetUserID)

	result := &dto.FollowResult{
		FollowerID: currentUserID,
		FollowID:   targetUserID,
	}
	if follower != nil {
		result.FollowCount = follower.FollowCount
	}
	if target != nil {
		result.FollowerCount = target.FollowerCount
	}

	return result, nil
}

// GetFollowingList 获取关注列表
func (s *RelationService) GetFollowingList(userID int64, page, pageSize int) (*dto.RelationListData, error) {
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	skip := (page - 1) * pageSize
	followIDs, err := s.relationRepo.GetFollowingList(userID, skip, pageSize)
	if err != nil {
		return nil, err
	}

	total, err := s.relationRepo.CountFollowing(userID)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.GetByIDs(followIDs)
	if err != nil {
		return nil, err
	}

	return buildRelationListData(users, followIDs, total, page, pageSize), nil
}

// GetFollowerList 获取粉丝列表
func (s *RelationService) GetFollowerList(userID int64, page, pageSize int) (*dto.RelationListData, error) {
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	skip := (page - 1) * pageSize
	followerIDs, err := s.relationRepo.GetFollowerList(userID, skip, pageSize)
	if err != nil {
		return nil, err
	}

	total, err := s.relationRepo.CountFollowers(userID)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.GetByIDs(followerIDs)
	if err != nil {
		return nil, err
	}

	return buildRelationListData(users, followerIDs, total, page, pageSize), nil
}

// GetFollowStatus 查询关注状态
func (s *RelationService) GetFollowStatus(currentUserID, targetUserID int64) (bool, error) {
	return s.relationRepo.Exists(currentUserID, targetUserID)
}

// GetMutualFollows 获取互相关注列表
func (s *RelationService) GetMutualFollows(userID int64, page, pageSize int) (*dto.RelationListData, error) {
	skip := (page - 1) * pageSize
	mutualIDs, err := s.relationRepo.GetMutualFollowIDs(userID, skip, pageSize)
	if err != nil {
		return nil, err
	}

	total, err := s.relationRepo.CountMutualFollows(userID)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.GetByIDs(mutualIDs)
	if err != nil {
		return nil, err
	}

	return buildRelationListData(users, mutualIDs, total, page, pageSize), nil
}

// BatchCheckFollowStatus 批量查询关注状态
func (s *RelationService) BatchCheckFollowStatus(currentUserID int64, targetIDs []int64) (map[int64]bool, error) {
	return s.relationRepo.BatchCheckFollowing(currentUserID, targetIDs)
}

// buildRelationListData 构建关注/粉丝列表响应，按 orderedIDs 排序
func buildRelationListData(users []model.User, orderedIDs []int64, total int64, page, pageSize int) *dto.RelationListData {
	// 先建 map 方便按 ID 查找
	userMap := make(map[int64]dto.RelationUserInfo, len(users))
	for i := range users {
		userMap[users[i].ID] = dto.RelationUserInfo{
			ID:            users[i].ID,
			Username:      users[i].UserName,
			Avatar:        users[i].Avatar,
			FollowCount:   users[i].FollowCount,
			FollowerCount: users[i].FollowerCount,
		}
	}

	// 按原始顺序输出
	userList := make([]dto.RelationUserInfo, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		if info, ok := userMap[id]; ok {
			userList = append(userList, info)
		}
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.RelationListData{
		Users:      userList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
