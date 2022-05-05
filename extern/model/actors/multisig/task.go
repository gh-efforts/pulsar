package multisig

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
)

type TaskResult struct {
	TransactionModel TransactionList
}

func (mtr *TaskResult) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(mtr.TransactionModel) > 0 {
		return mtr.TransactionModel.Persist(ctx, s)
	}
	return nil
}

type TaskResultList []*TaskResult

func (ml TaskResultList) Persist(ctx context.Context, s model.StorageBatch) error {
	for _, res := range ml {
		if err := res.Persist(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
