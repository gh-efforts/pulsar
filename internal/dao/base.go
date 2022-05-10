package dao

import (
	"github.com/bitrainforest/filmeta-hic/core/store"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetMongoDatabase() *mongo.Database {
	return store.GetMongoDB("v1_pulsar")
}
