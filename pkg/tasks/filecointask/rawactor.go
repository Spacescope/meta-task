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
	"github.com/sirupsen/logrus"
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

func (r *RawActor) Run(ctx context.Context, rpc *lotus.Rpc, version int, tipSet *types.TipSet, storage storage.Storage) error {
	if tipSet.Height() == 0 {
		return nil
	}

	var err error

	// lazy init actor map
	if err = utils.InitActorCodeCidMap(ctx, rpc.Node()); err != nil {
		return errors.Wrap(err, "InitActorCodeCidMap failed")
	}

	parentTs, err := rpc.Node().ChainGetTipSet(ctx, tipSet.Parents())
	if err != nil {
		return errors.Wrap(err, "ChainGetTipSet failed")
	}

	existed, err := storage.Existed(r.Model(), int64(parentTs.Height()), version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		logrus.Infof("task [%s] has been process (%d,%d), ignore it", r.Name(),
			int64(parentTs.Height()), version)
		return nil
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
			logrus.Warn("can not find cid:[%s] actor name")
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
		if err := storage.DelOldVersionAndWriteMany(ctx, new(filecoinmodel.RawActor), int64(parentTs.Height()), version,
			&rawActors); err != nil {
			return errors.Wrap(err, "storage.WriteMany failed")
		}
	}

	logrus.Debugf("process %d raw actor", len(rawActors))

	return nil
}
