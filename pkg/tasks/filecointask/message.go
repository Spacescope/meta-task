package filecointask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	log "github.com/sirupsen/logrus"

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

func (m *Message) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(m.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", m.Name(), int64(parentTs.Height()), version)
			return nil
		}
	}

	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, parentTs.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}
	var messageModels []*filecoinmodel.Message
	for _, message := range messages {
		messageModels = append(messageModels, &filecoinmodel.Message{
			Height:     int64(parentTs.Height()),
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
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.Message), int64(parentTs.Height()), version, &messageModels); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	log.Infof("Tipset[%v] has been process %d message", tipSet.Height(), len(messageModels))

	return nil
}
