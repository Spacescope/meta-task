package evmtask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
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
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, tp.Api); err != nil {
		log.Errorf("InitActorCodeCidMap err: %v", err)
		return err
	}

	if !tp.Force {
		// existed, err := storage.Existed(c.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", c.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

	changedActors, err := tp.Api.StateChangedActors(ctx, tp.AncestorTs.ParentState(), tp.CurrentTs.ParentState())
	if err != nil {
		log.Errorf("StateChangedActors err: %v", err)
		return err
	}

	var contracts []*evmmodel.Contract

	for _, actor := range changedActors {
		if utils.IsEVMActor(actor.Code) && actor.Address != nil {
			address := *actor.Address
			actorState, err := tp.Api.StateGetActor(ctx, address, tp.AncestorTs.Key())
			if err != nil {
				log.Errorf("get actor err:%s", err)
				continue
			}
			if actorState != nil {
				ethAddress, err := ethtypes.EthAddressFromFilecoinAddress(address)
				if err != nil {
					log.Errorf("EthAddressFromFilecoinAddress: %v failed: %v", address, err)
					continue
				}
				byteCode, err := tp.Api.EthGetCode(ctx, ethAddress, "pending")
				if err != nil {
					log.Errorf("Get EthGetCode failed: %v, address: %v", err, ethAddress.String())
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
	}

	if len(contracts) > 0 {
		// if err := storage.Inserts(ctx, new(evmmodel.Contract), int64(parentTs.Height()), version, &contracts); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}

	log.Infof("Tipset[%v] has been process %d evm_contract", tp.AncestorTs.Height(), len(contracts))

	return nil
}
