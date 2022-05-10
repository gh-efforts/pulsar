package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppDao interface {
	GetByUserId(ctx context.Context,
		userId string, AppType model.AppType) (appModel model.UserApp, err error)
	Create(ctx context.Context, userApp *model.UserApp) (err error)
	GetByAppId(ctx context.Context,
		appId string) (appModel model.UserApp, err error)
}
