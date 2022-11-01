package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	TaskFailedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "task_failed",
		}, []string{"name"},
	)

	TaskCost = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "task_cost",
		}, []string{"name"},
	)
)

// InitRegistry init prometheus registry
func InitRegistry() {
	registry := prometheus.NewRegistry()
	registry.MustRegister(TaskFailedGauge, TaskCost)
}
