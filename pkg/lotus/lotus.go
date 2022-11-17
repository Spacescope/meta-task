package lotus

import (
	"context"
	"sync"
	"time"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
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
	node, closer, err := client.NewFullNodeRPCV1(ctx, addr, nil)
	if err != nil {
		return nil, errors.Wrap(err, "NewFullNodeRPCV1 failed")
	}
	r := &Rpc{
		node:   node,
		addr:   addr,
		ctx:    ctx,
		closer: closer,
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

func (r *Rpc) reconnectLoop() {
	for {
		_, err := r.node.ChainHead(context.Background())
		if err != nil {
			r.mux.Lock()
			logrus.Warnf("lotus get ChainHead failed, need reconenct...")
			r.node, r.closer, err = client.NewFullNodeRPCV1(r.ctx, r.addr, nil)
			if err != nil {
				logrus.Errorf("reconnect lotus failed, err:%s", err)
				r.mux.Unlock()
				time.Sleep(10 * time.Second)
				continue
			}
			logrus.Warnf("reconnect success")
			r.mux.Unlock()
		}
		time.Sleep(10 * time.Second)
	}
}
