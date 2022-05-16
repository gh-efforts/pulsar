package verifreg

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
)

const (
	_ = iota
	Added
	Removed
	Modified
)

const VerifiedRegistryVerifierCollection = "verifreg_verifiers"

type VerifiedRegistryVerifier struct {
	Height    abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	StateRoot cid.Cid         `bson:"state_root" json:"state_root"`
	Address   address.Address `bson:"address" json:"address"`
	Event     int             `bson:"event" json:"event,omitempty"`
	DataCap   abi.TokenAmount `bson:"data_cap" json:"data_cap"`
}

func (v *VerifiedRegistryVerifier) Collection() string {
	return VerifiedRegistryVerifierCollection
}

func (v *VerifiedRegistryVerifier) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(v.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"address", 1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (v *VerifiedRegistryVerifier) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "verified_registry_verifier"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	return s.PersistModel(ctx, v)
}

func (v *VerifiedRegistryVerifier) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if v == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(v)}
}

func (v *VerifiedRegistryVerifier) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if v == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(v.Collection()).BulkWrite(ctx,
		v.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type VerifiedRegistryVerifiersList []*VerifiedRegistryVerifier

func (v VerifiedRegistryVerifiersList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "verified_registry_verifier"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(v) == 0 {
		return nil
	}

	return s.PersistModel(ctx, v)
}

func (v VerifiedRegistryVerifiersList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(v) == 0 {
		return
	}

	for _, a := range v {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (v VerifiedRegistryVerifiersList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(v) == 0 {
		return nil
	}

	_, err := d.Collection(v[0].Collection()).BulkWrite(ctx,
		v.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

const VerifiedRegistryVerifiedClientCollection = "verifreg_clients"

type VerifiedRegistryVerifiedClient struct {
	Height    abi.ChainEpoch  `bson:"height"`
	StateRoot cid.Cid         `bson:"state_root"`
	Address   address.Address `bson:"address"`
	Event     int             `bson:"event"`
	DataCap   abi.TokenAmount `bson:"data_cap"`
}

func (v *VerifiedRegistryVerifiedClient) Collection() string {
	return VerifiedRegistryVerifiedClientCollection
}

func (v *VerifiedRegistryVerifiedClient) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(v.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"address", 1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (v *VerifiedRegistryVerifiedClient) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "verified_registry_verified_client"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	return s.PersistModel(ctx, v)
}

func (v *VerifiedRegistryVerifiedClient) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if v == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(v)}
}

func (v *VerifiedRegistryVerifiedClient) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if v == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(v.Collection()).BulkWrite(ctx,
		v.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type VerifiedRegistryVerifiedClientsList []*VerifiedRegistryVerifiedClient

func (v VerifiedRegistryVerifiedClientsList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "verified_registry_verified_client"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(v) == 0 {
		return nil
	}
	return s.PersistModel(ctx, v)
}

func (v VerifiedRegistryVerifiedClientsList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(v) == 0 {
		return
	}

	for _, a := range v {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (v VerifiedRegistryVerifiedClientsList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(v) == 0 {
		return nil
	}

	_, err := d.Collection(v[0].Collection()).BulkWrite(ctx,
		v.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

var _ model.Persistable = (*VerifiedRegistryVerifier)(nil)
var _ model.Persistable = (*VerifiedRegistryVerifiersList)(nil)
var _ model.Persistable = (*VerifiedRegistryVerifiedClient)(nil)
var _ model.Persistable = (*VerifiedRegistryVerifiedClientsList)(nil)
