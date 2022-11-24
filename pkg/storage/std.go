package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/errors"
)

var _ Storage = (*Std)(nil)

// Std output terminal for debug
type Std struct {
}

func (s *Std) Name() string {
	return "std"
}

func (s *Std) InitFromConfig(ctx context.Context, storageCFG *config.Storage) error {
	return nil
}

func (s *Std) Existed(m interface{}, height int64, version int) (bool, error) {
	return false, nil
}

func (s *Std) DelOldVersionAndWrite(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
	b, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return errors.Wrap(err, "son.MarshalIndent failed")
	}
	fmt.Println(string(b))
	return nil
}

func (s *Std) DelOldVersionAndWriteMany(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
	return nil
}
