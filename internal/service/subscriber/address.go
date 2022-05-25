package subscriber

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
)

type Address interface {
	GetActorAddress(ctx context.Context, next *types.TipSet,
		a address.Address) (address.Address, error)
}
