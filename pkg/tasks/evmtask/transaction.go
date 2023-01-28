package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/sirupsen/logrus"
)

type Transaction struct {
}

func (e *Transaction) Name() string {
	return "evm_transaction"
}

func (e *Transaction) Model() interface{} {
	return new(evmmodel.Transaction)
}

func (e *Transaction) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool,
	storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(e.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			logrus.Infof("task [%s] has been process (%d,%d), ignore it", e.Name(),
				int64(parentTs.Height()), version)
			return nil
		}
	}

	tipSetCid, err := parentTs.Key().Cid()
	if err != nil {
		return errors.Wrap(err, "tipSetCid failed")
	}

	hash, err := ethtypes.EthHashFromCid(tipSetCid)
	if err != nil {
		return errors.Wrap(err, "rpc EthHashFromCid failed")
	}
	ethBlock, err := rpc.Node().EthGetBlockByHash(ctx, hash, true)
	if err != nil {
		return errors.Wrap(err, "rpc EthGetBlockByHash failed")
	}

	if ethBlock.Number == 0 {
		return errors.Wrap(err, "block number must greater than zero")
	}

	transactions := ethBlock.Transactions
	if len(transactions) == 0 {
		logrus.Debugf("can not find any transaction")
		return nil
	}

	var evmTransaction []*evmmodel.Transaction
	for _, transaction := range transactions {
		tm := transaction.(map[string]interface{})

		et := &evmmodel.Transaction{
			Height:               int64(parentTs.Height()),
			Version:              version,
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
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.Nonce, err = utils.ParseHexToUint64(tm["nonce"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.BlockNumber, err = utils.ParseHexToUint64(tm["blockNumber"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.TransactionIndex, err = utils.ParseHexToUint64(tm["transactionIndex"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.Type, err = utils.ParseHexToUint64(tm["type"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}
		et.GasLimit, err = utils.ParseHexToUint64(tm["gas"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseHexToUint64 failed")
		}

		et.V, err = utils.ParseStrToHex(tm["v"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}
		et.R, err = utils.ParseStrToHex(tm["r"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}
		et.S, err = utils.ParseStrToHex(tm["s"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}
		et.Input, err = utils.ParseStrToHex(tm["input"].(string))
		if err != nil {
			return errors.Wrap(err, "ParseStrToHex failed")
		}

		evmTransaction = append(evmTransaction, et)
	}

	if len(evmTransaction) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.Transaction), int64(parentTs.Height()), version,
			&evmTransaction); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d transaction", len(evmTransaction))

	return nil
}
