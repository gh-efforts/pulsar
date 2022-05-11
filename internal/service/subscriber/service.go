package subscriber

import (
	"context"
	"fmt"
	"sync"

	model2 "github.com/bitrainforest/pulsar/internal/model"

	"github.com/bitrainforest/filmeta-hic/core/threading"

	"github.com/bitrainforest/filmeta-hic/core/log"

	"github.com/bitrainforest/filmeta-hic/model"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/ipfs/go-cid"

	"github.com/nats-io/nats.go"
)

type (
	Core struct {
		opts      *Opts
		connect   *nats.Conn
		msgBuffer chan *model.Trading
		lock      sync.Mutex
	}
)

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
	core.msgBuffer = make(chan *model.Trading, opts.msgBuffer)
	connect, err := nats.Connect(uri)
	if err != nil {
		return nil, err
	}
	core.connect = connect
	threading.GoSafe(func() {
		core.processing()
	})
	return core, nil
	// todo start a g to receive messages
}

func (core *Core) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	fmt.Println("[Core MessageApplied] new message:", mcid)
	trading := model.Trading{
		TipSet: ts,
		MCid:   mcid,
		Msg:    msg,
	}

	// for test
	watchModel := model2.NewDefaultAppWatch()
	watchModel.AppId = "56e92447-53a3-48d3-820d-a28c5876f050"
	watchModel.Address = msg.From.String()

	if err := core.opts.appWatchDao.Create(context.TODO(), &watchModel); err != nil {
		log.Errorf("Create err:%v", err)
		return nil
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
	for msg := range core.msgBuffer {
		// we should to cache the subList
		// todo cache
		// todo de-duplication
		list, err := core.opts.appWatchDao.FindByAddresses(context.Background(),
			[]string{msg.Msg.From.String(), msg.Msg.To.String()})
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

		// we are not sure how many subscribers to sub this address,
		// so we do that can't control the number of goroutines
		// todo work pool to control the number of goroutines
		var wg sync.WaitGroup
		for i := range list {
			wg.Add(1)
			subject := list[i].AppId
			threading.GoSafe(func() {
				defer wg.Done()
				if err = core.connect.Publish(subject, msgByte); err != nil {
					log.Errorf("[core.processing] publish msg:%+v err: %s", msg, err)
				}
			})
		}
		wg.Wait()
	}
}

func (core *Core) Stop() {
	core.lock.Lock()
	defer core.lock.Unlock()
	if core.connect.IsClosed() {
		return
	}
	core.connect.Close()
}
