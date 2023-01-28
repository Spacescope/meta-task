package evmtask

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Receipt parse evm transaction receipt
type Receipt struct {
}

func (e *Receipt) Name() string {
	return "evm_receipt"
}

func (e *Receipt) Model() interface{} {
	return new(evmmodel.Receipt)
}

func (e *Receipt) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool,
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

	// TODO Should use pool be used to limit concurrency?
	grp := new(errgroup.Group)
	var (
		receipts []*evmmodel.Receipt
		lock     sync.Mutex
	)
	for _, transaction := range transactions {
		tm, ok := transaction.(map[string]interface{})
		if ok {
			tm := tm
			grp.Go(func() error {
				ethHash, err := ethtypes.ParseEthHash(tm["hash"].(string))
				if err != nil {
					return errors.Wrap(err, "EthAddressFromHex failed")
				}
				receipt, err := rpc.Node().EthGetTransactionReceipt(ctx, ethHash)
				if err != nil {
					return errors.Wrap(err, "EthGetTransactionReceipt failed")
				}

				r := &evmmodel.Receipt{
					Height:            int64(parentTs.Height()),
					Version:           version,
					TransactionHash:   receipt.TransactionHash.String(),
					TransactionIndex:  int64(receipt.TransactionIndex),
					BlockHash:         receipt.BlockHash.String(),
					BlockNumber:       int64(receipt.BlockNumber),
					From:              receipt.From.String(),
					Status:            int64(receipt.Status),
					CumulativeGasUsed: int64(receipt.CumulativeGasUsed),
					GasUsed:           int64(receipt.GasUsed),
					EffectiveGasPrice: receipt.EffectiveGasPrice.Int64(),
					LogsBloom:         hex.EncodeToString(receipt.LogsBloom),
				}

				b, _ := json.Marshal(receipt.Logs)
				r.Logs = string(b)
				if receipt.ContractAddress != nil {
					r.ContractAddress = receipt.ContractAddress.String()
				}
				if receipt.To != nil {
					r.To = receipt.To.String()
				}

				lock.Lock()
				receipts = append(receipts, r)
				lock.Unlock()
				return nil
			})
		}
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	if len(receipts) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.Receipt), int64(parentTs.Height()), version,
			&receipts); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d receipt", len(receipts))

	return nil
}
