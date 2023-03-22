package common

import (
	"log"

	"github.com/Spacescore/observatory-task/pkg/models/evmmodel"
	"github.com/Spacescore/observatory-task/pkg/utils"
	lotusapi "github.com/filecoin-project/lotus/api"

	"context"
	"sync"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

type TaskParameters struct {
	ReportServer string
	Api          *lotusapi.FullNodeStruct
	Version      int
	CurrentTs    *types.TipSet
	AncestorTs   *types.TipSet
	Force        bool
}

func GetAncestorTipset(ctx context.Context, lotus *lotusapi.FullNodeStruct, ts *types.TipSet, level uint64) (*types.TipSet, error) {
	var (
		ancestor *types.TipSet
		err      error
	)

	for ; level > 0; level-- {
		ancestor, err = lotus.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return nil, err
		}
	}

	return ancestor, nil
}

type CidCache struct {
	CidCache map[string]cid.Cid
}

var (
	once sync.Once
	cc   CidCache
)

// InitActorCodeCidMap init actor code map
func NewCidCache(ctx context.Context, lotus *lotusapi.FullNodeStruct) *CidCache {
	var err error

	once.Do(func() {
		version, err := lotus.StateNetworkVersion(ctx, types.EmptyTSK)
		if err != nil {
			return
		}
		cc.CidCache, err = lotus.StateActorCodeCIDs(ctx, version)
		if err != nil {
			return
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	return &cc
}

// IsEVMActor judge is evm actor
func (c *CidCache) IsEVMActor(codeCid cid.Cid) bool {
	return c.CidCache["evm"] == codeCid
}

// FindActorNameByCodeCid find actor name by code cide
func (c *CidCache) FindActorNameByCodeCid(codeCid cid.Cid) string {
	for name, c := range c.CidCache {
		if c.KeyString() == codeCid.KeyString() {
			return name
		}
	}
	return ""
}

// Existed judge model exist or not
func Existed(m interface{}, height int64, version int) (bool, error) {
	count, err := utils.EngineGroup[utils.TaskDB].Where("height = ? and version = ?", height, version).Count(m)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func InsertOne(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
	session := utils.EngineGroup[utils.TaskDB].NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	_, err := session.Where("height = ? and version <= ?", height, version).Delete(t)
	if err != nil {
		return err
	}

	_, err = session.InsertOne(m)
	if err != nil {
		return err
	}
	session.Commit()

	return nil
}

func InsertMany(ctx context.Context, t interface{}, height int64, version int, m interface{}) error {
	session := utils.EngineGroup[utils.TaskDB].NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	_, err := session.Where("height = ? and version <= ?", height, version).Delete(t)
	if err != nil {
		return err
	}

	_, err = session.Insert(m)
	if err != nil {
		return err
	}
	session.Commit()

	return nil
}

// only use for evm_contracts
func InsertContracts(ctx context.Context, contracts []*evmmodel.Contract) error {
	session := utils.EngineGroup[utils.TaskDB].NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	for _, contract := range contracts {
		_, err := session.NoAutoCondition().Where("address = ?", contract.Address).Delete(contract) // del the same contract address that historical extracted.
		if err != nil {
			return err
		}

		_, err = session.Insert(contract)
		if err != nil {
			return err
		}
	}

	session.Commit()

	return nil
}
