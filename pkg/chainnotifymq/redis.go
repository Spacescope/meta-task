package chainnotifymq

import (
	"context"
	"time"

	"github.com/Spacescore/observatory-task/config"
	"github.com/Spacescore/observatory-task/pkg/chainnotifymq/mqmessage"
	"github.com/Spacescore/observatory-task/pkg/errors"

	vredis "github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
)

var _ MQ = (*Redis)(nil)

type RedisParams struct {
	DSN       string
	QueueName string `mapstructure:"queue_name"`
}

type Redis struct {
	client    *vredis.Client
	queueName string
}

func (r *Redis) Name() string {
	return "redis"
}

func (r *Redis) InitFromConfig(ctx context.Context, cfg *config.ChainNotify) error {
	var (
		err    error
		params RedisParams
	)
	if err = mapstructure.Decode(cfg.MQ.Params, &params); err != nil {
		return errors.Wrap(err, "mapstructure.Decode failed")
	}
	if params.DSN == "" {
		return errors.New("dsn can not empty")
	}
	if params.QueueName == "" {
		return errors.New("queue name can not empty")
	}

	opt, err := vredis.ParseURL(params.DSN)
	if err != nil {
		return err
	}

	r.queueName = params.QueueName

	r.client = vredis.NewClient(opt)
	reply := r.client.Ping(ctx)
	if reply.Err() != nil {
		return reply.Err()
	}
	return nil
}

func (r *Redis) QueueName() string {
	return r.queueName
}

func (r *Redis) FetchMessage(ctx context.Context) (mqmessage.Message, error) {
	reply := r.client.BRPop(ctx, 10*time.Second, r.queueName)
	result, err := reply.Result()
	if err != nil {
		if err == vredis.Nil {
			return nil, nil
		}
		return nil, errors.Wrap(err, "brpop failed")
	}
	if len(result) > 1 {
		return &mqmessage.NormalMessage{Value: []byte(result[1])}, nil
	}
	return nil, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
