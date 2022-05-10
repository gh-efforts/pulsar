package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppWatchDao interface {
	FindByAddress(ctx context.Context,
		address string) (list []*model.UserAppWatch, err error)
	FindByAddresses(ctx context.Context,
		address []string) (list []*model.SpecialUserAppWatch, err error)
	Create(ctx context.Context,
		appWatchModel *model.UserAppWatch) (err error)
	GetByAppId(ctx context.Context,
		appId, address string) (appWatchModel model.UserAppWatch, err error)
	Cancel(ctx context.Context,
		appId, address string) (err error)
}
