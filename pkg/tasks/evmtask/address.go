package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
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
	var (
		evmAddresses []*evmmodel.Address
		m            sync.Map
	)

	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset[ts: %v, height: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), err)
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
			log.Errorf("StateGetActor[ts: %v, height: %v] err: %v", tp.AncestorTs.Key(), tp.AncestorTs.Height(), err)
			continue
		}
		if to != builtintypes.EthereumAddressManagerActorAddr && !common.NewCidCache(ctx, tp.Api).IsEVMActor(actor.Code) {
			continue
		}

		// remove duplicates
		from := msg.From
		_, loaded := m.LoadOrStore(from, true)
		if loaded {
			continue
		}

		ethFromAddress, err := ethtypes.EthAddressFromFilecoinAddress(from)
		if err != nil {
			log.Errorf("EthAddressFromFilecoinAddress[from: %v] err: %v", from.String(), err)
			continue
		}
		fromActor, err := tp.Api.StateGetActor(ctx, from, tp.AncestorTs.Key())
		if err != nil {
			log.Errorf("StateGetActor[from: %v, ts: %v, height: %v] err: %v", from.String(), tp.AncestorTs.String(), tp.AncestorTs.Height(), err)
			continue
		}
		if common.NewCidCache(ctx, tp.Api).IsEVMActor(fromActor.Code) {
			log.Infof("from[%v] is evm, ignore", ethFromAddress.String())
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
		if err = common.InsertMany(ctx, new(evmmodel.Address), int64(tp.CurrentTs.Height()), tp.Version, &evmAddresses); err != nil {
			log.Errorf("Sql Engine err: %v", err)
			return err
		}
	}
	log.Infof("has process %v evm_address", len(evmAddresses))
	return nil
}
