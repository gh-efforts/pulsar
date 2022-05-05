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

const (
	_ = iota
	PreCommitAdded
	PreCommitExpired
	CommitCapacityAdded
	SectorAdded
	SectorExtended
	SectorFaulted
	SectorRecovering
	SectorRecovered
	SectorExpired
	SectorTerminated
)

const SectorEventCollection = "miner_sector_events"

type SectorEvent struct {
	Height       abi.ChainEpoch   `bson:"height" json:"height,omitempty"`
	MinerID      address.Address  `bson:"miner_id" json:"miner_id"`
	SectorNumber abi.SectorNumber `bson:"sector_number" json:"sector_number,omitempty"`
	StateRoot    cid.Cid          `bson:"state_root" json:"state_root"`
	Event        int              `bson:"event" json:"event,omitempty"`
}

func (mse *SectorEvent) Collection() string {
	return SectorEventCollection
}

func (mse *SectorEvent) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(mse.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"miner_id", 1}, {"event", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (mse *SectorEvent) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, mse.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (mse *SectorEvent) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_events"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, mse)
}

func (mse *SectorEvent) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if mse == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(mse)}
}

func (mse *SectorEvent) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if mse == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(mse.Collection()).BulkWrite(ctx,
		mse.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type SectorEventList []*SectorEvent

func (l SectorEventList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerSectorEventList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(l)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_events"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(l) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(l))
	return s.PersistModel(ctx, l)
}

func (l SectorEventList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(l) == 0 {
		return
	}

	for _, a := range l {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (l SectorEventList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(l) == 0 {
		return nil
	}

	_, err := d.Collection(l[0].Collection()).BulkWrite(ctx,
		l.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
