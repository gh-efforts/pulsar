package miner

import (
	"context"
	"fmt"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const SectorDealCollection = "miner_sector_deals"

type SectorDeal struct {
	Height       abi.ChainEpoch   `bson:"height" json:"height,omitempty"`
	MinerID      address.Address  `bson:"miner_id" json:"miner_id"`
	SectorNumber abi.SectorNumber `bson:"sector_number" json:"sector_number,omitempty"`
	DealID       abi.DealID       `bson:"_id" json:"deal_id,omitempty"` // default index
}

func (ds *SectorDeal) Collection() string {
	return SectorDealCollection
}

func (ds *SectorDeal) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(ds.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"miner_id", 1}, {"sector_number", -1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (ds *SectorDeal) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, ds.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (ds *SectorDeal) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_deals"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, ds)
}

func (ds *SectorDeal) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if ds == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(ds)}
}

func (ds *SectorDeal) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if ds == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(ds.Collection()).BulkWrite(ctx,
		ds.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type SectorDealList []*SectorDeal

func (ml SectorDealList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerSectorDealList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ml)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_deals"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml SectorDealList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml SectorDealList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
