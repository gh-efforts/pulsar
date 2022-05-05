package chain

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

const EconomicsCollection = "chain_economics"

type Economics struct {
	Height              abi.ChainEpoch  `bson:"_id" json:"height,omitempty"`
	ParentStateRoot     cid.Cid         `bson:"parent_state_root" json:"parent_state_root"`
	CirculatingFil      abi.TokenAmount `bson:"circulating_fil" json:"circulating_fil"`
	VestedFil           abi.TokenAmount `bson:"vested_fil" json:"vested_fil"`
	MinedFil            abi.TokenAmount `bson:"mined_fil" json:"mined_fil"`
	BurntFil            abi.TokenAmount `bson:"burnt_fil" json:"burnt_fil"`
	LockedFil           abi.TokenAmount `bson:"locked_fil" json:"locked_fil"`
	FilReserveDisbursed abi.TokenAmount `bson:"fil_reserve_disbursed" json:"fil_reserve_disbursed"`
}

func (c *Economics) Collection() string {
	return EconomicsCollection
}

func (c *Economics) CreateIndexes(_ context.Context, _ *mongo.Database) error {

	return nil
}

func (c *Economics) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_economics"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, c)
}

func (c *Economics) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if c == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(c)}
}

func (c *Economics) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if c == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(c.Collection()).BulkWrite(ctx,
		c.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type EconomicsList []*Economics

func (l EconomicsList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(l) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "EconomicsList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(l)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_economics"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(l) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(l))
	return s.PersistModel(ctx, l)
}

func (l EconomicsList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(l) == 0 {
		return
	}

	for _, a := range l {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (l EconomicsList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(l) == 0 {
		return nil
	}

	_, err := d.Collection(l[0].Collection()).BulkWrite(ctx,
		l.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
