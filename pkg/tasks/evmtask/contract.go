package evmtask

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/types"
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

func (c *Contract) Run(ctx context.Context, lotusAddr string, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	var err error

	node, closer, err := client.NewFullNodeRPCV1(ctx, lotusAddr, nil)
	if err != nil {
		return errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	defer closer()

	// lazy init actor map
	if err = utils.InitActorCodeCidMap(ctx, node); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	parentTs, err := node.ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	actors, err := node.StateChangedActors(ctx, parentTs.ParentState(), tipSet.ParentState())
	if err != nil {
		return errors.Wrap(err, "StateChangedActors failed")
	}

	// TODO Should use pool to limit concurrency?
	grp := new(errgroup.Group)
	var (
		contracts []interface{}
		lock      sync.Mutex
	)
	for _, actor := range actors {
		if !utils.IsEVMActor(actor.Code) {
			continue
		}
		if actor.Address == nil {
			continue
		}
		actor := actor
		grp.Go(func() error {
			addr := *actor.Address
			ethAddress, err := api.EthAddressFromFilecoinAddress(addr)
			if err != nil {
				return errors.Wrap(err, "EthAddressFromFilecoinAddress failed")
			}
			byteCode, err := node.EthGetCode(ctx, ethAddress, "")
			if err != nil {
				return errors.Wrap(err, "EthGetCode failed")
			}
			lock.Lock()
			contracts = append(contracts, &evmmodel.Contract{
				Height:          int64(tipSet.Height()),
				Version:         version,
				FilecoinAddress: actor.Address.String(),
				Address:         ethAddress.String(),
				Balance:         actor.Balance.Int64(),
				Nonce:           actor.Nonce,
				ByteCode:        hex.EncodeToString(byteCode),
			})
			lock.Unlock()
			return nil
		})
	}

	if err = grp.Wait(); err != nil {
		return err
	}

	if len(contracts) > 0 {
		if err = storage.WriteMany(ctx, contracts...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d contract", len(contracts))

	return nil
}
