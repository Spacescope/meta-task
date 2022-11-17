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

// BlockParent parse block parent
type BlockParent struct {
}

func (b *BlockParent) Name() string {
	return "block_parent"
}

func (b *BlockParent) Model() interface{} {
	return new(filecoinmodel.BlockParent)
}

func (b *BlockParent) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	var blockParents []interface{}
	for _, bh := range tipSet.Blocks() {
		for _, parent := range bh.Parents {
			blockParents = append(blockParents, &filecoinmodel.BlockParent{
				Height:    int64(bh.Height),
				Version:   version,
				Cid:       bh.Cid().String(),
				ParentCid: parent.String(),
			})
		}
	}
	if len(blockParents) > 0 {
		if err := storage.WriteMany(ctx, blockParents...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		}
	}
	return nil
}
