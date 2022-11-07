package busi

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/internal/busi/core"
	"github.com/Spacescore/observatory-task/pkg/metrics"

	"github.com/sirupsen/logrus"
)

// Start manager
func Start() error {
	cfg, err := config.InitConfFile(Flags.Config)
	if err != nil {
		return err
	}

	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		return err
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)

	metrics.InitRegistry()

	manager := core.NewManager(cfg)
	if err != nil {
		return err
	}
	if err = manager.Start(context.Background()); err != nil {
		return err
	}

	return nil
}
