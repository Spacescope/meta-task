package evmmodel

import "time"

type BlockHeader struct {
	Height           int64     `xorm:"bigint notnull pk"`
	Version          int       `xorm:"integer notnull pk"`
	Number           int64     `xorm:"bigint pk notnull default 0"`
	Hash             string    `xorm:"varchar(255) index notnull default ''"`
	ParentHash       string    `xorm:"varchar(255) index notnull default ''"`
	Sha3Uncles       string    `xorm:"varchar(255) notnull default ''"`
	Miner            string    `xorm:"varchar(255) notnull default ''"`
	StateRoot        string    `xorm:"varchar(255) notnull default ''"`
	TransactionsRoot string    `xorm:"varchar(255) notnull default ''"`
	ReceiptsRoot     string    `xorm:"varchar(255) notnull default ''"`
	Difficulty       int64     `xorm:"bigint notnull default 0"`
	GasLimit         int64     `xorm:"bigint notnull default 0"`
	GasUsed          int64     `xorm:"bigint notnull default 0"`
	Timestamp        int64     `xorm:"integer notnull default 0"`
	ExtraData        string    `xorm:"text notnull default ''"`
	MixHash          string    `xorm:"varchar(255) notnull default ''"`
	Nonce            string    `xorm:"varchar(255) notnull default ''"`
	BaseFeePerGas    string    `xorm:"varchar(100) notnull default '0'"`
	Size             uint64    `xorm:"bigint notnull default 0"`
	CreatedAt        time.Time `xorm:"created"`
}

func (b *BlockHeader) TableName() string {
	return "evm_block_header"
}
