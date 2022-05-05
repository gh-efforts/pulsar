package miner

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type TaskResult struct {
	Posts SectorPostList

	MinerInfoModel           *Info
	FeeDebtModel             *FeeDebt
	LockedFundsModel         *LockedFund
	CurrentDeadlineInfoModel *CurrentDeadlineInfo
	PreCommitsModel          PreCommitInfoList
	SectorsModel             SectorInfoList
	SectorEventsModel        SectorEventList
	SectorDealsModel         SectorDealList
}

func (res *TaskResult) Persist(ctx context.Context, s model.StorageBatch) error {
	if res.PreCommitsModel != nil {
		if err := res.PreCommitsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.SectorsModel != nil {
		if err := res.SectorsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if len(res.SectorEventsModel) > 0 {
		if err := res.SectorEventsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.MinerInfoModel != nil {
		if err := res.MinerInfoModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.LockedFundsModel != nil {
		if err := res.LockedFundsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.FeeDebtModel != nil {
		if err := res.FeeDebtModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.CurrentDeadlineInfoModel != nil {
		if err := res.CurrentDeadlineInfoModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.SectorDealsModel != nil {
		if err := res.SectorDealsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if res.Posts != nil {
		if err := res.Posts.Persist(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

type TaskResultList []*TaskResult

func (ml TaskResultList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerTaskResultList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ml)))
	}
	defer span.End()

	for _, res := range ml {
		if err := res.Persist(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

// TaskLists allow better batched insertion of Miner-related models.
type TaskLists struct {
	MinerInfoModel           InfoList
	FeeDebtModel             FeeDebtList
	LockedFundsModel         LockedFundsList
	CurrentDeadlineInfoModel CurrentDeadlineInfoList
	PreCommitsModel          PreCommitInfoList
	SectorsModel             SectorInfoList
	SectorEventsModel        SectorEventList
	SectorDealsModel         SectorDealList
	SectorPostModel          SectorPostList
}

// Persist PersistModel with every field of MinerTasklists
func (mtl *TaskLists) Persist(ctx context.Context, s model.StorageBatch) error {
	if mtl.PreCommitsModel != nil {
		if err := mtl.PreCommitsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.SectorsModel != nil {
		if err := mtl.SectorsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if len(mtl.SectorEventsModel) > 0 {
		if err := mtl.SectorEventsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.MinerInfoModel != nil {
		if err := mtl.MinerInfoModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.LockedFundsModel != nil {
		if err := mtl.LockedFundsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.FeeDebtModel != nil {
		if err := mtl.FeeDebtModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.CurrentDeadlineInfoModel != nil {
		if err := mtl.CurrentDeadlineInfoModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.SectorDealsModel != nil {
		if err := mtl.SectorDealsModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	if mtl.SectorPostModel != nil {
		if err := mtl.SectorPostModel.Persist(ctx, s); err != nil {
			return err
		}
	}
	return nil
}
