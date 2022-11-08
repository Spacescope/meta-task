package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/chainnotifyclient"
	"github.com/Spacescore/observatory-task/pkg/chainnotifymq"
	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/Spacescore/observatory-task/pkg/metrics"
	"github.com/Spacescore/observatory-task/pkg/storage"
	"github.com/Spacescore/observatory-task/pkg/tasks"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

// Message from chain notify server
type Message struct {
	Version int           `json:"version"`
	TipSet  *types.TipSet `json:"tipset"`
}

// Manager task manage
type Manager struct {
	cfg           *config.CFG
	chainNotifyMQ chainnotifymq.MQ
	storage       storage.Storage
	task          tasks.Task
	message       *Message
}

// NewManager new manager
func NewManager(cfg *config.CFG) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) initStorage(ctx context.Context) error {
	m.storage = storage.GetStorage(m.cfg.Storage.Name)
	if m.storage == nil {
		return errors.New("can not found storage")
	}
	if err := m.storage.InitFromConfig(ctx, m.cfg.Storage); err != nil {
		return errors.Wrap(err, fmt.Sprintf("InitFromConfig failed"))
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

func (m *Manager) runTask(ctx context.Context, version int, tipSet *types.TipSet) error {
	timer := prometheus.NewTimer(metrics.TaskCost.WithLabelValues(m.task.Name()))
	defer timer.ObserveDuration()

	var err error

	defer func() {
		state := 1
		desc := ""
		notFoundState := 0
		if err != nil {
			state = 2
			desc = err.Error()
			// TODO only test
			if strings.Contains(err.Error(), "cannot find tipset") {
				logrus.Warnf("task name %s height %d need retry", m.task.Name(), int(tipSet.Height()))
				notFoundState = 1
			}
		}
		err = chainnotifyclient.ReportTipsetState(m.cfg.ChainNotify.Host, m.task.Name(),
			int(tipSet.Height()), version, state, notFoundState, desc)
		if err != nil {
			logrus.Errorf("ReportTipsetState err:%s", err)
		}
	}()

	existed, err := m.storage.Existed(m.task.Model(), int64(m.message.TipSet.Height()), m.message.Version)
	if err != nil {
		return errors.Wrap(err, "storage.Existed failed")
	}
	if existed {
		logrus.Infof("task [%s] has been process (%d,%d), ignore it", m.task.Name(),
			int64(m.message.TipSet.Height()),
			m.message.Version)
		return nil
	}

	if err = m.task.Run(ctx, m.cfg.Lotus.Addr, version, tipSet, m.storage); err != nil {
		return errors.Wrap(err, "task.Run failed")
	}
	return nil
}

func (m *Manager) init(ctx context.Context) error {
	var err error

	logrus.Info("init storage")
	if err = m.initStorage(ctx); err != nil {
		return errors.Wrap(err, "initStorage failed")
	}

	logrus.Info("init task")
	if err = m.initTask(); err != nil {
		return errors.Wrap(err, "initTask failed")
	}

	logrus.Info("init initChainNotifyMQ")
	if err = m.initChainNotifyMQ(ctx); err != nil {
		return errors.Wrap(err, "initChainNotifyMQ failed")
	}

	logrus.Info("topic sign in")
	if err = m.topicSignIn(); err != nil {
		return errors.Wrap(err, "topicSignIn failed")
	}

	logrus.Info("sync storage")
	if err = m.syncStorage(); err != nil {
		return errors.Wrap(err, "syncTable failed")
	}

	return err
}

func (m *Manager) handleSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	<-c
	if m.message != nil {
		chainnotifyclient.ReportTipsetState(m.cfg.ChainNotify.Host, m.task.Name(),
			int(m.message.TipSet.Height()), m.message.Version, 2, 2, "sigterm")
	}
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
	defer m.chainNotifyMQ.Close()

	go m.handleSignal()

	for {
		// reset
		m.message = nil

		message, err := m.chainNotifyMQ.FetchMessage(ctx)
		if err != nil {
			logrus.Errorf("%+v", errors.Wrap(err, "fetch message failed"))
			time.Sleep(1 * time.Second)
			continue
		}
		if message == nil {
			logrus.Debugf("can not get any message, waiting...")
			continue
		}

		logrus.Infof("get message, tipset height:%d, version:%d", m.message.TipSet.Height(), m.message.Version)

		if err = json.Unmarshal(message.Val(), &m.message); err != nil {
			logrus.Errorf("%+v", errors.Wrap(err, "json.Unmarshal failed"))
			continue
		}

		if err = m.runTask(ctx, m.message.Version, m.message.TipSet); err != nil {
			logrus.Errorf("%+v", errors.Wrap(err, "runTask failed"))
			continue
		}

		// do it if mq can commit
		if committableMQ, ok := m.chainNotifyMQ.(chainnotifymq.CommittableMQ); ok {
			if err = committableMQ.Commit(ctx, message); err != nil {
				logrus.Errorf("%+v", errors.Wrap(err, "message commit failed"))
				continue
			}
		}
	}
}
