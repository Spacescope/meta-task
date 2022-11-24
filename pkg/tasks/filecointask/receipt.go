package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
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

func (r *Receipt) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, storage storage.Storage) error {
	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}

	var receiptModels []*filecoinmodel.Receipt
	for idx, message := range messages {
		msgLookup, err := rpc.Node().StateSearchMsg(ctx, tipSet.Key(), message.Cid, -1, false)
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
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.Receipt), int64(tipSet.Height()), version, &receiptModels); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}
	return nil
}
