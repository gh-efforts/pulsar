package subscriber

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/bitrainforest/filmeta-hic/core/fx"

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

func (sub *Subscriber) GetActorAddress(ctx context.Context, from,
	to address.Address, tip *types.TipSet) (address.Address, address.Address) {
	var (
		err error
	)
	from, err = sub.opts.actorAddress.GetActorAddress(ctx, tip, from)
	if err != nil {
		// just to log getActorAddress error
		log.Errorf("[processing] from address:%v,err:%v", from, err.Error())
	}
	to, err = sub.opts.actorAddress.GetActorAddress(ctx, tip, to)
	if err != nil {
		// just to log getActorAddress error
		log.Errorf("[processing] to address:%v,err:%v", to, err.Error())
	}
	return from, to
}

func (sub *Subscriber) Notify(ctx context.Context, msg *model.Message) error {
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
		defer sub.wg.Done()
		// step1 notify subscribers who have  sub to  all appIds
		if !msg.IsImplicit() {
			err := sub.notify.Notify(sub.GetAppIdsSubAll(), msg)
			if err != nil {
				log.Errorf("[MessageApplied] Notify %s failed: %v", msg.MCid.String(), err)
			}
		}

		fx.From(func(source chan<- interface{}) {
			if msg.Ret == nil {
				log.Warnf("[MessageApplied] message %s ret is nil", msg.MCid.String())
				return
			}
			for item := range RangMsg(msg.Ret.ExecutionTrace) {
				source <- item
			}
		}).Walk(func(item interface{}, pipe chan<- interface{}) {
			if item == nil {
				return
			}
			typeMessage, ok := item.(types.Message)
			if !ok {
				return
			}
			// get actor address
			from, to := sub.GetActorAddress(ctx, typeMessage.From, typeMessage.To, msg.TipSet)
			appIds := sub.GetAppIdsByAddress()(context.Background(),
				from.String(), to.String())
			if len(appIds) == 0 {
				return
			}

			isSubCall := false
			if typeMessage.Cid().String() == msg.Msg.Cid().String() {
				isSubCall = true
			}
			oneMsg := model.OneMessage{
				Msg:       typeMessage,
				Implicit:  msg.IsImplicit(),
				IsSubCall: isSubCall,
			}
			pipe <- NewPackMsg(oneMsg, appIds)

		}).Walk(func(item interface{}, pipe chan<- interface{}) {
			if item == nil {
				return
			}
			packMsg, ok := item.(PackMsg)
			if !ok {
				return
			}
			if err := sub.notify.Notify(packMsg.AppIds, &packMsg.Msg); err != nil {
				log.Errorf("[MessageApplied] Notify %s failed: %v", msg.MCid.String(), err)
			}
		}).Done()
	})
}

func (sub *Subscriber) GetAppIdsSubAll() (appIds []string) {
	sub.subAllAppIds.Range(func(key, value interface{}) bool {
		appIds = append(appIds, key.(string))
		return true
	})
	return
}

func (sub *Subscriber) GetAppIdsByAddress() func(ctx context.Context, from, to string) (appIds []string) {
	return func(ctx context.Context, from, to string) (appIds []string) {
		markCache := sub.opts.addressMarkCache
		if markCache.ExistAddress(ctx, to) ||
			markCache.ExistAddress(ctx, from) {
			list, err := sub.opts.appSubDao.FindByAddresses(context.Background(),
				[]string{from, to})
			if err != nil {
				log.Errorf("[core.processing]: find by addresses err: %s", err)
			}
			for _, item := range list {
				appIds = append(appIds, item.AppId)
			}
		}
		return
	}
}

func (sub *Subscriber) Close() {
	sub.wg.Wait()
	// wait for all work msgDone
	sub.workPool.Release()
	sub.notify.Close()
}

// no Need for now !!!!!
// getAllAddress get all address that subscribe the message.
//include subs who have  subscriber to  all appIds and the appId that subscribe the address
//func (sub *Subscriber) getAllAddress(ctx context.Context, from, to string) (allAppIds []string) {
//	var (
//		appMap map[string]bool
//	)
//	filterFn := func(list []string) []string {
//		var (
//			appIds []string
//		)
//		for _, appId := range list {
//			if _, ok := appMap[appId]; !ok {
//				appIds = append(appIds, appId)
//				appMap[appId] = true
//			}
//		}
//		return appIds
//	}
//	// the appIds who subscribe the address
//	allAppIds = append(allAppIds, filterFn(sub.GetAppIdsByAddress()(ctx, from, to))...)
//	// the appIds who subscribe all address
//	allAppIds = append(allAppIds, sub.GetAppIdsSubAll()...)
//	return
//}
