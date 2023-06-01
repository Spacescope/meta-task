package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/go-address"
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

	var evmInternalTxns []*evmmodel.InternalTX

	for _, message := range messages {
		if message.Message == nil {
			continue
		}

		// ----- only handle the following cases: from -> to -----
		// a.deploy contract:              msg.sender -> f10(0x00) -- creation txn
		// b.call contract:                msg.sender -> contract
		// c.contract call contract:       contract -> contract
		// d.contract call normal address: contract -> address
		if message.Message.To != utils.MustMakeAddress(10) { // case b,c,d //builtintypes.EthereumAddressManagerActorID
			if isEVMActor, err := common.NewCidLRU(ctx, tp.Api).AtLeastOneAddressIsEVMActor(ctx, []address.Address{message.Message.From, message.Message.To}, tp.AncestorTs); err != nil || !isEVMActor {
				continue
			}
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

		// https://filecoinproject.slack.com/archives/CP50PPW2X/p1683294037978979?thread_ts=1683255378.588109&cid=CP50PPW2X
		// hash, err := ethtypes.EthHashFromCid(invocs.MsgCid)
		// if err != nil {
		// 	log.Errorf("EthHashFromCid[cid: %v] err: %v", invocs.MsgCid.String(), err)
		// 	continue
		// }

		for _, subCall := range invocs.ExecutionTrace.Subcalls {
			subInput := subCall.Msg

			from, err := ethtypes.EthAddressFromFilecoinAddress(subInput.From)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[from: %v] err: %v", subInput.From.String(), err)
				continue
			}
			to, err := ethtypes.EthAddressFromFilecoinAddress(subInput.To)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[to: %v] err: %v", subInput.To.String(), err)
				continue
			}

			internalTx := &evmmodel.InternalTX{
				Height:      int64(tp.AncestorTs.Height()),
				Version:     tp.Version,
				ParentHash:  parentHash.String(),
				From:        from.String(),
				To:          to.String(),
				Type:        uint64(subInput.Method),
				Value:       subInput.Value.String(),
				Params:      ethtypes.EthBytes(subInput.Params).String(),
				ParamsCodec: subInput.ParamsCodec,
			}

			evmInternalTxns = append(evmInternalTxns, internalTx)
		}
	}

	if err = common.InsertMany(ctx, new(evmmodel.InternalTX), int64(tp.AncestorTs.Height()), tp.Version, &evmInternalTxns); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v evm_internal_tx", len(evmInternalTxns))
	return nil
}
