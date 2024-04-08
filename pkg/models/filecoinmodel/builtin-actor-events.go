package filecoinmodel

import "time"

type BuiltinActorEvents struct {
	Height     int64     `xorm:"bigint notnull index"`
	Version    int       `xorm:"integer notnull index"`
	MessageCid string    `xorm:"varchar(255) notnull index default ''"`
	Emitter    string    `xorm:"varchar(255) notnull index default ''"`
	EventEntry string    `xorm:"text"`
	CreatedAt  time.Time `xorm:"created"`
}

func (a *BuiltinActorEvents) TableName() string {
	return "builtin_actor_events"
}
