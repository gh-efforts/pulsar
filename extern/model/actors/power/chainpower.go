package power

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
	"go.opentelemetry.io/otel/attribute"
)

const ChainPowerCollection = "chain_powers"

type ChainPower struct {
	Height                     abi.ChainEpoch  `bson:"_id" json:"height,omitempty"`
	StateRoot                  cid.Cid         `bson:"state_root" json:"state_root"`
	TotalRawBytesPower         abi.TokenAmount `bson:"total_raw_bytes_power" json:"total_raw_bytes_power"`
	TotalQABytesPower          abi.TokenAmount `bson:"total_qa_bytes_power" json:"total_qa_bytes_power"`
	TotalRawBytesCommitted     abi.TokenAmount `bson:"total_raw_bytes_committed" json:"total_raw_bytes_committed"`
	TotalQABytesCommitted      abi.TokenAmount `bson:"total_qa_bytes_committed" json:"total_qa_bytes_committed"`
	TotalPledgeCollateral      abi.TokenAmount `bson:"total_pledge_collateral" json:"total_pledge_collateral"`
	QASmoothedPositionEstimate abi.TokenAmount `bson:"qa_smoothed_position_estimate" json:"qa_smoothed_position_estimate"`
	QASmoothedVelocityEstimate abi.TokenAmount `bson:"qa_smoothed_velocity_estimate" json:"qa_smoothed_velocity_estimate"`
	MinerCount                 uint64          `bson:"miner_count" json:"miner_count,omitempty"`
	ParticipatingMinerCount    uint64          `bson:"participating_miner_count" json:"participating_miner_count,omitempty"`
}

func (cp *ChainPower) Collection() string {
	return ChainPowerCollection
}

func (cp *ChainPower) CreateIndexes(_ context.Context, _ *mongo.Database) error {
	return nil
}

func (cp *ChainPower) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "ChainPower.PersistWithTx")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_powers"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, cp)
}

func (cp *ChainPower) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if cp == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(cp)}
}

func (cp *ChainPower) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if cp == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(cp.Collection()).BulkWrite(ctx,
		cp.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

// ChainPowerList is a slice of ChainPowers for batch insertion.
type ChainPowerList []*ChainPower

// Persist makes a batch insertion of the list using the given
// transaction.
func (cpl ChainPowerList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "ChainPowerList.PersistWithTx")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(cpl)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_powers"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(cpl) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(cpl))
	return s.PersistModel(ctx, cpl)
}

func (cpl ChainPowerList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(cpl) == 0 {
		return
	}

	for _, a := range cpl {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (cpl ChainPowerList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(cpl) == 0 {
		return nil
	}

	_, err := d.Collection(cpl[0].Collection()).BulkWrite(ctx,
		cpl.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
