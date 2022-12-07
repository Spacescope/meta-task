package filecoinmodel

type RawActor struct {
	Height    int64  `xorm:"bigint notnull pk"`
	Version   int    `xorm:"integer notnull pk"`
	Address   string `xorm:"varchar(255) notnull pk"`
	StateRoot string `xorm:"varchar(255) notnull pk"`
	Code      string `xorm:"varchar(255) notnull pk"`
	Name      string `xorm:"varchar(100) notnull default ''"`
	Head      string `xorm:"varchar(255) notnull default ''"`
	Balance   string `xorm:"varchar(100) notnull default ''"`
	Nonce     uint64 `xorm:"bigint notnull default 0"`
}

func (r *RawActor) TableName() string {
	return "raw_actor"
}
