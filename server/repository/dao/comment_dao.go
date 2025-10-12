package dao

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/log"
	myMongo "github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/mongo"
	myRedis "github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/redis"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/model"
	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentDao interface {
	Save(ctx context.Context, comment *model.Comment) error
	Get(ctx context.Context, id string) (*model.Comment, error)
	HIncr(ctx context.Context, key string, member string, deta int) (err error)
	SAdd(ctx context.Context, key string, member string) (err error)
	GetListByUserID(ctx context.Context, userID int) (list []*model.Comment, err error)
	GetListByProductID(ctx context.Context, productId int) (list []*model.Comment, err error)
	HMGet(ctx context.Context, key string, members []string) (likesCntMap map[string]int, err error)
	SMembers(ctx context.Context, key string) (likedReviewIds []string, err error)
}

var (
	commentDaoInstance CommentDao
	commentSyncOnce    sync.Once
)

func GetCommentDao() CommentDao {
	commentSyncOnce.Do(func() {
		commentDaoInstance = &CommentDaoImpl{
			collection:  myMongo.CommentCollection,
			redisClient: myRedis.RedisClient,
		}
	})
	return commentDaoInstance
}

type CommentDaoImpl struct {
	collection  *mongo.Collection
	redisClient *redis.Client
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

func (c *CommentDaoImpl) HIncr(ctx context.Context, key string, member string, deta int) (err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil
	}
	cmd := c.redisClient.HIncrBy(ctx, key, member, int64(deta))
	if cmd.Err() != nil {
		log.Logger.Errorf("HIncr failed\tkey=%s\tmember=%s\tdeta=%d\terr=%v", key, member, deta, cmd.Err())
		return cmd.Err()
	}
	return nil
}

func (c *CommentDaoImpl) SAdd(ctx context.Context, key string, member string) (err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil
	}
	cmd := c.redisClient.SAdd(ctx, key, member)
	if cmd.Err() != nil {
		log.Logger.Errorf("SAdd failed\tkey=%s\tmember=%s\terr=%v", key, member, cmd.Err())
		return cmd.Err()
	}
	return nil
}

func (c *CommentDaoImpl) GetListByUserID(ctx context.Context, userID int) (list []*model.Comment, err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil, nil
	}
	cursor, err := c.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		log.Logger.Errorf("Find by user_id failed\tuser_id=%d\terr=%v", userID, err)
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()
	var results []*model.Comment
	for cursor.Next(ctx) {
		var cm model.Comment
		if err := cursor.Decode(&cm); err != nil {
			log.Logger.Errorf("Decode comment failed\terr=%v", err)
			return nil, err
		}
		log.Logger.Infof("cm = %+v\n", cm)
		results = append(results, &cm)
	}
	if err := cursor.Err(); err != nil {
		log.Logger.Errorf("cursor iteration error\terr=%v", err)
		return nil, err
	}
	return results, nil
}

func (c *CommentDaoImpl) GetListByProductID(ctx context.Context, productId int) (list []*model.Comment, err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil, nil
	}
	cursor, err := c.collection.Find(ctx, bson.M{"product_id": productId})
	if err != nil {
		log.Logger.Errorf("Find by product_id failed\tproduct_id=%d\terr=%v", productId, err)
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()
	var results []*model.Comment
	for cursor.Next(ctx) {
		var cm model.Comment
		if err := cursor.Decode(&cm); err != nil {
			log.Logger.Errorf("Decode comment failed\terr=%v", err)
			return nil, err
		}
		results = append(results, &cm)
	}
	if err := cursor.Err(); err != nil {
		log.Logger.Errorf("cursor iteration error\terr=%v", err)
		return nil, err
	}
	return results, nil
}

func (c *CommentDaoImpl) HMGet(ctx context.Context, key string, members []string) (likesCntMap map[string]int, err error) {
	likesCntMap = make(map[string]int, len(members))
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return likesCntMap, nil
	}
	if len(members) == 0 {
		return likesCntMap, nil
	}
	// redis HMGet accepts variadic keys
	vals, err := c.redisClient.HMGet(ctx, key, members...).Result()
	if err != nil {
		log.Logger.Errorf("HMGet failed\tkey=%s\tmembers=%v\terr=%v", key, members, err)
		return nil, err
	}
	for i, v := range vals {
		member := members[i]
		if v == nil {
			likesCntMap[member] = 0
			continue
		}
		// v can be string or []byte
		var s string
		switch t := v.(type) {
		case string:
			s = t
		case []byte:
			s = string(t)
		default:
			// fallback to sprint
			s = fmt.Sprint(v)
		}
		// parse int
		cnt, perr := strconv.Atoi(s)
		if perr != nil {
			log.Logger.Errorf("parse HMGet value failed\tkey=%s\tmember=%s\tvalue=%v\terr=%v", key, member, v, perr)
			likesCntMap[member] = 0
			continue
		}
		likesCntMap[member] = cnt
	}
	return likesCntMap, nil
}

func (c *CommentDaoImpl) SMembers(ctx context.Context, key string) (likedReviewIds []string, err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil, nil
	}
	vals, err := c.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		log.Logger.Errorf("SMembers failed\tkey=%s\terr=%v", key, err)
		return nil, err
	}
	return vals, nil
}
