package mqmessage

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
