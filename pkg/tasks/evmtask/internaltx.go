package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
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

func (i *InternalTx) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(i.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", i.Name(), int64(parentTs.Height()), version)
			return nil
		}
	}

	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, parentTs.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}

	var (
		internalTxs []*evmmodel.InternalTX
		sm          sync.Map
	)

	for _, message := range messages {
		message := message
		invocs, err := rpc.Node().StateReplay(ctx, types.EmptyTSK, message.Cid)
		if err != nil {
			log.Errorf("StateReplay[message.Cid: %v] failed: %v", message.Cid.String(), err)
			continue
		}
		parentHash, err := rpc.Node().EthGetTransactionHashByCid(ctx, message.Cid)
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
				Height:     int64(parentTs.Height()),
				Version:    version,
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
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.InternalTX), int64(parentTs.Height()), version, &internalTxs); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	log.Infof("process %d evm internal transactions", len(internalTxs))
	return nil
}
