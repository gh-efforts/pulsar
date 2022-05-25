package subscriber

import (
	"context"
	"fmt"

	"github.com/bitrainforest/pulsar/internal/utils/locker"
	"github.com/pkg/errors"

	"github.com/panjf2000/ants/v2"

	"github.com/bitrainforest/filmeta-hic/core/log"

	"github.com/bitrainforest/filmeta-hic/model"

	"sync"
)

var (
	// Sub when the service is started, it will be initialized.
	Sub *Subscriber
)

type (
	Subscriber struct {
		opts         *Opts
		subAllAppIds sync.Map
		workPool     *ants.Pool
		wg           sync.WaitGroup
		notify       Notify
	}
)

func NewSub(initAppIds []string, notify Notify, optFns ...OptFn) (*Subscriber, error) {
	if notify == nil {
		return nil, errors.New("notify is nil")
	}
	opts := defaultOpts()
	for _, opt := range optFns {
		opt(&opts)
	}
	if opts.workPoolNum <= 0 {
		opts.workPoolNum = DefaultWorkPoolNum
	}
	if opts.workPoolNum > MaxWorkPoolNum {
		opts.workPoolNum = MaxWorkPoolNum
	}

	if opts.lockerExpire > MaxLockExpire {
		opts.lockerExpire = MaxLockExpire
	}

	Sub = &Subscriber{
		subAllAppIds: sync.Map{},
		opts:         &opts,
		wg:           sync.WaitGroup{},
	}
	var (
		err error
	)
	if Sub.workPool, err = ants.NewPool(int(Sub.opts.workPoolNum)); err != nil {
		return nil, err
	}

	for _, appId := range initAppIds {
		Sub.AppendAppId(appId)
	}
	Sub.notify = notify
	return Sub, err
}

func (sub *Subscriber) AppendAppId(appId string) {
	sub.subAllAppIds.LoadOrStore(appId, true)
}

func (sub *Subscriber) RemoveAppId(appId string) {
	sub.subAllAppIds.Delete(appId)
}

func (sub *Subscriber) Notify(ctx context.Context, from, to string, msg *model.Message) error {
	sub.wg.Add(1)

	lockerCli := locker.NewRedisLock(ctx, msg.MCid.String(), sub.opts.lockerExpire)
	ok, err := lockerCli.Acquire(ctx)
	if err != nil {
		sub.wg.Done()
		return fmt.Errorf("[MessageApplied] Notify Mcid: %s failed: %v", msg.MCid.String(), err)
	}
	if !ok {
		sub.wg.Done()
		log.Infof("[MessageApplied] locked message %s", msg.MCid.String())
		return nil
	}

	// if lockerExpire=0, that means that when the message processing is complete,
	// the msg key is  deleted
	if sub.opts.lockerExpire == 0 {
		defer lockerCli.Release(ctx)
	}
	return sub.workPool.Submit(func() {
		appIds := sub.getSubsByAddress(ctx, from, to)
		if len(appIds) == 0 {
			sub.wg.Done()
			return
		}
		if err := sub.notify.Notify(&sub.wg, appIds, msg); err != nil {
			log.Errorf("notify error:%v", err)
		}
	})
}

func (sub *Subscriber) getSubsByAddress(ctx context.Context, from, to string) (allAppIds []string) {
	markCache := sub.opts.addressMarkCache
	if markCache.ExistAddress(ctx, to) ||
		markCache.ExistAddress(ctx, from) {
		list, err := sub.opts.appSubDao.FindByAddresses(context.Background(),
			[]string{from, to})
		if err != nil {
			log.Errorf("[core.processing]: find by addresses err: %s", err)
		}
		for _, item := range list {
			sub.AppendAppId(item.AppId)
		}
	}
	sub.subAllAppIds.Range(func(key, value interface{}) bool {
		allAppIds = append(allAppIds, key.(string))
		return true
	})
	return allAppIds
}

func (sub *Subscriber) Close() {
	sub.wg.Wait()
	// wait for all work msgDone
	sub.workPool.Release()
	sub.notify.Close()
}
