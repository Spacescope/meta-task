package filecointask

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
)

// BlockHeader extract block header
type BlockHeader struct {
}

func (b *BlockHeader) Name() string {
	return "block_header"
}

func (b *BlockHeader) Model() interface{} {
	return filecoinmodel.BlockHeader{}
}

func (b *BlockHeader) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	var blockHeaders []*filecoinmodel.BlockHeader
	for _, bh := range tipSet.Blocks() {
		blockHeaders = append(
			blockHeaders, &filecoinmodel.BlockHeader{
				Version:         version,
				Cid:             bh.Cid().String(),
				Miner:           bh.Miner.String(),
				ParentWeight:    bh.ParentWeight.String(),
				ParentBaseFee:   bh.ParentBaseFee.String(),
				ParentStateRoot: bh.ParentStateRoot.String(),
				Height:          int64(bh.Height),
				WinCount:        bh.ElectionProof.WinCount,
				Timestamp:       bh.Timestamp,
				ForkSignaling:   bh.ForkSignaling,
			},
		)
	}

	if len(blockHeaders) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.BlockHeader), int64(tipSet.Height()), version, &blockHeaders); err != nil {
			return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		}
	}
	return nil
}
