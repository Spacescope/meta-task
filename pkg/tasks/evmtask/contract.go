package evmtask

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Contract struct {
}

func (c *Contract) Name() string {
	return "evm_contract"
}

func (c *Contract) Model() interface{} {
	return new(evmmodel.Contract)
}

func (c *Contract) Run(ctx context.Context, rpc *lotus.Rpc,
	version int, tipSet *types.TipSet, storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	var err error

	// lazy init actor map
	if err = utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	existed, err := storage.Existed(c.Model(), int64(parentTs.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		logrus.Infof("task [%s] has been process (%d,%d), ignore it", c.Name(),
			int64(parentTs.Height()), version)
		return nil
	}

	changedActors, err := rpc.Node().StateChangedActors(ctx, parentTs.ParentState(), tipSet.ParentState())
	if err != nil {
		return errors.Wrap(err, "StateChangedActors failed")
	}

	var (
		contracts []*evmmodel.Contract
		lock      sync.Mutex
	)
	// TODO Should use pool be used to limit concurrency?
	grp := new(errgroup.Group)
	for _, actor := range changedActors {
		if utils.IsEVMActor(actor.Code) && actor.Address != nil {
			actor := actor
			grp.Go(func() error {
				address := *actor.Address
				actorState, err := rpc.Node().StateGetActor(ctx, address, tipSet.Key())
				if err != nil {
					logrus.Errorf("get actor err:%s", err)
					return nil
				}
				if actorState != nil {
					ethAddress, err := ethtypes.EthAddressFromFilecoinAddress(address)
					if err != nil {
						return errors.Wrap(err, "EthAddressFromFilecoinAddress failed")
					}
					byteCode, err := rpc.Node().EthGetCode(ctx, ethAddress, "latest")
					if err != nil {
						return errors.Wrap(err, "EthGetCode failed")
					}
					lock.Lock()
					contracts = append(contracts, &evmmodel.Contract{
						Height:          int64(parentTs.Height()),
						Version:         version,
						FilecoinAddress: address.String(),
						Address:         ethAddress.String(),
						Balance:         actorState.Balance.String(),
						Nonce:           actorState.Nonce,
						ByteCode:        hex.EncodeToString(byteCode),
					})
					lock.Unlock()
				}
				return nil
			})
		}
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	if len(contracts) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.Contract), int64(parentTs.Height()), version,
			&contracts); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d contract", len(contracts))

	return nil
}
