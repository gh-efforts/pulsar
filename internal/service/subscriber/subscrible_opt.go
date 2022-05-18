package subscriber

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/cache"
	"github.com/bitrainforest/pulsar/internal/dao"
)

const (
	DefaultWorkPoolNum = 500
	MaxWorkPoolNum     = 1000
)

type (
	OptFn func(opts *Opts)
	Opts  struct {
		workPoolNum      int64
		appSubDao        dao.UserAppSubDao
		addressMarkCache cache.AddressMark
	}
)

func defaultOpts() Opts {
	return Opts{
		addressMarkCache: cache.NewAddressMark(context.Background()),
		appSubDao:        dao.NewUserAppSubDao(),
		workPoolNum:      DefaultWorkPoolNum,
	}
}

func WithUserAppSubDao(appSub dao.UserAppSubDao) OptFn {
	return func(opts *Opts) {
		if appSub != nil {
			opts.appSubDao = appSub
		}
	}
}

func WithAddressMarkCache(mark cache.AddressMark) OptFn {
	return func(opts *Opts) {
		if mark != nil {
			opts.addressMarkCache = mark
		}
	}
}

func WithWorkPoolNum(num int64) OptFn {
	return func(opts *Opts) {
		opts.workPoolNum = num
	}
}
