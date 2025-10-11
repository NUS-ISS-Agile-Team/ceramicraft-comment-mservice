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

	reviewID := 99
	reviewIdStr := strconv.Itoa(reviewID)
	userID := 77

	// Expect HIncr called first
	mockDao.EXPECT().HIncr(gomock.Any(), "review_likes", reviewIdStr, 1).Return(nil)
	// Then expect SAdd called
	mockDao.EXPECT().SAdd(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes", reviewIdStr).Return(nil)

	err := svc.Like(context.Background(), types.LikeRequest{ReviewID: reviewID}, userID)
	assert.NoError(t, err)
}

func TestLike_HIncrFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := 100
	reviewIdStr := strconv.Itoa(reviewID)
	userID := 77

	// HIncr returns error
	mockDao.EXPECT().HIncr(gomock.Any(), "review_likes", reviewIdStr, 1).Return(assert.AnError)

	err := svc.Like(context.Background(), types.LikeRequest{ReviewID: reviewID}, userID)
	assert.Error(t, err)
}

func TestLike_SAddFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := mocks.NewMockCommentDao(ctrl)
	svc := &ReviewServiceImpl{reviewDao: mockDao}

	reviewID := 101
	reviewIdStr := strconv.Itoa(reviewID)
	userID := 88

	// HIncr succeeds
	mockDao.EXPECT().HIncr(gomock.Any(), "review_likes", reviewIdStr, 1).Return(nil)
	// SAdd fails
	mockDao.EXPECT().SAdd(gomock.Any(), "user:"+strconv.Itoa(userID)+":likes", reviewIdStr).Return(assert.AnError)

	err := svc.Like(context.Background(), types.LikeRequest{ReviewID: reviewID}, userID)
	assert.Error(t, err)
}
