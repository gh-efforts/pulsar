package subscriber

import (
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
		natsUri   string
		msgBuffer int64
		//userAppDao dao.UserAppDao
		appWatchDao dao.UserAppWatchDao
	}
)

func defaultOpts() Opts {
	return Opts{msgBuffer: DefaultMsgBuffer, natsUri: nats.DefaultURL}
}

func WithUserAppWatchDao(appWatch dao.UserAppWatchDao) OptFn {
	return func(opts *Opts) {
		if appWatch != nil {
			opts.appWatchDao = appWatch
		}
	}
}

func WithMsgBuffer(buffer int64) OptFn {
	return func(opts *Opts) {
		opts.msgBuffer = buffer
	}
}
