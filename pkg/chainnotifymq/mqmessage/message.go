package mqmessage

import "github.com/segmentio/kafka-go"

type Message interface {
	Val() []byte
}

// NormalMessage normal message
type NormalMessage struct {
	Value []byte
}

func (n *NormalMessage) Val() []byte {
	return n.Value
}

// KafkaMessage kafka message
type KafkaMessage struct {
	kafka.Message
}

func (k *KafkaMessage) Val() []byte {
	return k.Value
}
