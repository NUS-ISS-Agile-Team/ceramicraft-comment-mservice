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
		ParentID:    0,
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
		ParentID:    0,
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
		ParentID:    0,
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

	list, err := svc.GetListByProductID(context.Background(), productID, userID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	ri := list[0]
	assert.Equal(t, cm.ID, ri.ID)
	assert.Equal(t, 2, ri.Likes)
	assert.False(t, ri.CurrentUserLiked)
}
