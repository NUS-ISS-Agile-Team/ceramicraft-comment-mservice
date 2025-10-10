package model

import (
	"time"
)

type Comment struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Content   string    `bson:"content" json:"content"`
	UserID    int       `bson:"user_id" json:"user_id"`
	ProductID int       `bson:"product_id" json:"product_id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
