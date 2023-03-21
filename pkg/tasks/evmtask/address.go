package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
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

func (a *Address) Run(ctx context.Context, tp *common.TaskParameters) error {
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, tp.Api); err != nil {
		log.Errorf("InitActorCodeCidMap err: %v", err)
		return err
	}

	if !tp.Force {
		// existed, err := storage.Existed(a.Model(), int64(tipSet.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", a.Name(), int64(tipSet.Height()), version)
		// 	return nil
		// }
	}

	var (
		evmAddresses []*evmmodel.Address
		m            sync.Map
	)

	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset err: %v", err)
		return err
	}
	for _, message := range messages {
		if message.Message == nil {
			continue
		}
		msg := message.Message

		// -----------
		to := msg.To
		actor, err := tp.Api.StateGetActor(ctx, to, tp.AncestorTs.Key())
		if err != nil {
			log.Errorf("StateGetActor[tipset: %v, height: %v] failed err:%s", tp.AncestorTs.Key(), tp.AncestorTs.Height(), err)
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
		fromActor, err := tp.Api.StateGetActor(ctx, from, tp.AncestorTs.Key())
		if err != nil {
			log.Errorf("StateGetActor[from: %v, tipset: %v] failed: %v", from, tp.AncestorTs.Height(), err)
			continue
		}
		if utils.IsEVMActor(fromActor.Code) {
			log.Infof("from [%s] is evm, ignore", ethFromAddress)
			continue
		}
		address := &evmmodel.Address{
			Height:          int64(tp.AncestorTs.Height()),
			Version:         tp.Version,
			Address:         ethFromAddress.String(),
			FilecoinAddress: from.String(),
			Balance:         fromActor.Balance.String(),
			Nonce:           fromActor.Nonce,
		}
		evmAddresses = append(evmAddresses, address)
	}

	if len(evmAddresses) > 0 {
		// if err := storage.Inserts(ctx, new(evmmodel.Address), int64(tipSet.Height()), version, &evmAddresses); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}
	log.Infof("Tipset[%v] has been process %v evm_address", tp.AncestorTs.Height(), len(evmAddresses))
	return nil
}
