package multisig

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
)

const TransactionCollection = "multisig_transactions"

type Transaction struct {
	ID         int64             `bson:"id" json:"id,omitempty"`
	MultiSigID address.Address   `bson:"multi_sig_id" json:"multi_sig_id"`
	StateRoot  cid.Cid           `bson:"state_root" json:"state_root"`
	Height     abi.ChainEpoch    `bson:"height" json:"height,omitempty"`
	To         address.Address   `bson:"to" json:"to"`
	Value      abi.TokenAmount   `bson:"value" json:"value"`
	Method     abi.MethodNum     `bson:"method" json:"method,omitempty"`
	Params     []byte            `bson:"params" json:"params,omitempty"`
	Approved   []address.Address `bson:"approved" json:"approved,omitempty"`
}

func (m *Transaction) Collection() string {
	return TransactionCollection
}

func (m *Transaction) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(m.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"id", 1}}},
		{
			Keys: bson.D{{"from", 1}}},
		{
			Keys: bson.D{{"to", 1}, {"method", 1}}},
		{
			Keys: bson.D{{"approved", 1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Transaction) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, m.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (m *Transaction) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "multisig_transactions"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, m)
}

func (m *Transaction) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if m == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(m)}
}

func (m *Transaction) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if m == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(m.Collection()).BulkWrite(ctx,
		m.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type TransactionList []*Transaction

func (ml TransactionList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "multisig_transactions"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ml) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(ml))
	return s.PersistModel(ctx, ml)
}

func (ml TransactionList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ml) == 0 {
		return
	}

	for _, a := range ml {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ml TransactionList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ml) == 0 {
		return nil
	}

	_, err := d.Collection(ml[0].Collection()).BulkWrite(ctx,
		ml.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
