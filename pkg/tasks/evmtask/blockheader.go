package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	log "github.com/sirupsen/logrus"
)

// BlockHeader extract block header for evm
type BlockHeader struct {
}

func (b *BlockHeader) Name() string {
	return "evm_block_header"
}

func (b *BlockHeader) Model() interface{} {
	return new(evmmodel.BlockHeader)
}

func (b *BlockHeader) Run(ctx context.Context, tp *common.TaskParameters) error {
	if !tp.Force {
		// existed, err := storage.Existed(b.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", b.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

	tipSetCid, err := tp.AncestorTs.Key().Cid()
	if err != nil {
		log.Errorf("ts.Key().Cid() err: %v", err)
		return err
	}

	hash, err := ethtypes.EthHashFromCid(tipSetCid)
	if err != nil {
		log.Errorf("ethtypes.EthHashFromCid err: %v", err)
		return err
	}

	ethBlock, err := tp.Api.EthGetBlockByHash(ctx, hash, false)
	if err != nil {
		log.Errorf("EthGetBlockByHash err: %v", err)
		return err
	}
	if ethBlock.Number == 0 {
		log.Infof("block number == 0")
		return nil
	}

	// blockHeader := &evmmodel.BlockHeader{
	_ = &evmmodel.BlockHeader{
		Height:           int64(tp.AncestorTs.Height()),
		Version:          tp.Version,
		Hash:             hash.String(),
		ParentHash:       ethBlock.ParentHash.String(),
		Miner:            ethBlock.Miner.String(),
		StateRoot:        ethBlock.StateRoot.String(),
		TransactionsRoot: ethBlock.TransactionsRoot.String(),
		ReceiptsRoot:     ethBlock.ReceiptsRoot.String(),
		Difficulty:       int64(ethBlock.Difficulty),
		Number:           int64(ethBlock.Number),
		GasLimit:         int64(ethBlock.GasLimit),
		GasUsed:          int64(ethBlock.GasUsed),
		Timestamp:        int64(ethBlock.Timestamp),
		ExtraData:        string(ethBlock.Extradata),
		MixHash:          ethBlock.MixHash.String(),
		Nonce:            ethBlock.Nonce.String(),
		BaseFeePerGas:    ethBlock.BaseFeePerGas.String(),
		Size:             uint64(ethBlock.Size),
		Sha3Uncles:       ethBlock.Sha3Uncles.String(),
	}

	// if err = storage.Insert(ctx, new(evmmodel.BlockHeader),
	// 	int64(parentTs.Height()), version, blockHeader); err != nil {
	// 	return errors.Wrap(err, "storageWrite failed")
	// }

	return nil
}
