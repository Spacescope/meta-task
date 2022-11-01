package blockheader

import (
	"context"

	"github.com/Spacescore/observatory-task-server/pkg/errors"
	"github.com/Spacescore/observatory-task-server/pkg/metrics"
	"github.com/Spacescore/observatory-task-server/pkg/models"
	"github.com/Spacescore/observatory-task-server/pkg/storage"
	"github.com/filecoin-project/lotus/api/client"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/prometheus/client_golang/prometheus"
)

// EVMBlockHeader extract block header for evm
type EVMBlockHeader struct {
}

func (b *EVMBlockHeader) Name() string {
	return "evm_block_header"
}

func (b *EVMBlockHeader) Models() []interface{} {
	return []interface{}{new(models.EVMBlockHeader)}
}

func (b *EVMBlockHeader) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	timer := prometheus.NewTimer(metrics.TaskCost.WithLabelValues(b.Name()))
	defer timer.ObserveDuration()

	existed, err := storage.Existed(new(models.EVMBlockHeader), int64(tipSet.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		return nil
	}

	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewGatewayRPCV1 failed")
	}
	ethBlock, err := node.EthGetBlockByNumber(ctx, tipSet.Height().String(), true)
	if err != nil {
		return errors.Wrap(err, "rpc EthGetBlockByNumber failed")
	}
	defer closer()

	if ethBlock.Number > 0 {
		blockHeader := &models.EVMBlockHeader{
			Height:           int64(tipSet.Height()),
			Version:          version,
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
			BaseFeePerGas:    ethBlock.BaseFeePerGas.Int64(),
			Size:             uint64(ethBlock.Size),
			Sha3Uncles:       ethBlock.Sha3Uncles.String(),
		}
		if err = storage.Write(ctx, blockHeader); err != nil {
			return errors.Wrap(err, "storageWrite failed")
		}
	}

	return nil
}
