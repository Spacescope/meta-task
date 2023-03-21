package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
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

func (b *BlockParent) Run(ctx context.Context, tp *common.TaskParameters) error {
	if !tp.Force {
		// existed, err := storage.Existed(b.Model(), int64(tipSet.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(), int64(tipSet.Height()), version)
		// 	return nil
		// }
	}

	var blockParents []*filecoinmodel.BlockParent
	for _, bh := range tp.CurrentTs.Blocks() {
		for _, parent := range bh.Parents {
			blockParents = append(blockParents, &filecoinmodel.BlockParent{
				Height:    int64(bh.Height),
				Version:   tp.Version,
				Cid:       bh.Cid().String(),
				ParentCid: parent.String(),
			})
		}
	}
	if len(blockParents) > 0 {
		// if err := storage.Inserts(ctx, new(filecoinmodel.BlockParent), int64(tipSet.Height()), version, &blockParents); err != nil {
		// 	return errors.Wrap(err, fmt.Sprintf("storage %s write failed", storage.Name()))
		// }
	}

	log.Infof("Tipset[%v] has been process %d blockParents", tp.CurrentTs.Height(), len(blockParents))

	return nil
}
