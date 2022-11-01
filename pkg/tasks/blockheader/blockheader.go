package blockheader

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task-server/pkg/errors"
	"github.com/Spacescore/observatory-task-server/pkg/metrics"
	"github.com/Spacescore/observatory-task-server/pkg/models"
	"github.com/Spacescore/observatory-task-server/pkg/storage"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/filecoin-project/lotus/chain/types"
)

// BlockHeader extract block header
type BlockHeader struct {
}

func (b *BlockHeader) Name() string {
	return "block_header"
}

func (b *BlockHeader) Models() []interface{} {
	return []interface{}{new(models.BlockHeader)}
}

func (b *BlockHeader) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	timer := prometheus.NewTimer(metrics.TaskCost.WithLabelValues(b.Name()))
	defer timer.ObserveDuration()

	existed, err := storage.Existed(new(models.BlockHeader), int64(tipSet.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		return nil
	}

	var bhs []interface{}
	for _, bh := range tipSet.Blocks() {
		bhs = append(
			bhs, &models.BlockHeader{
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

	if len(bhs) > 0 {
		if err := storage.WriteMany(ctx, bhs...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		}
	}
	return nil
}
