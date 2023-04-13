package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	log "github.com/sirupsen/logrus"
)

// InternalTx task for parse internal transaction
type InternalTx struct {
}

func (i *InternalTx) Name() string {
	return "evm_internal_tx"
}

func (i *InternalTx) Model() interface{} {
	return new(evmmodel.InternalTX)
}

func (i *InternalTx) Run(ctx context.Context, tp *common.TaskParameters) error {
	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset[ts: %v, height: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), err)
		return err
	}

	var (
		evmInternalTxns []*evmmodel.InternalTX
		sm              sync.Map
	)

	for _, message := range messages {
		if message.Message == nil {
			continue
		}

		isEVMActor, err := common.NewCidLRU(ctx, tp.Api).IsEVMActor(ctx, message.Message.To, tp.AncestorTs)
		if err != nil || !isEVMActor {
			continue
		}

		invocs, err := tp.Api.StateReplay(ctx, tp.AncestorTs.Key(), message.Cid)
		if err != nil {
			log.Errorf("StateReplay[ts: %v, height: %v, cid: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), message.Cid.String(), err)
			continue
		}
		parentHash, err := tp.Api.EthGetTransactionHashByCid(ctx, message.Cid)
		if err != nil {
			log.Errorf("EthGetTransactionHashByCid[ts: %v, height: %v, cid: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), message.Cid.String(), err)
			continue
		}
		for _, subCall := range invocs.ExecutionTrace.Subcalls {
			subMessage := subCall.Msg
			_, loaded := sm.LoadOrStore(subMessage.Cid().String(), true)
			if loaded {
				continue
			}

			from, err := ethtypes.EthAddressFromFilecoinAddress(subMessage.From)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[from: %v] err: %v", subMessage.From.String(), err)
				continue
			}
			to, err := ethtypes.EthAddressFromFilecoinAddress(subMessage.To)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[to: %v] err: %v", subMessage.To.String(), err)
				continue
			}

			hash, err := ethtypes.EthHashFromCid(subMessage.Cid())
			if err != nil {
				log.Errorf("EthHashFromCid[cid: %v] err: %v", subMessage.Cid().String(), err)
				continue
			}
			internalTx := &evmmodel.InternalTX{
				Height:     int64(tp.AncestorTs.Height()),
				Version:    tp.Version,
				Hash:       hash.String(),
				ParentHash: parentHash.String(),
				From:       from.String(),
				To:         to.String(),
				Type:       uint64(subMessage.Method),
				Value:      subMessage.Value.String(),
			}
			evmInternalTxns = append(evmInternalTxns, internalTx)
		}
	}

	if len(evmInternalTxns) > 0 {
		if err = common.InsertMany(ctx, new(evmmodel.InternalTX), int64(tp.AncestorTs.Height()), tp.Version, &evmInternalTxns); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v evm_internal_tx", len(evmInternalTxns))
	return nil
}
