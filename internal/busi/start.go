package busi

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/internal/busi/core"

	"github.com/sirupsen/logrus"
)

// Start manager
func Start() error {
	cfg, err := config.InitConfFile(Flags.Config)
	if err != nil {
		return err
	}
	logrus.SetReportCaller(true)

	go HttpServerStart(cfg.Listen.Addr)

	if err = core.NewManager(cfg).Start(context.Background()); err != nil {
		return err
	}

	return nil
}
