package master

import (
	"context"
	"gcron/common"
	"gcron/common/constant"
	"gcron/common/zap"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type WorkerManager struct {
	Client *clientv3.Client
	KV     clientv3.KV
	Lease  clientv3.Lease
}

var (
	GlobalWorkerMgr *WorkerManager
)

// Init worker Manager to view the workers.
func InitWorkerManager(c *EtcdConf) (err error) {
	var (
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
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

	GlobalWorkerMgr = &WorkerManager{
		Client: client,
		KV:     kv,
		Lease:  lease,
	}

	return
ERR:
	zap.Logf(zap.ERROR, "master.InitWorkerManager() panic, error is:%+v", err)
	return
}

func (w *WorkerManager) ListWorkers() (list []string, err error) {
	var (
		getResp *clientv3.GetResponse
		kv      *mvccpb.KeyValue
	)
	list = make([]string, 0)

	if getResp, err = w.KV.Get(context.TODO(), constant.TASK_WORKER_PATH, clientv3.WithPrefix()); err != nil {
		goto ERR
	}

	for _, kv = range getResp.Kvs {
		// kv: /cron/workers/127.0.0.1
		list = append(list, common.ExtractNameFromPath(string(kv.Key), constant.TASK_WORKER_PATH))
	}

	return
ERR:
	zap.Logf(zap.ERROR, "master.ListWorkers() panic, error is:%+v", err)
	return
}
