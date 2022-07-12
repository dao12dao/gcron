package master

import (
	"context"
	"crontab/common"
	"crontab/common/model"
	"encoding/json"
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

	return
}

// list all tasks
func (m *TaskManager) ListTask() (list []*model.Task, err error) {
	var (
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		task    *model.Task
		idx     int
	)

	if getResp, err = m.KV.Get(context.TODO(), common.SaveTaskKeyPrefix, clientv3.WithPrefix()); err != nil {
		return
	}

	list = make([]*model.Task, len(getResp.Kvs))
	for idx, kvPair = range getResp.Kvs {
		task = &model.Task{}
		if err = json.Unmarshal(kvPair.Value, &task); err != nil {
			fmt.Printf("%v, err:%v", idx, err)
			err = nil
			continue
		}
		fmt.Printf("%v, data:%v", idx, task)
		list[idx] = task
	}
	return
}

// save task. when task is exists, to update; otherwise create new task.
func (m *TaskManager) SaveTask(task *model.Task) (oldTask *model.Task, err error) {
	var (
		putResp   *clientv3.PutResponse
		taskValue []byte
		taskPath  string
	)

	taskPath = common.SaveTaskKeyPrefix + task.Name
	if taskValue, err = json.Marshal(task); err != nil {
		return
	}

	// WithPrevKV means get old value before put new value.
	if putResp, err = m.KV.Put(context.TODO(), taskPath, string(taskValue), clientv3.WithPrevKV()); err != nil {
		return
	}

	// if have old value, return old value; otherwise return nil
	if putResp.PrevKv != nil {
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldTask); err != nil {
			err = nil
		}
	}
	return
}

// remove task by name
func (m *TaskManager) RemoveTask(name string) (oldTask *model.Task, err error) {
	var (
		delResp *clientv3.DeleteResponse
	)

	taskPath := common.SaveTaskKeyPrefix + name
	if delResp, err = m.KV.Delete(context.TODO(), taskPath, clientv3.WithPrevKV()); err != nil {
		return
	}

	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldTask); err != nil {
			err = nil
		}
	}

	return
}

// kill task by name
func (m *TaskManager) KillTask(name string) (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseID        clientv3.LeaseID
	)

	taskPath := common.KillTaskKeyPrefix + name
	// to create lease just use for: trigger the change event to listener, then do kill.
	if leaseGrantResp, err = m.Lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	leaseID = leaseGrantResp.ID

	if _, err = m.KV.Put(context.TODO(), taskPath, "", clientv3.WithLease(leaseID)); err != nil {
		return
	}

	return
}
