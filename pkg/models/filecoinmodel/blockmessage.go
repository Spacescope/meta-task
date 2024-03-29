package filecoinmodel

import "time"

type BlockMessage struct {
	Height     int64     `xorm:"bigint notnull pk"`
	Version    int       `xorm:"integer notnull pk"`
	BlockCid   string    `xorm:"varchar(255) notnull pk"`
	MessageCid string    `xorm:"varchar(255) notnull pk"`
	CreatedAt  time.Time `xorm:"created"`
}

func (b *BlockMessage) TableName() string {
	return "block_message"
}
