package models

type Receipt struct {
	Height    int64
	Version   int64
	Message   string
	StateRoot string
	Idx       int
	ExitCode  int64
	GasUsed   int64
}

func (r *Receipt) TableName() string {
	return "receipt"
}

type EVMReceipt struct {
	Height           int64
	Version          int64
	TransactionHash  string
	TransactionIndex int64
	BlockHash        string
	BlockNumber      int64
	From             string
	To               string
	// Logs
	// LogsBloom
	StateRoot         string
	Status            int64
	ContractAddress   string
	CumulativeGasUsed int64
	GasUsed           int64
	EffectiveGasPrice int64
	LogsBloom         string
	Logs              []string
}

func (r *EVMReceipt) TableName() string {
	return "evm_receipt"
}
