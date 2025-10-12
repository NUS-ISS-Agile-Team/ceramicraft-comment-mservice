package service

import (
	"context"
	"fmt"
	"time"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/log"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/model"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/types"
)

type ReviewService interface {
	CreateReview(ctx context.Context, req types.CreateReviewRequest, userID int) (err error)
	Like(ctx context.Context, req types.LikeRequest, userID int) (err error)
	GetListByUserID(ctx context.Context, userID int) (list []types.ReviewInfo, err error)
	GetListByProductID(ctx context.Context, productId int, userID int) (list []types.ReviewInfo, err error)
}

const reviewLikesCntKey = "review_likes"

type ReviewServiceImpl struct {
	reviewDao dao.CommentDao
}

func GetReviewServiceInstance() *ReviewServiceImpl {
	return &ReviewServiceImpl{
		reviewDao: dao.GetCommentDao(),
	}
}

func (r *ReviewServiceImpl) GetListByUserID(ctx context.Context, userID int) (list []types.ReviewInfo, err error) {
	listRaw, err := r.reviewDao.GetListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return r.buildReviewInfoList(ctx, listRaw, userID)
}

func ExistInSlice(lists []string, tar string) bool {
	for _, i := range lists {
		if i == tar {
			return true
		}
	}
	return false
}

func (r *ReviewServiceImpl) buildReviewInfoList(ctx context.Context, listRaw []*model.Comment, userID int) (list []types.ReviewInfo, err error) {
	// get likes from redis
	// like count
	members := make([]string, len(listRaw))
	for idx, review := range listRaw {
		members[idx] = review.ID
	}

	likes, err := r.reviewDao.HMGet(ctx, reviewLikesCntKey, members)
	if err != nil {
		return nil, err
	}

	// current user liked
	userLikesReviewSetKey := fmt.Sprintf("user:%d:likes", userID)
	likedReviewList, err := r.reviewDao.SMembers(ctx, userLikesReviewSetKey)
	if err != nil {
		return nil, err
	}

	ans := make([]types.ReviewInfo, len(listRaw))
	for idx, review := range listRaw {
		curUserLiked := ExistInSlice(likedReviewList, review.ID)
		ans[idx] = types.ReviewInfo{
			ID:               review.ID,
			Content:          review.Content,
			UserID:           review.UserID,
			PicInfo:          review.PicInfo,
			ProductID:        review.ProductID,
			ParentID:         review.ParentID,
			Stars:            review.Stars,
			IsAnonymous:      review.IsAnonymous,
			CreatedAt:        review.CreatedAt,
			Likes:            likes[review.ID],
			CurrentUserLiked: curUserLiked,
		}
	}

	return ans, nil
}

func (r *ReviewServiceImpl) GetListByProductID(ctx context.Context, productId int, userID int) (list []types.ReviewInfo, err error) {
	listRaw, err := r.reviewDao.GetListByProductID(ctx, productId)
	if err != nil {
		return nil, err
	}

	return r.buildReviewInfoList(ctx, listRaw, userID)
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
	err = r.reviewDao.HIncr(ctx, reviewLikesCntKey, req.ReviewID, 1)
	if err != nil {
		log.Logger.Errorf("Like: failed, err %s", err.Error())
		return err
	}

	userLikesReviewSetKey := fmt.Sprintf("user:%d:likes", userID)
	err = r.reviewDao.SAdd(ctx, userLikesReviewSetKey, req.ReviewID)
	if err != nil {
		log.Logger.Errorf("Like: failed, err %s", err.Error())
		return err
	}
	return nil
}
