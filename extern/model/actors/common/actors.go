package common

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

const ActorCollection = "actors"

type Actor struct {
	Head      cid.Cid         `bson:"head" json:"head"`
	ID        address.Address `bson:"id" json:"id"`
	Height    abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	StateRoot cid.Cid         `bson:"state_root" json:"state_root"`
	Code      cid.Cid         `bson:"code" json:"code"`
	Balance   abi.TokenAmount `bson:"balance" json:"balance"`
	Nonce     uint64          `bson:"nonce" json:"nonce,omitempty"`
	State     string          `bson:"state" json:"state,omitempty"`
}

func (a *Actor) Collection() string {
	return ActorCollection
}

func (a *Actor) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(a.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"id", 1}},
		},
		{
			Keys: bson.D{{"id", 1}, {"head", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *Actor) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, a.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (a *Actor) Persist(ctx context.Context, s model.StorageBatch) error {
	if a == nil {
		// Nothing to do
		return nil
	}

	ctx, span := otel.Tracer("").Start(ctx, "Actor.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "actors"))
	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	return s.PersistModel(ctx, a)
}

func (a *Actor) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if a == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(a)}
}

func (a *Actor) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if a == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(a.Collection()).BulkWrite(ctx,
		a.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

// ActorList is a slice of Actors persistable in a single batch.
type ActorList []*Actor

func (actors ActorList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "ActorList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(actors)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "actors"))
	metrics.RecordCount(ctx, metrics.PersistModel, len(actors))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(actors) == 0 {
		return nil
	}
	return s.PersistModel(ctx, actors)
}

func (actors ActorList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(actors) == 0 {
		return
	}

	for _, a := range actors {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (actors ActorList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(actors) == 0 {
		return nil
	}
	_, err := d.Collection(actors[0].Collection()).BulkWrite(ctx,
		actors.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
