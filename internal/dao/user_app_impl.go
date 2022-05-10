package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/helper"
	"github.com/bitrainforest/pulsar/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserAppDaoImpl struct {
}

func NewUserAppDao() UserAppDao {
	return &UserAppDaoImpl{}
}

func (app UserAppDaoImpl) GetCollection() *mongo.Collection {
	return GetMongoDatabase().Collection("user_app")
}

func (app UserAppDaoImpl) GetByUserId(ctx context.Context,
	userId string, AppType model.AppType) (appModel model.UserApp, err error) {
	filter := bson.M{"user_id": userId, "app_type": AppType}
	result := app.GetCollection().FindOne(ctx, filter)
	err = helper.WarpMongoErr(result.Decode(&appModel))
	return
}

func (app UserAppDaoImpl) GetByAppId(ctx context.Context,
	appId string) (appModel model.UserApp, err error) {
	filter := bson.M{"app_id": appId}
	result := app.GetCollection().FindOne(ctx, filter)
	err = helper.WarpMongoErr(result.Decode(&appModel))
	return
}

func (app UserAppDaoImpl) Create(ctx context.Context, userApp *model.UserApp) (err error) {
	_, err = app.GetCollection().InsertOne(ctx, userApp)
	return
}
