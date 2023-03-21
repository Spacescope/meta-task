package filecointask

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/lotus/chain/types"
	log "github.com/sirupsen/logrus"
)

// RawActor extract raw actor
type RawActor struct {
}

func (r *RawActor) Name() string {
	return "raw_actor"
}

func (r *RawActor) Model() interface{} {
	return new(filecoinmodel.RawActor)
}

func (r *RawActor) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, force bool, storage storage.Storage) error {
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	if !force {
		existed, err := storage.Existed(r.Model(), int64(parentTs.Height()), version)
		if err != nil {
			return errors.Wrap(err, "storage.Existed failed")
		}
		if existed {
			log.Infof("task [%s] has been process (%d,%d), ignore it", r.Name(), int64(parentTs.Height()), version)
			return nil
		}
	}

	changedActors, err := rpc.Node().StateChangedActors(ctx, parentTs.ParentState(), tipSet.ParentState())
	if err != nil {
		return errors.Wrap(err, "StateChangedActors failed")
	}

	var (
		rawActors []*filecoinmodel.RawActor
		exitRa    = make(map[string]*filecoinmodel.RawActor)
	)
	for _, ac := range changedActors {
		actorName := utils.FindActorNameByCodeCid(ac.Code)
		if actorName == "" {
			log.Warn("can not find cid:[%s] actor name")
			continue
		}
		ra := &filecoinmodel.RawActor{
			Height:    int64(parentTs.Height()),
			Version:   version,
			StateRoot: tipSet.ParentState().String(),
			Name:      actorName,
			Code:      ac.Code.String(),
			Head:      ac.Head.String(),
			Balance:   ac.Balance.String(),
			Nonce:     ac.Nonce,
		}
		if ac.Address != nil {
			ra.Address = ac.Address.String()
		}
		key := fmt.Sprintf("%d-%d-%s-%s-%s", ra.Height, ra.Version, ra.Address, ra.StateRoot, ra.Code)
		exitRa[key] = ra
	}

	for _, actor := range exitRa {
		rawActors = append(rawActors, actor)
	}

	if len(rawActors) > 0 {
		if err := storage.Inserts(ctx, new(filecoinmodel.RawActor), int64(parentTs.Height()), version, &rawActors); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	log.Infof("Tipset[%v] has been process %d raw actor", tipSet.Height(), len(rawActors))

	return nil
}
