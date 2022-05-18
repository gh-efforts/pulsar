package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/helper"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitrainforest/pulsar/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

type UserAppSubAllDaoImpl struct {
}

func NewUserAppSubAllDao() UserAppSubAllDao {
	return &UserAppSubAllDaoImpl{}
}

func (all UserAppSubAllDaoImpl) GetCollection() *mongo.Collection {
	return GetMongoDatabase().Collection("user_app_sub_all")
}

func (all UserAppSubAllDaoImpl) Create(ctx context.Context,
	watchAllModel *model.UserAppSubAll) (err error) {
	_, err = all.GetCollection().InsertOne(ctx, watchAllModel)
	// if err is mongo duplicate key error, it means the user has already watched the app
	if err != nil && mongo.IsDuplicateKeyError(err) {
		err = nil
	}
	return
}

func (all UserAppSubAllDaoImpl) ListByAllType(ctx context.Context,
	allType model.AllType) (watchAllList []*model.UserAppSubAll, err error) {
	filter := bson.M{"all_type": allType}
	cursor, err := all.GetCollection().Find(ctx, filter)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var watchAll model.UserAppSubAll
		err = cursor.Decode(&watchAll)
		if err != nil {
			return
		}
		watchAllList = append(watchAllList, &watchAll)
	}
	return
}

func (all UserAppSubAllDaoImpl) Cancel(ctx context.Context,
	appId string) (err error) {
	filter := bson.M{"app_id": appId}
	_, err = all.GetCollection().DeleteOne(ctx, filter)
	err = helper.WarpMongoErr(err)
	return
}

func (all UserAppSubAllDaoImpl) GetByAppId(ctx context.Context,
	appId string) (subAll model.UserAppSubAll, err error) {
	filter := bson.M{"app_id": appId}
	result := all.GetCollection().FindOne(ctx, filter)
	err = helper.WarpMongoErr(result.Decode(&subAll))
	return
}
