package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	log "github.com/sirupsen/logrus"
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

func (b *BlockHeader) Run(ctx context.Context, tp *common.TaskParameters) error {
	var blockHeaders []*filecoinmodel.BlockHeader
	for _, bh := range tp.AncestorTs.Blocks() {
		blockHeaders = append(
			blockHeaders, &filecoinmodel.BlockHeader{
				Version:         tp.Version,
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
		if err := common.InsertMany(ctx, new(filecoinmodel.BlockHeader), int64(tp.CurrentTs.Height()), tp.Version, &blockHeaders); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v block_header", len(blockHeaders))
	return nil
}
