package core

import (
	"context"
	"fmt"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/chainnotifyclient"
	"github.com/Spacescore/observatory-task/pkg/chainnotifymq"
	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/lotus"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/tasks"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
)

// Message from chain notify server
type Message struct {
	Version int           `json:"version"`
	TipSet  *types.TipSet `json:"tipset"`
	Force   bool          `json:"force"`
}

// Manager task manage
type Manager struct {
	cfg           *config.CFG
	chainNotifyMQ chainnotifymq.MQ
	storage       storage.Storage
	task          tasks.Task
	message       *Message
	rpc           *lotus.Rpc
}

// NewManager new manager
func NewManager(cfg *config.CFG) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) initRpc() error {
	var err error
	m.rpc, err = lotus.NewRPC(context.Background(), m.cfg.Lotus.Addr)
	if err != nil {
		return errors.Wrap(err, "NewRPC")
	}
	return nil
}

func (m *Manager) initStorage(ctx context.Context) error {
	m.storage = storage.GetStorage(m.cfg.Storage.Name)
	if m.storage == nil {
		return errors.New("can not found storage")
	}
	if err := m.storage.InitFromConfig(ctx, m.cfg.Storage); err != nil {
		return errors.Wrap(err, "InitFromConfig failed")
	}
	return nil
}

func (m *Manager) initTask() error {
	m.task = tasks.GetTask(m.cfg.Task.Name)
	if m.task == nil {
		return errors.New(fmt.Sprintf("task name %s not support", m.cfg.Task.Name))
	}
	return nil
}

func (m *Manager) initChainNotifyMQ(ctx context.Context) error {
	m.chainNotifyMQ = chainnotifymq.GetMQ(m.cfg.ChainNotify.MQ.Name)
	if m.chainNotifyMQ == nil {
		return errors.New(fmt.Sprintf("chain notify mq %s not support", m.cfg.ChainNotify.MQ.Name))
	}
	if err := m.chainNotifyMQ.InitFromConfig(ctx, m.cfg.ChainNotify, m.task.Name()); err != nil {
		return errors.Wrap(err, "chainNotifyMQ.InitFromConfig failed")
	}
	return nil
}

func (m *Manager) topicSignIn() error {
	return chainnotifyclient.TopicSignIn(m.cfg.ChainNotify.Host, m.task.Name())
}

func (m *Manager) runTask(ctx context.Context, version int, tipSet *types.TipSet, force bool) error {
	var err error

	defer func() {
		state := 1
		desc := ""
		notFoundState := 0
		if err != nil {
			state = 2
			desc = err.Error()
		}
		err = chainnotifyclient.ReportTipsetState(m.cfg.ChainNotify.Host, force, m.task.Name(), int(tipSet.Height()), version, state, notFoundState, desc)
		if err != nil {
			log.Errorf("ReportTipsetState err: %s", err)
		}
	}()

	if err = m.task.Run(ctx, m.rpc, version, tipSet, force, m.storage); err != nil {
		log.Errorf("extract height: %v err: %v", tipSet.Height(), err)
		return err
	}
	return nil
}

func (m *Manager) init(ctx context.Context) error {
	var err error

	log.Info("init storage")
	if err = m.initStorage(ctx); err != nil {
		return errors.Wrap(err, "initStorage failed")
	}

	log.Info("init lotus rpc")
	if err = m.initRpc(); err != nil {
		return errors.Wrap(err, "initRpc failed")
	}

	log.Info("init task")
	if err = m.initTask(); err != nil {
		return errors.Wrap(err, "initTask failed")
	}

	log.Info("init initChainNotifyMQ")
	if err = m.initChainNotifyMQ(ctx); err != nil {
		return errors.Wrap(err, "initChainNotifyMQ failed")
	}

	log.Info("topic sign in")
	if err = m.topicSignIn(); err != nil {
		return errors.Wrap(err, "topicSignIn failed")
	}

	log.Info("sync storage")
	if err = m.syncStorage(); err != nil {
		return errors.Wrap(err, "syncTable failed")
	}

	return err
}

func (m *Manager) syncStorage() error {
	db, ok := m.storage.(storage.Database)
	if ok {
		if err := db.Sync(m.task.Model()); err != nil {
			return errors.Wrap(err, "Sync failed")
		}
	}
	return nil
}

// Start run task manager
func (m *Manager) Start(ctx context.Context) error {
	var err error
	if err = m.init(ctx); err != nil {
		return errors.Wrap(err, "init failed")
	}
	defer func() {
		m.chainNotifyMQ.Close()
		m.rpc.Close()
	}()

	for {
		message, err := m.chainNotifyMQ.FetchMessage(ctx)
		if err != nil {
			log.Fatalf("fetch message err: %v", err)
		}
		if message == nil {
			continue
		}

		if err = json.Unmarshal(message.Val(), &m.message); err != nil {
			log.Fatalf("json.Unmarshal err: %v", err)
		}

		log.Infof("consume tipset: %v/version: %d", m.message.TipSet.Height(), m.message.Version)
		m.runTask(ctx, m.message.Version, m.message.TipSet, m.message.Force)
	}
}
