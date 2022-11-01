package models

// Message raw filecoin message
type Message struct {
	Height     int64  `xorm:"integer notnull pk"`
	Version    int    `xorm:"integer notnull pk"`
	Cid        string `xorm:"varchar(255) pk notnull index default ''"`
	From       string `xorm:"varchar(255) notnull index default ''"`
	To         string `xorm:"varchar(255) notnull index default ''"`
	Value      int64  `xorm:"bigint notnull default 0"`
	GasFeeCap  int64  `xorm:"bigint notnull default 0"`
	GasPremium int64  `xorm:"bigint notnull default 0"`
	GasLimit   int64  `xorm:"bigint notnull default 0"`
	SizeBytes  int    `xorm:"integer notnull default 0"`
	Nonce      uint64 `xorm:"bigint notnull default 0"`
	Method     uint64 `xorm:"integer notnull default 0"`
	CreatedAt  int64  `xorm:"created"`
}

func (m *Message) TableName() string {
	return "message"
}

// EVMMessage evm transaction
type EVMMessage struct {
	Height               int64  `xorm:"integer notnull pk"`
	Version              int    `xorm:"integer notnull pk"`
	Hash                 string `xorm:"varchar(255) pk notnull index default ''"`
	ChainID              int64  `xorm:"integer notnull default 0"`
	Nonce                int64  `xorm:"bigint notnull default 0"`
	BlockHash            string `xorm:"varchar(255) notnull index default ''"`
	BlockNumber          int64  `xorm:"integer notnull default 0"`
	TransactionIndex     int64  `xorm:"integer notnull default 0"`
	From                 string `xorm:"varchar(255) notnull index default ''"`
	To                   string `xorm:"varchar(255) notnull index default ''"`
	Value                int64  `xorm:"bigint notnull default 0"`
	Type                 int64  `xorm:"integer notnull default 0"`
	Input                string `xorm:"varchar(255) notnull index default ''"`
	Gas                  int64  `xorm:"bigint notnull default 0"`
	GasLimit             int64  `xorm:"bigint notnull default 0"`
	MaxFeePerGas         int64  `xorm:"bigint notnull default 0"`
	MaxPriorityFeePerGas int64  `xorm:"bigint notnull default 0"`
	V                    string `xorm:"varchar(255) notnull index default ''"`
	R                    string `xorm:"varchar(255) notnull index default ''"`
	S                    string `xorm:"varchar(255) notnull index default ''"`
	CreatedAt            int64  `xorm:"created"`
}

func (m *EVMMessage) TableName() string {
	return "evm_message"
}
