package service

import (
	"errors"

	"vida-go/internal/api/dto"
	"vida-go/internal/model"
	"vida-go/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrCommentNotFound    = errors.New("评论不存在")
	ErrCommentNoPermission = errors.New("没有权限操作该评论")
	ErrParentNotFound     = errors.New("父评论不存在")
	ErrParentVideoMismatch = errors.New("父评论不属于该视频")
)

type CommentService struct {
	commentRepo *repository.CommentRepository
	videoRepo   *repository.VideoRepository
}

func NewCommentService(commentRepo *repository.CommentRepository, videoRepo *repository.VideoRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo, videoRepo: videoRepo}
}

// Create 发表评论
func (s *CommentService) Create(userID, videoID int64, req *dto.CommentCreateRequest) (*dto.CommentInfo, error) {
	if _, err := s.videoRepo.GetByID(videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	if req.ParentID != nil {
		parent, err := s.commentRepo.GetByID(*req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentNotFound
			}
			return nil, err
		}
		if parent.VideoID != videoID {
			return nil, ErrParentVideoMismatch
		}
	}

	comment := &model.Comment{
		UserID:   userID,
		VideoID:  videoID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	_ = s.videoRepo.IncrementCommentCount(videoID)

	return toCommentInfo(comment, 0), nil
}

// Update 更新评论
func (s *CommentService) Update(commentID, userID int64, req *dto.CommentUpdateRequest) (*dto.CommentInfo, error) {
	if err := s.commentRepo.Update(commentID, userID, req.Content); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNoPermission
		}
		return nil, err
	}

	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, err
	}

	return toCommentInfo(comment, 0), nil
}

// Delete 删除评论
func (s *CommentService) Delete(commentID, userID int64) (int64, error) {
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrCommentNotFound
		}
		return 0, err
	}

	videoID := comment.VideoID

	deleted, err := s.commentRepo.Delete(commentID, userID)
	if err != nil {
		return 0, err
	}
	if !deleted {
		return 0, ErrCommentNoPermission
	}

	_ = s.videoRepo.DecrementCommentCount(videoID)

	return videoID, nil
}

// ListByVideo 获取视频评论列表
func (s *CommentService) ListByVideo(videoID int64, parentID *int64, page, pageSize int) (*dto.CommentListData, error) {
	if _, err := s.videoRepo.GetByID(videoID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	skip := (page - 1) * pageSize
	comments, total, err := s.commentRepo.ListByVideo(videoID, parentID, skip, pageSize)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListData(comments, total, page, pageSize, false)
}

// ListReplies 获取评论的回复列表
func (s *CommentService) ListReplies(commentID int64, page, pageSize int) (*dto.CommentListData, error) {
	if _, err := s.commentRepo.GetByID(commentID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	skip := (page - 1) * pageSize
	comments, total, err := s.commentRepo.ListReplies(commentID, skip, pageSize)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListData(comments, total, page, pageSize, false)
}

// ListByUser 获取用户的评论列表
func (s *CommentService) ListByUser(userID int64, page, pageSize int) (*dto.CommentListData, error) {
	skip := (page - 1) * pageSize
	comments, total, err := s.commentRepo.ListByUser(userID, skip, pageSize)
	if err != nil {
		return nil, err
	}

	return s.buildCommentListData(comments, total, page, pageSize, true)
}

func (s *CommentService) buildCommentListData(comments []model.Comment, total int64, page, pageSize int, includeVideoTitle bool) (*dto.CommentListData, error) {
	items := make([]dto.CommentInfo, 0, len(comments))
	for i := range comments {
		repliesCount, _ := s.commentRepo.CountReplies(comments[i].ID)
		info := toCommentInfo(&comments[i], repliesCount)

		if comments[i].User.ID != 0 {
			info.Username = &comments[i].User.UserName
			info.Avatar = comments[i].User.Avatar
		}

		if includeVideoTitle && comments[i].Video.ID != 0 {
			info.VideoTitle = &comments[i].Video.Title
		}

		items = append(items, *info)
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return &dto.CommentListData{
		Comments:   items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func toCommentInfo(c *model.Comment, repliesCount int64) *dto.CommentInfo {
	return &dto.CommentInfo{
		ID:           c.ID,
		UserID:       c.UserID,
		VideoID:      c.VideoID,
		Content:      c.Content,
		ParentID:     c.ParentID,
		LikeCount:    c.LikeCount,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		RepliesCount: repliesCount,
	}
}
