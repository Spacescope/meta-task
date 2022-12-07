package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Address struct {
}

func (a *Address) Name() string {
	return "evm_address"
}

func (a *Address) Model() interface{} {
	return new(evmmodel.Address)
}

func (a *Address) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	existed, err := storage.Existed(a.Model(), int64(tipSet.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		logrus.Infof("task [%s] has been process (%d,%d), ignore it", a.Name(),
			int64(tipSet.Height()), version)
		return nil
	}

	var (
		evmAddresses []*evmmodel.Address
		lock         sync.Mutex
		m            sync.Map
	)

	grp := new(errgroup.Group)
	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}
	for _, message := range messages {
		if message.Message == nil {
			continue
		}
		msg := message.Message
		grp.Go(func() error {
			to := msg.To
			actor, err := rpc.Node().StateGetActor(ctx, to, tipSet.Key())
			if err != nil {
				return errors.Wrap(err, "StateGetActor failed")
			}
			if to != builtintypes.EthereumAddressManagerActorAddr && !utils.IsEVMActor(actor.Code) {
				return nil
			}
			from := msg.From
			// 去重
			_, loaded := m.LoadOrStore(from, true)
			if loaded {
				return nil
			}

			ethFromAddress, err := api.EthAddressFromFilecoinAddress(from)
			if err != nil {
				return errors.Wrap(err, "EthAddressFromFilecoinAddress failed")
			}
			fromActor, err := rpc.Node().StateGetActor(ctx, from, tipSet.Key())
			if err != nil {
				return errors.Wrap(err, "StateGetActor failed")
			}
			address := &evmmodel.Address{
				Height:          int64(tipSet.Height()),
				Version:         version,
				Address:         ethFromAddress.String(),
				FilecoinAddress: from.String(),
				Balance:         fromActor.Balance.String(),
				Nonce:           fromActor.Nonce,
			}
			lock.Lock()
			evmAddresses = append(evmAddresses, address)
			lock.Unlock()
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	if len(evmAddresses) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.Address),
			int64(tipSet.Height()), version, &evmAddresses); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}
	logrus.Debugf("process %d evm_address", len(evmAddresses))
	return nil
}
