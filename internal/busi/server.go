package busi

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Spacescore/observatory-task/pkg/models"
	"github.com/Spacescore/observatory-task/pkg/utils"
	log "github.com/sirupsen/logrus"
)

type Task struct {
	Ctx context.Context
	Cf  utils.TomlConfig
	wg  sync.WaitGroup
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

	{
		s.wg.Add(1)
		ctx, cancel := context.WithCancel(s.Ctx)
		go MetaTaskStart(ctx, s.wg.Done, &s.Cf.MetaTask)
		<-s.sigHandle()
		cancel()
		s.wg.Wait()
	}
}

func (s *Task) sigHandle() <-chan os.Signal {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	return sigChannel
}
