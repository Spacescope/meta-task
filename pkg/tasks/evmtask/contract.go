package evmtask

import (
	"context"
	"encoding/hex"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/chain/types"
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

func (c *Contract) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(c.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", c.Name(), int64(parentTs.Height()), version)
			return nil
		}
	}

	changedActors, err := rpc.Node().StateChangedActors(ctx, parentTs.ParentState(), tipSet.ParentState())
	if err != nil {
		return errors.Wrap(err, "StateChangedActors failed")
	}

	var contracts []*evmmodel.Contract

	for _, actor := range changedActors {
		if utils.IsEVMActor(actor.Code) && actor.Address != nil {
			address := *actor.Address
			actorState, err := rpc.Node().StateGetActor(ctx, address, tipSet.Key())
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
				byteCode, err := rpc.Node().EthGetCode(ctx, ethAddress, "pending")
				if err != nil {
					log.Errorf("Get EthGetCode failed: %v, address: %v", err, ethAddress.String())
					continue
				}
				contracts = append(contracts, &evmmodel.Contract{
					Height:          int64(parentTs.Height()),
					Version:         version,
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
		if err := storage.Inserts(ctx, new(evmmodel.Contract), int64(parentTs.Height()), version, &contracts); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	log.Infof("Tipset[%v] has been process %d evm_contract", tipSet.Height(), len(contracts))

	return nil
}
