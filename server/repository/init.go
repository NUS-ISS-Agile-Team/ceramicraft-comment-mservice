package repository

import (
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/mongo"
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/redis"
)

func Init() {
	mongo.Init()
	redis.Init()
}
