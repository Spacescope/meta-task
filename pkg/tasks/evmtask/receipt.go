package evmtask

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
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

func (e *Receipt) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	tipSetCid, err := tipSet.Key().Cid()
	if err != nil {
		return errors.Wrap(err, "tipSetCid failed")
	}

	hash, err := api.EthHashFromCid(tipSetCid)
	if err != nil {
		return errors.Wrap(err, "rpc EthHashFromCid failed")
	}
	ethBlock, err := node.EthGetBlockByHash(ctx, hash, true)
	if err != nil {
		return errors.Wrap(err, "rpc EthGetBlockByHash failed")
	}

	transactions := ethBlock.Transactions
	if len(transactions) == 0 {
		logrus.Debugf("can not find any transaction")
		return nil
	}

	// TODO Should use pool be used to limit concurrency?
	grp := new(errgroup.Group)
	var (
		receipts []interface{}
		lock     sync.Mutex
	)
	for _, transaction := range transactions {
		tm, ok := transaction.(map[string]interface{})
		if ok {
			tm := tm
			grp.Go(func() error {
				ethHash, err := api.EthHashFromHex(tm["hash"].(string))
				if err != nil {
					return errors.Wrap(err, "EthAddressFromHex failed")
				}
				receipt, err := node.EthGetTransactionReceipt(ctx, ethHash)
				if err != nil {
					return errors.Wrap(err, "EthGetTransactionReceipt failed")
				}

				r := &evmmodel.Receipt{
					Height:            int64(tipSet.Height()),
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
		if err := storage.WriteMany(ctx, receipts...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d receipt", len(receipts))

	return nil
}
