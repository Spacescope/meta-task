package tasks

import (
	"github.com/Spacescore/observatory-task-server/pkg/tasks/blockheader"
	"github.com/Spacescore/observatory-task-server/pkg/tasks/message"
)

func init() {
	Register(
		new(blockheader.BlockHeader),
		new(blockheader.EVMBlockHeader),
		new(message.Message),
		new(message.EVMMessage),
	)
}
