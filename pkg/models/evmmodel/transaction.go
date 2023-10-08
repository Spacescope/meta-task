package evmmodel

import "time"

// Transaction evm transaction
type Transaction struct {
	Height               int64     `xorm:"bigint notnull pk"`
	Version              int       `xorm:"integer notnull pk"`
	Hash                 string    `xorm:"varchar(255) pk notnull index default ''"`
	ChainID              uint64    `xorm:"integer notnull default 0"`
	Nonce                uint64    `xorm:"bigint notnull default 0"`
	BlockHash            string    `xorm:"varchar(255) notnull index default ''"`
	BlockNumber          uint64    `xorm:"bigint notnull default 0"`
	TransactionIndex     uint64    `xorm:"integer notnull default 0"`
	From                 string    `xorm:"varchar(255) notnull index default ''"`
	To                   string    `xorm:"varchar(255) notnull index default ''"`
	Value                string    `xorm:"varchar(100) notnull default '0'"`
	Type                 uint64    `xorm:"integer notnull default 0"`
	Input                string    `xorm:"text notnull default ''"`
	Gas                  uint64    `xorm:"bigint notnull default 0"`
	MaxFeePerGas         string    `xorm:"varchar(100) notnull default ''"`
	MaxPriorityFeePerGas string    `xorm:"varchar(100) notnull default ''"`
	V                    string    `xorm:"varchar(255) notnull index default ''"`
	R                    string    `xorm:"varchar(255) notnull index default ''"`
	S                    string    `xorm:"varchar(255) notnull index default ''"`
	MessageCid           string    `xorm:"varchar(255) index notnull default ''"`
	CreatedAt            time.Time `xorm:"created"`
}

func (m *Transaction) TableName() string {
	return "evm_transaction"
}
