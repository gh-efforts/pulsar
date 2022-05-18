package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppSubDao interface {
	FindByAddress(ctx context.Context,
		address string) (list []*model.UserAppSub, err error)
	FindByAddresses(ctx context.Context,
		address []string) (list []*model.SpecialUserAppSub, err error)
	Create(ctx context.Context,
		appWatchModel *model.UserAppSub) (err error)
	GetByAppId(ctx context.Context,
		appId, address string) (appWatchModel model.UserAppSub, err error)
	Cancel(ctx context.Context,
		appId, address string) (err error)
}
