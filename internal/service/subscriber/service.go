package subscriber

import (
	"context"
	"fmt"
	"sync"

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

	core := &Core{opts: &opts}
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
	fmt.Println("core新消息:", mcid)
	trading := model.Trading{
		TipSet: ts,
		MCid:   mcid,
		Msg:    msg,
	}
	// todo to Confirm whether the call is asynchronous or synchronous
	select {
	case <-ctx.Done():
		log.Errorf("core.MessageApplied: context done: %s", ctx.Err())
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
	if core.connect.IsClosed() {
		return
	}
	core.connect.Close()
}
