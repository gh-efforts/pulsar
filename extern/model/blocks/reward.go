package blocks

import (
	"context"

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

type Reward struct {
	Cid        cid.Cid         `bson:"_id" json:"cid"`
	Height     abi.ChainEpoch  `bson:"height" json:"height"`
	Miner      address.Address `bson:"miner" json:"miner"`
	Penalty    abi.TokenAmount `bson:"penalty" json:"penalty"`
	GasReward  abi.TokenAmount `bson:"gas_reward" json:"gas_reward"`
	MineReward abi.TokenAmount `bson:"mine_reward" json:"mine_reward"`
	WinCount   int64           `bson:"win_count" json:"win_count"`
}

func (b *Reward) Collection() string {
	return RewardCollection
}

func (b *Reward) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(b.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"miner", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *Reward) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "block_reward"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, b)
}

func (b *Reward) ToMongoWriteModel(upsert bool) []mongo.WriteModel {
	if b == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewReplaceOneModel().
		SetFilter(bson.M{"_id": b.Cid}).SetReplacement(b).SetUpsert(upsert)}
}

func (b *Reward) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if b == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(b.Collection()).BulkWrite(ctx,
		b.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type Rewards []*Reward

func (l Rewards) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(l) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "Rewards.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(l)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "rewards"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(l) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(l))
	return s.PersistModel(ctx, l)
}

func (l Rewards) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(l) == 0 {
		return
	}

	for _, a := range l {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (l Rewards) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(l) == 0 {
		return nil
	}

	_, err := d.Collection(l[0].Collection()).BulkWrite(ctx,
		l.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
