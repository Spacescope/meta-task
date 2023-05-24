package filecointask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
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
	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset[ts: %v]: %v", tp.AncestorTs.String(), err)
		return err
	}

	var receiptModels []*filecoinmodel.Receipt
	for idx, message := range messages {
		msgLookup, err := tp.Api.StateSearchMsg(ctx, tp.AncestorTs.Key(), message.Cid, -1, false)
		if err != nil {
			log.Errorf("StateSearchMsg[ts: %v, cid: %v] err: %v", tp.AncestorTs.String(), message.Cid.String(), err)
			return err
		}
		if msgLookup == nil {
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

	if err = common.InsertMany(ctx, new(filecoinmodel.Receipt), int64(tp.AncestorTs.Height()), tp.Version, &receiptModels); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v receipt", len(receiptModels))
	return nil
}
