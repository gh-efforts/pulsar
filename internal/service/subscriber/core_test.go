package subscriber

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"unsafe"

	"github.com/filecoin-project/go-address"

	"github.com/bitrainforest/filmeta-hic/model"

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
	// todo replace
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
		closed    bool
		sub       *Subscriber
		msgBuffer int64
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "TestNewCore1", args: args{sub: sub, opts: []CoreOpt{}}, want: want{
			closed:    false,
			sub:       sub,
			msgBuffer: DefaultMsgBuffer}},
		{name: "TestNewCore2", args: args{sub: sub, opts: []CoreOpt{WithMsgBuffer(100)}}, want: want{
			closed:    false,
			sub:       sub,
			msgBuffer: 100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCore(tt.args.sub, tt.args.opts...)
			assert.Equal(t, tt.want.closed, got.closed)
			assert.Equal(t, tt.want.sub, got.sub)
			assert.Equal(t, tt.want.msgBuffer, got.msgBuffer)
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
	core.Stop()
	id := cid.Undef
	err = core.MessageApplied(context.Background(), nil, id, nil, nil, false)
	assert.Equal(t, err, ErrClosed)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	core2, err := NewTestCore()
	assert.Nil(t, err)
	err = core2.MessageApplied(ctx, nil, id, nil, nil, false)
	assert.Equal(t, err, ErrCtxDone)

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

func NewTestCore(opts ...CoreOpt) (*Core, error) {
	sub, err := NewSub([]string{"test"}, &MockNotify{})
	if err != nil {
		return nil, err
	}
	core := NewCore(sub, opts...)
	return core, nil
}

func NewDefaultTrace() types.ExecutionTrace {
	t0123, _ := address.NewFromString("t0123") //nolint:errcheck
	return types.ExecutionTrace{
		Subcalls: []types.ExecutionTrace{
			{Subcalls: []types.ExecutionTrace{
				{Subcalls: nil, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
			}, Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1}},
		},
		Msg: &types.Message{To: t0123, From: t0123, Method: 5, Nonce: 1},
	}
}

func TestProcessing(t *testing.T) {
	mockNotify := &MockNotify{}
	initAppIds := []string{"test", "test2"}
	sub, err := NewSub(initAppIds, mockNotify, WithLockerExpire(0))
	assert.Nil(t, err)
	core := NewCore(sub)
	var (
		wg sync.WaitGroup
	)
	count := rand.Intn(20) + 30
	for i := 0; i < count; i++ {
		wg.Add(1)
		msg := &model.Message{
			MCid: RandCId(strconv.Itoa(i) + "Processing22"),
			Msg:  &types.Message{To: builtin.ReserveAddress, From: builtin.CronActorAddr},
			Ret:  &vm.ApplyRet{ExecutionTrace: NewDefaultTrace()},
		}
		go func() {
			defer wg.Done()
			err := core.processing(msg)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
	core.Stop()
	assert.Equal(t, int64(count*len(initAppIds)), mockNotify.count)

}

func BenchmarkMessageApplied(b *testing.B) {
	core, err := NewTestCore()
	assert.Nil(b, err)
	defer func() {
		err := recover()
		assert.Nil(b, err)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	msg := types.Message{
		Version: 1,
		To:      builtin.ReserveAddress,
		From:    builtin.RootVerifierAddress,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := core.MessageApplied(ctx, nil, cid.Undef, &msg, nil, false)
		assert.Nil(b, err)
	}
}

func RandCId(v string) cid.Cid {
	NewCid := cid.Undef
	val := reflect.ValueOf(&NewCid).Elem().FieldByName("str")
	val = reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem()
	nv := reflect.ValueOf(v)
	val.Set(nv)
	return NewCid
}

func BenchmarkProcessing(b *testing.B) {
	core, err := NewTestCore()
	assert.Nil(b, err)
	defer func() {
		err := recover()
		assert.Nil(b, err)
	}()
	msg := model.Message{
		Msg: &types.Message{
			To:   builtin.ReserveAddress,
			From: builtin.CronActorAddr,
		},
		TipSet: &types.TipSet{},
		Ret:    &vm.ApplyRet{ExecutionTrace: NewDefaultTrace()},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.MCid = RandCId(strconv.Itoa(i) + "processing")
		err := core.processing(&msg)
		assert.Nil(b, err)
	}
}
