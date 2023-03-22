package evmtask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
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

func (e *Receipt) Run(ctx context.Context, tp *common.TaskParameters) error {
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
	receipts := make([]*evmmodel.Receipt, 0)

	for _, transaction := range transactions {
		tm, ok := transaction.(map[string]interface{})
		if !ok {
			continue
		}

		ethHash, err := ethtypes.ParseEthHash(tm["hash"].(string))
		if err != nil {
			log.Errorf("ParseEthHash[hash: %v] err: %v", tm["hash"].(string), err)
			continue
		}
		receipt, err := tp.Api.EthGetTransactionReceipt(ctx, ethHash)
		if err != nil {
			log.Errorf("EthGetTransactionReceipt[hash: %v] err: %v", ethHash.String(), err)
			continue
		}
		if receipt == nil {
			continue
		}

		r := &evmmodel.Receipt{
			Height:            int64(tp.AncestorTs.Height()),
			Version:           tp.Version,
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

		receipts = append(receipts, r)
	}

	if len(receipts) > 0 {
		if err = common.InsertMany(ctx, new(evmmodel.Receipt), int64(tp.AncestorTs.Height()), tp.Version, &receipts); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v evm_receipt", len(receipts))
	return nil
}
