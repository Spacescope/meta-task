package filecointask

import (
	"context"
	"fmt"

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

func (r *Receipt) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}

	var receiptModels []interface{}
	for idx, message := range messages {
		msgLookup, err := rpc.Node().StateSearchMsg(ctx, types.EmptyTSK, message.Cid, -1, true)
		if err != nil {
			return errors.Wrap(err, "rpcv1/StateSearchMsg failed")
		}
		if msgLookup == nil || (msgLookup.Message.String() != message.Message.Cid().String()) {
			if msgLookup != nil {
				return errors.New(fmt.Sprintf("msg look may be nil or message id not equal, old:%s, new:%s",
					message.Message.Cid(), msgLookup.Message.String()))
			}
			return errors.New("msglookup is nil")
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
