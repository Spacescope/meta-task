package lotus

import (
	"context"
	"sync"
	"time"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/lestrrat-go/backoff/v2"
	"github.com/sirupsen/logrus"
)

type Rpc struct {
	mux    sync.Mutex
	node   api.FullNode
	addr   string
	ctx    context.Context
	closer jsonrpc.ClientCloser
}

// NewRPC new lotus rpc
func NewRPC(ctx context.Context, addr string) (*Rpc, error) {
	r := &Rpc{
		addr: addr,
		ctx:  ctx,
	}
	if err := r.connect(); err != nil {
		return nil, errors.Wrap(err, "connect failed")
	}
	go r.reconnectLoop()
	return r, nil
}

func (r *Rpc) Node() api.FullNode {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.node
}

func (r *Rpc) Close() {
	r.closer()
}

func (r *Rpc) connect() error {
	var err error
	var node v1api.FullNodeStruct
	r.closer, err = jsonrpc.NewMergeClient(r.ctx, r.addr, "Filecoin",
		api.GetInternalStructs(&node), nil)
	if err != nil {
		return errors.Wrap(err, "NewMergeClient failed")
	}
	r.node = &node
	return nil
}

func (r *Rpc) reconnectLoop() {
	bp := backoff.Exponential(
		backoff.WithMinInterval(2*time.Second),
		backoff.WithMaxInterval(60*time.Minute),
		backoff.WithJitterFactor(0.05),
	)
	for {
		_, err := r.node.ChainHead(context.Background())
		if err != nil {
			logrus.Warnf("lotus get ChainHead failed, need reconenct...")
			r.mux.Lock()

			ctx, cancel := context.WithCancel(r.ctx)
			bps := bp.Start(ctx)
			for backoff.Continue(bps) {
				err = r.connect()
				if err == nil {
					cancel()
				} else {
					logrus.Errorf("reconnect lotus failed, err:%s", err)
				}
			}

			r.mux.Unlock()
			logrus.Warnf("reconnect success")
		}
		time.Sleep(10 * time.Second)
	}
}
