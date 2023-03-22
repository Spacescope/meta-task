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
	var blockParents []*filecoinmodel.BlockParent
	for _, bh := range tp.AncestorTs.Blocks() {
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
		if err := common.InsertMany(ctx, new(filecoinmodel.BlockParent), int64(tp.AncestorTs.Height()), tp.Version, &blockParents); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v block_parent", len(blockParents))
	return nil
}
