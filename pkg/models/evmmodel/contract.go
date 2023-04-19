package evmmodel

import "time"

// Contract evm smart contract
type Contract struct {
	Height          int64     `xorm:"bigint notnull pk"`
	Version         int       `xorm:"integer notnull pk"`
	Address         string    `xorm:"varchar(255) notnull pk"`
	FilecoinAddress string    `xorm:"varchar(255) notnull default ''"`
	Balance         string    `xorm:"varchar(100) notnull default '0'"`
	Nonce           uint64    `xorm:"bigint notnull default 0"`
	ByteCode        string    `xorm:"text notnull default ''"`
	CreatedAt       time.Time `xorm:"created"`
}

func (c *Contract) TableName() string {
	return "evm_contract"
}
