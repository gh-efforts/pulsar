package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/helper"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitrainforest/pulsar/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

type UserAppWatchDaoImpl struct {
}

func NewUserAppWatchDao() UserAppWatchDao {
	return &UserAppWatchDaoImpl{}
}

func (appWatch UserAppWatchDaoImpl) GetCollection() *mongo.Collection {
	return GetMongoDatabase().Collection("user_app_watch")
}

func (appWatch UserAppWatchDaoImpl) FindByAddress(ctx context.Context,
	address string) (list []*model.UserAppWatch, err error) {
	filter := bson.M{"address": address}
	cur, err := appWatch.GetCollection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(ctx) {
		var appWatch model.UserAppWatch
		err := cur.Decode(&appWatch)
		if err != nil {
			return nil, err
		}
		list = append(list, &appWatch)
	}
	return
}

func (appWatch UserAppWatchDaoImpl) Create(ctx context.Context,
	appWatchModel *model.UserAppWatch) (err error) {
	_, err = appWatch.GetCollection().InsertOne(ctx, appWatchModel)
	return
}

func (appWatch UserAppWatchDaoImpl) GetByAppId(ctx context.Context,
	appId, address string) (appWatchModel *model.UserAppWatch, err error) {
	filter := bson.M{"app_id": appId, "address": address}
	result := appWatch.GetCollection().FindOne(ctx, filter)
	err = helper.WarpMongoErr(result.Decode(&appWatchModel))
	return
}
