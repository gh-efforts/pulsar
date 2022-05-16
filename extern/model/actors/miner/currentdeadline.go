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
)

const CurrentDeadlineInfoCollection = "miner_current_deadline_infos"

type CurrentDeadlineInfo struct {
	MinerID      address.Address `bson:"miner_id" json:"miner_id"`
	Height       abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	StateRoot    cid.Cid         `bson:"state_root" json:"state_root"`
	CurrentEpoch abi.ChainEpoch  `bson:"current_epoch" json:"current_epoch,omitempty"`
	Index        uint64          `bson:"index" json:"index,omitempty"`
	PeriodStart  abi.ChainEpoch  `bson:"period_start" json:"period_start,omitempty"`
	Open         abi.ChainEpoch  `bson:"open" json:"open,omitempty"`
	Close        abi.ChainEpoch  `bson:"close" json:"close,omitempty"`
	FaultCutoff  abi.ChainEpoch  `bson:"fault_cutoff" json:"fault_cutoff,omitempty"`
	Challenge    abi.ChainEpoch  `bson:"challenge" json:"challenge,omitempty"`
}

func (m *CurrentDeadlineInfo) Collection() string {
	return CurrentDeadlineInfoCollection
}

func (m *CurrentDeadlineInfo) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(m.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"miner_id", 1}, {"index", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *CurrentDeadlineInfo) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, m.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (m *CurrentDeadlineInfo) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerCurrentDeadlineInfo.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_current_deadline_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, m)
}

func (m *CurrentDeadlineInfo) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if m == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(m)}
}

func (m *CurrentDeadlineInfo) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if m == nil {
		return nil
	}

	//d.Collection(m.Collection())

	_, err := d.Collection(m.Collection()).BulkWrite(ctx,
		m.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type CurrentDeadlineInfoList []*CurrentDeadlineInfo

func (ml CurrentDeadlineInfoList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerCurrentDeadlineInfoList.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_current_deadline_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml CurrentDeadlineInfoList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml CurrentDeadlineInfoList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
