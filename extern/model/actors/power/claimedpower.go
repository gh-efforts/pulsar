package power

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
)

const ActorClaimCollection = "actor_powers"

type ActorClaim struct {
	Height          abi.ChainEpoch  `bson:"height" json:"height,omitempty"`
	MinerID         address.Address `bson:"miner_id" json:"miner_id"`
	StateRoot       cid.Cid         `bson:"state_root" json:"state_root"`
	RawBytePower    abi.TokenAmount `bson:"raw_byte_power" json:"raw_byte_power"`
	QualityAdjPower abi.TokenAmount `bson:"quality_adj_power" json:"quality_adj_power"`
}

func (p *ActorClaim) Collection() string {
	return ActorClaimCollection
}

func (p *ActorClaim) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(p.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"miner_id", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *ActorClaim) EnableShard(ctx context.Context, dbName string, adminDB *mongo.Database) error {
	cmd := bson.D{
		{"shardCollection", fmt.Sprintf("%s.%s", dbName, p.Collection())},
		{Key: "key", Value: bson.D{bson.E{Key: "height", Value: 1}}},
	}
	return adminDB.RunCommand(ctx, cmd).Err()
}

func (p *ActorClaim) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "PowerActorClaim.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "power_actor_claims"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, p)
}

func (p *ActorClaim) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if p == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(p)}
}

func (p *ActorClaim) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if p == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(p.Collection()).BulkWrite(ctx,
		p.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type ActorClaimList []*ActorClaim

func (pl ActorClaimList) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "PowerActorClaimList.Persist")
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "power_actor_claims"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(pl) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(pl))
	return s.PersistModel(ctx, pl)
}

func (pl ActorClaimList) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(pl) == 0 {
		return
	}

	for _, a := range pl {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (pl ActorClaimList) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(pl) == 0 {
		return nil
	}

	_, err := d.Collection(pl[0].Collection()).BulkWrite(ctx,
		pl.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
