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

	tipSetCid, err := tipSet.Key().Cid()
	if err != nil {
		return errors.Wrap(err, "tipSetCid failed")
	}

	hash, err := api.EthHashFromCid(tipSetCid)
	if err != nil {
		return errors.Wrap(err, "rpc EthHashFromCid failed")
	}
	ethBlock, err := node.EthGetBlockByHash(ctx, hash, true)
	if err != nil {
		return errors.Wrap(err, "rpc EthGetBlockByHash failed")
	}

	transactions := ethBlock.Transactions
	if len(transactions) == 0 {
		logrus.Debugf("can not find any transaction")
		return nil
	}

	// lazy init actor map
	if err = utils.InitActorCodeCidMap(ctx, node); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	// TODO Should use pool be used to limit concurrency?
	grp := new(errgroup.Group)
	var (
		contracts []interface{}
		lock      sync.Mutex
	)

	for _, transaction := range transactions {
		tm, ok := transaction.(map[string]interface{})
		if ok {
			tm := tm
			grp.Go(func() error {
				ethHash, err := api.EthHashFromHex(tm["hash"].(string))
				if err != nil {
					return errors.Wrap(err, "EthAddressFromHex failed")
				}
				receipt, err := node.EthGetTransactionReceipt(ctx, ethHash)
				if err != nil {
					return errors.Wrap(err, "EthGetTransactionReceipt failed")
				}

				// first, judge to address is evm actor
				// second, judge from address is evm actor
				// finally, it may be contract creation
				var (
					evmActor *types.Actor
				)

				if receipt.To != nil {
					toFilecoinAddress, err := receipt.To.ToFilecoinAddress()
					if err != nil {
						return errors.Wrap(err, "ToFilecoinAddress failed")
					}
					toActor, err := node.StateGetActor(ctx, toFilecoinAddress, tipSet.Key())
					if err != nil {
						return errors.Wrap(err, "StateGetActor failed")
					}
					if utils.IsEVMActor(toActor.Code) {
						evmActor = toActor
					}
				}
				if evmActor == nil {
					fromFilecoinAddress, err := receipt.From.ToFilecoinAddress()
					if err != nil {
						return errors.Wrap(err, "ToFilecoinAddress failed")
					}
					fromActor, err := node.StateGetActor(ctx, fromFilecoinAddress, tipSet.Key())
					if err != nil {
						return errors.Wrap(err, "StateGetActor failed")
					}
					if utils.IsEVMActor(fromActor.Code) {
						evmActor = fromActor
					}
				}
				if evmActor == nil {
					// it means contract creation
					if receipt.ContractAddress != nil && receipt.To == nil {
						filecoinAddress, err := receipt.ContractAddress.ToFilecoinAddress()
						if err != nil {
							return errors.Wrap(err, "ToFilecoinAddress failed")
						}
						// current height tipset not have actor state, so init evm actor
						evmActor = &types.Actor{
							Nonce:   0,
							Balance: types.NewInt(0),
							Address: &filecoinAddress,
						}
					}
				}

				if evmActor != nil && evmActor.Address != nil {
					ethAddress, err := api.EthAddressFromFilecoinAddress(*evmActor.Address)
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
						FilecoinAddress: evmActor.Address.String(),
						Address:         receipt.ContractAddress.String(),
						Balance:         evmActor.Balance.Int64(),
						Nonce:           evmActor.Nonce,
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
		if err = storage.WriteMany(ctx, contracts...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d contract", len(contracts))

	return nil
}
