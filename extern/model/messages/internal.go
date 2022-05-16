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

// InternalMessage Use auto-generated id
type InternalMessage struct {
	Cid           cid.Cid         `bson:"cid" json:"cid"`
	Version       uint64          `bson:"version" json:"version,omitempty"`
	Nonce         uint64          `bson:"nonce" json:"nonce,omitempty"`
	Height        abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	StateRoot     cid.Cid         `bson:"state_root" json:"state_root"`
	SourceMessage cid.Cid         `bson:"source_message" json:"source_message"`
	From          address.Address `bson:"from" json:"from"`
	FromActorID   address.Address `bson:"from_actor_id" json:"from_actor_id"`
	To            address.Address `bson:"to" json:"to"`
	ToActorID     address.Address `bson:"to_actor_id" json:"to_actor_id"`
	Value         abi.TokenAmount `bson:"value" json:"value"`
	Method        abi.MethodNum   `bson:"method" json:"method,omitempty"`
	ActorName     string          `bson:"actor_name" json:"actor_name,omitempty"`
	ActorFamily   string          `bson:"actor_family" json:"actor_family,omitempty"`
	ExitCode      int64           `bson:"exit_code" json:"exit_code,omitempty"`
	GasUsed       int64           `bson:"gas_used" json:"gas_used,omitempty"`
	Return        []byte          `bson:"return" json:"return"`
	Params        string          `bson:"params" json:"params,omitempty"`
	GasLimit      int64           `bson:"gas_limit" json:"gas_limit,omitempty"`
	GasFeeCap     abi.TokenAmount `bson:"gas_fee_cap" json:"gas_fee_cap"`
	GasPremium    abi.TokenAmount `bson:"gas_premium" json:"gas_premium"`
}

func (im *InternalMessage) Collection() string {
	return InternalMessageCollection
}

func (im *InternalMessage) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(im.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"source_message", 1}, {"cid", 1}},
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

func (im *InternalMessage) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, im.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (im *InternalMessage) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "internal_messages"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, im)
}

func (im *InternalMessage) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if im == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(im)}
}

func (im *InternalMessage) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if im == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(im.Collection()).BulkWrite(ctx,
		im.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type InternalMessages []*InternalMessage

func (l InternalMessages) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(l) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "InternalMessageList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(l)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "internal_messages"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(l) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(l))
	return s.PersistModel(ctx, l)
}

func (l InternalMessages) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(l) == 0 {
		return
	}

	for _, a := range l {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (l InternalMessages) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(l) == 0 {
		return nil
	}

	_, err := d.Collection(l[0].Collection()).BulkWrite(ctx,
		l.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
