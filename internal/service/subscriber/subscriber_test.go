package subscriber

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/filecoin-project/lotus/chain/vm"

	"github.com/bitrainforest/pulsar/internal/service/subscriber/actoraddress"

	model2 "github.com/bitrainforest/filmeta-hic/model"

	"github.com/bitrainforest/pulsar/internal/cache"

	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/bitrainforest/pulsar/internal/model"

	"github.com/stretchr/testify/assert"

	"github.com/nats-io/nats.go"
)

var _ dao.UserAppSubDao = &MockUserAppSubDao{}

type MockUserAppSubDao struct {
	appIds []string
}

func (m MockUserAppSubDao) FindByAddress(ctx context.Context, address string) (list []*model.UserAppSub, err error) {
	//TODO implement me
	panic("implement me")
}

func (m MockUserAppSubDao) FindByAddresses(ctx context.Context, address []string) (list []*model.SpecialUserAppSub, err error) {
	for i := range m.appIds {
		list = append(list, &model.SpecialUserAppSub{
			AppId: m.appIds[i],
		})
	}
	return list, nil
}

func (m MockUserAppSubDao) Create(ctx context.Context, appWatchModel *model.UserAppSub) (err error) {
	//TODO implement me
	panic("implement me")
}

func (m MockUserAppSubDao) GetByAppId(ctx context.Context, appId, address string) (appWatchModel model.UserAppSub, err error) {
	//TODO implement me
	panic("implement me")
}

func (m MockUserAppSubDao) Cancel(ctx context.Context, appId, address string) (err error) {
	//TODO implement me
	panic("implement me")
}

func GenerateAddress(item string) address.Address {
	a, _ := address.NewFromString(item) ////nolint:errcheck
	return a
}

var (
	_ cache.AddressMark = &MockAddressMark{}

	// three msgs
	DefaultTrace = types.ExecutionTrace{
		Msg: &types.Message{To: GenerateAddress("t0555"), From: GenerateAddress("t0666"), Method: 5, Nonce: 1},
		Subcalls: []types.ExecutionTrace{
			{
				Msg: &types.Message{To: GenerateAddress("t0111"), From: GenerateAddress("t0222"), Method: 5, Nonce: 1},
				Subcalls: []types.ExecutionTrace{
					{
						Msg: &types.Message{To: GenerateAddress("t0333"), From: GenerateAddress("t0444"), Method: 5, Nonce: 1},
					},
				},
			},
		},
	}
	initAppIds = []string{"wq", "wq2"}
)

type MockAddressMark struct {
	markMap map[string]struct{}
}

func NewMockMockAddressMark(m map[string]struct{}) *MockAddressMark {
	return &MockAddressMark{
		markMap: m,
	}
}

func (m *MockAddressMark) ExistAddress(ctx context.Context, address string) bool {
	if m.markMap == nil {
		return true
	}
	_, ok := m.markMap[address]
	return ok
}

func (m *MockAddressMark) MarkAddress(ctx context.Context, address string) bool {
	//TODO implement me
	panic("implement me")
}

func TestNewSub(t *testing.T) {
	type args struct {
		initAppIds    []string
		natsServerUri string
		optFns        []OptFn
	}
	type want struct {
		errIsNil   bool
		initAppIds []string
		poolNum    int64
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestNewSub1",
			args: args{
				initAppIds:    []string{"test1", "test2"},
				natsServerUri: nats.DefaultURL,
			},
			want: want{
				errIsNil:   false,
				initAppIds: []string{},
				poolNum:    DefaultWorkPoolNum,
			},
		},
		{
			name: "TestNewSubOpt",
			args: args{
				initAppIds:    []string{"test1", "test2"},
				natsServerUri: nats.DefaultURL,
				optFns: []OptFn{
					WithUserAppSubDao(&MockUserAppSubDao{}),
					WithAddressMarkCache(&MockAddressMark{}),
					WithWorkPoolNum(4000),
				},
			},
			want: want{
				errIsNil:   false,
				initAppIds: []string{},
				poolNum:    MaxWorkPoolNum,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notify := &MockNotify{}
			got, err := NewSub(tt.args.initAppIds, notify, tt.args.optFns...)
			assert.Equal(t, err != nil, tt.want.errIsNil, "NewSub() error = %v, wantErr %v", err, tt.want.errIsNil)
			for _, appid := range tt.want.initAppIds {
				_, ok := got.subAllAppIds.Load(appid)
				assert.Equal(t, ok, true, "appid %s not found", appid)
			}
			assert.Equal(t, got.opts.workPoolNum, tt.want.poolNum)
			if len(tt.args.optFns) > 0 {
				_, ok := got.opts.addressMarkCache.(*MockAddressMark)
				assert.Equal(t, ok, true)
				_, ok = got.opts.appSubDao.(*MockUserAppSubDao)
				assert.Equal(t, ok, true)
			}
		})
	}
}

func TestSubscriber_AppendAppId_AND_RemoveAppID(t *testing.T) {
	notify, err := NewNotify(nats.DefaultURL)
	assert.Equal(t, err, nil)
	sub, err := NewSub(initAppIds, notify, WithAddress(actoraddress.NewProxyActorAddress()))
	assert.Nil(t, err)

	addCount := rand.Intn(100) + 50
	removeCount := rand.Intn(30) + 10

	for i := 0; i < addCount; i++ {
		sub.AppendAppId(fmt.Sprintf("test%d", i))
	}

	var wg sync.WaitGroup

	for i := 0; i < removeCount; i++ {
		wg.Add(1)
		go func(item int) {
			defer wg.Done()
			sub.RemoveAppId(fmt.Sprintf("test%d", item))
		}(i)
	}

	wg.Wait()
	var (
		appIdCount int
	)
	sub.subAllAppIds.Range(func(key, value interface{}) bool {
		appIdCount++
		return true
	})

	assert.Equal(t, appIdCount, addCount-removeCount+len(initAppIds))

}

func TestSubscriber_NoNotify(t *testing.T) {
	m := &MockNotify{}
	sub, err := NewSub(initAppIds, m, WithAddress(actoraddress.NewProxyActorAddress()))
	assert.Nil(t, err)
	var (
		wg sync.WaitGroup
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		msg := &model2.Message{
			MCid:     RandCId(strconv.Itoa(i) + "NoNotify"),
			Msg:      &types.Message{To: builtin.ReserveAddress, From: builtin.CronActorAddr},
			Ret:      &vm.ApplyRet{ExecutionTrace: NewDefaultTrace()},
			Implicit: true, // if Implicit is true, then the message will not be sent to the initAppIds
		}
		go func() {
			defer wg.Done()
			err := sub.Notify(context.Background(), msg)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
	sub.Close()
	assert.Equal(t, 0, int(m.count))
}

func TestSubscriber_NotifyOnlyInitAppIds(t *testing.T) {
	m := &MockNotify{}
	appIds := initAppIds
	sub, err := NewSub(appIds, m, WithAddress(actoraddress.NewProxyActorAddress()))
	assert.Nil(t, err)
	var (
		wg sync.WaitGroup
	)
	// three msgs
	trace := DefaultTrace

	for i := 0; i < 10; i++ {
		wg.Add(1)
		msg := &model2.Message{
			MCid:     RandCId(strconv.Itoa(i) + "NotifyOnlyInitAppIds"),
			Msg:      &types.Message{To: builtin.ReserveAddress, From: builtin.CronActorAddr},
			Ret:      &vm.ApplyRet{ExecutionTrace: trace},
			Implicit: false,
		}
		go func() {
			defer wg.Done()
			err := sub.Notify(context.Background(), msg)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
	sub.Close()
	assert.Equal(t, int64(10*len(appIds)), m.count)
}

func TestSubscriber_NotifyOnlySpecialAppId(t *testing.T) {
	m := &MockNotify{}

	// mark t0111 address  that means address is subscribed by some appIds
	t111 := GenerateAddress("t0111")
	markAddressList := map[string]struct{}{
		t111.String(): {},
	}

	// init appIds is empty
	appIds := []string{}

	// three msgs
	traceMsg := DefaultTrace

	// append a message to traceMsg
	traceMsg.Subcalls = append(traceMsg.Subcalls, types.ExecutionTrace{
		Msg: &types.Message{To: GenerateAddress("t0111"), From: GenerateAddress("t0666"), Method: 5, Nonce: 1},
	})

	// "wq", "wq2",wq3 subscribe to t0111
	subAddressList := []string{"wq", "wq2", "wq3"}
	// two appIds subscribe address of ReserveAddress
	subDao := MockUserAppSubDao{appIds: subAddressList}

	sub, err := NewSub(appIds, m,
		WithAddress(actoraddress.NewProxyActorAddress()),
		WithUserAppSubDao(subDao),
		WithAddressMarkCache(NewMockMockAddressMark(markAddressList)),
		WithLockerExpire(0))
	assert.Nil(t, err)

	var (
		wg sync.WaitGroup
	)
	randCount := rand.Intn(10) + 10
	for i := 0; i < randCount; i++ {
		wg.Add(1)
		msg := &model2.Message{
			MCid:     RandCId(strconv.Itoa(i) + "NotifyOnlySpecialAppId"),
			Msg:      &types.Message{To: builtin.ReserveAddress, From: builtin.CronActorAddr},
			Ret:      &vm.ApplyRet{ExecutionTrace: traceMsg},
			Implicit: false,
		}
		go func() {
			defer wg.Done()
			err := sub.Notify(context.Background(), msg)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
	sub.Close()
	// msgs contain two msg for t0111
	subAddressCount := len(subAddressList) * randCount * 2
	assert.Equal(t, int64(subAddressCount), m.count, "rand count: "+strconv.Itoa(randCount))
}

func TestSubscriber_NotifyMixedMode(t *testing.T) {
	m := &MockNotify{}
	// mark t0111 address  that means address is subscribed by some appIds
	t111 := GenerateAddress("t0111")
	markAddressList := map[string]struct{}{
		t111.String(): {},
	}
	// init appIds is empty
	appIds := initAppIds
	appIds = append(appIds, "addOne")

	// three msgs
	traceMsg := DefaultTrace

	// "wq", "wq2",wq3 subscribe to t0111
	subAddressList := []string{"wq", "wq2", "wq3"}
	// two appIds subscribe address of ReserveAddress
	subDao := MockUserAppSubDao{appIds: subAddressList}

	sub, err := NewSub(appIds, m,
		WithAddress(actoraddress.NewProxyActorAddress()),
		WithUserAppSubDao(subDao),
		WithAddressMarkCache(NewMockMockAddressMark(markAddressList)),
		WithLockerExpire(0))
	assert.Nil(t, err)

	var (
		wg sync.WaitGroup
	)
	randCount := rand.Intn(10) + 10
	for i := 0; i < randCount; i++ {
		wg.Add(1)
		msg := &model2.Message{
			MCid:     RandCId(strconv.Itoa(i) + "NotifyMixedMode"),
			Msg:      &types.Message{To: GenerateAddress("t0555"), From: GenerateAddress("t0666"), Method: 5, Nonce: 1},
			Ret:      &vm.ApplyRet{ExecutionTrace: traceMsg},
			Implicit: false,
		}
		go func() {
			defer wg.Done()
			err := sub.Notify(context.Background(), msg)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()
	sub.Close()
	// msgs contain two msg for t0111
	initAppIds := randCount * len(appIds)
	subAddressCount := len(subAddressList) * randCount
	assert.Equal(t, int64(initAppIds+subAddressCount), m.count, "rand count: "+strconv.Itoa(randCount))
}

func BenchmarkNotify(b *testing.B) {
	notify, err := NewNotify(nats.DefaultURL)
	assert.Nil(b, err)

	t0123, err := address.NewFromString("t0123")
	assert.Nil(b, err)
	sub, err := NewSub([]string{"test1", "test2"},
		notify, WithUserAppSubDao(MockUserAppSubDao{appIds: []string{"test1", "test2"}}),
		WithAddressMarkCache(&MockAddressMark{}),
		WithAddress(actoraddress.NewProxyActorAddress()))

	assert.Nil(b, err)
	defer func() {
		sub.Close()
		err := recover()
		assert.Nil(b, err)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &model2.Message{Msg: &types.Message{
			To:   t0123,
			From: t0123,
		}}
		msg.MCid = RandCId(strconv.Itoa(i) + "Notify")
		err = sub.Notify(context.Background(), msg)
		assert.Nil(b, err)
	}
}
