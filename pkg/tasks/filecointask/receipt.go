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

// Receipt message receipt
type Receipt struct {
}

func (r *Receipt) Name() string {
	return "receipt"
}

func (r *Receipt) Model() interface{} {
	return new(filecoinmodel.Receipt)
}

func (r *Receipt) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(r.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", r.Name(), int64(parentTs.Height()), version)
			return nil
		}
	}

	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, parentTs.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}

	var receiptModels []*filecoinmodel.Receipt
	for idx, message := range messages {
		msgLookup, err := rpc.Node().StateSearchMsg(ctx, types.EmptyTSK, message.Cid, -1, false)
		if err != nil {
			return errors.Wrap(err, "rpcv1/StateSearchMsg failed")
		}

		if msgLookup == nil {
			log.Infof("filecoin task, receipt StateSearchMsg return nil, height: %v, message.Cid: %v", parentTs.Height(), message.Cid.String())
			continue
		}

		receiptModels = append(receiptModels, &filecoinmodel.Receipt{
			Height:     int64(parentTs.Height()),
			Version:    version,
			MessageCID: message.Message.Cid().String(),
			StateRoot:  parentTs.ParentState().String(),
			Idx:        idx,
			ExitCode:   int64(msgLookup.Receipt.ExitCode),
			GasUsed:    msgLookup.Receipt.GasUsed,
		})
	}

	if len(receiptModels) > 0 {
		if err := storage.Inserts(ctx, new(filecoinmodel.Receipt), int64(parentTs.Height()), version, &receiptModels); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	log.Infof("Tipset[%v] has been process %d receipt", tipSet.Height(), len(receiptModels))

	return nil
}
