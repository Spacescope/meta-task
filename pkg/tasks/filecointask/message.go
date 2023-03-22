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
		if err = common.InsertMany(ctx, new(filecoinmodel.Message), int64(tp.AncestorTs.Height()), tp.Version, &messageModels); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v message", len(messageModels))
	return nil
}
