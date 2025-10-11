package service

import (
	"context"
	"time"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/mongo"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/model"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/types"
)

type ReviewService interface {
	CreateReview(ctx context.Context, req types.CreateReviewRequest, userID int) (err error)
}

type ReviewServiceImpl struct {
	reviewDao mongo.CommentDao
}

func GetReviewServiceInstance() *ReviewServiceImpl {
	return &ReviewServiceImpl{
		reviewDao: mongo.GetCommentDao(),
	}
}

func (r *ReviewServiceImpl) CreateReview(ctx context.Context, req types.CreateReviewRequest, userID int) (err error) {
	return r.reviewDao.Save(ctx, &model.Comment{
		Content:     req.Content,
		UserID:      userID,
		ProductID:   req.ProductID,
		ParentID:    req.ParentID,
		CreatedAt:   time.Now(),
		IsAnonymous: req.IsAnonymous,
		Stars:       req.Stars,
		PicInfo:     req.PicInfo,
	})
}
