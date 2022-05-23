package util

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
	"github.com/go-kratos/kratos/v2/log"

	builtininit "github.com/bitrainforest/pulsar/chain/actors/builtin/init"
)

var (
	actorAddressMap sync.Map
)

func MakeGetActorIDFunc(_ context.Context, store adt.Store, next *types.TipSet) (func(a address.Address) (address.Address, bool), error) {
	nextStateTree, err := state.LoadStateTree(store, next.ParentState())
	if err != nil {
		return nil, fmt.Errorf("load state tree: %w", err)
	}
	nextInitActor, err := nextStateTree.GetActor(builtininit.Address)
	if err != nil {
		return nil, fmt.Errorf("getting init actor: %w", err)
	}

	nextInitActorState, err := builtininit.Load(store, nextInitActor)
	if err != nil {
		return nil, fmt.Errorf("loading init actor state: %w", err)
	}

	return func(a address.Address) (address.Address, bool) {
		// Shortcut lookup before resolving
		c, ok := actorAddressMap.Load(a)
		if ok {
			log.Infof("[MakeGetActorIDFunc] actorAddressMap.Load(%s) = %s", a, c)
			return c.(address.Address), true
		}

		ra, found, err := nextInitActorState.ResolveAddress(a)
		if err != nil || !found {
			log.Warnf("failed to resolve actor address:%v,err:%v", a.String(), err)
			return a, false
		}
		actorAddressMap.Store(a, ra)
		return ra, true
	}, nil
}
