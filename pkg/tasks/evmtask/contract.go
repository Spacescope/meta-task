package evmtask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	log "github.com/sirupsen/logrus"
)

type Contract struct {
}

func (c *Contract) Name() string {
	return "evm_contract"
}

func (c *Contract) Model() interface{} {
	return new(evmmodel.Contract)
}

func (c *Contract) Run(ctx context.Context, tp *common.TaskParameters) error {
	changedActors, err := tp.Api.StateChangedActors(ctx, tp.AncestorTs.ParentState(), tp.CurrentTs.ParentState())
	if err != nil {
		log.Errorf("StateChangedActors[pTs: %v, pHeight: %v, cTs: %v, cHeight: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), tp.CurrentTs.String(), tp.CurrentTs.Height(), err)
		return err
	}

	var contracts []*evmmodel.Contract

	for _, actor := range changedActors {
		if !common.NewCidCache(ctx, tp.Api).IsEVMActor(actor.Code) || actor.Address == nil {
			continue
		}

		address := *actor.Address
		actorState, err := tp.Api.StateGetActor(ctx, address, tp.AncestorTs.Key())
		if err != nil {
			log.Errorf("StateGetActor[addr: %v, ts: %v, height: %v] err: %v", address.String(), tp.AncestorTs.String(), tp.AncestorTs.Height(), err)
			continue
		}
		if actorState != nil {
			ethAddress, err := ethtypes.EthAddressFromFilecoinAddress(address)
			if err != nil {
				log.Errorf("EthAddressFromFilecoinAddress[addr: %v] err: %v", address.String(), err)
				continue
			}
			byteCode, err := tp.Api.EthGetCode(ctx, ethAddress, "pending")
			if err != nil {
				log.Errorf("EthGetCode[addr: %v] err: %v, ", ethAddress.String(), err)
				continue
			}
			contracts = append(contracts, &evmmodel.Contract{
				Height:          int64(tp.AncestorTs.Height()),
				Version:         tp.Version,
				FilecoinAddress: address.String(),
				Address:         ethAddress.String(),
				Balance:         actorState.Balance.String(),
				Nonce:           actorState.Nonce,
				ByteCode:        hex.EncodeToString(byteCode),
			})
		}
	}

	if len(contracts) > 0 {
		if err = common.InsertContracts(ctx, contracts); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has been process %v evm_contract", len(contracts))
	return nil
}
