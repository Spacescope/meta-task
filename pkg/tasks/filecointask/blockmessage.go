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
	var blockMessages []*filecoinmodel.BlockMessage

	for _, block := range tp.AncestorTs.Blocks() {
		msg, err := tp.Api.ChainGetBlockMessages(ctx, block.Cid())
		if err != nil {
			log.Errorf("ChainGetBlockMessages error: %v", err)
			continue
		}

		for _, message := range msg.SecpkMessages {
			bm := &filecoinmodel.BlockMessage{
				Height:     int64(block.Height),
				Version:    tp.Version,
				BlockCid:   block.Cid().String(),
				MessageCid: message.Cid().String(),
			}
			blockMessages = append(blockMessages, bm)
		}

		for _, message := range msg.BlsMessages {
			bm := &filecoinmodel.BlockMessage{
				Height:     int64(block.Height),
				Version:    tp.Version,
				BlockCid:   block.Cid().String(),
				MessageCid: message.Cid().String(),
			}
			blockMessages = append(blockMessages, bm)
		}
	}

	if err := common.InsertMany(ctx, new(filecoinmodel.BlockMessage), int64(tp.AncestorTs.Height()), tp.Version, &blockMessages); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v block_message", len(blockMessages))
	return nil
}
