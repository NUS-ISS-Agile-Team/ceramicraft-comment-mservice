package repository

import (
	"github.com/NUS-ISS-Agile-Team/ceramicraft-comment-mservice/server/repository/dao/mongo"
)

func Init() {
	mongo.Init()
}
