package subscriber

import (
	"context"
	"sync"

	"github.com/bitrainforest/pulsar/lens/util"

	"github.com/filecoin-project/lotus/chain/store"

	"github.com/bitrainforest/filmeta-hic/core/log"
	"github.com/bitrainforest/filmeta-hic/core/threading"
	"github.com/bitrainforest/filmeta-hic/model"
	"github.com/bitrainforest/pulsar/internal/utils/locker"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/ipfs/go-cid"
	"github.com/smallnest/chanx"
)

const (
	DefaultMsgBuffer = 500
)

type CoreOpt func(*Core)

func WithMsgBuffer(buffer int64) CoreOpt {
	return func(c *Core) {
		if buffer > 0 {
			c.msgBuffer = buffer
		}
	}
}

type Core struct {
	cs        *store.ChainStore
	closed    bool
	msgDone   chan struct{}
	lock      sync.RWMutex
	sub       *Subscriber
	ch        *chanx.UnboundedChan
	wgStop    sync.WaitGroup
	msgBuffer int64
}

func NewCore(sub *Subscriber, opts ...CoreOpt) *Core {
	core := &Core{
		lock: sync.RWMutex{}, msgDone: make(chan struct{}),
		wgStop:    sync.WaitGroup{},
		msgBuffer: DefaultMsgBuffer,
	}

	for _, opt := range opts {
		opt(core)
	}
	core.ch = chanx.NewUnboundedChan(int(core.msgBuffer))
	core.sub = sub
	threading.GoSafe(func() {
		core.processing()
	})
	return core
}

func (core *Core) GetExecMonitor(cs *store.ChainStore) *Core {
	core.cs = cs
	return core
}

func (core *Core) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	if core.IsClosed() {
		//log.Infof("[MessageApplied] core is closed, ignore message")
		return nil
	}
	log.Infof("[Core]Received  message:%v,from:%v,to:%v", mcid.String(), msg.From.String(), msg.To.String())
	core.wgStop.Add(1)
	defer core.wgStop.Done()
	ok, err := locker.NewRedisLock(ctx, mcid.String(), 20).Acquire(ctx)
	if err != nil {
		log.Errorf("[MessageApplied] Acquire %s failed: %v", mcid.String(), err)
		return err
	}
	if !ok {
		log.Infof("[MessageApplied] locked message %s", mcid.String())
		return nil
	}
	trading := model.Message{
		TipSet:   ts,
		MCid:     mcid,
		Msg:      msg,
		Ret:      ret,
		Implicit: implicit,
	}
	select {
	case <-ctx.Done():
		log.Errorf("[Core MessageApplied] core.MessageApplied: context msgDone: %s", ctx.Err())
		return nil
	default:
		core.ch.In <- &trading
	}
	return nil
}

func (core *Core) processing() {
	for item := range core.ch.Out {
		msg := item.(*model.Message)
		ctx := context.Background()
		// todo get address
		getActorID, err := util.MakeGetActorIDFunc(ctx, core.cs.ActorStore(ctx), msg.TipSet)
		if err != nil {
			log.Errorf("[processing] get actor id failed: %v", err)
			continue
		}
		to, ok := getActorID(msg.Msg.To)
		if !ok {
			log.Errorf("[processing] to address:%v called  getActorID false", msg.Msg.To)
			continue
		}
		from, ok := getActorID(msg.Msg.From)
		if !ok {
			log.Errorf("[processing] from address:%v called  getActorID false", msg.Msg.From)
			continue
		}
		log.Infof("[processing] to:%v,from:%v", to, from)

		if err := core.sub.Notify(ctx, to.String(), from.String(), msg); err != nil {
			log.Errorf("[Core processing] notify failed: %v", err)
		}
	}
	core.msgDone <- struct{}{}
}

func (core *Core) IsClosed() bool {
	core.lock.RLock()
	defer core.lock.RUnlock()
	return core.closed
}

func (core *Core) Stop() {
	log.Infof("[]Core Stop")
	if core.closed {
		return
	}
	core.lock.Lock()
	if core.closed {
		return
	}
	// wait for processing goroutine
	core.wgStop.Wait()
	core.closed = true
	// core.IsClosed==true, so no message will get to MessageApplied,
	//and no message send to UnboundedChan,we can close the UnboundedChan in this way.
	close(core.ch.In)
	core.lock.Unlock()

	// wait msgDone
	<-core.msgDone
	core.sub.Close()
}
