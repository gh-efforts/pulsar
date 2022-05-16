package subscriber

import (
	"context"

	"github.com/bitrainforest/pulsar/internal/cache"
	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/nats-io/nats.go"
)

const (
	DefaultMsgBuffer = 300
	MaxMsgBuffer     = 1000
)

type (
	OptFn func(*Opts)

	Opts struct {
		natsUri          string
		msgBuffer        int64
		appWatchDao      dao.UserAppWatchDao
		addressMarkCache cache.AddressMark
	}
)

func defaultOpts() Opts {
	return Opts{
		msgBuffer: DefaultMsgBuffer, natsUri: nats.DefaultURL,
		addressMarkCache: cache.NewAddressMark(context.Background()),
	}
}

func WithUserAppWatchDao(appWatch dao.UserAppWatchDao) OptFn {
	return func(opts *Opts) {
		if appWatch != nil {
			opts.appWatchDao = appWatch
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

func WithMsgBuffer(buffer int64) OptFn {
	return func(opts *Opts) {
		opts.msgBuffer = buffer
	}
}
