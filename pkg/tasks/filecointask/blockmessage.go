package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	log "github.com/sirupsen/logrus"
)

// BlockMessage block and message
type BlockMessage struct {
}

func (b *BlockMessage) Name() string {
	return "block_message"
}

func (b *BlockMessage) Model() interface{} {
	return new(filecoinmodel.BlockMessage)
}

func (b *BlockMessage) Run(ctx context.Context, tp *common.TaskParameters) error {
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

	var blockMessages []*filecoinmodel.BlockMessage

	for _, block := range tp.AncestorTs.Blocks() {
		msg, err := tp.Api.ChainGetBlockMessages(ctx, block.Cid())
		if err != nil {
			log.Errorf("ChainGetBlockMessages error: %v", err)
			continue
		}
		for _, message := range msg.SecpkMessages {
			bm := &filecoinmodel.BlockMessage{
				Height:     int64(tp.AncestorTs.Height()),
				Version:    tp.Version,
				BlockCid:   block.Cid().String(),
				MessageCid: message.Cid().String(),
			}
			blockMessages = append(blockMessages, bm)
		}

		for _, message := range msg.BlsMessages {
			bm := &filecoinmodel.BlockMessage{
				Height:     int64(tp.AncestorTs.Height()),
				Version:    tp.Version,
				BlockCid:   block.Cid().String(),
				MessageCid: message.Cid().String(),
			}
			blockMessages = append(blockMessages, bm)
		}
	}

	if len(blockMessages) > 0 {
		// if err := storage.Inserts(ctx, new(filecoinmodel.BlockMessage), int64(parentTs.Height()), version, &blockMessages); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}

	log.Infof("Tipset[%v] has been process %d messages", tp.AncestorTs.Height(), len(blockMessages))

	return nil
}
