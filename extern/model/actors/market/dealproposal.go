package market

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
	"go.opentelemetry.io/otel/attribute"

	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel"
)

const DealProposalCollection = "market_deal_proposals"

type DealProposal struct {
	DealID               abi.DealID            `bson:"_id" json:"deal_id,omitempty"`
	Height               abi.ChainEpoch        `bson:"height" json:"height,omitempty"`
	StateRoot            cid.Cid               `bson:"state_root" json:"state_root"`
	PaddedPieceSize      abi.PaddedPieceSize   `bson:"padded_piece_size" json:"padded_piece_size,omitempty"`
	UnPaddedPieceSize    abi.UnpaddedPieceSize `bson:"un_padded_piece_size" json:"un_padded_piece_size,omitempty"`
	StartEpoch           abi.ChainEpoch        `bson:"start_epoch" json:"start_epoch,omitempty"`
	EndEpoch             abi.ChainEpoch        `bson:"end_epoch" json:"end_epoch,omitempty"`
	ClientID             address.Address       `bson:"client_id" json:"client_id"`
	ProviderID           address.Address       `bson:"provider_id" json:"provider_id"`
	ClientCollateral     abi.TokenAmount       `bson:"client_collateral" json:"client_collateral"`
	ProviderCollateral   abi.TokenAmount       `bson:"provider_collateral" json:"provider_collateral"`
	StoragePricePerEpoch abi.TokenAmount       `bson:"storage_price_per_epoch" json:"storage_price_per_epoch"`
	PieceCID             cid.Cid               `bson:"piece_cid" json:"piece_cid"`
	IsVerified           bool                  `bson:"is_verified" json:"is_verified,omitempty"`
	Label                string                `bson:"label" json:"label,omitempty"`
}

func (dp *DealProposal) Collection() string {
	return DealProposalCollection
}

func (dp *DealProposal) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(dp.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}}},
		{
			Keys: bson.D{{"client_id", 1}}},
		{
			Keys: bson.D{{"provider_id", 1}}},
	})
	if err != nil {
		return err
	}
	return nil
}

func (dp *DealProposal) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "market_deal_proposals"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, dp)
}

func (dp *DealProposal) ToMongoWriteModel(_ bool) []mongo.WriteModel {
	if dp == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewInsertOneModel().
		SetDocument(dp)}
}

func (dp *DealProposal) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if dp == nil {
		return nil
	}
	_, err := d.Collection(dp.Collection()).BulkWrite(ctx,
		dp.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}

type DealProposals []*DealProposal

func (dps DealProposals) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, span := otel.Tracer("").Start(ctx, "MarketDealProposals.Persist")
	if span.IsRecording() {
		span.SetAttributes(attribute.Int("count", len(dps)))
	}
	defer span.End()

	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "market_deal_proposals"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	if len(dps) == 0 {
		return nil
	}

	metrics.RecordCount(ctx, metrics.PersistModel, len(dps))
	return s.PersistModel(ctx, dps)
}

func (dps DealProposals) ToMongoWriteModel(upsert bool) (resp []mongo.WriteModel) {
	if len(dps) == 0 {
		return
	}

	for _, a := range dps {
		resp = append(resp, a.ToMongoWriteModel(upsert)...)
	}
	return
}

func (dps DealProposals) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if len(dps) == 0 {
		return nil
	}

	_, err := d.Collection(dps[0].Collection()).BulkWrite(ctx,
		dps.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
