package messages

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
	MessageCollection           = "messages"
	InternalMessageCollection   = "internal_messages"
	MessageGasEconomyCollection = "message_gas_economy"
)

type Message struct {
	Cid                cid.Cid         `bson:"_id" json:"cid"`
	Version            uint64          `bson:"version" json:"version"`
	Nonce              uint64          `bson:"nonce" json:"nonce"`
	Height             abi.ChainEpoch  `bson:"height" json:"height"`
	StateRoot          cid.Cid         `bson:"state_root" json:"state_root"`
	From               address.Address `bson:"from" json:"from"`
	FromActorID        address.Address `bson:"from_actor_id" json:"from_actor_id"`
	To                 address.Address `bson:"to" json:"to"`
	ToActorID          address.Address `bson:"to_actor_id" json:"to_actor_id"`
	Value              abi.TokenAmount `bson:"value" json:"value"`
	Method             abi.MethodNum   `bson:"method" json:"method"`
	ActorName          string          `bson:"actor_name" json:"actor_name"`
	ActorFamily        string          `bson:"actor_family" json:"actor_family"`
	ExitCode           int64           `bson:"exit_code" json:"exit_code"`
	GasUsed            int64           `bson:"gas_used" json:"gas_used"`
	Return             []byte          `bson:"return" json:"return"`
	Params             string          `bson:"params" json:"params"`
	GasLimit           int64           `bson:"gas_limit" json:"gas_limit"`
	GasFeeCap          abi.TokenAmount `bson:"gas_fee_cap" json:"gas_fee_cap"`
	GasPremium         abi.TokenAmount `bson:"gas_premium" json:"gas_premium"`
	ParentBaseFee      abi.TokenAmount `bson:"parent_base_fee" json:"parent_base_fee"`
	BaseFeeBurn        abi.TokenAmount `bson:"base_fee_burn" json:"base_fee_burn"`
	OverEstimationBurn abi.TokenAmount `bson:"over_estimation_burn" json:"over_estimation_burn"`
	MinerPenalty       abi.TokenAmount `bson:"miner_penalty" json:"miner_penalty"`
	MinerTip           abi.TokenAmount `bson:"miner_tip" json:"miner_tip"`
	Refund             abi.TokenAmount `bson:"refund" json:"refund"`
	GasRefund          int64           `bson:"gas_refund" json:"gas_refund"`
	GasBurned          int64           `bson:"gas_burned" json:"gas_burned"`
	Size               int             `bson:"size" json:"size"`
}

func (m *Message) Collection() string {
	return MessageCollection
}

func (m *Message) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(m.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"from_actor_id", 1}},
		},
		{
			Keys: bson.D{{"to_actor_id", 1}, {"method", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Message) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, m.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (m *Message) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "messages"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, m)
}

func (m *Message) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if m == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(m)}
}

func (m *Message) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if m == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(m.Collection()).BulkWrite(ctx,
		m.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type Messages []*Message

func (ms Messages) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ms) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "Messages.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ms)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "messages"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ms) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(ms))
	return s.PersistModel(ctx, ms)
}

func (ms Messages) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ms) == 0 {
		return
	}

	for _, a := range ms {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ms Messages) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ms) == 0 {
		return nil
	}

	_, err := d.Collection(ms[0].Collection()).BulkWrite(ctx,
		ms.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
