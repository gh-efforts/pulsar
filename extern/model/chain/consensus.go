package chain

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const ConsensusCollection = "chain_consensuses"

type Consensus struct {
	Height          abi.ChainEpoch `bson:"_id" json:"height,omitempty"`
	ParentStateRoot cid.Cid        `bson:"parent_state_root" json:"parent_state_root"`
	ParentTipSet    []cid.Cid      `bson:"parent_tip_set" json:"parent_tip_set,omitempty"`
	TipSet          []cid.Cid      `bson:"tip_set" json:"tip_set,omitempty"`
}

func (c *Consensus) Collection() string {
	return ConsensusCollection
}

func (c *Consensus) CreateIndexes(_ context.Context, _ *mongo.Database) error {
	return nil
}

func (c *Consensus) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "ChainConsensus.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_consensus"))
	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	return s.PersistModel(ctx, c)
}

func (c *Consensus) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if c == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(c)}
}

func (c *Consensus) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if c == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(c.Collection()).BulkWrite(ctx,
		c.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type ConsensusList []*Consensus

func (c ConsensusList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "ChainConsensusList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(c)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "chain_consensus"))
	metrics.RecordCount(ctx, metrics.PersistModel, len(c))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(c) == 0 {
		return nil
	}
	return s.PersistModel(ctx, c)
}

func (c ConsensusList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(c) == 0 {
		return
	}

	for _, a := range c {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (c ConsensusList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(c) == 0 {
		return nil
	}

	_, err := d.Collection(c[0].Collection()).BulkWrite(ctx,
		c.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
