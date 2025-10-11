package api

import (
	"net/http"

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
