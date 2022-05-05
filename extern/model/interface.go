package model

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// A Storage can marshal models into a serializable format and persist them.
type Storage interface {
	PersistBatch(ctx context.Context, ps ...Persistable) error
}

// A StorageBatch persists a model to storage as part of a batch such as a transaction.
type StorageBatch interface {
	PersistModel(ctx context.Context, m interface{}) error
}

// A Persistable can persist a full copy of itself or its components as part of a storage batch using a specific
// version of a schema. Persist should call PersistModel on s with a model containing data that should be persisted.
// ErrUnsupportedSchemaVersion should be retuned if the Persistable cannot provide a model compatible with the requested
// schema version. If the model does not exist in the schema version because it has been removed or was added in a later
// version then Persist should be a no-op and return nil.
type Persistable interface {
	Persist(ctx context.Context, s StorageBatch) error
}

// A PersistableList is a list of Persistables that should be persisted together
type PersistableList []Persistable

// Ensure that a PersistableList can be used as a Persistable
var _ Persistable = (PersistableList)(nil)

func (pl PersistableList) Persist(ctx context.Context, s StorageBatch) error {
	if len(pl) == 0 {
		return nil
	}
	for _, p := range pl {
		if p == nil {
			continue
		}
		if err := p.Persist(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

type MongoIndexer interface {
	CreateIndexes(ctx context.Context, d *mongo.Database) error
}

type MongoPersist interface {
	ToMongoWriteModel(upsert bool) []mongo.WriteModel
	PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error
}

type MongoModel interface {
	MongoIndexer
	Collection() string
	ToMongoWriteModel(upsert bool) []mongo.WriteModel
	PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error
}

type MongoShardModel interface {
	MongoPersist
	EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error
}
