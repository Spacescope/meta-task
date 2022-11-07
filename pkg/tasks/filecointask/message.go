package filecointask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/sirupsen/logrus"

	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
)

type Message struct {
}

func (m *Message) Name() string {
	return "message"
}

func (m *Message) Model() interface{} {
	return new(filecoinmodel.Message)
}

func (m *Message) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	messages, err := node.ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}
	var messageModels []interface{}
	for _, message := range messages {
		messageModels = append(messageModels, &filecoinmodel.Message{
			Height:     int64(tipSet.Height()),
			Version:    version,
			Cid:        message.Cid.String(),
			From:       message.Message.From.String(),
			To:         message.Message.To.String(),
			Value:      message.Message.Value.String(),
			GasFeeCap:  message.Message.GasFeeCap.String(),
			GasPremium: message.Message.GasPremium.String(),
			GasLimit:   message.Message.GasLimit,
			SizeBytes:  message.Message.ChainLength(),
			Nonce:      message.Message.Nonce,
			Method:     uint64(message.Message.Method),
			Params:     hex.EncodeToString(message.Message.Params),
		})
	}

	if len(messageModels) > 0 {
		if err := storage.WriteMany(ctx, messageModels...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d message", len(messageModels))

	return nil
}
