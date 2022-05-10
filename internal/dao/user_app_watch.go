package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppWatchDao interface {
	FindByAddress(ctx context.Context,
		address string) (list []*model.UserAppWatch, err error)
	Create(ctx context.Context,
		appWatchModel *model.UserAppWatch) (err error)
	GetByAppId(ctx context.Context,
		appId, address string) (appWatchModel *model.UserAppWatch, err error)
}
