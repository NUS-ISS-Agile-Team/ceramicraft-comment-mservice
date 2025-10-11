package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/log"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/model"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/types"
)

type ReviewService interface {
	CreateReview(ctx context.Context, req types.CreateReviewRequest, userID int) (err error)
	Like(ctx context.Context, req types.LikeRequest, userID int) (err error)
}

type ReviewServiceImpl struct {
	reviewDao dao.CommentDao
}

func GetReviewServiceInstance() *ReviewServiceImpl {
	return &ReviewServiceImpl{
		reviewDao: dao.GetCommentDao(),
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

func (r *ReviewServiceImpl) Like(ctx context.Context, req types.LikeRequest, userID int) (err error) {
	
	reviewLikesCntKey := "review_likes"
	reviewIdStr := strconv.Itoa(req.ReviewID)
	err = r.reviewDao.HIncr(ctx, reviewLikesCntKey, reviewIdStr, 1)
	if err != nil {
		log.Logger.Errorf("Like: failed, err %s", err.Error())
		return err
	}

	userLikesReviewSetKey := fmt.Sprintf("user:%d:likes", userID)
	err = r.reviewDao.SAdd(ctx, userLikesReviewSetKey, reviewIdStr)
	if err != nil {
		log.Logger.Errorf("Like: failed, err %s", err.Error())
		return err
	}
	return nil
}
