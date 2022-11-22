package evmtask

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/api"
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

func (c *Contract) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	var err error

	tipSetCid, err := tipSet.Key().Cid()
	if err != nil {
		return errors.Wrap(err, "tipSetCid failed")
	}

	hash, err := api.NewEthHashFromCid(tipSetCid)
	if err != nil {
		return errors.Wrap(err, "rpc EthHashFromCid failed")
	}

	ethBlock, err := rpc.Node().EthGetBlockByHash(ctx, hash, true)
	if ethBlock.Number == 0 {
		return errors.Wrap(err, "block number must greater than zero")
	}

	transactions := ethBlock.Transactions
	if len(transactions) == 0 {
		logrus.Debugf("can not find any transaction")
		return nil
	}

	// lazy init actor map
	if err = utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	// TODO Should use pool be used to limit concurrency?
	grp := new(errgroup.Group)
	var (
		contracts []*evmmodel.Contract
		lock      sync.Mutex
	)

	exist := make(map[string]bool)

	for _, transaction := range transactions {
		tm, ok := transaction.(map[string]interface{})
		if ok {
			tm := tm
			grp.Go(func() error {
				ethHash, err := api.EthHashFromHex(tm["hash"].(string))
				if err != nil {
					return errors.Wrap(err, "EthAddressFromHex failed")
				}

				receipt, err := rpc.Node().EthGetTransactionReceipt(ctx, ethHash)
				if err != nil {
					return errors.Wrap(err, "EthGetTransactionReceipt failed")
				}

				// first, judge to address is evm actor
				// second, judge from address is evm actor
				// finally, it may be contract creation
				var (
					evmActors []*types.Actor
				)

				if receipt.To != nil {
					toFilecoinAddress, err := receipt.To.ToFilecoinAddress()
					if err != nil {
						return errors.Wrap(err, "ToFilecoinAddress failed")
					}
					toActor, err := rpc.Node().StateGetActor(ctx, toFilecoinAddress, types.EmptyTSK)
					if err != nil {
						return errors.Wrap(err, "StateGetActor failed")
					}
					if utils.IsEVMActor(toActor.Code) {
						evmActors = append(evmActors, toActor)
					}
				}
				fromFilecoinAddress, err := receipt.From.ToFilecoinAddress()
				if err != nil {
					return errors.Wrap(err, "ToFilecoinAddress failed")
				}
				fromActor, err := rpc.Node().StateGetActor(ctx, fromFilecoinAddress, types.EmptyTSK)
				if err != nil {
					return errors.Wrap(err, "StateGetActor failed")
				}
				if utils.IsEVMActor(fromActor.Code) {
					evmActors = append(evmActors, fromActor)
				}
				// it means contract creation
				if receipt.ContractAddress != nil && receipt.To == nil {
					filecoinAddress, err := receipt.ContractAddress.ToFilecoinAddress()
					if err != nil {
						return errors.Wrap(err, "ToFilecoinAddress failed")
					}
					// current height tipset not have actor state, so init evm actor
					evmActor := &types.Actor{
						Nonce:   0,
						Balance: types.NewInt(0),
						Address: &filecoinAddress,
					}
					evmActors = append(evmActors, evmActor)
				}

				for _, evmActor := range evmActors {
					if evmActor != nil && evmActor.Address != nil {
						ethAddress, err := api.EthAddressFromFilecoinAddress(*evmActor.Address)
						if err != nil {
							return errors.Wrap(err, "EthAddressFromFilecoinAddress failed")
						}
						byteCode, err := rpc.Node().EthGetCode(ctx, ethAddress, "")
						if err != nil {
							return errors.Wrap(err, "EthGetCode failed")
						}
						lock.Lock()
						key := fmt.Sprintf("%d-%d-%s", tipSet.Height(), version, ethAddress.String())
						_, ok := exist[key]
						if !ok {
							contracts = append(contracts, &evmmodel.Contract{
								Height:          int64(tipSet.Height()),
								Version:         version,
								FilecoinAddress: evmActor.Address.String(),
								Address:         ethAddress.String(),
								Balance:         evmActor.Balance.String(),
								Nonce:           evmActor.Nonce,
								ByteCode:        hex.EncodeToString(byteCode),
							})
							exist[key] = true
						}
						lock.Unlock()
					}
				}

				return nil
			})
		}
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	if len(contracts) > 0 {
		if err = storage.WriteMany(ctx, &contracts); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d contract", len(contracts))

	return nil
}
