package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	log "github.com/sirupsen/logrus"
)

type Transaction struct {
}

func (e *Transaction) Name() string {
	return "evm_transaction"
}

func (e *Transaction) Model() interface{} {
	return new(evmmodel.Transaction)
}

func (e *Transaction) Run(ctx context.Context, tp *common.TaskParameters) error {
	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset[ts: %v, height: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), err)
		return err
	}

	evmTransactions := make([]*evmmodel.Transaction, 0)
	for _, message := range messages {
		if message.Message == nil {
			continue
		}

		// determine if "to" is evm actor.
		isEVMActor, err := common.NewCidLRU(ctx, tp.Api).IsEVMActor(ctx, message.Message.To, tp.AncestorTs)
		if err != nil || !isEVMActor {
			continue
		}

		ethHash, err := ethtypes.EthHashFromCid(message.Cid)
		if err != nil {
			log.Errorf("EthHashFromCid[cid: %v] err: %v", message.Cid.String(), err)
			return err
		}

		evmTxn, err := tp.Api.EthGetTransactionByHash(ctx, &ethHash)
		if err != nil {
			log.Errorf("0 EthGetTransactionByHash[hash: %v] err: %v", ethHash)
			continue
		}

		evmTransaction := &evmmodel.Transaction{
			Height:               int64(tp.AncestorTs.Height()),
			Version:              tp.Version,
			Hash:                 evmTxn.Hash.String(),
			ChainID:              uint64(evmTxn.ChainID),
			Nonce:                uint64(evmTxn.Nonce),
			BlockHash:            evmTxn.BlockHash.String(),
			BlockNumber:          uint64(*evmTxn.BlockNumber),
			TransactionIndex:     uint64(*evmTxn.TransactionIndex),
			From:                 evmTxn.From.String(),
			To:                   evmTxn.To.String(),
			Value:                evmTxn.Value.String(),
			Type:                 uint64(evmTxn.Type),
			Input:                evmTxn.Input.String(),
			GasLimit:             uint64(evmTxn.Gas),
			MaxFeePerGas:         evmTxn.MaxFeePerGas.String(),
			MaxPriorityFeePerGas: evmTxn.MaxPriorityFeePerGas.String(),
			V:                    evmTxn.V.String(),
			R:                    evmTxn.R.String(),
			S:                    evmTxn.S.String(),
		}

		evmTransactions = append(evmTransactions, evmTransaction)
	}

	if len(evmTransactions) > 0 {
		if err = common.InsertMany(ctx, new(evmmodel.Transaction), int64(tp.AncestorTs.Height()), tp.Version, &evmTransactions); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v evm_transaction", len(evmTransactions))
	return nil
}
