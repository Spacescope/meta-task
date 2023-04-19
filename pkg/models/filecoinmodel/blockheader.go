package filecoinmodel

import "time"

type BlockHeader struct {
	Height          int64     `xorm:"bigint notnull pk"`
	Version         int       `xorm:"integer notnull pk"`
	Cid             string    `xorm:"varchar(255) pk notnull index default ''"`
	Miner           string    `xorm:"varchar(255) notnull default ''"`
	ParentWeight    string    `xorm:"varchar(255) notnull default ''"`
	ParentBaseFee   string    `xorm:"varchar(255) notnull default ''"`
	ParentStateRoot string    `xorm:"varchar(255) notnull default ''"`
	WinCount        int64     `xorm:"integer notnull default 0"`
	Timestamp       uint64    `xorm:"integer notnull default 0"`
	ForkSignaling   uint64    `xorm:"integer notnull default 0"`
	CreatedAt       time.Time `xorm:"created"`
}

func (b *BlockHeader) TableName() string {
	return "block_header"
}
