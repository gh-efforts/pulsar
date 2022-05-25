package subscriber

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

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

var _ cache.AddressMark = &MockAddressMark{}

type MockAddressMark struct{}

func (m MockAddressMark) ExistAddress(ctx context.Context, address string) bool {
	return true
}

func (m MockAddressMark) MarkAddress(ctx context.Context, address string) bool {
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
	initAppIds := []string{"wq", "wq2"}
	notify, err := NewNotify(nats.DefaultURL)
	assert.Equal(t, err, nil)
	sub, err := NewSub(initAppIds, notify)
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

func TestSubscriber_Notify(t *testing.T) {
	type args struct {
		initAppIds []string
		userSubDao dao.UserAppSubDao
		markCache  cache.AddressMark
		round      int
		notify     *MockNotify
	}
	type want struct {
		initAppIds  []string
		notifyCount int64
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestSubscriber_Notify1",
			args: args{
				initAppIds: []string{"test1", "test2"},
				userSubDao: &MockUserAppSubDao{appIds: []string{"test1", "test2"}},
				markCache:  &MockAddressMark{},
				notify:     &MockNotify{},
				round:      150,
			},
			want: want{
				initAppIds:  []string{"test1", "test2"},
				notifyCount: 300,
			},
		},
		{
			name: "TestSubscriber_Notify2",
			args: args{
				initAppIds: []string{"test2", "test3"},
				userSubDao: &MockUserAppSubDao{appIds: []string{"test1", "test2"}},
				markCache:  &MockAddressMark{},
				notify:     &MockNotify{},
				round:      200,
			},
			want: want{
				initAppIds:  []string{"test2", "test3"},
				notifyCount: 600,
			},
		},

		{
			name: "TestSubscriber_Notify3",
			args: args{
				initAppIds: []string{"test4", "test5"},
				userSubDao: &MockUserAppSubDao{appIds: []string{"test6", "test7"}},
				markCache:  &MockAddressMark{},
				notify:     &MockNotify{},
				round:      100,
			},
			want: want{
				initAppIds:  []string{"test4", "test5"},
				notifyCount: 400,
			},
		},
	}

	var cidval int64 = 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := NewSub(tt.args.initAppIds, tt.args.notify,
				WithUserAppSubDao(tt.args.userSubDao),
				WithAddressMarkCache(tt.args.markCache),
			)
			assert.Nil(t, err)
			for i := 0; i < tt.args.round; i++ {
				atomic.AddInt64(&cidval, 1)
				msg := &model2.Message{}
				msg.MCid = RandCId(strconv.Itoa(int(cidval)) + "Notify")
				err = sub.Notify(context.Background(), fmt.Sprintf("test%d", i), "test", msg)
				assert.Nil(t, err)
			}
			sub.Close()
			assert.Equal(t, tt.args.initAppIds, tt.want.initAppIds)
			assert.Equal(t, tt.args.notify.count, tt.want.notifyCount)
		})
	}
}
