package subscriber

import (
	"context"
	"sync"

	"github.com/bitrainforest/pulsar/internal/service/subscriber/actoraddress"

	"github.com/panjf2000/ants/v2"

	"github.com/filecoin-project/lotus/chain/store"

	"github.com/bitrainforest/filmeta-hic/core/log"
	"github.com/bitrainforest/filmeta-hic/core/threading"
	"github.com/bitrainforest/filmeta-hic/model"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/ipfs/go-cid"
	"github.com/smallnest/chanx"
)

const (
	DefaultMsgBuffer = 500
	CoreWorkerNum    = 100
)

type CoreOpt func(*Core)

func WithMsgBuffer(buffer int64) CoreOpt {
	return func(c *Core) {
		if buffer > 0 {
			c.msgBuffer = buffer
		}
	}
}

func WithAddress(a Address) CoreOpt {
	return func(c *Core) {
		if a != nil {
			c.address = a
		}
	}
}

type Core struct {
	closed      bool
	msgDone     chan struct{}
	lock        sync.RWMutex
	sub         *Subscriber
	ch          *chanx.UnboundedChan
	lockWait    sync.WaitGroup
	processWait sync.WaitGroup
	msgBuffer   int64
	// why we should use pool?
	//Although we have an unbounded channel which ensures that the messages are sent without block,
	//we are slow to process message if we receive message and process them synchronously.
	//this can cause a backlog of messages
	processPool *ants.Pool
	address     Address
}

func NewCore(sub *Subscriber, opts ...CoreOpt) *Core {
	core := &Core{
		lock: sync.RWMutex{}, msgDone: make(chan struct{}),
		lockWait:    sync.WaitGroup{},
		processWait: sync.WaitGroup{},
		msgBuffer:   DefaultMsgBuffer,
	}
	core.processPool, _ = ants.NewPool(CoreWorkerNum) //nolint:errcheck
	for _, opt := range opts {
		opt(core)
	}
	core.ch = chanx.NewUnboundedChan(int(core.msgBuffer))
	core.sub = sub
	threading.GoSafe(func() {
		core.Rec()
	})
	return core
}

func (core *Core) OverrideExecMonitor(cs *store.ChainStore) *Core {
	actor := actoraddress.NewActorAddress(cs)
	core.address = actor
	return core
}

func (core *Core) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	if core.IsClosed() {
		log.Infof("[MessageApplied] core is closed, ignore message：%v", msg.Cid())
		return nil
	}
	//log.Infof("[Core]Received  message:%v,from:%v,to:%v", mcid.String(), msg.From.String(), msg.To.String())
	core.lockWait.Add(1)
	defer core.lockWait.Done()
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

func (core *Core) Rec() {
	for item := range core.ch.Out {
		msg, ok := item.(*model.Message)
		if !ok {
			log.Errorf("[core.Rec()] msg:%v", item)
			continue
		}
		if err := core.processing(msg); err != nil {
			log.Errorf("[core.Rec()] processing msg:+%v,err:%v", msg, err)
		}
	}
	core.msgDone <- struct{}{}
}

func (core *Core) processing(msg *model.Message) error {
	core.processWait.Add(1)
	ctx := context.Background()
	from := msg.Msg.From
	to := msg.Msg.To

	return core.processPool.Submit(func() {
		defer core.processWait.Done()
		if core.address != nil {
			var (
				wg sync.WaitGroup
			)
			wg.Add(2)
			threading.GoSafe(func() {
				defer wg.Done()
				var (
					err error
				)
				from, err = core.address.GetActorAddress(ctx, msg.TipSet, from)
				if err != nil {
					// just to log getActorAddress error
					log.Errorf("[processing] from address:%v MCid:%v,err:%v", msg.Msg.From, msg.MCid.String(), err.Error())
				}
			})

			threading.GoSafe(func() {
				defer wg.Done()
				var (
					err error
				)
				to, err = core.address.GetActorAddress(ctx, msg.TipSet, to)
				if err != nil {
					// just to log getActorAddress error
					log.Errorf("[processing] to address:%v, MCid:%v,err:%v", msg.Msg.To, msg.MCid.String(), err.Error())
				}
			})
			wg.Wait()
		}
		//log.Infof("[processing] from:%v,to:%v", from.String(), to.String())
		err := core.sub.Notify(ctx, from.String(), to.String(), msg)
		if err != nil {
			log.Errorf("[processing] sub.Notify err:%v", err)
		}
	})
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
		core.lock.Unlock()
		return
	}
	// wait for processing goroutine
	core.lockWait.Wait()
	core.closed = true
	// core.IsClosed==true, so no message will get to MessageApplied,
	//and no message send to UnboundedChan,we can close the UnboundedChan in this way.
	close(core.ch.In)
	core.lock.Unlock()

	// wait msgDone
	<-core.msgDone
	// wait for processPool goroutine
	core.processWait.Wait()

	core.sub.Close()
}
