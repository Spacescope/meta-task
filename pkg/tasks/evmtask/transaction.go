package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
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
	tipSetCid, err := tp.AncestorTs.Key().Cid()
	if err != nil {
		log.Errorf("ts.Key().Cid()[ts: %v] err: %v", tp.AncestorTs.String(), err)
		return err
	}

	hash, err := ethtypes.EthHashFromCid(tipSetCid)
	if err != nil {
		log.Errorf("EthHashFromCid[cid: %v] err: %v", tipSetCid.String(), err)
		return err
	}
	ethBlock, err := tp.Api.EthGetBlockByHash(ctx, hash, true)
	if err != nil {
		log.Errorf("EthGetBlockByHash[hash: %v] err: %v", hash.String(), err)
		return err
	}

	if ethBlock.Number == 0 {
		log.Warn("block number == 0")
		return nil
	}

	transactions := ethBlock.Transactions
	evmTransaction := make([]*evmmodel.Transaction, 0)

	for _, transaction := range transactions {
		tm := transaction.(map[string]interface{})

		et := &evmmodel.Transaction{
			Height:               int64(tp.AncestorTs.Height()),
			Version:              tp.Version,
			Hash:                 tm["hash"].(string),
			BlockHash:            tm["blockHash"].(string),
			From:                 tm["from"].(string),
			Value:                utils.ParseHexToBigInt(tm["value"].(string)).String(),
			MaxFeePerGas:         utils.ParseHexToBigInt(tm["maxFeePerGas"].(string)).String(),
			MaxPriorityFeePerGas: utils.ParseHexToBigInt(tm["maxPriorityFeePerGas"].(string)).String(),
		}

		if _, ok := tm["to"]; ok {
			v, ok := tm["to"].(string)
			if ok {
				et.To = v
			}
		}

		et.ChainID, err = utils.ParseHexToUint64(tm["chainId"].(string))
		if err != nil {
			log.Errorf("ParseHexToUint64 err: %v", err)
			continue
		}
		et.Nonce, err = utils.ParseHexToUint64(tm["nonce"].(string))
		if err != nil {
			log.Errorf("ParseHexToUint64 err: %v", err)
			continue
		}
		et.BlockNumber, err = utils.ParseHexToUint64(tm["blockNumber"].(string))
		if err != nil {
			log.Errorf("ParseHexToUint64 err: %v", err)
			continue
		}
		et.TransactionIndex, err = utils.ParseHexToUint64(tm["transactionIndex"].(string))
		if err != nil {
			log.Errorf("ParseHexToUint64 err: %v", err)
			continue
		}
		et.Type, err = utils.ParseHexToUint64(tm["type"].(string))
		if err != nil {
			log.Errorf("ParseHexToUint64 err: %v", err)
			continue
		}
		et.GasLimit, err = utils.ParseHexToUint64(tm["gas"].(string))
		if err != nil {
			log.Errorf("ParseHexToUint64 err: %v", err)
			continue
		}
		et.V, err = utils.ParseStrToHex(tm["v"].(string))
		if err != nil {
			log.Errorf("ParseStrToHex err: %v", err)
			continue
		}
		et.R, err = utils.ParseStrToHex(tm["r"].(string))
		if err != nil {
			log.Errorf("ParseStrToHex err: %v", err)
			continue
		}
		et.S, err = utils.ParseStrToHex(tm["s"].(string))
		if err != nil {
			log.Errorf("ParseStrToHex err: %v", err)
			continue
		}
		et.Input, err = utils.ParseStrToHex(tm["input"].(string))
		if err != nil {
			log.Errorf("ParseStrToHex err: %v", err)
			continue
		}

		evmTransaction = append(evmTransaction, et)
	}

	if len(evmTransaction) > 0 {
		if err = common.InsertMany(ctx, new(evmmodel.Transaction), int64(tp.CurrentTs.Height()), tp.Version, &evmTransaction); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v evm_transaction", len(evmTransaction))
	return nil
}
