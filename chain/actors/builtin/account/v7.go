// Code generated by: `make actors-gen`. DO NOT EDIT.
package account

import (
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"

	"github.com/bitrainforest/pulsar/chain/actors/adt"

	account7 "github.com/filecoin-project/specs-actors/v7/actors/builtin/account"
)

var _ State = (*state7)(nil)

func load7(store adt.Store, root cid.Cid) (State, error) {
	out := state7{store: store}
	err := store.Get(store.Context(), root, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type state7 struct {
	account7.State
	store adt.Store
}

func (s *state7) PubkeyAddress() (address.Address, error) {
	return s.Address, nil
}
