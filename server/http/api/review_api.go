package api

import (
	"net/http"
	"strconv"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/http/data"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/service"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/types"
	"github.com/gin-gonic/gin"
)

// CreateReview Create.
//
// @Summary Review Create
// @Description Create an review record.
// @Tags Review
// @Accept json
// @Produce json
// @Param user body types.CreateReviewRequest true "CreateReviewRequest"
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200	{object} data.BaseResponse{data=types.CreateReviewRequest}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/create [post]
func CreateReview(c *gin.Context) {
	var req types.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	userID := c.Value("userID").(int)
	err := service.GetReviewServiceInstance().CreateReview(c, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "create review success"))
}

// Like review
// @Summary Like a review
// @Description Like a review by id
// @Tags Review
// @Accept json
// @Produce json
// @Param user body types.LikeRequest true "LikeRequest"
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200 {object} data.BaseResponse{data=string}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/like [post]
func Like(c *gin.Context) {
	var req types.LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	userID := c.Value("userID").(int)
	err := service.GetReviewServiceInstance().Like(c, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "like success"))
}

// Get reviews by current user
// @Summary Get reviews by user
// @Description Get list of reviews created by current authenticated user
// @Tags Review
// @Accept json
// @Produce json
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200 {object} data.BaseResponse{data=[]types.ReviewInfo}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/list/user [get]
func GetListByUserID(c *gin.Context) {
	userID := c.Value("userID").(int)
	list, err := service.GetReviewServiceInstance().GetListByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, list))
}

// Get reviews by product id
// @Summary Get reviews by product
// @Description Get list of reviews for a product
// @Tags Review
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200 {object} data.BaseResponse{data=[]types.ReviewInfo}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/list/product/{product_id} [get]
func GetListByProductID(c *gin.Context) {
	pidStr := c.Param("product_id")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "invalid product_id"})
		return
	}
	userID := 0
	if v := c.Value("userID"); v != nil {
		userID = v.(int)
	}
	list, err := service.GetReviewServiceInstance().GetListByProductID(c, pid, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, list))
}
