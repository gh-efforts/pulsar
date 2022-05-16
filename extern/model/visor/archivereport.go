package visor

import (
	"context"
	"time"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
)

const (
	_ = iota
	ArchiveMessage
	ArchiveSnapshot
)

const (
	ArchiveReportCollection = "visor_archive_reports"
)

type ArchiveReport struct {
	Height      abi.ChainEpoch `bson:"height" json:"height,omitempty"`
	Task        int            `bson:"task" json:"task,omitempty"`
	Status      int            `bson:"status" json:"status,omitempty"`
	StartedAt   time.Time      `bson:"started_at" json:"started_at"`
	CompletedAt time.Time      `bson:"completed_at" json:"completed_at"`
	Error       string         `bson:"error" json:"error,omitempty"`
}

func (p *ArchiveReport) Collection() string {
	return ArchiveReportCollection
}

func (p *ArchiveReport) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(p.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"task", 1}},
		},
		{
			Keys: bson.D{{"task", 1}, {"status", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *ArchiveReport) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "visor_archive_reports"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, p)
}

func (p *ArchiveReport) ToMongoWriteModel(upsert bool) []mongo.WriteModel {
	if p == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewReplaceOneModel().
		SetFilter(bson.D{{"height", p.Height}, {"task", p.Task}}).
		SetReplacement(p).SetUpsert(upsert),
	}
}

func (p *ArchiveReport) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if p == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(p.Collection()).BulkWrite(ctx,
		p.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
