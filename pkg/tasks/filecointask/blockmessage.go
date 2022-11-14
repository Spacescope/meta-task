package filecointask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"

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

func (b *BlockMessage) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	var (
		blockMessages []interface{}
		lock          sync.Mutex
	)

	grp := new(errgroup.Group)

	for _, block := range tipSet.Blocks() {
		grp.Go(func() error {
			msg, err := node.ChainGetBlockMessages(ctx, block.Cid())
			if err != nil {
				return errors.Wrap(err, "ChainGetBlockMessages failed")
			}
			for _, message := range msg.SecpkMessages {
				bm := &filecoinmodel.BlockMessage{
					Height:     int64(tipSet.Height()),
					Version:    version,
					BlockCid:   block.Cid().String(),
					MessageCid: message.Cid().String(),
				}
				lock.Lock()
				blockMessages = append(blockMessages, bm)
				lock.Unlock()
			}

			for _, message := range msg.BlsMessages {
				bm := &filecoinmodel.BlockMessage{
					Height:     int64(tipSet.Height()),
					Version:    version,
					BlockCid:   block.Cid().String(),
					MessageCid: message.Cid().String(),
				}
				lock.Lock()
				blockMessages = append(blockMessages, bm)
				lock.Unlock()
			}
			return nil
		})
	}

	if err = grp.Wait(); err != nil {
		return err
	}

	if len(blockMessages) > 0 {
		if err := storage.WriteMany(ctx, blockMessages...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	return nil
}
