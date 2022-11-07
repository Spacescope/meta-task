package filecoinmodel

// Message raw filecoin message
type Message struct {
	Height     int64  `xorm:"bigint notnull pk"`
	Version    int    `xorm:"integer notnull pk"`
	Cid        string `xorm:"varchar(255) pk notnull index default ''"`
	From       string `xorm:"varchar(255) notnull index default ''"`
	To         string `xorm:"varchar(255) notnull index default ''"`
	Value      string `xorm:"varchar(100) notnull default '0'"`
	GasFeeCap  string `xorm:"varchar(100) notnull default '0'"`
	GasPremium string `xorm:"varchar(100) notnull default '0'"`
	GasLimit   int64  `xorm:"bigint notnull default 0"`
	SizeBytes  int    `xorm:"bigint notnull default 0"`
	Nonce      uint64 `xorm:"bigint notnull default 0"`
	Method     uint64 `xorm:"integer notnull default 0"`
	Params     string `xorm:"text notnull default ''"`
	CreatedAt  int64  `xorm:"created"`
}

func (m *Message) TableName() string {
	return "message"
}
