package visor

import (
	"context"
	"fmt"
	"time"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	_            = iota
	GapStatusGap //"GAP"
	GapStatusFilled
)

const GapReportCollection = "visor_gap_reports"

type GapReport struct {
	Height     abi.ChainEpoch `bson:"height" json:"height,omitempty"`
	Task       string         `bson:"task" json:"task,omitempty"`
	Status     int            `bson:"status" json:"status,omitempty"`
	Reporter   string         `bson:"reporter" json:"reporter,omitempty"`
	ReportedAt time.Time      `bson:"reported_at" json:"reported_at"`
}

func (p *GapReport) Collection() string {
	return GapReportCollection
}

func (p *GapReport) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(p.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"task", 1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *GapReport) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if p == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(p.Collection()).BulkWrite(ctx,
		p.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

func (p *GapReport) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, p.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (p *GapReport) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "visor_gap_reports"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, p)
}

func (p *GapReport) ToMongoWriteModel(upsert bool) []mongo.WriteModel {
	if p == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewReplaceOneModel().
		SetFilter(bson.D{{"height", p.Height}, {"task", p.Task}}).
		SetReplacement(p).SetUpsert(upsert),
	}
}

type GapReportList []*GapReport

func (pl GapReportList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(pl) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "GapReportList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(pl)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "visor_gap_reports"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(pl) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(pl))
	return s.PersistModel(ctx, pl)
}

func (pl GapReportList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(pl) == 0 {
		return
	}

	for _, a := range pl {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (pl GapReportList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(pl) == 0 {
		return nil
	}

	_, err := d.Collection(pl[0].Collection()).BulkWrite(ctx,
		pl.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
