package subscriber

import (
	"context"
	"sync"

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

type Core struct {
	closed bool
	done   chan struct{}
	lock   sync.RWMutex
	sub    *Subscriber
	ch     *chanx.UnboundedChan
	wg     sync.WaitGroup
}

func NewCore(sub *Subscriber, opts ...CoreOpt) *Core {
	core := &Core{
		lock: sync.RWMutex{}, done: make(chan struct{}),
		wg: sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(core)
	}
	core.ch = chanx.NewUnboundedChan(DefaultMsgBuffer)
	core.sub = sub
	threading.GoSafe(func() {
		core.processing()
	})
	return core
}

func (core *Core) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	if core.IsClosed() {
		//log.Infof("[MessageApplied] core is closed, ignore message")
		return nil
	}
	core.wg.Add(1)
	defer core.wg.Done()
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
	// todo to Confirm whether the call is asynchronous or synchronous
	select {
	case <-ctx.Done():
		log.Errorf("[Core MessageApplied] core.MessageApplied: context done: %s", ctx.Err())
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
		to := msg.Msg.To.String()
		from := msg.Msg.From.String()
		if err := core.sub.Notify(ctx, to, from, msg); err != nil {
			log.Errorf("[Core processing] notify failed: %v", err)
		}
	}
	core.done <- struct{}{}
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
	core.wg.Wait()
	core.closed = true
	// core.IsClosed==true, so no message will get to MessageApplied,
	//and no message send to UnboundedChan,we can close the UnboundedChan in this way.
	close(core.ch.In)
	core.lock.Unlock()

	// wait done
	<-core.done
	core.sub.Close()
}
