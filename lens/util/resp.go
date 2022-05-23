package util

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v2/actors/util/adt"
	"github.com/go-kratos/kratos/v2/log"

	builtininit "github.com/bitrainforest/pulsar/chain/actors/builtin/init"

	"golang.org/x/xerrors"
)

func MakeGetActorIDFunc(_ context.Context, store adt.Store, next *types.TipSet) (func(a address.Address) (address.Address, bool), error) {
	nextStateTree, err := state.LoadStateTree(store, next.ParentState())
	if err != nil {
		return nil, xerrors.Errorf("load state tree: %w", err)
	}

	actorIDs := map[address.Address]address.Address{}

	nextInitActor, err := nextStateTree.GetActor(builtininit.Address)
	if err != nil {
		return nil, xerrors.Errorf("getting init actor: %w", err)
	}

	nextInitActorState, err := builtininit.Load(store, nextInitActor)
	if err != nil {
		return nil, xerrors.Errorf("loading init actor state: %w", err)
	}

	return func(a address.Address) (address.Address, bool) {
		// Shortcut lookup before resolving
		c, ok := actorIDs[a]
		if ok {
			return c, true
		}

		ra, found, err := nextInitActorState.ResolveAddress(a)
		if err != nil || !found {
			log.Warnw("failed to resolve actor address", "address", a.String())
			return a, false
		}
		actorIDs[a] = ra

		return ra, true
	}, nil
}
