package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
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

func (b *BlockMessage) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
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
			log.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(), int64(parentTs.Height()), version)
			return nil
		}
	}

	var (
		blockMessages []*filecoinmodel.BlockMessage
	)

	for _, block := range parentTs.Blocks() {
		msg, err := rpc.Node().ChainGetBlockMessages(ctx, block.Cid())
		if err != nil {
			log.Errorf("ChainGetBlockMessages error: %v", err)
			continue
		}
		for _, message := range msg.SecpkMessages {
			bm := &filecoinmodel.BlockMessage{
				Height:     int64(parentTs.Height()),
				Version:    version,
				BlockCid:   block.Cid().String(),
				MessageCid: message.Cid().String(),
			}
			blockMessages = append(blockMessages, bm)
		}

		for _, message := range msg.BlsMessages {
			bm := &filecoinmodel.BlockMessage{
				Height:     int64(parentTs.Height()),
				Version:    version,
				BlockCid:   block.Cid().String(),
				MessageCid: message.Cid().String(),
			}
			blockMessages = append(blockMessages, bm)
		}
	}

	if len(blockMessages) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.BlockMessage), int64(parentTs.Height()), version, &blockMessages); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	log.Infof("Tipset[%v] has been process %d messages", tipSet.Height(), len(blockMessages))

	return nil
}
