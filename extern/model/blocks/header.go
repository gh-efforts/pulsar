package blocks

import (
	"bytes"
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

const (
	HeaderCollection  = "block_headers"
	MessageCollection = "block_messages"
	RewardCollection  = "block_rewards"
)

type BlockHeader struct {
	Cid                   cid.Cid         `bson:"_id" json:"cid"`
	Height                abi.ChainEpoch  `bson:"height" json:"height"`
	Miner                 address.Address `bson:"miner" json:"miner"`
	Parents               []cid.Cid       `bson:"parents" json:"parents"`
	ParentWeight          abi.TokenAmount `bson:"parent_weight" json:"parent_weight"`
	ParentBaseFee         abi.TokenAmount `bson:"parent_base_fee" json:"parent_base_fee"`
	ParentStateRoot       cid.Cid         `bson:"parent_state_root" json:"parent_state_root"`
	ParentMessageReceipts cid.Cid         `bson:"parent_message_receipts" json:"parent_message_receipts"`
	Messages              cid.Cid         `bson:"messages" json:"messages"`
	MessagesCount         int64           `bson:"messages_count" json:"messages_count"`
	Ticket                []byte          `bson:"ticket" json:"ticket"`
	ElectionProof         []byte          `bson:"election_proof" json:"election_proof"`
	WinCount              int64           `bson:"win_count" json:"win_count"`
	ForkSignaling         uint64          `bson:"fork_signaling" json:"fork_signaling"`
	BlockSig              []byte          `bson:"block_sig" json:"block_sig"`
	BLSAggregate          []byte          `bson:"bls_aggregate" json:"bls_aggregate"`
	BeaconEntries         [][]byte        `bson:"beacon_entries" json:"beacon_entries"`
	WinPoStProof          [][]byte        `bson:"win_po_st_proof" json:"win_po_st_proof"`
	Size                  int64           `bson:"size" json:"size"`
	Timestamp             uint64          `bson:"timestamp" json:"timestamp"`
	Validated             bool            `bson:"validated" json:"-"`
}

func (b *BlockHeader) Collection() string {
	return HeaderCollection
}

func (b *BlockHeader) CreateIndexes(ctx context.Context, d *mongo.Database) error {
	_, err := d.Collection(b.Collection()).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{"height", 1}},
		},
		{
			Keys: bson.D{{"miner", 1}},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func NewBlockHeader(h *types.BlockHeader, blsCids, secpkCids []cid.Cid, validated bool) (*BlockHeader, error) {
	var ticket, ep, bs, ba bytes.Buffer
	if h.Ticket != nil {
		if err := h.Ticket.MarshalCBOR(&ticket); err != nil {
			return nil, err
		}
	}
	if h.ElectionProof != nil {
		if err := h.ElectionProof.MarshalCBOR(&ep); err != nil {
			return nil, err
		}
	}
	if h.BlockSig != nil {
		if err := h.BlockSig.MarshalCBOR(&bs); err != nil {
			return nil, err
		}
	}
	if h.BLSAggregate != nil {
		if err := h.BLSAggregate.MarshalCBOR(&ba); err != nil {
			return nil, err
		}
	}

	header := BlockHeader{
		Cid:                   h.Cid(),
		Height:                h.Height,
		Miner:                 h.Miner,
		Parents:               h.Parents,
		ParentWeight:          h.ParentWeight,
		ParentBaseFee:         h.ParentBaseFee,
		ParentStateRoot:       h.ParentStateRoot,
		ParentMessageReceipts: h.ParentMessageReceipts,
		Messages:              h.Messages,
		MessagesCount:         int64(len(blsCids) + len(secpkCids)),
		Ticket:                ticket.Bytes(),
		ElectionProof:         ep.Bytes(),
		WinCount:              h.ElectionProof.WinCount,
		ForkSignaling:         h.ForkSignaling,
		BlockSig:              bs.Bytes(),
		BLSAggregate:          ba.Bytes(),
		Timestamp:             h.Timestamp,
		Validated:             validated,
	}

	for _, bn := range h.BeaconEntries {
		var b bytes.Buffer
		if err := bn.MarshalCBOR(&b); err != nil {
			return nil, err
		}
		header.BeaconEntries = append(header.BeaconEntries, b.Bytes())
	}
	for _, wp := range h.WinPoStProof {
		var b bytes.Buffer
		if err := wp.MarshalCBOR(&b); err != nil {
			return nil, err
		}
		header.WinPoStProof = append(header.WinPoStProof, b.Bytes())
	}

	if sb, err := h.Serialize(); err != nil {
		return nil, err
	} else {
		header.Size = int64(len(sb))
	}

	return &header, nil
}

func (b *BlockHeader) Persist(ctx context.Context, s model.StorageBatch) error {
	ctx, _ = tag.New(ctx, tag.Upsert(metrics.Table, "block_headers"))
	stop := metrics.Timer(ctx, metrics.PersistDuration)
	defer stop()

	metrics.RecordCount(ctx, metrics.PersistModel, 1)
	return s.PersistModel(ctx, b)
}

func (b *BlockHeader) ToMongoWriteModel(upsert bool) []mongo.WriteModel {
	if b == nil {
		// Nothing to do
		return nil
	}
	return []mongo.WriteModel{mongo.NewReplaceOneModel().
		SetFilter(bson.D{{"_id", b.Cid}}).
		SetReplacement(b).SetUpsert(upsert)}
}

func (b *BlockHeader) PersistToMongo(ctx context.Context, d *mongo.Database, upsert bool) error {
	if b == nil {
		// Nothing to do
		return nil
	}

	_, err := d.Collection(b.Collection()).BulkWrite(ctx,
		b.ToMongoWriteModel(upsert), options.BulkWrite().SetOrdered(false))
	return err
}
