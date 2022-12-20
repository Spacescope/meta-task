package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/tasks/evmtask/gasoutput"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/chain/types"
	"xorm.io/xorm"

	"github.com/sirupsen/logrus"
)

type GasOutput struct {
}

func (g *GasOutput) Name() string {
	return "evm_gas_output"
}

func (g *GasOutput) Model() interface{} {
	return new(evmmodel.GasOutputs)
}

func ancestralTipset(ctx context.Context, rpc *lotus.Rpc, t *types.TipSet, delay int) (*types.TipSet, error) {
	if t == nil || delay == 0 {
		return t, nil
	}

	if t.Height() == 0 {
		return t, nil
	}

	pTs, err := rpc.Node().ChainGetTipSet(ctx, t.Parents())
	if err != nil {
		return t, err
	}

	delay--
	return ancestralTipset(ctx, rpc, pTs, delay)
}

// In order to reuse the EVMBlockHeader/EVMMessage/EVMReceipt task's results, the GasOutput job may delay several Epochs to make sure three tasks are finished.
// There will raise a flaw: this will work fine in incremental synchronization, but not in walk(full synchronization), gas_output should walk after three tasks are finished.
func (g *GasOutput) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, storage storage.Storage) error {
	ts, err := ancestralTipset(ctx, rpc, tipSet, 120)
	if err != nil {
		return err
	}

	existed, err := storage.Existed(g.Model(), int64(ts.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		logrus.Infof("task [%s] has been process (%d,%d), ignore it", g.Name(), int64(ts.Height()), version)
		return nil
	}

	btr := new(gasoutput.BlockTransactionReceipt)
	err = btr.GetBlockTransactionReceipt(ctx, storage.ExposeEngine().(*xorm.Engine), int64(ts.Height()))
	if err != nil {
		return err
	}

	gasOutputSlice := make([]*evmmodel.GasOutputs, 0)

	for btr.HashNext() {
		t, r := btr.Next()

		baseFeePerGas, _ := big.FromString(btr.BlockHeader.BaseFeePerGas)
		gasFeeCap, _ := big.FromString(t.MaxFeePerGas)
		gasPremium, _ := big.FromString(t.MaxPriorityFeePerGas)

		gasOutputs := gasoutput.ComputeGasOutputs(r.GasUsed, int64(t.GasLimit), baseFeePerGas, gasFeeCap, gasPremium, true)

		gasOutput := evmmodel.GasOutputs{
			Height:        int64(ts.Height()),
			Version:       version,
			StateRoot:     btr.BlockHeader.StateRoot,
			ParentBaseFee: btr.BlockHeader.BaseFeePerGas,

			Cid:        t.Hash,
			From:       t.From,
			To:         t.To,
			Value:      t.Value,
			GasFeeCap:  t.MaxFeePerGas,
			GasPremium: t.MaxPriorityFeePerGas,
			GasLimit:   int64(t.GasLimit),
			Nonce:      t.Nonce,
			// Method:             r.Method,
			// SizeBytes:          message.SizeBytes,

			Status:             r.Status,
			GasUsed:            r.GasUsed,
			BaseFeeBurn:        gasOutputs.BaseFeeBurn.String(),
			OverEstimationBurn: gasOutputs.OverEstimationBurn.String(),
			MinerPenalty:       gasOutputs.MinerPenalty.String(),
			MinerTip:           gasOutputs.MinerTip.String(),
			Refund:             gasOutputs.Refund.String(),
			GasRefund:          gasOutputs.GasRefund,
			GasBurned:          gasOutputs.GasBurned,
			ActorName:          "",
			ActorFamily:        "",
		}

		gasOutputSlice = append(gasOutputSlice, &gasOutput)
	}

	if len(gasOutputSlice) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.GasOutputs), int64(ts.Height()), version, &gasOutputSlice); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	return nil
}
