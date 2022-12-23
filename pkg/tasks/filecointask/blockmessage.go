package filecointask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
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

func (b *BlockMessage) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	existed, err := storage.Existed(b.Model(), int64(parentTs.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		logrus.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(),
			int64(parentTs.Height()), version)
		return nil
	}

	var (
		blockMessages []*filecoinmodel.BlockMessage
		lock          sync.Mutex
	)

	grp := new(errgroup.Group)

	for _, block := range parentTs.Blocks() {
		blockTmp := block
		grp.Go(func() error {
			msg, err := rpc.Node().ChainGetBlockMessages(ctx, blockTmp.Cid())
			if err != nil {
				return errors.Wrap(err, "ChainGetBlockMessages failed")
			}
			for _, message := range msg.SecpkMessages {
				bm := &filecoinmodel.BlockMessage{
					Height:     int64(parentTs.Height()),
					Version:    version,
					BlockCid:   blockTmp.Cid().String(),
					MessageCid: message.Cid().String(),
				}
				lock.Lock()
				blockMessages = append(blockMessages, bm)
				lock.Unlock()
			}

			for _, message := range msg.BlsMessages {
				bm := &filecoinmodel.BlockMessage{
					Height:     int64(parentTs.Height()),
					Version:    version,
					BlockCid:   blockTmp.Cid().String(),
					MessageCid: message.Cid().String(),
				}
				lock.Lock()
				blockMessages = append(blockMessages, bm)
				lock.Unlock()
			}
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	if len(blockMessages) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.BlockMessage), int64(parentTs.Height()),
			version, &blockMessages); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	return nil
}
