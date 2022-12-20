package evmmodel

// derived_gas_outputs message receipt
type GasOutputs struct {
	Height        int64  `xorm:"notnull default 0 pk"`
	Version       int    `xorm:"integer notnull default 0"`
	StateRoot     string `xorm:"notnull"`
	ParentBaseFee string `xorm:"numeric notnull"`
	Cid           string `xorm:"notnull"`
	From          string `xorm:"notnull"`
	To            string `xorm:"notnull"`
	Value         string `xorm:"numeric notnull"`
	GasFeeCap     string `xorm:"numeric notnull"`
	GasPremium    string `xorm:"numeric notnull"`
	GasLimit      int64  `xorm:"notnull default 0"`
	Nonce         uint64 `xorm:"notnull default 0"`
	// Method             uint64 `xorm:"notnull default 0"`
	// SizeBytes          int    `xorm:"notnull default 0"`
	Status             int64  `xorm:"notnull default 0"`
	GasUsed            int64  `xorm:"notnull default 0"`
	BaseFeeBurn        string `xorm:"numeric notnull"`
	OverEstimationBurn string `xorm:"numeric notnull"`
	MinerPenalty       string `xorm:"numeric notnull"`
	MinerTip           string `xorm:"numeric notnull"`
	Refund             string `xorm:"numeric notnull"`
	GasRefund          int64  `xorm:"notnull default 0"`
	GasBurned          int64  `xorm:"notnull default 0r"`
	ActorName          string `xorm:"notnull"`
	ActorFamily        string `xorm:"notnull"`
	CreatedAt          int64  `xorm:"created default CURRENT_TIMESTAMP"`
}

func (g *GasOutputs) TableName() string {
	return "evm_derived_gas_outputs"
}
