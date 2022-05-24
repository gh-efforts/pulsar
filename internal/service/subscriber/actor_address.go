package subscriber

import (
	"context"
	"fmt"
	"sync"

	"github.com/bitrainforest/filmeta-hic/core/log"

	builtininit "github.com/bitrainforest/pulsar/chain/actors/builtin/init"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
)

type Address interface {
	GetActorAddress(ctx context.Context, next *types.TipSet,
		a address.Address) (address.Address, error)
}

var _ Address = (*ActorAddress)(nil)

type ActorAddress struct {
	cs *store.ChainStore
	m  sync.Map
}

func NewActorAddress(cs *store.ChainStore) *ActorAddress {
	return &ActorAddress{
		cs: cs, m: sync.Map{},
	}
}

func (actor *ActorAddress) GetActorAddress(ctx context.Context, next *types.TipSet,
	a address.Address) (address.Address, error) {
	c, ok := actor.m.Load(a)
	if ok {
		log.Infof("[GetActorAddress] actorAddressMap.Load(%s) = %s", a, c)
		return c.(address.Address), nil
	}

	idFunc, err := actor.GetActorIDFunc(ctx, next)
	if err != nil {
		return a, err
	}
	addressRes, ok, err := idFunc(a)
	if err != nil {
		return a, err
	}
	//if existed, return
	if ok && !addressRes.Empty() {
		// todo : check if addressRes is the same as a
		actor.m.Store(a, addressRes)
		return addressRes, nil
	}
	// otherwise, maybe it's  first transaction
	return a, nil
}

func (actor *ActorAddress) GetActorIDFunc(ctx context.Context, next *types.TipSet) (func(a address.Address) (address.Address, bool, error), error) {
	adtStore := actor.cs.ActorStore(ctx)
	nextStateTree, err := state.LoadStateTree(adtStore, next.ParentState())
	if err != nil {
		return nil, fmt.Errorf("load state tree: %w", err)
	}
	nextInitActor, err := nextStateTree.GetActor(builtininit.Address)
	if err != nil {
		return nil, fmt.Errorf("getting init actor: %w", err)
	}

	nextInitActorState, err := builtininit.Load(adtStore, nextInitActor)
	if err != nil {
		return nil, fmt.Errorf("loading init actor state: %w", err)
	}

	return func(a address.Address) (address.Address, bool, error) {
		ra, found, err := nextInitActorState.ResolveAddress(a)
		if err != nil || !found {
			log.Warnf("failed to resolve actor address:%v,err:%v", a.String(), err)
			return a, false, err
		}
		return ra, true, nil
	}, nil
}
