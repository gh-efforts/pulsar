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

const SectorInfoCollection = "miner_sector_infos"

type SectorInfo struct {
	SealedCID             cid.Cid          `bson:"sealed_cid" json:"sealed_cid"`
	Height                abi.ChainEpoch   `bson:"height" json:"height,omitempty"`
	MinerID               address.Address  `bson:"miner_id" json:"miner_id"`
	SectorNumber          abi.SectorNumber `bson:"sector_number" json:"sector_number,omitempty"`
	StateRoot             cid.Cid          `bson:"state_root" json:"state_root"`
	ActivationEpoch       abi.ChainEpoch   `bson:"activation_epoch" json:"activation_epoch,omitempty"`
	ExpirationEpoch       abi.ChainEpoch   `bson:"expiration_epoch" json:"expiration_epoch,omitempty"`
	DealWeight            abi.TokenAmount  `bson:"deal_weight" json:"deal_weight"`
	VerifiedDealWeight    abi.TokenAmount  `bson:"verified_deal_weight" json:"verified_deal_weight"`
	InitialPledge         abi.TokenAmount  `bson:"initial_pledge" json:"initial_pledge"`
	ExpectedDayReward     abi.TokenAmount  `bson:"expected_day_reward" json:"expected_day_reward"`
	ExpectedStoragePledge abi.TokenAmount  `bson:"expected_storage_pledge" json:"expected_storage_pledge"`
}

func (msi *SectorInfo) Collection() string {
	return SectorInfoCollection
}

func (msi *SectorInfo) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(msi.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
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

func (msi *SectorInfo) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, msi.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (msi *SectorInfo) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, msi)
}

func (msi *SectorInfo) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if msi == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(msi)}
}

func (msi *SectorInfo) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if msi == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(msi.Collection()).BulkWrite(ctx,
		msi.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type (
	SectorInfoList []*SectorInfo
)

func (ml SectorInfoList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerSectorInfoList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ml)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_sector_infos"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml SectorInfoList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml SectorInfoList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
