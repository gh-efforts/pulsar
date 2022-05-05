package observed

import (
	"context"
	"time"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"

	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
)

type PeerAgent struct {
	SurveyerPeerID  string    `bson:"surveyer_peer_id" json:"surveyer_peer_id,omitempty"`
	ObservedAt      time.Time `bson:"observed_at" json:"observed_at"`
	RawAgent        string    `bson:"raw_agent" json:"raw_agent,omitempty"`
	NormalizedAgent string    `bson:"normalized_agent" json:"normalized_agent,omitempty"`
	Count           int64     `bson:"count" json:"count,omitempty"`
}

func (p *PeerAgent) Collection() string {
	return "surveyed_peer_agents"
}

func (p *PeerAgent) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(p.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"surveyer_peer_id", 1}, {"observed_at", -1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *PeerAgent) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "surveyed_peer_agents"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	return s.PersistModel(ctx, p)
}

func (p *PeerAgent) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if p == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().SetDocument(p)}
}

type PeerAgentList []*PeerAgent

func (l PeerAgentList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(l) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "PeerAgentList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(l)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "surveyed_peer_agents"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(l) == 0 {
		return nil
	}
	return s.PersistModel(ctx, l)
}

func (l PeerAgentList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(l) == 0 {
		return
	}

	for _, a := range l {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (l PeerAgentList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(l) == 0 {
		return nil
	}

	_, err := d.Collection(l[0].Collection()).BulkWrite(ctx,
		l.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
