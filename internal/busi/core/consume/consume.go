package consume

import (
	"context"

	"github.com/Spacescore/observatory-task/internal/busi/core/consume/notifyClient"
	"github.com/Spacescore/observatory-task/pkg/tasks"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"

	log "github.com/sirupsen/logrus"
)

func ConsumeTipset(ctx context.Context, tp *common.TaskParameters, taskPlugin tasks.Task) {
	var (
		err         error
		whetherToGo bool
	)

	defer func() {
		state := 1
		desc := ""
		notFoundState := 0
		if err != nil {
			state = 2
			desc = err.Error()
		}

		err = notifyClient.ReportTipsetState(tp.ReportServer,
			tp.Force,
			taskPlugin.Name(),
			int64(tp.CurrentTs.Height()),
			tp.Version,
			state,
			notFoundState,
			desc)
		if err != nil {
			log.Errorf("ReportTipsetState err: %s", err)
		}
	}()
	log.Infof("receive tipset: %v/version: %v", tp.CurrentTs.Height(), tp.Version)

	if whetherToGo, err = DecideWhetherToProceed(ctx, tp, taskPlugin.Model()); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return
	}
	if !whetherToGo {
		log.Infof("tipset has been extracted successfully: %v/version: %v", tp.CurrentTs.Height(), tp.Version)
		return
	}

	// get parent tipset
	tp.AncestorTs, err = common.GetAncestorTipset(ctx, tp.Api, tp.CurrentTs, 1)
	if err != nil {
		log.Errorf("GetAncestorTipset err: %v", err)
		return
	}

	if err = taskPlugin.Run(ctx, tp); err != nil {
		log.Errorf("extract height: %v, err: %v", tp.CurrentTs.Height(), err)
		return
	}
	log.Infof("tipset has been extracted successfully: %v/version: %v", tp.CurrentTs.Height(), tp.Version)
}

func DecideWhetherToProceed(ctx context.Context, tp *common.TaskParameters, model interface{}) (bool, error) {
	if tp.Force {
		return true, nil
	}

	if b, err := common.Existed(model, int64(tp.CurrentTs.Height()), tp.Version); err != nil {
		return false, err
	} else if b {
		return false, nil
	} else {
		return true, nil
	}
}
