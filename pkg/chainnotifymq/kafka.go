package chainnotifymq

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/chainnotifymq/mqmessage"
	"github.com/Spacescore/observatory-task/pkg/errors"

	"github.com/mitchellh/mapstructure"
	"github.com/segmentio/kafka-go"
)

type KafkaParams struct {
	Brokers []string
	GroupID string
}

type Kafka struct {
	reader *kafka.Reader
	topic  string
}

func (k *Kafka) Name() string {
	return "kafka"
}

func (k *Kafka) InitFromConfig(ctx context.Context, cfg *config.ChainNotify, queueName string) error {
	var (
		err    error
		params KafkaParams
	)

	if err = mapstructure.Decode(cfg.MQ.Params, &params); err != nil {
		return errors.Wrap(err, "mapstructure.Decode failed")
	}

	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: params.Brokers,
		GroupID: params.GroupID,
		Topic:   queueName,
	})

	return nil
}

func (k *Kafka) FetchMessage(ctx context.Context) (mqmessage.Message, error) {
	msg, err := k.reader.FetchMessage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "reader.FetchMessage failed")
	}
	return &mqmessage.KafkaMessage{Message: msg}, nil
}

func (k *Kafka) Close() error {
	return k.reader.Close()
}

func (k *Kafka) Commit(ctx context.Context, message mqmessage.Message) error {
	msg, ok := message.(*mqmessage.KafkaMessage)
	if !ok {
		return nil
	}
	if err := k.reader.CommitMessages(ctx, msg.Message); err != nil {
		return errors.Wrap(err, "CommitMessages failed")
	}
	return nil
}
