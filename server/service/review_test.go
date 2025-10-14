package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/log"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/mocks"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/model"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/types"
)

func init() {
	// 初始化测试用logger
	logger, _ := zap.NewDevelopment()
	log.Logger = logger.Sugar()
}

func TestCreateReview_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)

	svc := &ReviewServiceImpl{reviewDao: mockDao}

	req := types.CreateReviewRequest{
		ProductID:   42,
		Content:     "great",
		ParentID:    "0",
		Stars:       5,
		PicInfo:     []string{"a.jpg"},
		IsAnonymous: false,
	}
	userID := 123

	// Expect Save called with a Comment that has fields from req and userID
	mockDao.EXPECT().Save(gomock.Any(), gomock.AssignableToTypeOf(&model.Comment{})).DoAndReturn(
		func(ctx context.Context, c *model.Comment) error {
			assert.Equal(t, req.Content, c.Content)
			assert.Equal(t, userID, c.UserID)
			assert.Equal(t, req.ProductID, c.ProductID)
			assert.Equal(t, req.ParentID, c.ParentID)
			assert.Equal(t, req.Stars, c.Stars)
			assert.Equal(t, req.PicInfo, c.PicInfo)
			// CreatedAt should be set near now
			assert.WithinDuration(t, time.Now(), c.CreatedAt, time.Second*5)
			return nil
		})

	err := svc.CreateReview(context.Background(), req, userID)
	assert.NoError(t, err)
}

func TestLike_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "99"
	userID := 77

	// Expect HIncr called first
	mockDao.EXPECT().HIncr(gomock.Any(), "review_likes", reviewID, 1).Return(nil)
	// Then expect SAdd called
	mockDao.EXPECT().SAdd(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes", reviewID).Return(nil)

	err := svc.Like(context.Background(), types.LikeRequest{ReviewID: reviewID}, userID)
	assert.NoError(t, err)
}

func TestLike_HIncrFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "100"
	userID := 77

	// HIncr returns error
	mockDao.EXPECT().HIncr(gomock.Any(), "review_likes", reviewID, 1).Return(assert.AnError)

	err := svc.Like(context.Background(), types.LikeRequest{ReviewID: reviewID}, userID)
	assert.Error(t, err)
}

func TestLike_SAddFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "101"
	userID := 88

	// HIncr succeeds
	mockDao.EXPECT().HIncr(gomock.Any(), "review_likes", reviewID, 1).Return(nil)
	// SAdd fails
	mockDao.EXPECT().SAdd(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes", reviewID).Return(assert.AnError)

	err := svc.Like(context.Background(), types.LikeRequest{ReviewID: reviewID}, userID)
	assert.Error(t, err)
}

func TestGetListByUserID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	userID := 200
	// prepare one comment
	cm := &model.Comment{
		ID:          "c1",
		Content:     "nice",
		UserID:      userID,
		ProductID:   10,
		ParentID:    "0",
		Stars:       4,
		PicInfo:     []string{"p1.jpg"},
		IsAnonymous: false,
		CreatedAt:   time.Now(),
	}

	// Expect GetListByUserID
	mockDao.EXPECT().GetListByUserID(gomock.Any(), userID).Return([]*model.Comment{cm}, nil)

	// Expect HMGet called with members ["c1"] and return likes
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).DoAndReturn(
		func(ctx context.Context, key string, members []string) (map[string]int, error) {
			assert.Equal(t, []string{"c1"}, members)
			return map[string]int{"c1": 5}, nil
		})

	// Expect SMembers for current user's liked set
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{"c1"}, nil)

	list, err := svc.GetListByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	ri := list[0]
	assert.Equal(t, cm.ID, ri.ID)
	assert.Equal(t, cm.Content, ri.Content)
	assert.Equal(t, 5, ri.Likes)
	assert.True(t, ri.CurrentUserLiked)
}

func TestGetListByProductID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	userID := 300
	productID := 55
	cm := &model.Comment{
		ID:          "p1",
		Content:     "good product",
		UserID:      999,
		ProductID:   productID,
		ParentID:    "0",
		Stars:       5,
		PicInfo:     []string{},
		IsAnonymous: false,
		CreatedAt:   time.Now(),
	}

	mockDao.EXPECT().GetListByProductID(gomock.Any(), productID).Return([]*model.Comment{cm}, nil)

	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).DoAndReturn(
		func(ctx context.Context, key string, members []string) (map[string]int, error) {
			assert.Equal(t, []string{"p1"}, members)
			return map[string]int{"p1": 2}, nil
		})

	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{}, nil)

	// No pinned review for this product
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return("", nil)

	resp, err := svc.GetListByProductID(context.Background(), productID, userID)
	assert.NoError(t, err)
	assert.Len(t, resp.ReviewList, 1)
	ri := resp.ReviewList[0]
	assert.Equal(t, cm.ID, ri.ID)
	assert.Equal(t, 2, ri.Likes)
	assert.False(t, ri.CurrentUserLiked)
	assert.Nil(t, resp.PinnedReview)
}

func TestPinReview_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "r123"
	productID := 42

	// Expect Get to return comment with ProductID
	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	// Expect HSet to be called with pinnedReviewKey, productIdStr, reviewID
	mockDao.EXPECT().HSet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID), reviewID).Return(nil)

	err := svc.PinReview(context.Background(), reviewID)
	assert.NoError(t, err)
}

func TestDeleteReview_Success_Pinned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "del123"
	productID := 77

	// Get returns comment with ProductID
	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	// Delete from mongo
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(nil)
	// remove likes hash field
	mockDao.EXPECT().HDel(gomock.Any(), reviewLikesCntKey, reviewID).Return(nil)
	// pinned mapping returns this review id
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(reviewID, nil)
	// remove pinned mapping
	mockDao.EXPECT().HDel(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(nil)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.NoError(t, err)
}

func TestGetReviewDetail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "g1"
	userID := 999
	now := time.Now()

	// mock Get
	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{
		ID:          reviewID,
		Content:     "detail content",
		ParentID:    "0",
		ProductID:   11,
		UserID:      123,
		Stars:       4,
		IsAnonymous: false,
		PicInfo:     []string{"img1"},
		CreatedAt:   now,
	}, nil)

	// mock HGet for likes
	mockDao.EXPECT().HGet(gomock.Any(), reviewLikesCntKey, reviewID).Return("7", nil)

	// mock SMembers for current user liked set
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{reviewID}, nil)

	detail, err := svc.getReviewDetail(context.Background(), reviewID, userID)
	assert.NoError(t, err)
	assert.Equal(t, reviewID, detail.ID)
	assert.Equal(t, "detail content", detail.Content)
	assert.Equal(t, 7, detail.Likes)
	assert.True(t, detail.CurrentUserLiked)
	assert.WithinDuration(t, now, detail.CreatedAt, time.Second)
}

func TestGetListByProductID_WithPinned_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	userID := 50
	productID := 101
	cm := &model.Comment{
		ID:          "rA",
		Content:     "product review",
		UserID:      12,
		ProductID:   productID,
		ParentID:    "0",
		Stars:       5,
		PicInfo:     []string{},
		IsAnonymous: false,
		CreatedAt:   time.Now(),
	}

	// list
	mockDao.EXPECT().GetListByProductID(gomock.Any(), productID).Return([]*model.Comment{cm}, nil)
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).DoAndReturn(
		func(ctx context.Context, key string, members []string) (map[string]int, error) {
			assert.Equal(t, []string{"rA"}, members)
			return map[string]int{"rA": 4}, nil
		})
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{}, nil)

	// pinned mapping exists
	pinnedID := "pinned1"
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(pinnedID, nil)

	// getReviewDetail calls
	mockDao.EXPECT().Get(gomock.Any(), pinnedID).Return(&model.Comment{
		ID:        pinnedID,
		Content:   "pinned content",
		ProductID: productID,
		UserID:    99,
		CreatedAt: time.Now(),
	}, nil)
	mockDao.EXPECT().HGet(gomock.Any(), reviewLikesCntKey, pinnedID).Return("3", nil)
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{pinnedID}, nil)

	resp, err := svc.GetListByProductID(context.Background(), productID, userID)
	assert.NoError(t, err)
	assert.Len(t, resp.ReviewList, 1)
	assert.NotNil(t, resp.PinnedReview)
	assert.Equal(t, pinnedID, resp.PinnedReview.ID)
	assert.Equal(t, 3, resp.PinnedReview.Likes)
	assert.True(t, resp.PinnedReview.CurrentUserLiked)
}

func TestGetListByProductID_HMGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	productID := 202
	// make HMGet return error
	mockDao.EXPECT().GetListByProductID(gomock.Any(), productID).Return([]*model.Comment{}, nil)
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).Return(nil, assert.AnError)

	_, err := svc.GetListByProductID(context.Background(), productID, 0)
	assert.Error(t, err)
}

func TestGetReviewDetail_HGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "gx"
	userID := 11

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID}, nil)
	mockDao.EXPECT().HGet(gomock.Any(), reviewLikesCntKey, reviewID).Return("", assert.AnError)

	_, err := svc.getReviewDetail(context.Background(), reviewID, userID)
	assert.Error(t, err)
}

func TestGetReviewDetail_SMembersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "gy"
	userID := 12

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID}, nil)
	mockDao.EXPECT().HGet(gomock.Any(), reviewLikesCntKey, reviewID).Return("2", nil)
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return(nil, assert.AnError)

	_, err := svc.getReviewDetail(context.Background(), reviewID, userID)
	assert.Error(t, err)
}

func TestGetListByQuery_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	userID := 123
	req := types.ListReviewRequest{
		ProductID: 88,
		Stars:     5,
	}
	// prepare comments
	cm1 := &model.Comment{
		ID:          "c1",
		Content:     "good",
		UserID:      123,
		ProductID:   req.ProductID,
		ParentID:    "0",
		Stars:       req.Stars,
		PicInfo:     []string{"img1.jpg"},
		IsAnonymous: false,
		CreatedAt:   time.Now(),
	}
	cm2 := &model.Comment{
		ID:          "c2",
		Content:     "excellent",
		UserID:      456,
		ProductID:   req.ProductID,
		ParentID:    "0",
		Stars:       req.Stars,
		PicInfo:     []string{"img2.jpg"},
		IsAnonymous: true,
		CreatedAt:   time.Now().Add(-time.Hour),
	}

	// Expect DAO method called with correct params
	mockDao.EXPECT().GetListByQuery(gomock.Any(), req.ProductID, req.Stars).Return([]*model.Comment{cm1, cm2}, nil)
	// Expect HMGet called with both IDs
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).DoAndReturn(
		func(ctx context.Context, key string, members []string) (map[string]int, error) {
			assert.ElementsMatch(t, []string{"c1", "c2"}, members)
			return map[string]int{"c1": 10, "c2": 5}, nil
		})
	// Expect SMembers for current user's liked set
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{"c2"}, nil)

	list, err := svc.GetListByQuery(context.Background(), req, userID)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	// Check order: should be sorted by CreatedAt desc (cm1 newer)
	assert.Equal(t, "c1", list[0].ID)
	assert.Equal(t, "c2", list[1].ID)
	// Likes count
	assert.Equal(t, 10, list[0].Likes)
	assert.Equal(t, 5, list[1].Likes)
	// CurrentUserLiked
	assert.False(t, list[0].CurrentUserLiked)
	assert.True(t, list[1].CurrentUserLiked)
}

// DeleteReview tests: cover Get failure, Delete failure, HDel(likes) failure,
// pinned HGet error, pinned HDel failure, and non-pinned success.
func TestDeleteReview_GetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "dgetfail"
	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(nil, assert.AnError)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.Error(t, err)
}

func TestDeleteReview_DeleteFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "ddelfail"
	productID := 10

	// Get returns comment
	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	// Delete fails
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(assert.AnError)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.Error(t, err)
}

func TestDeleteReview_HDelLikesFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "dhdel"
	productID := 11

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(nil)
	// HDel for likes fails
	mockDao.EXPECT().HDel(gomock.Any(), reviewLikesCntKey, reviewID).Return(assert.AnError)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.Error(t, err)
}

func TestDeleteReview_PinnedHGetFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "dpinnederr"
	productID := 12

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(nil)
	mockDao.EXPECT().HDel(gomock.Any(), reviewLikesCntKey, reviewID).Return(nil)
	// HGet for pinned mapping fails
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return("", assert.AnError)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.Error(t, err)
}

func TestDeleteReview_PinnedHDelFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "dpinnedhdel"
	productID := 13

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(nil)
	mockDao.EXPECT().HDel(gomock.Any(), reviewLikesCntKey, reviewID).Return(nil)
	// pinned mapping matches and then HDel fails
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(reviewID, nil)
	mockDao.EXPECT().HDel(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(assert.AnError)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.Error(t, err)
}

func TestDeleteReview_NonPinned_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "dok"
	productID := 21

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(nil)
	mockDao.EXPECT().HDel(gomock.Any(), reviewLikesCntKey, reviewID).Return(nil)
	// pinned mapping returns different id
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return("otherid", nil)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.NoError(t, err)
}

func TestDeleteReview_Pinned_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := "dpinsuccess"
	productID := 31

	mockDao.EXPECT().Get(gomock.Any(), reviewID).Return(&model.Comment{ID: reviewID, ProductID: productID}, nil)
	mockDao.EXPECT().Delete(gomock.Any(), reviewID).Return(nil)
	mockDao.EXPECT().HDel(gomock.Any(), reviewLikesCntKey, reviewID).Return(nil)
	// pinned mapping matches and then remove it successfully
	mockDao.EXPECT().HGet(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(reviewID, nil)
	mockDao.EXPECT().HDel(gomock.Any(), pinnedReviewKey, strconv.Itoa(productID)).Return(nil)

	err := svc.DeleteReview(context.Background(), reviewID)
	assert.NoError(t, err)
}

func TestGetListByQuery_HMGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	req := types.ListReviewRequest{ProductID: 300, Stars: 4}

	mockDao.EXPECT().GetListByQuery(gomock.Any(), req.ProductID, req.Stars).Return([]*model.Comment{{ID: "a1"}}, nil)
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).Return(nil, assert.AnError)

	_, err := svc.GetListByQuery(context.Background(), req, 0)
	assert.Error(t, err)
}

func TestGetListByQuery_SMembersError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	req := types.ListReviewRequest{ProductID: 301, Stars: 5}

	mockDao.EXPECT().GetListByQuery(gomock.Any(), req.ProductID, req.Stars).Return([]*model.Comment{{ID: "b1"}}, nil)
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).Return(map[string]int{"b1": 1}, nil)
	mockDao.EXPECT().SMembers(gomock.Any(), gomock.AssignableToTypeOf("")).Return(nil, assert.AnError)

	_, err := svc.GetListByQuery(context.Background(), req, 0)
	assert.Error(t, err)
}

func TestGetListByQuery_DAOError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	req := types.ListReviewRequest{ProductID: 302, Stars: 3}

	mockDao.EXPECT().GetListByQuery(gomock.Any(), req.ProductID, req.Stars).Return(nil, assert.AnError)

	_, err := svc.GetListByQuery(context.Background(), req, 0)
	assert.Error(t, err)
}

func TestGetListByQuery_StarsZero_AllStars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	userID := 77
	req := types.ListReviewRequest{ProductID: 400, Stars: 0}

	cm1 := &model.Comment{ID: "s1", CreatedAt: time.Now()}
	cm2 := &model.Comment{ID: "s2", CreatedAt: time.Now().Add(-time.Minute)}

	mockDao.EXPECT().GetListByQuery(gomock.Any(), req.ProductID, req.Stars).Return([]*model.Comment{cm1, cm2}, nil)
	mockDao.EXPECT().HMGet(gomock.Any(), reviewLikesCntKey, gomock.AssignableToTypeOf([]string{})).Return(map[string]int{"s1": 2, "s2": 0}, nil)
	mockDao.EXPECT().SMembers(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes").Return([]string{"s1"}, nil)

	list, err := svc.GetListByQuery(context.Background(), req, userID)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "s1", list[0].ID)
	assert.Equal(t, 2, list[0].Likes)
	assert.True(t, list[0].CurrentUserLiked)
}
