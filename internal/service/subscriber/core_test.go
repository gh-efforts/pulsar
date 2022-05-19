package subscriber

import (
	"context"
	"sync"
	"testing"

	"github.com/filecoin-project/lotus/chain/actors/builtin"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/ipfs/go-cid"

	assert2 "github.com/bitrainforest/filmeta-hic/core/assert"
	"github.com/bitrainforest/filmeta-hic/core/config"
	"github.com/bitrainforest/filmeta-hic/core/log"
	"github.com/bitrainforest/filmeta-hic/core/store"
	mongox2 "github.com/bitrainforest/filmeta-hic/core/store/mongox"
	kconf "github.com/go-kratos/kratos/v2/config"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func init() {
	MustLoadTestEnv()
}

func TestNewCore(t *testing.T) {
	type args struct {
		sub  *Subscriber
		opts []CoreOpt
	}

	sub, err := NewSub([]string{"test"}, &MockNotify{})
	assert.Nil(t, err)

	type want struct {
		closed bool
		sub    *Subscriber
		chCap  int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "TestNewCore1", args: args{sub: sub, opts: []CoreOpt{}}, want: want{
			closed: false,
			sub:    sub,
			chCap:  DefaultWorkPoolNum}},
		{name: "TestNewCore2", args: args{sub: sub}, want: want{
			closed: false,
			sub:    sub,
			chCap:  DefaultWorkPoolNum}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCore(tt.args.sub, tt.args.opts...)
			assert.Equal(t, tt.want.closed, got.closed)
			assert.Equal(t, tt.want.sub, got.sub)
			assert.Equal(t, tt.want.chCap, got.sub.workPool.Cap())
		})
	}
}

func TestCore_MessageApplied(t *testing.T) {
	core, err := NewTestCore()
	assert.Nil(t, err)

	type args struct {
		ctx      context.Context
		ts       *types.TipSet
		mcid     cid.Cid
		msg      *types.Message
		ret      *vm.ApplyRet
		implicit bool
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg := types.Message{
		Version: 1,
		To:      builtin.ReserveAddress,
		From:    builtin.RootVerifierAddress,
	}
	tests := []struct {
		name     string
		args     args
		errIsNil bool
	}{
		{
			name:     "TestMessageApplied1",
			args:     args{msg: &msg, ctx: ctx},
			errIsNil: true,
		},
		{
			name:     "TestMessageApplied2",
			args:     args{msg: &msg, ctx: ctx},
			errIsNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := core.MessageApplied(tt.args.ctx, tt.args.ts, tt.args.mcid,
				tt.args.msg, tt.args.ret, tt.args.implicit)
			if tt.errIsNil {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
	assert.Equal(t, core.closed, false)
}

func TestCore_Stop(t *testing.T) {
	core, err := NewTestCore()
	assert.Nil(t, err)
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()

	var (
		wg sync.WaitGroup
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			core.Stop()
			assert.Equal(t, core.IsClosed(), true)
		}()
	}
	wg.Wait()
}

func MustLoadTestEnv() {
	serverName := "pulsar"
	// log
	log.SetUp(serverName, log.LevelInfo)
	var (
		conf kconf.Config
		err  error
	)
	if conf, err = config.LoadConfigAndInitData(serverName, func() (schema config.Schema, host string) {
		return config.Etcd, "etcd://127.0.0.1:2379"
	}); err != nil {
		assert2.CheckErr(err)
	}

	// must init mongo
	store.MustLoadMongoDB(conf, func(cfg kconf.Config) (*mongox2.Conf, error) {
		v := cfg.Value("data.mongo.uri")
		mongoUri, err := v.String()
		if err != nil {
			assert2.CheckErr(err)
		}
		if mongoUri == "" {
			assert2.CheckErr(errors.New("mongoUri must not be empty"))
		}
		return &mongox2.Conf{Uri: mongoUri}, nil
	})
	// mustLoadRedis
	store.MustLoadRedis(conf)
}

func NewTestCore() (*Core, error) {
	sub, err := NewSub([]string{"test"}, &MockNotify{})
	if err != nil {
		return nil, err
	}
	core := NewCore(sub)
	return core, nil
}
