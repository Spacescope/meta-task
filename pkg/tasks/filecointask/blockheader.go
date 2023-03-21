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
	if !tp.Force {
		// existed, err := storage.Existed(b.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

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
		// if err := storage.Inserts(ctx, new(filecoinmodel.BlockHeader), int64(parentTs.Height()), version, &blockHeaders); err != nil {
		// 	return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		// }
	}

	log.Infof("Tipset[%v] has been process", tp.AncestorTs.Height())

	return nil
}
