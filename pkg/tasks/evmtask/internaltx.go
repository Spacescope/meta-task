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
	if !tp.Force {
		// existed, err := storage.Existed(i.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", i.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset err: %v", err)
		return err
	}

	var (
		internalTxs []*evmmodel.InternalTX
		sm          sync.Map
	)

	for _, message := range messages {
		invocs, err := tp.Api.StateReplay(ctx, tp.AncestorTs.Key(), message.Cid)
		if err != nil {
			log.Errorf("StateReplay[message.Cid: %v] failed: %v", message.Cid.String(), err)
			continue
		}
		parentHash, err := tp.Api.EthGetTransactionHashByCid(ctx, message.Cid)
		if err != nil {
			log.Errorf("EthGetTransactionHashByCid[message.Cid: %v] failed: %v", message.Cid.String(), err)
			continue
		}
		for _, subCall := range invocs.ExecutionTrace.Subcalls {
			subMessage := subCall.Msg
			// filter same sub message
			_, loaded := sm.LoadOrStore(subMessage.Cid().String(), true)
			if loaded {
				continue
			}

			from, err := ethtypes.EthAddressFromFilecoinAddress(subMessage.From)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[From]: %v failed: %v", subMessage.From, err)
				continue
			}
			to, err := ethtypes.EthAddressFromFilecoinAddress(subMessage.To)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[To]: %v failed: %v", subMessage.To, err)
				continue
			}

			hash, err := ethtypes.EthHashFromCid(subMessage.Cid())
			if err != nil {
				log.Errorf("EthHashFromCid[%v] failed: %v", subMessage.Cid().String(), err)
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
			internalTxs = append(internalTxs, internalTx)
		}
	}

	if len(internalTxs) > 0 {
		// if err := storage.Inserts(ctx, new(evmmodel.InternalTX), int64(parentTs.Height()), version, &internalTxs); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}

	log.Infof("Tipset[%v] has been process %d evm internal transactions", tp.AncestorTs.Height(), len(internalTxs))
	return nil
}
