package storage

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
)

var storageMap = make(map[string]Storage)

func register(storages ...Storage) {
	for _, s := range storages {
		storageMap[s.Name()] = s
	}
}

// GetStorage get storage by name
func GetStorage(name string) Storage {
	return storageMap[name]
}

// Storage factory
type Storage interface {
	Name() string
	InitFromConfig(ctx context.Context, storageCFG *config.Storage) error
	Existed(m interface{}, height int64, version int) (bool, error)
	DelOldVersionAndWrite(ctx context.Context, t interface{}, height int64, version int, m interface{}) error
	DelOldVersionAndWriteMany(ctx context.Context, t interface{}, height int64, version int, m interface{}) error

	ExposeEngine() interface{}
}

// Database sync table when storage is database
type Database interface {
	Storage
	Sync(m ...interface{}) error
}
