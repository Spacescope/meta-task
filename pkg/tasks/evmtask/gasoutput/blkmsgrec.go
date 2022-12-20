package gasoutput

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"
)

type BlockTransactionReceipt struct {
	BlockHeader evmmodel.BlockHeader
	Transaction []evmmodel.Transaction
	Receipt     []evmmodel.Receipt
	idx         int
}

func (btr *BlockTransactionReceipt) getEVMBlockHeader(ctx context.Context, engine *xorm.Engine, height int64) (*evmmodel.BlockHeader, bool) {
	evmBlockHeader := new(evmmodel.BlockHeader)
	b, err := engine.Where("height = ?", height).Get(evmBlockHeader)
	if err != nil {
		logrus.Infof("[%v] Executed sql error: %v", height, err)
		return nil, false
	}

	if b {
		return evmBlockHeader, true
	} else {
		return nil, false
	}
}

func (btr *BlockTransactionReceipt) getEVMTransaction(ctx context.Context, engine *xorm.Engine, height int64) ([]*evmmodel.Transaction, error) {
	evmTransactions := make([]*evmmodel.Transaction, 0)
	err := engine.Where("height = ?", height).Find(&evmTransactions)
	return evmTransactions, err
}

func (btr *BlockTransactionReceipt) getEVMReceipt(ctx context.Context, engine *xorm.Engine, height int64, txnHash string) (*evmmodel.Receipt, bool) {
	evmReceipt := new(evmmodel.Receipt)
	b, err := engine.Where("height = ? and transaction_hash = ?", height, txnHash).Get(evmReceipt)
	if err != nil {
		logrus.Infof("[%v] Executed sql error: %v", height, err)
		return nil, false
	}

	if b {
		return evmReceipt, true
	} else {
		return nil, false
	}
}

func (btr *BlockTransactionReceipt) GetBlockTransactionReceipt(ctx context.Context, engine *xorm.Engine, height int64) error {
	bh, b := btr.getEVMBlockHeader(ctx, engine, height)
	if !b {
		return nil
	}
	btr.BlockHeader = *bh

	txns, err := btr.getEVMTransaction(ctx, engine, height)
	if err != nil {
		return err
	}

	for _, txn := range txns {
		receipt, b := btr.getEVMReceipt(ctx, engine, height, txn.Hash)
		if !b {
			continue
		}

		btr.Transaction = append(btr.Transaction, *txn)
		btr.Receipt = append(btr.Receipt, *receipt)
	}

	return nil
}

func (btr *BlockTransactionReceipt) HashNext() bool {
	if btr.idx < len(btr.Transaction) {
		return true
	}

	return false
}

func (btr *BlockTransactionReceipt) Next() (*evmmodel.Transaction, *evmmodel.Receipt) {
	if btr.HashNext() {
		txn := btr.Transaction[btr.idx]
		receipt := btr.Receipt[btr.idx]
		btr.idx++
		return &txn, &receipt
	}

	return nil, nil
}
