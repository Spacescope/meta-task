package busi

import (
	"context"

	"github.com/Spacescore/observatory-task/pkg/models"
	"github.com/Spacescore/observatory-task/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Task struct {
	Ctx context.Context
	Cf  utils.TomlConfig
}

func NewServer(ctx context.Context) *Task {
	return &Task{Ctx: ctx}
}

func (s *Task) initconfig() {
	if err := utils.InitConfFile(Flags.Config, &s.Cf); err != nil {
		log.Fatalf("Load configuration file err: %v", err)
	}

	utils.EngineGroup = utils.NewEngineGroup(s.Ctx, &[]utils.EngineInfo{{utils.TaskDB, s.Cf.MetaTask.DB, models.Tables}})
}

func (s *Task) setLogTimeformat() {
	timeFormater := new(log.TextFormatter)
	timeFormater.FullTimestamp = true
	log.SetFormatter(timeFormater)
}

func (s *Task) Start() {
	s.initconfig()
	s.setLogTimeformat()

	go HttpServerStart(s.Cf.MetaTask.Addr)
	MetaTaskStart(s.Ctx, &s.Cf.MetaTask)
}
