package power

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
)

type TaskResult struct {
	ChainPowerModel *ChainPower    `db:"chain_power_model" cql:"chain_power_model"`
	ClaimStateModel ActorClaimList `db:"claim_state_model" cql:"claim_state_model"`
}

func (p *TaskResult) Persist(ctx context.Context, s model.StorageBatch) error {
	if p.ChainPowerModel != nil {
		if err := p.ChainPowerModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if p.ClaimStateModel != nil {
		if err := p.ClaimStateModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
