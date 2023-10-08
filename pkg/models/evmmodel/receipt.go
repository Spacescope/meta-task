package evmmodel

import "time"

// Receipt evm transaction receipt
type Receipt struct {
	Height            int64     `xorm:"bigint notnull pk"`
	Version           int       `xorm:"integer notnull pk"`
	TransactionHash   string    `xorm:"varchar(255) notnull pk"`
	TransactionIndex  int64     `xorm:"integer notnull default 0"`
	BlockHash         string    `xorm:"varchar(255) notnull default ''"`
	BlockNumber       int64     `xorm:"bigint notnull default 0"`
	From              string    `xorm:"varchar(255) index notnull default ''"`
	To                string    `xorm:"varchar(255) index notnull default ''"`
	StateRoot         string    `xorm:"varchar(255) notnull default ''"`
	Status            int64     `xorm:"integer notnull default 0"`
	ContractAddress   string    `xorm:"varchar(255) index notnull default ''"`
	CumulativeGasUsed int64     `xorm:"bigint notnull default 0"`
	GasUsed           int64     `xorm:"bigint notnull default 0"`
	EffectiveGasPrice int64     `xorm:"bigint notnull default 0"`
	LogsBloom         string    `xorm:"text notnull default ''"`
	Logs              string    `xorm:"text notnull default ''"`
	MessageCid        string    `xorm:"varchar(255) index notnull default ''"`
	CreatedAt         time.Time `xorm:"created"`
}

func (r *Receipt) TableName() string {
	return "evm_receipt"
}
