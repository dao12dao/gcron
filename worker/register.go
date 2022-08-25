package worker

import (
	"context"
	"gcron/common"
	"gcron/common/constant"
	"gcron/common/zap"
	"net"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// auto register worker node to etcd.
// path: /cron/workers/:ip
type Register struct {
	Client  *clientv3.Client
	KV      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher

	LocalIP string
}

var (
	GlobalRegister *Register
)

func InitRegister(c *EtcdConf) (err error) {
	var (
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
		ipv4    string
	)

	config := clientv3.Config{
		Endpoints:   c.EndPoints,
		DialTimeout: time.Duration(c.DialTimeout) * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		goto ERR
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	GlobalRegister = &Register{
		Client:  client,
		KV:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	if ipv4, err = getLocalIP(); err != nil {
		goto ERR
	}

	GlobalRegister.LocalIP = ipv4

	go GlobalRegister.KeepAlive()

	return
ERR:
	zap.Logf(zap.ERROR, "worker.InitRegister() panic, error is:%+v", err)
	return
}

func (r *Register) KeepAlive() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)
	regKey = constant.TASK_WORKER_PATH + r.LocalIP
	for {
		if leaseGrantResp, err = r.Lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		if keepAliveChan, err = r.Lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// register key into etcd, if false, cancel the lease.
		if _, err = r.KV.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		// handle keep alive resp
		for {
			select {
			case keepAliveResp := <-keepAliveChan:
				if keepAliveResp == nil {
					goto RETRY
				}
			}
		}

	RETRY:
		zap.Logf(zap.INFO, "worker.KeepAlive() retry.")
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

func getLocalIP() (ipv4 string, err error) {
	var (
		addrs []net.Addr
	)
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}

	for _, addr := range addrs {
		// if ipnet, maybe ipv4 or ipv6.
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.IsLoopback() {
			// convert to IPV4
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = common.ErrorNoLocalIPFound

	return
}
