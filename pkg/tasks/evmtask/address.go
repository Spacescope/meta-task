package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/go-address"
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

		// ----- only handle the following cases: from -> to -----
		// a.deploy contract:              msg.sender -> f10(0x00) -- creation txn
		// b.call contract:                msg.sender -> contract
		// c.contract call contract:       contract -> contract
		// d.contract call normal address: contract -> address
		if message.Message.To != utils.MustMakeAddress(10) { // case b,c,d //builtintypes.EthereumAddressManagerActorID
			if isEVMActor, err := common.NewCidLRU(ctx, tp.Api).AtLeastOneAddressIsEVMActor(ctx, []address.Address{message.Message.From, message.Message.To}, tp.AncestorTs); err != nil || !isEVMActor {
				continue
			}
		}

		// remove duplicates
		from := message.Message.From
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

	if err = common.InsertMany(ctx, new(evmmodel.Address), int64(tp.AncestorTs.Height()), tp.Version, &evmAddresses); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has process %v evm_address", len(evmAddresses))
	return nil
}
