package worker

import (
	"context"
	"crontab/common"
	"crontab/common/model"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type TaskManager struct {
	Client  *clientv3.Client
	KV      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

var (
	GlobalTaskMgr *TaskManager
)

// Init Task Manager to manage the task.
func InitTaskManager(c *EtcdConf) (err error) {
	var (
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	config := clientv3.Config{
		Endpoints:   c.EndPoints,
		DialTimeout: time.Duration(c.DialTimeout) * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	GlobalTaskMgr = &TaskManager{
		Client:  client,
		KV:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	// start watch task after init.
	GlobalTaskMgr.watchTasks()

	return
}

// watch the tasks changes.
func (m *TaskManager) watchTasks() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		watchStartRevivion int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		taskEvent          *Event
	)

	if getResp, err = m.KV.Get(context.TODO(), common.SAVE_TASK_PATH, clientv3.WithPrefix()); err != nil {
		goto ERR
	}

	for _, kvPair = range getResp.Kvs {
		if task, err := common.UnpackJsonToTask(kvPair.Value); err == nil {
			taskEvent = NewEvent(TaskEventSave, task)
			//TODO: push to scheduler.
			fmt.Printf("kvpair:taskEvent:%+v", taskEvent)
		}
	}

	// start to watch
	go func() {
		watchStartRevivion = getResp.Header.Revision
		watchChan = m.Watcher.Watch(context.TODO(), common.SAVE_TASK_PATH, clientv3.WithRev(watchStartRevivion), clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //task save event.
					task, err := common.UnpackJsonToTask(watchEvent.Kv.Value)
					if err != nil {
						continue
					}

					taskEvent = NewEvent(TaskEventSave, task)
				case mvccpb.DELETE: // task delete event.
					taskName := common.ExtractNameFromPath(string(watchEvent.Kv.Key))
					taskEvent = NewEvent(TaskEventDelete, &model.Task{
						Name: taskName,
					})
				}

				//TODO:unmarshal and push delete event to scheduler.
				fmt.Printf("watch:taskEvent:%+v", taskEvent)
			}
		}
	}()

	return
ERR:
	return
}
