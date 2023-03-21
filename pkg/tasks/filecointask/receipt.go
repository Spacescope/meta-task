package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
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

func (r *Receipt) Run(ctx context.Context, tp *common.TaskParameters) error {
	if !tp.Force {
		// existed, err := storage.Existed(r.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", r.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset err: %v", err)
		return err
	}

	var receiptModels []*filecoinmodel.Receipt
	for idx, message := range messages {
		msgLookup, err := tp.Api.StateSearchMsg(ctx, types.EmptyTSK, message.Cid, -1, false)
		if err != nil {
			log.Errorf("StateSearchMsg err: %v", err)
			return err
		}

		if msgLookup == nil {
			log.Infof("filecoin task, receipt StateSearchMsg return nil, height: %v, message.Cid: %v", tp.AncestorTs.Height(), message.Cid.String())
			continue
		}

		receiptModels = append(receiptModels, &filecoinmodel.Receipt{
			Height:     int64(tp.AncestorTs.Height()),
			Version:    tp.Version,
			MessageCID: message.Message.Cid().String(),
			StateRoot:  tp.AncestorTs.ParentState().String(),
			Idx:        idx,
			ExitCode:   int64(msgLookup.Receipt.ExitCode),
			GasUsed:    msgLookup.Receipt.GasUsed,
		})
	}

	if len(receiptModels) > 0 {
		// if err := storage.Inserts(ctx, new(filecoinmodel.Receipt), int64(parentTs.Height()), version, &receiptModels); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}

	log.Infof("Tipset[%v] has been process %d receipt", tp.AncestorTs.Height(), len(receiptModels))

	return nil
}
