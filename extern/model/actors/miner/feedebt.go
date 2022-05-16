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

const FeeDebtCollection = "miner_fee_debts"

type FeeDebt struct {
	Height    abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	MinerID   address.Address `bson:"miner_id" json:"miner_id"`
	StateRoot cid.Cid         `bson:"state_root" json:"state_root"`

	FeeDebt abi.TokenAmount `bson:"fee_debt" json:"fee_debt"`
}

func (m *FeeDebt) Collection() string {
	return FeeDebtCollection
}

func (m *FeeDebt) CreateIndexes(ctx context.Context, d *mongo.Database) error {
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

func (m *FeeDebt) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, m.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (m *FeeDebt) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerFeeDebt.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_fee_debts"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, m)
}

func (m *FeeDebt) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if m == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(m)}
}

func (m *FeeDebt) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if m == nil {
		return nil
	}

	_, err := d.Collection(m.Collection()).BulkWrite(ctx,
		m.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type FeeDebtList []*FeeDebt

func (ml FeeDebtList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MinerFeeDebtList.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "miner_fee_debts"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml FeeDebtList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml FeeDebtList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
