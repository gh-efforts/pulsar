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
	"go.opentelemetry.io/otel/attribute"

	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
)

const PreCommitInfoCollection = "miner_pre_commit_infos"

type PreCommitInfo struct {
	Height                 abi.ChainEpoch   `bson:"height" json:"height,omitempty"`
	MinerID                address.Address  `bson:"miner_id" json:"miner_id"`
	SectorNumber           abi.SectorNumber `bson:"sector_number" json:"sector_number,omitempty"`
	StateRoot              cid.Cid          `bson:"state_root" json:"state_root"`
	SealedCID              cid.Cid          `bson:"_id" json:"sealed_cid"` // default index
	SealRandEpoch          abi.ChainEpoch   `bson:"seal_rand_epoch" json:"seal_rand_epoch,omitempty"`
	ExpirationEpoch        abi.ChainEpoch   `bson:"expiration_epoch" json:"expiration_epoch,omitempty"`
	PreCommitDeposit       abi.TokenAmount  `bson:"pre_commit_deposit" json:"pre_commit_deposit"`
	PreCommitEpoch         abi.ChainEpoch   `bson:"pre_commit_epoch" json:"pre_commit_epoch,omitempty"`
	DealWeight             abi.DealWeight   `bson:"deal_weight" json:"deal_weight"`
	VerifiedDealWeight     abi.DealWeight   `bson:"verified_deal_weight" json:"verified_deal_weight"`
	IsReplaceCapacity      bool             `bson:"is_replace_capacity" json:"is_replace_capacity,omitempty"`
	ReplaceSectorDeadline  uint64           `bson:"replace_sector_deadline" json:"replace_sector_deadline,omitempty"`
	ReplaceSectorPartition uint64           `bson:"replace_sector_partition" json:"replace_sector_partition,omitempty"`
	ReplaceSectorNumber    abi.SectorNumber `bson:"replace_sector_number" json:"replace_sector_number,omitempty"`
}

func (mpi *PreCommitInfo) Collection() string {
	return PreCommitInfoCollection
}

func (mpi *PreCommitInfo) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(mpi.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
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

func (mpi *PreCommitInfo) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, mpi.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (mpi *PreCommitInfo) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_pre_commit_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, mpi)
}

func (mpi *PreCommitInfo) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if mpi == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(mpi)}
}

func (mpi *PreCommitInfo) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if mpi == nil {
		return nil
	}

	_, err := d.Collection(mpi.Collection()).BulkWrite(ctx,
		mpi.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type PreCommitInfoList []*PreCommitInfo

func (ml PreCommitInfoList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerPreCommitInfoList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ml)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_pre_commit_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml PreCommitInfoList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml PreCommitInfoList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
