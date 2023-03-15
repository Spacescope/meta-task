package busi

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/internal/busi/core"
)

// Start manager
func Start() error {
	cfg, err := config.InitConfFile(Flags.Config)
	if err != nil {
		return err
	}
	go HttpServerStart(cfg.Listen.Addr)

	if err = core.NewManager(cfg).Start(context.Background()); err != nil {
		return err
	}

	return nil
}
