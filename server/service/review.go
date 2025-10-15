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
	GetListByUserID(ctx context.Context, userID int) (list []types.ReviewInfo, err error)
	GetListByProductID(ctx context.Context, productId int, userID int) (resp types.ListReviewResponse, err error)
	PinReview(ctx context.Context, reviewID string) (err error)
	DeleteReview(ctx context.Context, reviewID string) (err error)
	GetListByQuery(ctx context.Context, req types.ListReviewRequest, userID int) (resp []types.ReviewInfo, err error)
}

const (
	reviewLikesCntKey = "review_likes"
	pinnedReviewKey   = "pinned_reviews"
)

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

func (r *ReviewServiceImpl) getReviewDetail(ctx context.Context, reviewID string, userID int) (detail types.ReviewInfo, err error) {
	reviewInfoRaw, err := r.reviewDao.Get(ctx, reviewID)
	if err != nil {
		return types.ReviewInfo{}, err
	}

	likesCntStr, err := r.reviewDao.HGet(ctx, reviewLikesCntKey, reviewID)
	if err != nil {
		return types.ReviewInfo{}, err
	}

	likesCnt, _ := strconv.Atoi(likesCntStr)

	userLikesReviewSetKey := fmt.Sprintf("user:%d:likes", userID)
	likedReviewList, err := r.reviewDao.SMembers(ctx, userLikesReviewSetKey)
	if err != nil {
		return types.ReviewInfo{}, err
	}

	curUserLiked := ExistInSlice(likedReviewList, reviewID)

	return types.ReviewInfo{
		ID:               reviewInfoRaw.ID,
		Content:          reviewInfoRaw.Content,
		ParentID:         reviewInfoRaw.ParentID,
		ProductID:        reviewInfoRaw.ProductID,
		UserID:           reviewInfoRaw.UserID,
		Stars:            reviewInfoRaw.Stars,
		IsAnonymous:      reviewInfoRaw.IsAnonymous,
		PicInfo:          reviewInfoRaw.PicInfo,
		CreatedAt:        reviewInfoRaw.CreatedAt,
		Likes:            likesCnt,
		CurrentUserLiked: curUserLiked,
	}, nil
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

func (r *ReviewServiceImpl) GetListByProductID(ctx context.Context, productId int, userID int) (resp types.ListReviewResponse, err error) {
	// 1. get review list
	listRaw, err := r.reviewDao.GetListByProductID(ctx, productId)
	if err != nil {
		return types.ListReviewResponse{}, err
	}

	list, err := r.buildReviewInfoList(ctx, listRaw, userID)
	if err != nil {
		return types.ListReviewResponse{}, err
	}

	// 2. get pinned review
	productIdStr := strconv.Itoa(productId)
	commentIdStr, err := r.reviewDao.HGet(ctx, pinnedReviewKey, productIdStr)
	if err != nil {
		return types.ListReviewResponse{}, err
	}

	// 3. build result
	var pinnedReviewDetail types.ReviewInfo
	if commentIdStr != "" {
		pinnedReviewDetail, err = r.getReviewDetail(ctx, commentIdStr, userID)
		if err != nil {
			return types.ListReviewResponse{}, err
		}
		return types.ListReviewResponse{
			ReviewList:   list,
			PinnedReview: &pinnedReviewDetail,
		}, nil
	}

	return types.ListReviewResponse{
		ReviewList:   list,
		PinnedReview: nil,
	}, nil
}

// GetListByQuery returns list filtered by product and stars (stars==0 means any)
func (r *ReviewServiceImpl) GetListByQuery(ctx context.Context, req types.ListReviewRequest, userID int) (resp []types.ReviewInfo, err error) {
	productId := req.ProductID
	stars := req.Stars
	listRaw, err := r.reviewDao.GetListByQuery(ctx, productId, stars)
	if err != nil {
		return nil, err
	}

	list, err := r.buildReviewInfoList(ctx, listRaw, userID)
	if err != nil {
		return nil, err
	}

	return list, nil
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

func (r *ReviewServiceImpl) PinReview(ctx context.Context, reviewID string) (err error) {
	commentRaw, err := r.reviewDao.Get(ctx, reviewID)
	if err != nil {
		return err
	}

	productIdStr := strconv.Itoa(commentRaw.ProductID)

	oldCommentID, err := r.reviewDao.HGet(ctx, pinnedReviewKey, productIdStr)
	if err != nil {
		return err
	}

	if oldCommentID != "" {
		err = r.reviewDao.UpdateIsPinnedByID(ctx, oldCommentID, false)
		if err != nil {
			return err
		}
	}

	err = r.reviewDao.UpdateIsPinnedByID(ctx, reviewID, true)
	if err != nil {
		return err
	}

	return r.reviewDao.HSet(ctx, pinnedReviewKey, productIdStr, reviewID)
}

func (r *ReviewServiceImpl) DeleteReview(ctx context.Context, reviewID string) (err error) {
	// get comment to know product id
	commentRaw, err := r.reviewDao.Get(ctx, reviewID)
	if err != nil {
		return err
	}

	// delete from mongo
	if err := r.reviewDao.Delete(ctx, reviewID); err != nil {
		return err
	}

	// remove likes hash field
	if err := r.reviewDao.HDel(ctx, reviewLikesCntKey, reviewID); err != nil {
		return err
	}

	// remove from any user's liked sets is optional (could be many) - skip

	// if this review was pinned for the product, delete pinned mapping
	productIdStr := strconv.Itoa(commentRaw.ProductID)
	pinnedId, err := r.reviewDao.HGet(ctx, pinnedReviewKey, productIdStr)
	if err != nil {
		return err
	}
	if pinnedId == reviewID {
		if err := r.reviewDao.HDel(ctx, pinnedReviewKey, productIdStr); err != nil {
			return err
		}
	}

	return nil
}
