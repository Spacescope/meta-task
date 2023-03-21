package filecointask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	log "github.com/sirupsen/logrus"
)

type Message struct {
}

func (m *Message) Name() string {
	return "message"
}

func (m *Message) Model() interface{} {
	return new(filecoinmodel.Message)
}

func (m *Message) Run(ctx context.Context, tp *common.TaskParameters) error {
	if !tp.Force {
		// existed, err := storage.Existed(m.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", m.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset err: %v", err)
		return err
	}
	var messageModels []*filecoinmodel.Message
	for _, message := range messages {
		messageModels = append(messageModels, &filecoinmodel.Message{
			Height:     int64(tp.AncestorTs.Height()),
			Version:    tp.Version,
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
		// if err := storage.Inserts(ctx, new(filecoinmodel.Message), int64(parentTs.Height()), version, &messageModels); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}

	log.Infof("Tipset[%v] has been process %d message", tp.AncestorTs.Height(), len(messageModels))

	return nil
}
