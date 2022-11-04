package message

import (
	"context"

	"github.com/Spacescore/observatory-task-server/pkg/errors"
	"github.com/Spacescore/observatory-task-server/pkg/metrics"
	"github.com/Spacescore/observatory-task-server/pkg/models"
	"github.com/Spacescore/observatory-task-server/pkg/storage"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
)

type Message struct {
}

func (m *Message) Name() string {
	return "message"
}

func (m *Message) Models() []interface{} {
	return []interface{}{new(models.Message), new(models.InternalMessage), new(models.Receipt)}
}

func (m *Message) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	timer := prometheus.NewTimer(metrics.TaskCost.WithLabelValues(m.Name()))
	defer timer.ObserveDuration()

	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	messages, err := node.ChainGetMessagesInTipset(ctx, tipSet.Key())

	g := new(errgroup.Group)

	g.Go(func() error {
		if err = m.insertMessage(ctx, storage, messages, version, int64(tipSet.Height())); err != nil {
			return errors.Wrap(err, "insertMessage failed")
		}
		return nil
	})

	g.Go(func() error {
		if err = m.insertReceipt(ctx, node, storage, messages, version, tipSet); err != nil {
			return errors.Wrap(err, "insertReceipt failed")
		}
		return nil
	})

	if err = g.Wait(); err != nil {
		return err
	}

	return nil
}

func (m *Message) insertReceipt(ctx context.Context, node api.FullNode, storage storage.Storage,
	messages []api.Message, version int, tipSet *types.TipSet) error {
	height := int64(tipSet.Height())
	existed, err := storage.Existed(new(models.Receipt), height, version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		return nil
	}

	// var receiptModels []interface{}
	// for _, message := range messages {
	// 	msgLookup, err := node.StateSearchMsg(ctx, types.EmptyTSK, message.Cid, -1, true)
	// 	if err != nil {
	// 		return errors.Wrap(err, "rpcv1/StateSearchMsg failed")
	// 	}
	// 	receipt := &models.Receipt{
	// 		Height:     height,
	// 		Version:    version,
	// 		MessageCID: message.Message.Cid().String(),
	// 		StateRoot:  tipSet.ParentState().String(),
	// 		Idx:        message.Message.in,
	// 		ExitCode:   0,
	// 		GasUsed:    0,
	// 		CreatedAt:  0,
	// 	}
	// }

	return nil
}

func (m *Message) insertMessage(ctx context.Context, storage storage.Storage, messages []api.Message, version int,
	height int64) error {
	existed, err := storage.Existed(new(models.Message), height, version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		return nil
	}

	var messageModels []interface{}
	for _, message := range messages {
		messageModels = append(messageModels, &models.Message{
			Height:     height,
			Version:    version,
			Cid:        message.Cid.String(),
			From:       message.Message.From.String(),
			To:         message.Message.To.String(),
			Value:      message.Message.Value.Int64(),
			GasFeeCap:  message.Message.GasFeeCap.Int64(),
			GasPremium: message.Message.GasPremium.Int64(),
			GasLimit:   message.Message.GasLimit,
			SizeBytes:  message.Message.ChainLength(),
			Nonce:      message.Message.Nonce,
			Method:     uint64(message.Message.Method),
		})
	}

	if len(messageModels) > 0 {
		if err := storage.WriteMany(ctx, messageModels...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	return nil
}
