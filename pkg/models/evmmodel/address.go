package evmmodel

// Address evm address
type Address struct {
	Height          int64  `xorm:"bigint notnull pk"`
	Version         int    `xorm:"integer notnull pk"`
	Address         string `xorm:"varchar(255) notnull pk"`
	FilecoinAddress string `xorm:"varchar(255) notnull default ''"`
	Balance         string `xorm:"varchar(100) notnull default '0'"`
	Nonce           uint64 `xorm:"bigint notnull default 0"`
	CreatedAt       int64  `xorm:"created"`
}

func (a *Address) TableName() string {
	return "evm_address"
}
