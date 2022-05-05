package visor

import (
	"context"
	"fmt"
	"time"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	_                     = iota
	ProcessingStatusOK    // "OK"
	ProcessingStatusInfo  // "INFO"  Processing was successful but the task reported information in the StatusInformation column
	ProcessingStatusError // "ERROR" one or more errors were encountered, data may be incomplete
	ProcessingStatusSkip  //  "SKIP"  no processing was attempted, a reason may be given in the StatusInformation column
)

const (
	// ProcessingStatusInformationNullRound is set byt the consensus task to indicate a null round
	ProcessingStatusInformationNullRound = "NullRound" // used by consensus task to indicate a null round
	// TODO this could likely be a status of its own, but the indexer isn't currently suited for tasks to set their own status.
)

type ProcessingReport struct {
	Height            abi.ChainEpoch `bson:"height" json:"height,omitempty"`
	StateRoot         cid.Cid        `bson:"state_root" json:"state_root"`
	Reporter          string         `bson:"reporter" json:"reporter,omitempty"`
	Task              string         `bson:"task" json:"task,omitempty"`
	StartedAt         time.Time      `bson:"started_at" json:"started_at"`
	CompletedAt       time.Time      `bson:"completed_at" json:"completed_at"`
	Status            int            `bson:"status" json:"status,omitempty"`
	StatusInformation string         `bson:"status_information" json:"status_information,omitempty"`
	ErrorsDetected    interface{}    `bson:"errors_detected" json:"errors_detected,omitempty"`
}

func (p *ProcessingReport) Collection() string {
	return "visor_processing_reports"
}

func (p *ProcessingReport) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(p.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}, {"task", 1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *ProcessingReport) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, p.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (p *ProcessingReport) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "visor_processing_reports"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, p)
}

func (p *ProcessingReport) ToMongoWriteModel(upsert bool) []mongo.WriteModel {
	if p == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewReplaceOneModel().
		SetFilter(bson.D{{"height", p.Height}, {"task", p.Task}}).
		SetReplacement(p).SetUpsert(upsert),
	}
}

func (p *ProcessingReport) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if p == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(p.Collection()).BulkWrite(ctx,
		p.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type ProcessingReportList []*ProcessingReport

func (pl ProcessingReportList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(pl) == 0 {
		return nil
	}
	ctx, span := otel.Tracer("").Start(ctx, "ProcessingReportList.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(pl)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "visor_processing_reports"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(pl) == 0 {
		return nil
	}
	metrics.RecordCount(ctx, metrics.PersistModel, len(pl))
	return s.PersistModel(ctx, pl)
}

func (pl ProcessingReportList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(pl) == 0 {
		return
	}

	for _, a := range pl {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (pl ProcessingReportList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(pl) == 0 {
		return nil
	}

	_, err := d.Collection(pl[0].Collection()).BulkWrite(ctx,
		pl.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
