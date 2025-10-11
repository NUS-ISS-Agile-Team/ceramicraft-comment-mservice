package model

import (
	"time"
)

type Comment struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	Content     string    `bson:"content" json:"content"`
	UserID      int       `bson:"user_id" json:"user_id"`
	ProductID   int       `bson:"product_id" json:"product_id"`
	ParentID    int       `bson:"parent_id" json:"parent_id"`
	Stars       int       `bson:"stars" json:"stars"`
	IsAnonymous bool      `bson:"is_anonymous" json:"is_anonymous"`
	PicInfo     []string  `bson:"pic_info" json:"pic_info"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
}
