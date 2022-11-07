package filecoinmodel

type BlockParent struct {
	Height    int64  `xorm:"bigint notnull pk"`
	Version   int    `xorm:"integer notnull pk"`
	Cid       string `xorm:"varchar(255) pk notnull index default ''"`
	ParentCid string `xorm:"varchar(255) notnull index default ''"`
	CreatedAt int64  `xorm:"created"`
}

func (b *BlockParent) TableName() string {
	return "block_parent"
}
