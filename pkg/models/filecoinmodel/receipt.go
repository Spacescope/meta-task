package filecoinmodel

import "time"

// Receipt message receipt
type Receipt struct {
	Height     int64     `xorm:"bigint notnull pk"`
	Version    int       `xorm:"integer notnull pk"`
	MessageCID string    `xorm:"varchar(255) notnull pk"`
	StateRoot  string    `xorm:"varchar(255) notnull default ''"`
	Idx        int       `xorm:"bigint notnull pk default 0"`
	ExitCode   int64     `xorm:"integer notnull pk default 0"`
	GasUsed    int64     `xorm:"bigint notnull pk default 0"`
	CreatedAt  time.Time `xorm:"created"`
}

func (r *Receipt) TableName() string {
	return "receipt"
}
