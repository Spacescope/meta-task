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

		ethHash, err := ethtypes.EthHashFromCid(message.Cid)
		if err != nil {
			log.Errorf("EthHashFromCid[cid: %v] err: %v", message.Cid.String(), err)
			return err
		}

		evmTxn, err := tp.Api.EthGetTransactionByHash(ctx, &ethHash)
		if err != nil {
			log.Errorf("EthGetTransactionByHash[hash: %v] err: %v", ethHash, err)
			continue
		}

		evmTransaction := &evmmodel.Transaction{
			Height:  int64(tp.AncestorTs.Height()),
			Version: tp.Version,
			Hash:    evmTxn.Hash.String(),
			ChainID: uint64(evmTxn.ChainID),
			Nonce:   uint64(evmTxn.Nonce),
			// BlockHash:            evmTxn.BlockHash.String(),
			// BlockNumber:          uint64(*evmTxn.BlockNumber),
			// TransactionIndex:     uint64(*evmTxn.TransactionIndex),
			From: evmTxn.From.String(),
			// To:                   evmTxn.To.String(),
			Value:                evmTxn.Value.String(),
			Type:                 uint64(evmTxn.Type),
			Input:                evmTxn.Input.String(),
			GasLimit:             uint64(evmTxn.Gas), // should be name GAS
			MaxFeePerGas:         evmTxn.MaxFeePerGas.String(),
			MaxPriorityFeePerGas: evmTxn.MaxPriorityFeePerGas.String(),
			V:                    evmTxn.V.String(),
			R:                    evmTxn.R.String(),
			S:                    evmTxn.S.String(),
		}

		if evmTxn.BlockHash != nil {
			evmTransaction.BlockHash = evmTxn.BlockHash.String()
		}
		if evmTxn.BlockNumber != nil {
			evmTransaction.BlockNumber = uint64(*evmTxn.BlockNumber)
		}
		if evmTxn.TransactionIndex != nil {
			evmTransaction.TransactionIndex = uint64(*evmTxn.TransactionIndex)
		}
		if evmTxn.To != nil {
			evmTransaction.To = evmTxn.To.String()
		}

		evmTransactions = append(evmTransactions, evmTransaction)
	}

	if err = common.InsertMany(ctx, new(evmmodel.Transaction), int64(tp.AncestorTs.Height()), tp.Version, &evmTransactions); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v evm_transaction", len(evmTransactions))
	return nil
}
