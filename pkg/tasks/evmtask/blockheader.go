package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
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

func (b *BlockHeader) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewGatewayRPCV1 failed")
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

	if ethBlock.Number > 0 {
		blockHeader := &evmmodel.BlockHeader{
			Height:           int64(tipSet.Height()),
			Version:          version,
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
		if err = storage.Write(ctx, blockHeader); err != nil {
			return errors.Wrap(err, "storageWrite failed")
		}
	}

	return nil
}
