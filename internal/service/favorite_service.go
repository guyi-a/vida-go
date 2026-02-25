package service

import (
	"errors"

	"vida-go/internal/api/dto"
	"vida-go/internal/model"
	"vida-go/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrAlreadyFavorited = errors.New("您已经点赞过该视频了")
	ErrNotFavorited     = errors.New("您尚未点赞该视频")
)

type FavoriteService struct {
	favoriteRepo *repository.FavoriteRepository
	videoRepo    *repository.VideoRepository
}

func NewFavoriteService(favoriteRepo *repository.FavoriteRepository, videoRepo *repository.VideoRepository) *FavoriteService {
	return &FavoriteService{favoriteRepo: favoriteRepo, videoRepo: videoRepo}
}

// Favorite 点赞视频
func (s *FavoriteService) Favorite(userID, videoID int64) (*dto.FavoriteInfo, int64, error) {
	if _, err := s.videoRepo.GetByID(videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrVideoNotFound
		}
		return nil, 0, err
	}

	exists, err := s.favoriteRepo.Exists(userID, videoID)
	if err != nil {
		return nil, 0, err
	}
	if exists {
		return nil, 0, ErrAlreadyFavorited
	}

	fav, err := s.favoriteRepo.Create(userID, videoID)
	if err != nil {
		return nil, 0, err
	}

	_ = s.videoRepo.IncrementFavoriteCount(videoID)

	totalFav, _ := s.favoriteRepo.CountByVideo(videoID)

	return toFavoriteInfo(fav), totalFav, nil
}

// Unfavorite 取消点赞
func (s *FavoriteService) Unfavorite(userID, videoID int64) (int64, error) {
	deleted, err := s.favoriteRepo.Delete(userID, videoID)
	if err != nil {
		return 0, err
	}
	if !deleted {
		return 0, ErrNotFavorited
	}

	_ = s.videoRepo.DecrementFavoriteCount(videoID)

	totalFav, _ := s.favoriteRepo.CountByVideo(videoID)
	return totalFav, nil
}

// GetStatus 查询点赞状态
func (s *FavoriteService) GetStatus(userID, videoID int64) (bool, int64, error) {
	if _, err := s.videoRepo.GetByID(videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, ErrVideoNotFound
		}
		return false, 0, err
	}

	isFav, err := s.favoriteRepo.Exists(userID, videoID)
	if err != nil {
		return false, 0, err
	}

	total, _ := s.favoriteRepo.CountByVideo(videoID)
	return isFav, total, nil
}

// ListByUser 获取用户点赞列表
func (s *FavoriteService) ListByUser(userID int64, page, pageSize int) (*dto.FavoriteListData, error) {
	skip := (page - 1) * pageSize
	favorites, total, err := s.favoriteRepo.ListByUser(userID, skip, pageSize)
	if err != nil {
		return nil, err
	}
	return buildFavoriteListData(favorites, total, page, pageSize), nil
}

// ListByVideo 获取视频点赞列表
func (s *FavoriteService) ListByVideo(videoID int64, page, pageSize int) (*dto.FavoriteListData, error) {
	if _, err := s.videoRepo.GetByID(videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	skip := (page - 1) * pageSize
	favorites, total, err := s.favoriteRepo.ListByVideo(videoID, skip, pageSize)
	if err != nil {
		return nil, err
	}
	return buildFavoriteListData(favorites, total, page, pageSize), nil
}

// BatchCheckStatus 批量查询点赞状态
func (s *FavoriteService) BatchCheckStatus(userID int64, videoIDs []int64) (map[int64]bool, error) {
	return s.favoriteRepo.BatchCheckFavorited(userID, videoIDs)
}

// GetFavoritedVideoIDs 获取用户点赞的视频 ID 列表
func (s *FavoriteService) GetFavoritedVideoIDs(userID int64, page, pageSize int) ([]int64, int64, error) {
	skip := (page - 1) * pageSize
	return s.favoriteRepo.GetFavoritedVideoIDs(userID, skip, pageSize)
}

func toFavoriteInfo(f *model.Favorite) *dto.FavoriteInfo {
	return &dto.FavoriteInfo{
		ID:        f.ID,
		UserID:    f.UserID,
		VideoID:   f.VideoID,
		CreatedAt: f.CreatedAt,
	}
}

func buildFavoriteListData(favorites []model.Favorite, total int64, page, pageSize int) *dto.FavoriteListData {
	items := make([]dto.FavoriteInfo, 0, len(favorites))
	for i := range favorites {
		items = append(items, *toFavoriteInfo(&favorites[i]))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.FavoriteListData{
		Favorites:  items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
