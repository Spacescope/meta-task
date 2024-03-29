package busi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Spacescore/observatory-task/pkg/tasks/common"
	"github.com/Spacescore/observatory-task/pkg/utils"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"

	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/Spacescore/observatory-task/internal/busi/core/consume"

	"github.com/Spacescore/observatory-task/internal/busi/core/consume/notifyClient"
	"github.com/Spacescore/observatory-task/pkg/tasks"
)

type MetaTask struct {
	ReportServer string
	Lotus        string
	Mq           string
	rdb          *redis.Client
	TaskName     string
	TaskPlugin   tasks.Task
}

// Message from chain notify server
type Message struct {
	Version int           `json:"version"`
	TipSet  *types.TipSet `json:"tipset"`
	Force   bool          `json:"force"`
}

func initTask(taskName string) tasks.Task {
	task := tasks.GetTask(taskName)
	if task == nil {
		log.Fatalf("task name %v not support", taskName)
	}
	return task
}

func newMetaTask(cf *utils.MetaTask) *MetaTask {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cf.MQ,
		Password: "",
		DB:       0,
	})

	log.Infof("connect to mq: %v", cf.MQ)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}

	if err := notifyClient.TopicSignIn(cf.ReportServer, cf.TaskName); err != nil {
		log.Fatal(err)
	}

	return &MetaTask{
		ReportServer: cf.ReportServer,
		Lotus:        cf.Lotus,
		Mq:           cf.MQ,
		rdb:          rdb,
		TaskName:     cf.TaskName,
		TaskPlugin:   initTask(cf.TaskName),
	}
}

func MetaTaskStart(ctx context.Context, done func(), cf *utils.MetaTask) {
	defer done()
	s := newMetaTask(cf)
	defer s.rdb.Close()

	for {
		cancelSignal, _ := s.Watcher(ctx)
		if cancelSignal { // cancel due to signal
			return
		}
	}
}

func (s *MetaTask) Watcher(ctx context.Context) (bool, error) {
	api, closer, _ := s.lotusHandshake(ctx)
	defer closer()

	for {
		select {
		case <-ctx.Done():
			log.Infof("meta-task, ctx done, receive signal: %s", ctx.Err().Error())
			return true, nil
		default:
			s.fetchMessage(ctx, api, s.TaskPlugin)

			if _, err := api.ChainHead(ctx); err != nil { // Due to Lotus didn't offer keepalive rpc, we call ChainHead method and treat it as keepalive RPC.
				log.Fatalf("keepalive err: %v", err)
			}
		}
	}
}

func (s *MetaTask) fetchMessage(ctx context.Context, api *lotusapi.FullNodeStruct, taskPlugin tasks.Task) {
	result, err := s.rdb.BRPop(ctx, time.Second*30, s.TaskName).Result()
	if err != nil {
		if err == redis.Nil {
			// log.Infof("BRPop nil: %v", err)
			return
		}
		log.Fatalf("consume redis error: %v", err)
	}
	if len(result) < 1 {
		return
	}

	var event Message
	if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
		log.Fatalf("json.Unmarshal err: %v", err)
	}

	tp := common.TaskParameters{
		ReportServer: s.ReportServer,
		Api:          api,
		Version:      event.Version,
		CurrentTs:    event.TipSet,
		Force:        event.Force,
	}

	consume.ConsumeTipset(ctx, &tp, taskPlugin)
}

func (s *MetaTask) lotusHandshake(ctx context.Context) (*lotusapi.FullNodeStruct, jsonrpc.ClientCloser, error) {
	log.Infof("connect to lotus0: %v", s.Lotus)

	var api lotusapi.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), s.Lotus, "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, nil)
	if err != nil {
		log.Fatalf("lotusHandshake[%v] error: %v", s.Lotus, err)
	}
	return &api, closer, nil
}
