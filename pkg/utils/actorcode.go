package utils

import (
	"context"
	"sync"

	"github.com/filecoin-project/lotus/api"
	apitypes "github.com/filecoin-project/lotus/api/types"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

var (
	once sync.Once
	m    map[string]cid.Cid
)

// InitActorCodeCidMap init actor code map
func InitActorCodeCidMap(ctx context.Context, node api.FullNode) error {
	var err error

	once.Do(func() {
		var (
			version apitypes.NetworkVersion
		)
		version, err = node.StateNetworkVersion(ctx, types.EmptyTSK)
		if err != nil {
			return
		}
		m, err = node.StateActorCodeCIDs(ctx, version)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}
	return nil
}

// IsEVMActor judge is evm actor
func IsEVMActor(codeCid cid.Cid) bool {
	return m["evm"] == codeCid
}
