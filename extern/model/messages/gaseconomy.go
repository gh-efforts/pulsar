package messages

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
)

type MessageGasEconomy struct {
	Height              abi.ChainEpoch `bson:"_id" json:"height,omitempty"`
	StateRoot           cid.Cid        `bson:"state_root" json:"state_root"`
	BaseFee             float64        `bson:"base_fee" json:"base_fee,omitempty"`
	BaseFeeChangeLog    float64        `bson:"base_fee_change_log" json:"base_fee_change_log,omitempty"`
	GasLimitTotal       int64          `bson:"gas_limit_total" json:"gas_limit_total,omitempty"`
	GasLimitUniqueTotal int64          `bson:"gas_limit_unique_total" json:"gas_limit_unique_total,omitempty"`
	GasFillRatio        float64        `bson:"gas_fill_ratio" json:"gas_fill_ratio,omitempty"`
	GasCapacityRatio    float64        `bson:"gas_capacity_ratio" json:"gas_capacity_ratio,omitempty"`
	GasWasteRatio       float64        `bson:"gas_waste_ratio" json:"gas_waste_ratio,omitempty"`
}

func (g *MessageGasEconomy) Collection() string {
	return MessageGasEconomyCollection
}

func (g *MessageGasEconomy) CreateIndexes(_ context.Context, _ *mongo.Database) error {
	return nil
}

func (g *MessageGasEconomy) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "message_gas_economy"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, g)
}

func (g *MessageGasEconomy) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if g == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(g)}
}

func (g *MessageGasEconomy) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if g == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(g.Collection()).BulkWrite(ctx,
		g.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
