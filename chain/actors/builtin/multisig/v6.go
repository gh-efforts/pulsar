// Code generated by: `make actors-gen`. DO NOT EDIT.
package multisig

import (
	"bytes"
	"encoding/binary"

	adt6 "github.com/filecoin-project/specs-actors/v6/actors/util/adt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"

	"bony/chain/actors/adt"

	builtin6 "github.com/filecoin-project/specs-actors/v6/actors/builtin"
	msig6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/multisig"
)

var _ State = (*state6)(nil)

func load6(store adt.Store, root cid.Cid) (State, error) {
	out := state6{store: store}
	err := store.Get(store.Context(), root, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type state6 struct {
	msig6.State
	store adt.Store
}

func (s *state6) Code() cid.Cid {
	return builtin6.MultisigActorCodeID
}

func (s *state6) LockedBalance(currEpoch abi.ChainEpoch) (abi.TokenAmount, error) {
	return s.State.AmountLocked(currEpoch - s.State.StartEpoch), nil
}

func (s *state6) StartEpoch() (abi.ChainEpoch, error) {
	return s.State.StartEpoch, nil
}

func (s *state6) UnlockDuration() (abi.ChainEpoch, error) {
	return s.State.UnlockDuration, nil
}

func (s *state6) InitialBalance() (abi.TokenAmount, error) {
	return s.State.InitialBalance, nil
}

func (s *state6) Threshold() (uint64, error) {
	return s.State.NumApprovalsThreshold, nil
}

func (s *state6) Signers() ([]address.Address, error) {
	return s.State.Signers, nil
}

func (s *state6) ForEachPendingTxn(cb func(id int64, txn Transaction) error) error {
	arr, err := adt6.AsMap(s.store, s.State.PendingTxns, builtin6.DefaultHamtBitwidth)
	if err != nil {
		return err
	}
	var out msig6.Transaction
	return arr.ForEach(&out, func(key string) error {
		txid, n := binary.Varint([]byte(key))
		if n <= 0 {
			return xerrors.Errorf("invalid pending transaction key: %v", key)
		}
		return cb(txid, (Transaction)(out)) //nolint:unconvert
	})
}

func (s *state6) PendingTxnChanged(other State) (bool, error) {
	other6, ok := other.(*state6)
	if !ok {
		// treat an upgrade as a change, always
		return true, nil
	}
	return !s.State.PendingTxns.Equals(other6.PendingTxns), nil
}

func (s *state6) transactions() (adt.Map, error) {
	return adt6.AsMap(s.store, s.PendingTxns, builtin6.DefaultHamtBitwidth)
}

func (s *state6) decodeTransaction(val *cbg.Deferred) (Transaction, error) {
	var tx msig6.Transaction
	if err := tx.UnmarshalCBOR(bytes.NewReader(val.Raw)); err != nil {
		return Transaction{}, err
	}
	return tx, nil
}
