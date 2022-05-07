package exec_monitor

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("bony/exec_monitor")
)

func NewBufferedExecMonitor(cs *store.ChainStore) *BufferedExecMonitor {
	// this only errors when a negative size is supplied...y u no accept unsigned ints :(
	cache, err := lru.New(64)
	if err != nil {
		panic(err)
	}
	filled, err := lru.New(64)
	if err != nil {
		panic(err)
	}
	return &BufferedExecMonitor{
		cs:     cs,
		cache:  cache,
		filled: filled,
		fillCh: make(chan struct{}, 1),
	}
}

type BufferedExecMonitor struct {
	cs     *store.ChainStore
	filled *lru.Cache
	fillCh chan struct{}

	cacheMu sync.Mutex
	cache   *lru.Cache
}

var _ stmgr.ExecMonitor = (*BufferedExecMonitor)(nil)

type BufferedExecution struct {
	TipSet        *types.TipSet
	Mcid          cid.Cid
	Msg           *types.Message
	Ret           *vm.ApplyRet
	Implicit      bool
	ToActorCode   cid.Cid
	ToActorID     address.Address
	FromActorCode cid.Cid
	FromActorID   address.Address
}

func (b *BufferedExecMonitor) MessageApplied(ctx context.Context, ts *types.TipSet, mcid cid.Cid, msg *types.Message, ret *vm.ApplyRet, implicit bool) error {
	fmt.Println("new message cid:", mcid)
	// todo handle message
	execution := &BufferedExecution{
		TipSet:   ts,
		Mcid:     mcid,
		Msg:      msg,
		Ret:      ret,
		Implicit: implicit,
	}

	b.cacheMu.Lock()
	defer b.cacheMu.Unlock()

	// if this is the first tipset we have seen a message applied for add it to the cache and bail.
	found := b.cache.Contains(ts.Key())
	if !found {
		b.cache.Add(ts.Key(), &Executions{
			Data:   []*BufferedExecution{execution},
			Filled: false,
		})
		return nil
	}
	// otherwise append to the current list of executions for this tipset.
	v, _ := b.cache.Get(ts.Key())
	exe := v.(*Executions)
	exe.Data = append(exe.Data, execution)
	evicted := b.cache.Add(ts.Key(), exe)
	// TODO it would be nice to know if we have extracted the buffered execution for this tipset already, maybe not important
	if evicted {
		log.Debugw("Evicting tipset from buffered exec monitor", "ts", ts.Key())
	}

	return nil
}

type Executions struct {
	Data   []*BufferedExecution
	Filled bool
}
