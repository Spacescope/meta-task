package message

import (
	"context"

	"github.com/Spacescore/observatory-task-server/pkg/errors"
	"github.com/Spacescore/observatory-task-server/pkg/metrics"
	"github.com/Spacescore/observatory-task-server/pkg/models"
	"github.com/Spacescore/observatory-task-server/pkg/storage"
	"github.com/filecoin-project/lotus/api/client"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/prometheus/client_golang/prometheus"
)

type EVMMessage struct {
}

func (e *EVMMessage) Name() string {
	return "evm_message"
}

func (e *EVMMessage) Models() []interface{} {
	return []interface{}{new(models.EVMMessage)}
}

func (e *EVMMessage) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	timer := prometheus.NewTimer(metrics.TaskCost.WithLabelValues(e.Name()))
	defer timer.ObserveDuration()

	existed, err := storage.Existed(new(models.EVMMessage), int64(tipSet.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		return nil
	}

	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	ethBlock, err := node.EthGetBlockByNumber(ctx, tipSet.Height().String(), true)
	if err != nil {
		return errors.Wrap(err, "rpc EthGetBlockByHash failed")
	}
	var messageModels []interface{}
	for _, transaction := range ethBlock.Transactions {
		ethTx, ok := transaction.(api.EthTx)
		if ok {
			messageModels = append(messageModels, &models.EVMMessage{
				Height:               int64(tipSet.Height()),
				Version:              version,
				ChainID:              int64(ethTx.ChainID),
				Nonce:                int64(ethTx.Nonce),
				Hash:                 ethTx.Hash.String(),
				BlockHash:            ethTx.BlockHash.String(),
				BlockNumber:          int64(ethTx.BlockNumber),
				TransactionIndex:     int64(ethTx.TransactionIndex),
				From:                 ethTx.From.String(),
				To:                   ethTx.To.String(),
				Value:                ethTx.Value.Int64(),
				Type:                 int64(ethTx.Type),
				Input:                string(ethTx.Input),
				Gas:                  int64(ethTx.Gas),
				GasLimit:             int64(*ethTx.GasLimit),
				MaxFeePerGas:         ethTx.MaxFeePerGas.Int64(),
				MaxPriorityFeePerGas: ethTx.MaxPriorityFeePerGas.Int64(),
				V:                    string(ethTx.V),
				R:                    string(ethTx.R),
				S:                    string(ethTx.S),
			})
		}
	}

	if len(messageModels) > 0 {
		if err = storage.WriteMany(ctx, messageModels...); err != nil {
			return errors.Wrap(err, "")
		}
	}
	return nil
}
