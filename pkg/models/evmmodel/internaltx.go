package evmmodel

import "time"

// InternalTX contract internal transaction
type InternalTX struct {
	Height           int64     `xorm:"bigint notnull index"`
	Version          int       `xorm:"integer notnull"`
	ParentHash       string    `xorm:"varchar(255) notnull index default ''"`
	ParentMessageCid string    `xorm:"varchar(255) notnull index default ''"`
	Type             uint64    `xorm:"bigint notnull default 0"`
	From             string    `xorm:"varchar(255) notnull index default ''"`
	To               string    `xorm:"varchar(255) notnull index default ''"`
	Value            string    `xorm:"varchar(100) notnull default '0'"`
	Params           string    `xorm:"text notnull default ''"`
	ParamsCodec      uint64    `xorm:"bigint notnull default 0"`
	CreatedAt        time.Time `xorm:"created"`
}

func (i *InternalTX) TableName() string {
	return "evm_internal_tx"
}
