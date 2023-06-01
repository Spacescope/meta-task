package common

import (
	"fmt"
	"reflect"

	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/go-address"
	lotusapi "github.com/filecoin-project/lotus/api"

	"context"
	"sync"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"

	lru "github.com/hashicorp/golang-lru"
	log "github.com/sirupsen/logrus"
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

type CidLRU struct {
	cidLRU *lru.Cache
	lotus  *lotusapi.FullNodeStruct
}

var (
	onceCidCache sync.Once
	onceCidLRU   sync.Once
	cc           CidCache
	cl           CidLRU
)

// -------------------------------------------------------------------------------

// InitActorCodeCidMap init actor code map
func NewCidCache(ctx context.Context, lotus *lotusapi.FullNodeStruct) *CidCache {
	onceCidCache.Do(func() {
		version, err := lotus.StateNetworkVersion(ctx, types.EmptyTSK)
		if err != nil {
			return
		}
		cc.CidCache, err = lotus.StateActorCodeCIDs(ctx, version)
		if err != nil {
			log.Fatal(err)
			return
		}
	})
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

// -------------------------------------------------------------------------------

// InitActorCodeCidLRU init actor code lru
func NewCidLRU(ctx context.Context, lotus *lotusapi.FullNodeStruct) *CidLRU {
	onceCidLRU.Do(func() {
		var err error
		cl.cidLRU, err = lru.New(256)
		if err != nil {
			log.Fatal(err)
		}
		cl.lotus = lotus
	})
	return &cl
}

func (c *CidLRU) getCacheActorCode(key string) (cid.Cid, bool) {
	code, ok := c.cidLRU.Get(key)
	if ok {
		return code.(cid.Cid), ok
	} else {
		return cid.Cid{}, ok
	}
}

func (c *CidLRU) setCacheActorCode(key string, value interface{}) {
	c.cidLRU.Add(key, value)
}

func (c *CidLRU) IsEVMActor(ctx context.Context, messageAddress address.Address, ts *types.TipSet) (bool, error) {
	var (
		actor *types.Actor
		err   error
	)
	messageActorKey := fmt.Sprintf("%v-%v", messageAddress.String(), ts.String())

	actorCode, ok := c.getCacheActorCode(messageActorKey)
	if ok {
		return NewCidCache(ctx, c.lotus).IsEVMActor(actorCode), nil
	} else {
		actor, err = c.lotus.StateGetActor(ctx, messageAddress, ts.Key())
		if err != nil {
			log.Errorf("StateGetActor[ts: %v, height: %v] err: %v", ts.Key(), ts.Height(), err)
			return false, err
		}
		c.setCacheActorCode(messageActorKey, actor.Code)
		return NewCidCache(ctx, c.lotus).IsEVMActor(actor.Code), nil
	}
}

func (c *CidLRU) AtLeastOneAddressIsEVMActor(ctx context.Context, ToAndFromAddresses []address.Address, ts *types.TipSet) (bool, error) {
	if len(ToAndFromAddresses) != 2 {
		panic("ToAndFromAddresses should contain both 'to' and 'from' addresses")
	}

	var (
		b   bool
		err error
	)

	for _, txnAddress := range ToAndFromAddresses {
		b, err = c.IsEVMActor(ctx, txnAddress, ts)
		if err != nil {
			return false, err
		}

		if b {
			return b, nil
		}
	}

	return false, nil
}

// -------------------------------------------------------------------------------

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

	sliceValue := reflect.Indirect(reflect.ValueOf(m))

	if sliceValue.Kind() == reflect.Slice && sliceValue.Len() > 0 {
		_, err = session.Insert(m)
		if err != nil {
			return err
		}
	}

	session.Commit()

	return nil
}
