package subscriber

import (
	"context"
	"sync"

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
	closed    bool
	msgDone   chan struct{}
	lock      sync.RWMutex
	sub       *Subscriber
	ch        *chanx.UnboundedChan
	wgStop    sync.WaitGroup
	msgBuffer int64

	actor *ActorAddress
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

func (core *Core) OverrideExecMonitor(cs *store.ChainStore) *Core {
	actor := NewActorAddress(cs)
	core.actor = actor
	return core
}

func (core *Core) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	if core.IsClosed() {
		log.Infof("[MessageApplied] core is closed, ignore message：%v", msg.Cid())
		return nil
	}
	log.Infof("[Core]Received  message:%v,from:%v,to:%v", mcid.String(), msg.From.String(), msg.To.String())
	core.wgStop.Add(1)
	defer core.wgStop.Done()
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

		from := msg.Msg.From
		to := msg.Msg.To

		if core.actor != nil {
			var (
				err error
			)
			from, err = core.actor.GetActorAddress(ctx, msg.TipSet, from)
			if err != nil {
				// just to log getActorAddress error
				log.Errorf("[processing] from address:%v called  getActorID,err:%v", msg.Msg.From, err)
			}

			to, err = core.actor.GetActorAddress(ctx, msg.TipSet, to)
			if err != nil {
				// just to log getActorAddress error
				log.Errorf("[processing] to address:%v called  getActorID,err:%v", msg.Msg.To, err)
			}
		}
		log.Infof("[processing] from:%v,to:%v", from.String(), to.String())
		if err := core.sub.Notify(ctx, from.String(), to.String(), msg); err != nil {
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
