package filecointask

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"
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
	changedActors, err := tp.Api.StateChangedActors(ctx, tp.AncestorTs.ParentState(), tp.CurrentTs.ParentState())
	if err != nil {
		log.Errorf("StateChangedActors[pTs: %v, pHeight: %v, cTs: %v, cHeight: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), tp.CurrentTs.String(), tp.CurrentTs.Height(), err)
		return err
	}

	var (
		rawActors []*filecoinmodel.RawActor
		exitRa    = make(map[string]*filecoinmodel.RawActor)
	)
	for _, ac := range changedActors {
		actorName := common.NewCidCache(ctx, tp.Api).FindActorNameByCodeCid(ac.Code)
		if actorName == "" {
			log.Warnf("FindActorNameByCodeCid[actor: %v] err: not found", ac.Code.String())
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

	if err = common.InsertMany(ctx, new(filecoinmodel.RawActor), int64(tp.AncestorTs.Height()), tp.Version, &rawActors); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v raw_actor", len(rawActors))
	return nil
}
