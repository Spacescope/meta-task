package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

type Contract struct {
	byteCodeCache *lru.Cache
}

func (c *Contract) Name() string {
	return "evm_contract"
}

func (c *Contract) Model() interface{} {
	return new(evmmodel.Contract)
}

func (c *Contract) getByteCode(ctx context.Context, lotus *lotusapi.FullNodeStruct, ethAddress ethtypes.EthAddress) (ethtypes.EthBytes, error) {
	if c.byteCodeCache == nil {
		var err error
		c.byteCodeCache, err = lru.New(512)
		if err != nil {
			log.Fatal(err)
		}
	}

	byteCodeFromCache, ok := c.byteCodeCache.Get(ethAddress)
	if ok {
		return byteCodeFromCache.(ethtypes.EthBytes), nil
	} else {
		pending := "pending"
		byteCode, err := lotus.EthGetCode(ctx, ethAddress, ethtypes.EthBlockNumberOrHash{PredefinedBlock: &pending})
		if err != nil {
			log.Errorf("EthGetCode[addr: %v] err: %v", ethAddress.String(), err)
			return nil, err
		} else {
			c.byteCodeCache.Add(ethAddress, byteCode)
			return byteCode, nil
		}
	}
}

func (c *Contract) Run(ctx context.Context, tp *common.TaskParameters) error {
	changedActors, err := tp.Api.StateChangedActors(ctx, tp.AncestorTs.ParentState(), tp.CurrentTs.ParentState())
	if err != nil {
		log.Errorf("StateChangedActors[pTs: %v, pHeight: %v, cTs: %v, cHeight: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), tp.CurrentTs.String(), tp.CurrentTs.Height(), err)
		return err
	}

	var evmContracts []*evmmodel.Contract

	for _, actor := range changedActors {
		if !common.NewCidCache(ctx, tp.Api).IsEVMActor(actor.Code) || actor.Address == nil {
			continue
		}

		address := *actor.Address
		ethAddress, err := ethtypes.EthAddressFromFilecoinAddress(address)
		if err != nil {
			log.Errorf("EthAddressFromFilecoinAddress[addr: %v] err: %v", address.String(), err)
			continue
		}
		byteCode, err := c.getByteCode(ctx, tp.Api, ethAddress)
		if err != nil {
			continue
		}

		contract := &evmmodel.Contract{
			Height:          int64(tp.CurrentTs.Height()),
			Version:         tp.Version,
			FilecoinAddress: address.String(),
			Address:         ethAddress.String(),
			ByteCode:        byteCode.String(),
		}

		actorState, err := tp.Api.StateGetActor(ctx, address, tp.CurrentTs.Key())
		if err != nil {
			log.Warnf("StateGetActor[addr: %v, ts: %v, height: %v] err: %v", address.String(), tp.CurrentTs.String(), tp.CurrentTs.Height(), err)
		} else if err == nil && actorState != nil {
			contract.Balance = actorState.Balance.String()
			contract.Nonce = actorState.Nonce
		}

		evmContracts = append(evmContracts, contract)
	}

	if err = common.InsertMany(ctx, new(evmmodel.Contract), int64(tp.CurrentTs.Height()), tp.Version, &evmContracts); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v evm_contract", len(evmContracts))
	return nil
}
