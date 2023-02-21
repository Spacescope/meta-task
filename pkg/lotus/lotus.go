package lotus

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/Spacescore/observatory-task/pkg/errors"
	"github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	log "github.com/sirupsen/logrus"
)

type Rpc struct {
	mux    sync.RWMutex
	node   lotusapi.FullNode
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
	if err := r.lotusHandshake(); err != nil {
		return nil, errors.Wrap(err, "connect failed")
	}
	go r.keepallive()
	return r, nil
}

func (r *Rpc) Node() lotusapi.FullNode {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return r.node
}

func (r *Rpc) Close() {
	r.closer()
}

func (r *Rpc) lotusHandshake() error {
	log.Infof("connect to lotus: %v", r.addr)

	const MAXSLEEP int = 512
	var (
		err  error
		node lotusapi.FullNodeStruct
	)

	for numsec := 1; numsec < MAXSLEEP; numsec <<= 1 {
		closer, err := jsonrpc.NewMergeClient(r.ctx, r.addr, "Filecoin", lotusapi.GetInternalStructs(&node), nil)
		if err == nil {
			r.mux.Lock()
			defer r.mux.Unlock()
			r.node = &node
			r.closer = closer
			return nil
		}
		log.Errorf("connecting to lotus failed: %v", err)
		if numsec <= MAXSLEEP/2 {
			time.Sleep(time.Duration(numsec) * time.Second)
		}
	}
	return err
}

func (r *Rpc) keepallive() {
	select {
	case <-time.After(time.Second * 30): // hearbeat
		if _, err := r.node.ChainHead(context.Background()); err != nil {
			log.Errorf("keepallive failed, err: %s", err)
			r.closer()
			if err = r.lotusHandshake(); err != nil {
				log.Error("system exit")
				os.Exit(1)
			}
			log.Info("reconnect success.")
		}
	default:
	}
}
