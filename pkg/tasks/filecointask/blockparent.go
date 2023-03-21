package filecointask

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
	log "github.com/sirupsen/logrus"
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

func (b *BlockParent) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	if !force {
		existed, err := storage.Existed(b.Model(), int64(tipSet.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(), int64(tipSet.Height()), version)
			return nil
		}
	}

	var blockParents []*filecoinmodel.BlockParent
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
		if err := storage.Inserts(ctx, new(filecoinmodel.BlockParent), int64(tipSet.Height()), version, &blockParents); err != nil {
			return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		}
	}

	log.Infof("Tipset[%v] has been process %d blockParents", tipSet.Height(), len(blockParents))

	return nil
}
