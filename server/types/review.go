package types

type CreateReviewRequest struct {
	ProductID   int
	Content     string
	ParentID    int
	Stars       int
	PicInfo     []string `json:"pic_info"`
	IsAnonymous bool     `json:"is_anonymous"`
}

type LikeRequest struct {
	ReviewID int `json:"review_id"`
}
