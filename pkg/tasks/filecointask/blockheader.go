package filecointask

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/sirupsen/logrus"
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

func (b *BlockHeader) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(b.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			logrus.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(),
				int64(parentTs.Height()), version)
			return nil
		}
	}

	var blockHeaders []*filecoinmodel.BlockHeader
	for _, bh := range parentTs.Blocks() {
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
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.BlockHeader), int64(parentTs.Height()),
			version, &blockHeaders); err != nil {
			return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		}
	}
	return nil
}
