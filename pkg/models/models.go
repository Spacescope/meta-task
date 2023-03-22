package models

import (
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
)

var (
	Tables []interface{}
)

func init() {
	Tables = append(Tables,
		new(evmmodel.Address),
		new(evmmodel.BlockHeader),
		new(evmmodel.Contract),
		new(evmmodel.InternalTX),
		new(evmmodel.Receipt),
		new(evmmodel.Transaction),

		new(filecoinmodel.BlockHeader),
		new(filecoinmodel.BlockMessage),
		new(filecoinmodel.BlockParent),
		new(filecoinmodel.Message),
		new(filecoinmodel.RawActor),
		new(filecoinmodel.Receipt),
	)
}
