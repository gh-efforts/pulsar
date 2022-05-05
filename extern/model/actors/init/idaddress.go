package init

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

const IdAddressCollection = "id_addresses"

type IdAddress struct {
	ID        address.Address `bson:"id" json:"id"`
	Address   address.Address `bson:"address" json:"address"`
	Height    abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	StateRoot cid.Cid         `bson:"state_root" json:"state_root"`
}

func (ia *IdAddress) Collection() string {
	return IdAddressCollection
}

// CreateIndexes no need to shard
func (ia *IdAddress) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(ia.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"address", 1}},
		},
		{
			Keys: bson.D{{"id", 1}},
		},
		{
			Keys: bson.D{{"height", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (ia *IdAddress) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "id_addresses"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, ia)
}

func (ia *IdAddress) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if ia == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(ia.Collection()).BulkWrite(ctx,
		ia.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

func (ia *IdAddress) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if ia == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(ia)}
}

type IdAddressList []*IdAddress

func (ias IdAddressList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "IdAddressList.PersistWithTx")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(ias)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "id_addresses"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(ias) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(ias))
	return s.PersistModel(ctx, ias)
}

func (ias IdAddressList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(ias) == 0 {
		return
	}

	for _, a := range ias {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (ias IdAddressList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(ias) == 0 {
		return nil
	}

	_, err := d.Collection(ias[0].Collection()).BulkWrite(ctx,
		ias.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
