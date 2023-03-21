package common

import (
	"context"

	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
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
