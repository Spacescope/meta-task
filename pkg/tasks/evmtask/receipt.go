package evmtask

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/goccy/go-json"
	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
)

// Receipt parse evm transaction receipt
type Receipt struct {
	actorCache *lru.Cache
}

func (e *Receipt) Name() string {
	return "evm_receipt"
}

func (e *Receipt) Model() interface{} {
	return new(evmmodel.Receipt)
}

func (e *Receipt) Run(ctx context.Context, tp *common.TaskParameters) error {
	messages, err := tp.Api.ChainGetMessagesInTipset(ctx, tp.AncestorTs.Key())
	if err != nil {
		log.Errorf("ChainGetMessagesInTipset[ts: %v, height: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), err)
		return err
	}

	evmReceipts := make([]*evmmodel.Receipt, 0)
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

		ethHash, err := ethtypes.EthHashFromCid(message.Cid)
		if err != nil {
			log.Errorf("EthHashFromCid[cid: %v] err: %v", message.Cid.String(), err)
			return err
		}
		receipt, err := tp.Api.EthGetTransactionReceipt(ctx, ethHash)
		if err != nil {
			log.Errorf("EthGetTransactionReceipt[hash: %v] err: %v", ethHash.String(), err)
			continue
		}
		if receipt == nil {
			continue
		}

		r := &evmmodel.Receipt{
			Height:            int64(tp.AncestorTs.Height()),
			Version:           tp.Version,
			TransactionHash:   receipt.TransactionHash.String(),
			TransactionIndex:  int64(receipt.TransactionIndex),
			BlockHash:         receipt.BlockHash.String(),
			BlockNumber:       int64(receipt.BlockNumber),
			From:              receipt.From.String(),
			Status:            int64(receipt.Status),
			CumulativeGasUsed: int64(receipt.CumulativeGasUsed),
			GasUsed:           int64(receipt.GasUsed),
			EffectiveGasPrice: receipt.EffectiveGasPrice.Int64(),
			LogsBloom:         receipt.LogsBloom.String(),
			MessageCid:        message.Cid.String(),
		}

		b, _ := json.Marshal(receipt.Logs)
		r.Logs = string(b)
		if receipt.ContractAddress != nil {
			r.ContractAddress = receipt.ContractAddress.String()
		}
		if receipt.To != nil {
			r.To = receipt.To.String()
		}

		evmReceipts = append(evmReceipts, r)
	}

	if err = common.InsertMany(ctx, new(evmmodel.Receipt), int64(tp.AncestorTs.Height()), tp.Version, &evmReceipts); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v evm_receipt", len(evmReceipts))
	return nil
}
