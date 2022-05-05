package model

import (
	"context"

	"github.com/filecoin-project/go-state-types/abi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Clean struct {
	epoch      abi.ChainEpoch
	collection string
	epochField string
}

func NewClean(collection string, epoch abi.ChainEpoch, epochField string) *Clean {
	return &Clean{
		epoch:      epoch,
		collection: collection,
		epochField: epochField,
	}
}

func (c *Clean) Persist(ctx context.Context, s StorageBatch) error {
	return s.PersistModel(ctx, c)
}

func (c *Clean) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	return []mongo.WriteModel{
		mongo.NewDeleteManyModel().SetFilter(bson.D{{c.epochField, c.epoch}}),
	}
}

func (c *Clean) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if c == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(c.collection).BulkWrite(ctx,
		c.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type CleanList []*Clean

func (c CleanList) Persist(ctx context.Context, s StorageBatch) error {
	if len(c) == 0 {
		return nil
	}
	return s.PersistModel(ctx, c)
}

func (c CleanList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(c) == 0 {
		return
	}

	for _, a := range c {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (c CleanList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(c) == 0 {
		return nil
	}
	_, err := d.Collection(c[0].collection).BulkWrite(ctx,
		c.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
