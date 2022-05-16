package miner

import (
	"context"
	"fmt"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const SectorPostCollection = "miner_sector_posts"

type SectorPost struct {
	MessageCID   cid.Cid            `bson:"_id" json:"message_cid"`
	Height       abi.ChainEpoch     `bson:"height" json:"height,omitempty"`
	MinerID      address.Address    `bson:"miner_id" json:"miner_id"`
	SectorNumber []abi.SectorNumber `bson:"sector_numbers" json:"sector_number,omitempty"`
}

func (msp *SectorPost) Collection() string {
	return SectorPostCollection
}

func (msp *SectorPost) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(msp.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"miner_id", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (msp *SectorPost) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, msp.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (msp *SectorPost) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if msp == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(msp)}
}

func (msp *SectorPost) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if msp == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(msp.Collection()).BulkWrite(ctx,
		msp.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type SectorPostList []*SectorPost

func (msp *SectorPost) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_posts"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, msp)
}

func (ml SectorPostList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerSectorPostList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ml)))
	}
	defer span.End()
	if len(ml) == 0 {
		return nil
	}

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_posts"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml SectorPostList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml SectorPostList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
