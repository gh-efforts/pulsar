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

const InfoCollection = "miner_infos"

type Info struct {
	Height                  abi.ChainEpoch    `bson:"height" json:"height,omitempty"`
	MinerID                 address.Address   `bson:"miner_id" json:"miner_id"`
	StateRoot               cid.Cid           `bson:"state_root" json:"state_root"`
	OwnerID                 address.Address   `bson:"owner_id" json:"owner_id"`
	WorkerID                address.Address   `bson:"worker_id" json:"worker_id"`
	NewWorker               address.Address   `bson:"new_worker" json:"new_worker"`
	WorkerChangeEpoch       abi.ChainEpoch    `bson:"worker_change_epoch" json:"worker_change_epoch,omitempty"`
	ConsensusFaultedElapsed abi.ChainEpoch    `bson:"consensus_faulted_elapsed" json:"consensus_faulted_elapsed,omitempty"`
	PeerID                  string            `bson:"peer_id" json:"peer_id,omitempty"`
	ControlAddresses        []address.Address `bson:"control_addresses" json:"control_addresses,omitempty"`
	MultiAddresses          []string          `bson:"multi_addresses" json:"multi_addresses,omitempty"`
	SectorSize              abi.SectorSize    `bson:"sector_size" json:"sector_size,omitempty"`
}

func (m *Info) Collection() string {
	return InfoCollection
}

func (m *Info) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(m.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"miner_id", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Info) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, m.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (m *Info) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerInfoModel.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, m)
}

func (m *Info) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if m == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(m)}
}

func (m *Info) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if m == nil {
		return nil
	}

	_, err := d.Collection(m.Collection()).BulkWrite(ctx,
		m.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type InfoList []*Info

func (ml InfoList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerInfoList.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml InfoList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml InfoList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
