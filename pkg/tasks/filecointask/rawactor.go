package filecointask

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
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

func (r *RawActor) Run(ctx context.Context, tp *common.TaskParameters) error {
	// lazy init actor map
	if err := utils.InitActorCodeCidMap(ctx, tp.Api); err != nil {
		log.Errorf("InitActorCodeCidMap err: %v", err)
		return err
	}

	if !tp.Force {
		// existed, err := storage.Existed(r.Model(), int64(parentTs.Height()), version)
		// if err != nil {
		// 	return errors.Wrap(err, "storage.Existed failed")
		// }
		// if existed {
		// 	log.Infof("task [%s] has been process (%d,%d), ignore it", r.Name(), int64(parentTs.Height()), version)
		// 	return nil
		// }
	}

	changedActors, err := tp.Api.StateChangedActors(ctx, tp.AncestorTs.ParentState(), tp.CurrentTs.ParentState())
	if err != nil {
		log.Errorf("StateChangedActors err: %v", err)
		return err
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
			Height:    int64(tp.AncestorTs.Height()),
			Version:   tp.Version,
			StateRoot: tp.AncestorTs.ParentState().String(),
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
		// if err := storage.Inserts(ctx, new(filecoinmodel.RawActor), int64(parentTs.Height()), version, &rawActors); err != nil {
		// 	return errors.Wrap(err, "storage.WriteMany failed")
		// }
	}

	log.Infof("Tipset[%v] has been process %d raw actor", tp.AncestorTs.Height(), len(rawActors))

	return nil
}
