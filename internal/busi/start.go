package busi

import (
	"context"

	"github.com/Spacescore/observatory-task-server/config"
	"github.com/Spacescore/observatory-task-server/internal/busi/core"
	"github.com/Spacescore/observatory-task-server/pkg/metrics"
	"github.com/sirupsen/logrus"
)

// Start manager
func Start() error {
	logrus.SetLevel(logrus.DebugLevel)

	metrics.InitRegistry()

	cfg, err := config.InitConfFile(Flags.Config)
	if err != nil {
		return err
	}
	manager := core.NewManager(cfg)
	if err != nil {
		return err
	}
	if err = manager.Start(context.Background()); err != nil {
		return err
	}

	return nil
}
