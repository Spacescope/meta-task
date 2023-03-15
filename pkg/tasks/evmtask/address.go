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
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	log "github.com/sirupsen/logrus"
)

type Address struct {
}

func (a *Address) Name() string {
	return "evm_address"
}

func (a *Address) Model() interface{} {
	return new(evmmodel.Address)
}

func (a *Address) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	if !force {
		existed, err := storage.Existed(a.Model(), int64(tipSet.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", a.Name(), int64(tipSet.Height()), version)
			return nil
		}
	}

	var (
		evmAddresses []*evmmodel.Address
		m            sync.Map
	)

	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}
	for _, message := range messages {
		if message.Message == nil {
			continue
		}
		msg := message.Message

		// -----------
		to := msg.To
		actor, err := rpc.Node().StateGetActor(ctx, to, tipSet.Key())
		if err != nil {
			log.Errorf("StateGetActor[tipset: %v, height: %v] failed err:%s", tipSet.Key(), tipSet.Height(), err)
			continue
		}
		if to != builtintypes.EthereumAddressManagerActorAddr && !utils.IsEVMActor(actor.Code) {
			continue
		}
		from := msg.From
		// 去重
		_, loaded := m.LoadOrStore(from, true)
		if loaded {
			continue
		}

		ethFromAddress, err := ethtypes.EthAddressFromFilecoinAddress(from)
		if err != nil {
			log.Errorf("EthAddressFromFilecoinAddress: %v failed: %v", from, err)
			continue
		}
		fromActor, err := rpc.Node().StateGetActor(ctx, from, tipSet.Key())
		if err != nil {
			log.Errorf("StateGetActor[from: %v, tipset: %v] failed: %v", from, tipSet.Height(), err)
			continue
		}
		if utils.IsEVMActor(fromActor.Code) {
			log.Infof("from [%s] is evm, ignore", ethFromAddress)
			continue
		}
		address := &evmmodel.Address{
			Height:          int64(tipSet.Height()),
			Version:         version,
			Address:         ethFromAddress.String(),
			FilecoinAddress: from.String(),
			Balance:         fromActor.Balance.String(),
			Nonce:           fromActor.Nonce,
		}
		evmAddresses = append(evmAddresses, address)
	}

	if len(evmAddresses) > 0 {
		if err := storage.DelOldVersionAndWriteMany(ctx, new(evmmodel.Address),
			int64(tipSet.Height()), version, &evmAddresses); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}
	log.Infof("Tipset[%v] has been process %d evm_address", tipSet.Height(), len(evmAddresses))
	return nil
}
