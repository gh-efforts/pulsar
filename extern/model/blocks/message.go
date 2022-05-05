package blocks

import (
	"context"

	model "github.com/BitRainforest/filmeta-model"
	"github.com/BitRainforest/filmeta-model/metrics"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/tag"
)

type BlockMessage struct {
	Cid       cid.Cid         `bson:"_id" json:"cid"`
	Miner     address.Address `bson:"miner" json:"miner"`
	Height    abi.ChainEpoch  `bson:"height" json:"height"`
	BlsCids   []cid.Cid       `bson:"bls_cids" json:"bls_cids"`
	SecpkCids []cid.Cid       `bson:"secpk_cids" json:"secpk_cids"`
}

func NewBlockMessage(h *types.BlockHeader, blsCids, secpkCids []cid.Cid) (*BlockMessage, error) {
	return &BlockMessage{
		Cid:       h.Cid(),
		Miner:     h.Miner,
		Height:    h.Height,
		BlsCids:   blsCids,
		SecpkCids: secpkCids,
	}, nil
}

func (b *BlockMessage) Collection() string {
	return MessageCollection
}

func (b *BlockMessage) CreateIndexes(context.Context, *mongo.Database) error {
	return nil
}

func (b *BlockMessage) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "block_messages"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, b)
}

func (b *BlockMessage) ToMongoWriteModel(upsert bool) []mongo.WriteModel {
	if b == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewReplaceOneModel().
		SetFilter(bson.M{"_id": b.Cid}).
		SetReplacement(b).SetUpsert(upsert)}
}

func (b *BlockMessage) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if b == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(b.Collection()).BulkWrite(ctx,
		b.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
