package reward

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
)

const ChainRewardCollection = "chain_rewards"

type ChainReward struct {
	Height                            abi.ChainEpoch  `bson:"_id" json:"height,omitempty"`
	StateRoot                         cid.Cid         `bson:"state_root" json:"state_root"`
	CumSumBaseline                    abi.TokenAmount `bson:"cum_sum_baseline" json:"cum_sum_baseline"`
	CumSumRealized                    abi.TokenAmount `bson:"cum_sum_realized" json:"cum_sum_realized"`
	EffectiveBaselinePower            abi.TokenAmount `bson:"effective_baseline_power" json:"effective_baseline_power"`
	NewBaselinePower                  abi.TokenAmount `bson:"new_baseline_power" json:"new_baseline_power"`
	NewRewardSmoothedPositionEstimate abi.TokenAmount `bson:"new_reward_smoothed_position_estimate" json:"new_reward_smoothed_position_estimate"`
	NewRewardSmoothedVelocityEstimate abi.TokenAmount `bson:"new_reward_smoothed_velocity_estimate" json:"new_reward_smoothed_velocity_estimate"`
	TotalMinedReward                  abi.TokenAmount `bson:"total_mined_reward" json:"total_mined_reward"`
	NewReward                         abi.TokenAmount `bson:"new_reward" json:"new_reward"`
	EffectiveNetworkTime              abi.ChainEpoch  `bson:"effective_network_time" json:"effective_network_time,omitempty"`
}

func (r *ChainReward) Collection() string {
	return ChainRewardCollection
}

func (r *ChainReward) CreateIndexes(_ context.Context, _ *mongo.Database) error {
	return nil
}

func (r *ChainReward) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "ChainReward.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_rewards"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, r)
}

func (r *ChainReward) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if r == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(r)}
}

func (r *ChainReward) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if r == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(r.Collection()).BulkWrite(ctx,
		r.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
