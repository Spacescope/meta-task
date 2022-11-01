package message

// import (
// 	"context"
//
// 	"github.com/Spacescore/observatory-task-server/pkg/errors"
// 	"github.com/Spacescore/observatory-task-server/pkg/metrics"
// 	"github.com/Spacescore/observatory-task-server/pkg/models"
// 	"github.com/Spacescore/observatory-task-server/pkg/storage"
// 	"github.com/filecoin-project/lotus/chain/types"
// 	"github.com/prometheus/client_golang/prometheus"
// )
//
// type Receipt struct {
// }
//
// func (r *Receipt) Name() string {
// 	return "receipt"
// }
//
// func (r *Receipt) Run(ctx context.Context, version int, tipSet *types.TipSet, lotusCluster lotuscluster.LotusCluster,
// 	storage storage.Storage) error {
// 	timer := prometheus.NewTimer(metrics.TaskCost.WithLabelValues(r.Name()))
// 	defer timer.ObserveDuration()
//
// 	existed, err := storage.Existed(new(models.Receipt), int64(tipSet.Height()), version)
// 	if err != nil {
// 		return errors.Wrap(err, "storage.Existed failed")
// 	}
// 	if existed {
// 		return nil
// 	}
//
// 	return nil
// 	// node := lotusCluster.Select()
// }
