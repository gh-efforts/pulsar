package market

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
)

type TaskResult struct {
	Proposals DealProposals
	States    DealStates
}

func (mtr *TaskResult) Persist(ctx context.Context, s model.StorageBatch) error {
	if err := mtr.Proposals.Persist(ctx, s); err != nil {
		return err
	}
	if err := mtr.States.Persist(ctx, s); err != nil {
		return err
	}
	return nil
}
