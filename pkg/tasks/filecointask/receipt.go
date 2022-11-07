package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/api/client"

	"github.com/filecoin-project/lotus/chain/types"
)

// Receipt message receipt
type Receipt struct {
}

func (r *Receipt) Name() string {
	return "receipt"
}

func (r *Receipt) Model() interface{} {
	return new(filecoinmodel.Receipt)
}

func (r *Receipt) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	messages, err := node.ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}

	var receiptModels []interface{}
	for idx, message := range messages {
		msgLookup, err := node.StateSearchMsg(ctx, types.EmptyTSK, message.Cid, -1, true)
		if err != nil {
			return errors.Wrap(err, "rpcv1/StateSearchMsg failed")
		}
		receiptModels = append(receiptModels, &filecoinmodel.Receipt{
			Height:     int64(tipSet.Height()),
			Version:    version,
			MessageCID: message.Message.Cid().String(),
			StateRoot:  tipSet.ParentState().String(),
			Idx:        idx,
			ExitCode:   int64(msgLookup.Receipt.ExitCode),
			GasUsed:    msgLookup.Receipt.GasUsed,
		})
	}

	if len(receiptModels) > 0 {
		if err = storage.WriteMany(ctx, receiptModels...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}
	return nil
}
