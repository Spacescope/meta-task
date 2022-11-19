package evmtask

import (
	"context"
	"sync"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// InternalTx task for parse internal transaction
type InternalTx struct {
}

func (i *InternalTx) Name() string {
	return "evm_internal_tx"
}

func (i *InternalTx) Model() interface{} {
	return new(evmmodel.InternalTX)
}

func (i *InternalTx) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet,
	storage storage.Storage) error {
	messages, err := rpc.Node().ChainGetMessagesInTipset(ctx, tipSet.Key())
	if err != nil {
		return errors.Wrap(err, "ChainGetMessagesInTipset failed")
	}

	var (
		internalTxs []interface{}
		lock        sync.Mutex
	)

	grp := new(errgroup.Group)
	for _, message := range messages {
		message := message
		grp.Go(func() error {
			invocs, err := rpc.Node().StateReplay(ctx, tipSet.Key(), message.Cid)
			if err != nil {
				return errors.Wrap(err, "StateReplay failed")
			}
			parentHash, err := api.EthHashFromCid(message.Cid)
			if err != nil {
				return errors.Wrap(err, "EthHashFromCid failed")
			}
			for _, subCall := range invocs.ExecutionTrace.Subcalls {
				subMessage := subCall.Msg
				from, err := api.EthAddressFromFilecoinAddress(subMessage.From)
				if err != nil {
					return errors.Wrap(err, "EthAddressFromFilecoinAddress failed")
				}
				to, err := api.EthAddressFromFilecoinAddress(subMessage.To)
				if err != nil {
					return errors.Wrap(err, "EthAddressFromFilecoinAddress failed")
				}
				hash, err := api.EthHashFromCid(subMessage.Cid())
				if err != nil {
					return errors.Wrap(err, "EthHashFromCid failed")
				}
				internalTx := &evmmodel.InternalTX{
					Height:     int64(tipSet.Height()),
					Version:    version,
					Hash:       hash.String(),
					ParentHash: parentHash.String(),
					From:       from.String(),
					To:         to.String(),
					Value:      subMessage.Value.String(),
				}
				lock.Lock()
				internalTxs = append(internalTxs, internalTx)
				lock.Unlock()
			}

			return nil
		})
	}

	if err = grp.Wait(); err != nil {
		return err
	}

	if len(internalTxs) > 0 {
		if err = storage.WriteMany(ctx, internalTxs...); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d internal transactions", len(internalTxs))
	return nil
}
