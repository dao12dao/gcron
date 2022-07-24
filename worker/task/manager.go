package task

import (
	"context"
	"crontab/common"
	"crontab/common/constant"
	"crontab/common/model"
	"crontab/common/zap"
	"crontab/worker"
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
func InitTaskManager(c *worker.EtcdConf) (err error) {
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
		goto ERR
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

	// start watch killer after init
	GlobalTaskMgr.watchKillers()

	return
ERR:
	zap.Zlogger.Errorf("task.InitTaskManager() panic, error is:%v", err)
	return
}

// watch the tasks and their changes.
func (m *TaskManager) watchTasks() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		taskEvent          *TaskEvent
	)

	if getResp, err = m.KV.Get(context.TODO(), constant.TASK_SAVE_PATH, clientv3.WithPrefix()); err != nil {
		goto ERR
	}

	// load all task and append its event into scheduler chan.
	for _, kvPair = range getResp.Kvs {
		if task, err := common.UnpackJsonToTask(kvPair.Value); err == nil {
			taskEvent = NewTaskEvent(TASK_EVENT_SAVE, task)
			// append task event into scheduler chan.
			GlobalScheduler.AppendTaskEvent(taskEvent)
		}
	}

	// start to watch
	go func() {
		watchStartRevision = getResp.Header.Revision
		watchChan = m.Watcher.Watch(context.TODO(), constant.TASK_SAVE_PATH, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //task save event.
					task, err := common.UnpackJsonToTask(watchEvent.Kv.Value)
					if err != nil {
						continue
					}

					taskEvent = NewTaskEvent(TASK_EVENT_SAVE, task)
				case mvccpb.DELETE: // task delete event.
					taskName := common.ExtractNameFromPath(string(watchEvent.Kv.Key), constant.TASK_SAVE_PATH)
					taskEvent = NewTaskEvent(TASK_EVENT_DELETE, &model.Task{
						Name: taskName,
					})
				}

				// append task event into scheduler chan
				GlobalScheduler.AppendTaskEvent(taskEvent)
			}
		}
	}()

	return
ERR:
	zap.Zlogger.Errorf("task.watchTasks() panic, error is:%v", err)
	return
}

func (m *TaskManager) CreateLock(taskName string) (taskLock *TaskLock) {
	taskLock = InitTaskLock(taskName, m.KV, m.Lease)
	return
}

// watch the killers and their changes.
func (m *TaskManager) watchKillers() (err error) {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		taskEvent  *TaskEvent
	)

	// start to watch
	go func() {
		watchChan = m.Watcher.Watch(context.TODO(), constant.TASK_KILL_PATH, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //task save event.
					taskName := common.ExtractNameFromPath(string(watchEvent.Kv.Key), constant.TASK_KILL_PATH)
					taskEvent = NewTaskEvent(TASK_EVENT_KILL, &model.Task{Name: taskName})
				case mvccpb.DELETE: // task delete event.
				}

				// append task event into scheduler chan
				GlobalScheduler.AppendTaskEvent(taskEvent)
			}
		}
	}()

	return
}
