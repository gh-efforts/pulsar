package service

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/service/subscriber"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/bitrainforest/pulsar/internal/cache"

	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppService interface {
	CreateUserApp(ctx context.Context,
		userAppModel *model.UserApp) error
	GetAppByUserId(ctx context.Context,
		userId string, appType model.AppType) (model.UserApp, error)
	FindByAddress(ctx context.Context,
		address string) ([]*model.UserAppSub, error)
	AddSubAddress(ctx context.Context, appWatch model.UserAppSub) error
	GetAppWatchByAppId(ctx context.Context,
		appId string, address string) (model.UserAppSub, error)
	GetAppByAppId(ctx context.Context,
		appId string) (model.UserApp, error)
	CancelSubAddress(ctx context.Context, appId, address string) error
	GetWatchAllByAppId(ctx context.Context,
		appId string) (model.UserAppSubAll, error)
	CancelAll(ctx context.Context, appId string) error
	CreateSubAll(ctx context.Context,
		watchAll model.UserAppSubAll) error
}

type UserAppServiceImpl struct {
	userApp   dao.UserAppDao
	appSub    dao.UserAppSubDao
	appSubAll dao.UserAppSubAllDao
}

func NewUserAppService(userApp dao.UserAppDao,
	appWatch dao.UserAppSubDao, appSubAll dao.UserAppSubAllDao) UserAppService {
	return UserAppServiceImpl{userApp: userApp, appSub: appWatch, appSubAll: appSubAll}
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
	address string) ([]*model.UserAppSub, error) {
	return userApp.appSub.FindByAddress(ctx, address)
}

func (userApp UserAppServiceImpl) CreateSubAll(ctx context.Context,
	watchAll model.UserAppSubAll) error {
	subscriber.Sub.AppendAppId(watchAll.AppId)
	return userApp.appSubAll.Create(ctx, &watchAll)
}

func (userApp UserAppServiceImpl) CancelAll(ctx context.Context, appId string) error {
	subscriber.Sub.RemoveAppId(appId)
	return userApp.appSubAll.Cancel(ctx, appId)
}

func (userApp UserAppServiceImpl) GetWatchAllByAppId(ctx context.Context,
	appId string) (model.UserAppSubAll, error) {
	return userApp.appSubAll.GetByAppId(ctx, appId)
}

func (userApp UserAppServiceImpl) AddSubAddress(ctx context.Context, appWatch model.UserAppSub) error {
	if err := userApp.appSub.Create(ctx, &appWatch); err != nil {
		return err
	}
	markCache := cache.NewAddressMark(ctx)
	if !markCache.ExistAddress(ctx, appWatch.Address) {
		if !markCache.MarkAddress(ctx, appWatch.Address) {
			log.Warnf("[AddSubAddress] mark address %v err", appWatch)
		}
	}
	return nil
}

func (userApp UserAppServiceImpl) CancelSubAddress(ctx context.Context, appId, address string) error {
	return userApp.appSub.Cancel(ctx, appId, address)
}

func (userApp UserAppServiceImpl) GetAppWatchByAppId(ctx context.Context,
	appId string, address string) (model.UserAppSub, error) {
	return userApp.appSub.GetByAppId(ctx, appId, address)
}
