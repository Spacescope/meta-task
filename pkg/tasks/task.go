package tasks

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/tasks/evmtask"
	"github.com/Spacescore/observatory-task/pkg/tasks/filecointask"
	"github.com/filecoin-project/lotus/chain/types"
)

var taskMap = make(map[string]Task)

func Register(tasks ...Task) {
	for _, task := range tasks {
		_, ok := taskMap[task.Name()]
		if ok {
			panic(fmt.Sprintf("task name %s conflict", task.Name()))
		}
		taskMap[task.Name()] = task
	}
}

// GetTask get task by name
func GetTask(name string) Task {
	return taskMap[name]
}

// Task interface for task
type Task interface {
	Name() string
	Model() interface{}
	Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, storage storage.Storage) error
}

func init() {
	Register(
		new(filecointask.BlockHeader),
		new(filecointask.BlockParent),
		new(filecointask.BlockMessage),
		new(filecointask.Message),
		new(filecointask.Receipt),

		new(evmtask.BlockHeader),
		new(evmtask.Transaction),
		new(evmtask.Receipt),
		new(evmtask.Contract),
		new(evmtask.InternalTx),
	)
}
