package dao

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/model"
)

type UserAppSubAllDao interface {
	Create(ctx context.Context,
		appWatchModel *model.UserAppSubAll) (err error)
	Cancel(ctx context.Context,
		appId string) (err error)
	GetByAppId(ctx context.Context,
		appId string) (subAll model.UserAppSubAll, err error)
	ListByAllType(ctx context.Context,
		allType model.AllType) (watchAllList []*model.UserAppSubAll, err error)
}
