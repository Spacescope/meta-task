package busi

import (
	"context"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/internal/busi/core"
	log "github.com/sirupsen/logrus"
)

func setLogTimeformat() {
	timeFormater := new(log.TextFormatter)
	timeFormater.FullTimestamp = true
	log.SetFormatter(timeFormater)
}

// Start manager
func Start() error {
	cfg, err := config.InitConfFile(Flags.Config)
	if err != nil {
		return err
	}
	setLogTimeformat()

	go HttpServerStart(cfg.Listen.Addr)

	if err = core.NewManager(cfg).Start(context.Background()); err != nil {
		return err
	}

	return nil
}
