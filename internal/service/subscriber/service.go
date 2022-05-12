package subscriber

import (
	"context"
	"fmt"
	"sync"

	"github.com/panjf2000/ants/v2"

	"github.com/bitrainforest/filmeta-hic/core/threading"

	"github.com/bitrainforest/filmeta-hic/core/log"

	"github.com/bitrainforest/filmeta-hic/model"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/ipfs/go-cid"

	"github.com/nats-io/nats.go"
)

const (
	DefaultWorkPoolNum = 100
)

type Core struct {
	opts      *Opts
	connect   *nats.Conn
	closed    bool
	msgBuffer chan *model.Trading
	lock      sync.Mutex
	workPool  *ants.Pool
}

func NewCore(uri string, fns ...OptFn) (*Core, error) {
	opts := defaultOpts()

	if uri != "" {
		opts.natsUri = uri
	}

	for _, fn := range fns {
		fn(&opts)
	}

	if opts.msgBuffer < 0 {
		opts.msgBuffer = DefaultMsgBuffer
	}
	if opts.msgBuffer > MaxMsgBuffer {
		opts.msgBuffer = MaxMsgBuffer
	}

	core := &Core{opts: &opts, lock: sync.Mutex{}}
	// msgBuffer
	core.msgBuffer = make(chan *model.Trading, opts.msgBuffer)

	var (
		err error
	)
	if core.workPool, err = ants.NewPool(DefaultWorkPoolNum); err != nil {
		return nil, err
	}

	connect, err := nats.Connect(uri)
	if err != nil {
		return nil, err
	}
	core.connect = connect
	threading.GoSafe(func() {
		core.processing()
	})
	return core, nil
}

func (core *Core) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	fmt.Println("[Core MessageApplied] message:", mcid, msg.From.String(), msg.From.String())
	trading := model.Trading{
		TipSet: ts,
		MCid:   mcid,
		Msg:    msg,
	}
	// todo to Confirm whether the call is asynchronous or synchronous
	select {
	case <-ctx.Done():
		log.Errorf("[Core MessageApplied] core.MessageApplied: context done: %s", ctx.Err())
	case core.msgBuffer <- &trading:
	}
	return nil
}

func (core *Core) processing() {
	markCache := core.opts.addressMarkCache
	for msg := range core.msgBuffer {
		ctx := context.Background()
		to := msg.Msg.To.String()
		from := msg.Msg.From.String()

		if !markCache.ExistAddress(ctx, to) &&
			!markCache.ExistAddress(ctx, from) {
			log.Infof("both address from %v,to %v have  no sub:", from, to)
			continue
		}

		list, err := core.opts.appWatchDao.FindByAddresses(context.Background(),
			[]string{from, to})
		if err != nil {
			log.Errorf("[core.processing]: find by addresses err: %s", err)
			continue
		}
		if len(list) == 0 {
			continue
		}

		msgByte, err := msg.Marshal()
		if err != nil {
			log.Errorf("[core.processing] marshal msg:%+v err: %s", msg, err)
			continue
		}

		publishFn := func(subject string, msgByte []byte) error {
			return core.connect.Publish(subject, msgByte)
		}
		var wg sync.WaitGroup

		for i := range list {
			wg.Add(1)
			subject := list[i].AppId
			_ = core.workPool.Submit(func() { //nolint
				defer wg.Done()
				if err := publishFn(subject, msgByte); err != nil {
					log.Errorf("[core.processing] publish msg:%+v err: %s", msg, err)
				}
			})
		}
		wg.Wait()
	}
}

func (core *Core) IsClosed() bool {
	return core.closed
}

func (core *Core) Stop() {
	core.lock.Lock()
	defer core.lock.Unlock()

	if core.IsClosed() {
		return
	}
	core.closed = true

	// work pol release
	core.workPool.Release()
	// nats.Conn close
	core.connect.Close()
}
