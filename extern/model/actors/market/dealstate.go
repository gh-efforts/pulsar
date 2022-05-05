package market

import (
	"context"
	"fmt"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const DealStateCollection = "market_deal_states"

type DealState struct {
	DealID           abi.DealID     `bson:"deal_id" json:"deal_id,omitempty"`
	Height           abi.ChainEpoch `bson:"height" json:"height,omitempty"`
	StateRoot        cid.Cid        `bson:"state_root" json:"state_root"`
	SectorStartEpoch abi.ChainEpoch `bson:"sector_start_epoch" json:"sector_start_epoch,omitempty"`
	LastUpdateEpoch  abi.ChainEpoch `bson:"last_update_epoch" json:"last_update_epoch,omitempty"`
	SlashEpoch       abi.ChainEpoch `bson:"slash_epoch" json:"slash_epoch,omitempty"`
}

func (ds *DealState) Collection() string {
	return DealStateCollection
}

func (ds *DealState) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(ds.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"deal_id", 1}},
		},
		{
			Keys: bson.D{{"sector_start_epoch", 1}},
		},
		{
			Keys: bson.D{{"last_update_epoch", 1}},
		},
		{
			Keys: bson.D{{"slash_epoch", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (ds *DealState) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, ds.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (ds *DealState) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "market_deal_states"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, ds)
}

func (ds *DealState) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if ds == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(ds)}
}

func (ds *DealState) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if ds == nil {
		return nil
	}

	_, err := d.Collection(ds.Collection()).BulkWrite(ctx,
		ds.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type DealStates []*DealState

func (dss DealStates) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MarketDealStates.PersistWithTx")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(dss)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "market_deal_states"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(dss) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(dss))
	return s.PersistModel(ctx, dss)
}

func (dss DealStates) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(dss) == 0 {
		return
	}

	for _, a := range dss {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (dss DealStates) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(dss) == 0 {
		return nil
	}

	_, err := d.Collection(dss[0].Collection()).BulkWrite(ctx,
		dss.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
