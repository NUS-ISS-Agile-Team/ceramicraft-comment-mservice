package mongo

import (
	"context"
	"sync"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/log"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentDao interface {
	Save(ctx context.Context, comment *model.Comment) error
	Get(ctx context.Context, id string) (*model.Comment, error)
}

var (
	commentDaoInstance CommentDao
	commentSyncOnce    sync.Once
)

func GetCommentDao() CommentDao {
	commentSyncOnce.Do(func() {
		commentDaoInstance = &CommentDaoImpl{
			collection: commentCollection,
		}
	})
	return commentDaoInstance
}

type CommentDaoImpl struct {
	collection *mongo.Collection
}

// Get implements CommentDao.
func (c *CommentDaoImpl) Get(ctx context.Context, id string) (*model.Comment, error) {
	var returnComment model.Comment
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Logger.Errorf("parse id failed.\terr=%v", err)
		return nil, err
	}
	err = c.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&returnComment)
	if err != nil {
		log.Logger.Errorf("failed to get comment by id %s: %v", id, err)
		return nil, err
	}
	return &returnComment, nil
}

// Save implements CommentDao.
func (c *CommentDaoImpl) Save(ctx context.Context, comment *model.Comment) error {
	ret, err := c.collection.InsertOne(ctx, comment)
	if err != nil {
		log.Logger.Errorf("failed to save comment: %v", err)
		return err
	}
	// 将 InsertedID 转换为字符串
	objectID, ok := ret.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Logger.Fatalf("InsertedID is not of type primitive.ObjectID")
	} else {
		comment.ID = objectID.Hex()
	}
	log.Logger.Infof("comment saved with id: %v", ret.InsertedID)
	return nil
}
