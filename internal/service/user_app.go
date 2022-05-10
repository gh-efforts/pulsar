package service

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppService interface {
	CreateUserApp(ctx context.Context,
		userAppModel *model.UserApp) error
	GetAppByUserId(ctx context.Context,
		userId string, appType model.AppType) (model.UserApp, error)
	FindByAddress(ctx context.Context,
		address string) ([]*model.UserAppWatch, error)
	CreateAppWatch(ctx context.Context, appWatch model.UserAppWatch) error
	GetAppWatchByAppId(ctx context.Context,
		appId string, address string) (*model.UserAppWatch, error)
	GetAppByAppId(ctx context.Context,
		appId string) (model.UserApp, error)
}

type UserAppServiceImpl struct {
	userApp  dao.UserAppDao
	appWatch dao.UserAppWatchDao
}

func NewUserAppService(userApp dao.UserAppDao, appWatch dao.UserAppWatchDao) UserAppService {
	return UserAppServiceImpl{userApp: userApp, appWatch: appWatch}
}

func (userApp UserAppServiceImpl) CreateUserApp(ctx context.Context,
	userAppModel *model.UserApp) error {
	return userApp.userApp.Create(ctx, userAppModel)
}

func (userApp UserAppServiceImpl) GetAppByUserId(ctx context.Context,
	userId string, appType model.AppType) (model.UserApp, error) {
	return userApp.userApp.GetByUserId(ctx, userId, appType)
}

func (userApp UserAppServiceImpl) GetAppByAppId(ctx context.Context,
	appId string) (model.UserApp, error) {
	return userApp.userApp.GetByAppId(ctx, appId)
}

func (userApp UserAppServiceImpl) FindByAddress(ctx context.Context,
	address string) ([]*model.UserAppWatch, error) {
	return userApp.appWatch.FindByAddress(ctx, address)
}

func (userApp UserAppServiceImpl) CreateAppWatch(ctx context.Context, appWatch model.UserAppWatch) error {
	return userApp.appWatch.Create(ctx, &appWatch)
}

func (userApp UserAppServiceImpl) GetAppWatchByAppId(ctx context.Context,
	appId string, address string) (*model.UserAppWatch, error) {
	return userApp.appWatch.GetByAppId(ctx, appId, address)
}
