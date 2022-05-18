package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bitrainforest/pulsar/internal/helper"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitrainforest/pulsar/internal/model"
	"go.mongodb.org/mongo-driver/bson"
)

type UserAppSubDaoImpl struct {
}

func NewUserAppSubDao() UserAppSubDao {
	return &UserAppSubDaoImpl{}
}

func (appWatch UserAppSubDaoImpl) GetCollection() *mongo.Collection {
	return GetMongoDatabase().Collection("user_app_sub")
}

func (appWatch UserAppSubDaoImpl) FindByAddress(ctx context.Context,
	address string) (list []*model.UserAppSub, err error) {
	filter := bson.M{"address": address}
	cur, err := appWatch.GetCollection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	for cur.Next(ctx) {
		var appWatch model.UserAppSub
		err := cur.Decode(&appWatch)
		if err != nil {
			return nil, err
		}
		list = append(list, &appWatch)
	}
	return
}

func (appWatch UserAppSubDaoImpl) FindByAddresses(ctx context.Context,
	address []string) (list []*model.SpecialUserAppSub, err error) {

	matchStage := bson.D{{Key: "$match", Value: bson.M{"address": bson.M{"$in": address}}}}
	groupStage := bson.D{{Key: "$group", Value: bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "app_id", Value: "$app_id"},
		}},
		{Key: "app_id", Value: bson.D{
			{Key: "$first", Value: "$app_id"},
		}},
	}},
	}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "app_id", Value: 1},
		}},
	}
	opts := options.Aggregate().SetMaxTime(15 * time.Second)
	cur, err := appWatch.GetCollection().Aggregate(ctx,
		mongo.Pipeline{matchStage, groupStage, projectStage}, opts)
	if err != nil {
		return nil, err
	}
	for cur.Next(ctx) {
		var appWatch model.SpecialUserAppSub
		err := cur.Decode(&appWatch)
		if err != nil {
			return nil, err
		}
		list = append(list, &appWatch)
	}
	return
}

func (appWatch UserAppSubDaoImpl) Create(ctx context.Context,
	appWatchModel *model.UserAppSub) (err error) {
	_, err = appWatch.GetCollection().InsertOne(ctx, appWatchModel)
	// if err is mongo duplicate key error, it means the user has already watched the app
	if err != nil && mongo.IsDuplicateKeyError(err) {
		err = nil
	}
	return
}

func (appWatch UserAppSubDaoImpl) Cancel(ctx context.Context,
	appId, address string) (err error) {
	filter := bson.M{"app_id": appId, "address": address}
	_, err = appWatch.GetCollection().DeleteOne(ctx, filter)
	err = helper.WarpMongoErr(err)
	return
}

func (appWatch UserAppSubDaoImpl) GetByAppId(ctx context.Context,
	appId, address string) (appWatchModel model.UserAppSub, err error) {
	filter := bson.M{"app_id": appId, "address": address}
	result := appWatch.GetCollection().FindOne(ctx, filter)
	err = helper.WarpMongoErr(result.Decode(&appWatchModel))
	return
}
