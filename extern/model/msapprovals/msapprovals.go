package msapprovals

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

const MultiSigApprovalCollection = "multisig_approvals"

type MultiSigApproval struct {
	Message        cid.Cid           `bson:"_id" json:"message"`
	Height         abi.ChainEpoch    `bson:"height" json:"height,omitempty"`
	StateRoot      cid.Cid           `bson:"state_root" json:"state_root"`
	MultiSigID     address.Address   `bson:"multi_sig_id" json:"multi_sig_id"`
	Method         abi.MethodNum     `bson:"method" json:"method,omitempty"`
	Approver       address.Address   `bson:"approver" json:"approver"`
	Threshold      uint64            `bson:"threshold" json:"threshold,omitempty"`
	InitialBalance abi.TokenAmount   `bson:"initial_balance" json:"initial_balance"`
	Signers        []address.Address `bson:"signers" json:"signers,omitempty"`
	GasUsed        int64             `bson:"gas_used" json:"gas_used,omitempty"`
	TransactionID  int64             `bson:"transaction_id" json:"transaction_id,omitempty"`
	To             address.Address   `bson:"to" json:"to"`
	Value          abi.TokenAmount   `bson:"value" json:"value"`
}

func (ma *MultiSigApproval) Collection() string {
	return MultiSigApprovalCollection
}

func (ma *MultiSigApproval) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(ma.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"multi_sig_id", 1}},
		},
		{
			Keys: bson.D{{"approver", 1}},
		},
		{
			Keys: bson.D{{"signers", 1}},
		},
		{
			Keys: bson.D{{"transaction_id", 1}},
		},
		{
			Keys: bson.D{{"to", 1}, {"method", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (ma *MultiSigApproval) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, ma.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (ma *MultiSigApproval) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "multisig_approvals"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, ma)
}

func (ma *MultiSigApproval) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if ma == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(ma)}
}

func (ma *MultiSigApproval) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if ma == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(ma.Collection()).BulkWrite(ctx,
		ma.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type MultiSigApprovalList []*MultiSigApproval

func (mal MultiSigApprovalList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(mal) == 0 {
		return nil
	}
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "multisig_approvals"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(mal) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(mal))
	return s.PersistModel(ctx, mal)
}

func (mal MultiSigApprovalList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(mal) == 0 {
		return
	}

	for _, a := range mal {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (mal MultiSigApprovalList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(mal) == 0 {
		return nil
	}

	_, err := d.Collection(mal[0].Collection()).BulkWrite(ctx,
		mal.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
