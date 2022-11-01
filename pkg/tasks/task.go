package tasks

import (
	"context"

	"github.com/Spacescore/observatory-task-server/pkg/storage"
	"github.com/filecoin-project/lotus/chain/types"
)

var taskMap = make(map[string]Task)

func Register(tasks ...Task) {
	for _, task := range tasks {
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
	Models() []interface{}
	Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet, storage storage.Storage) error
}
