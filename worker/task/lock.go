package task

import (
	"context"
	"crontab/common"
	"crontab/common/constant"
	"crontab/common/zap"

	"github.com/coreos/etcd/clientv3"
)

type TaskLock struct {
	KV    clientv3.KV
	Lease clientv3.Lease

	Name       string
	IsLocked   bool
	CancelFunc context.CancelFunc
	LeaseID    clientv3.LeaseID
}

func InitTaskLock(name string, kv clientv3.KV, lease clientv3.Lease) *TaskLock {
	return &TaskLock{
		KV:    kv,
		Lease: lease,
		Name:  name,
	}
}

// try to lock, if failed it will rollback.
func (t *TaskLock) TryLock() (err error) {
	var (
		leaseGrandResp    *clientv3.LeaseGrantResponse
		leaseID           clientv3.LeaseID
		cancelCtx         context.Context
		cancelFunc        context.CancelFunc
		keepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txn               clientv3.Txn
		lockKey           string
		txnResp           *clientv3.TxnResponse
	)

	// create ctx and func use to cancel/unlock the lease.
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	// create a new lease.
	if leaseGrandResp, err = t.Lease.Grant(context.TODO(), 5); err != nil {
		goto ERR
	}
	leaseID = leaseGrandResp.ID

	if keepAliveRespChan, err = t.Lease.KeepAlive(cancelCtx, leaseID); err != nil {
		goto ERR
	}

	// read resp from keepAliveRespChan, if nil means stop keep alive.
	go func() {
		for {
			select {
			case keepAliveResp := <-keepAliveRespChan:
				if keepAliveResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	// start txn
	txn = t.KV.Txn(context.TODO())

	// lock key
	lockKey = constant.TASK_LOCK_PATH + t.Name

	// txn do things.
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockKey))

	// txn commit.
	if txnResp, err = txn.Commit(); err != nil {
		goto ERR
	}

	// lock failed.
	if !txnResp.Succeeded {
		err = common.ErrorLockIsOccupied
		goto ERR
	}

	// lock succeeded.
	t.LeaseID = leaseID
	t.CancelFunc = cancelFunc
	t.IsLocked = true
	return
ERR:
	zap.Zlogger.Errorf("task.TryLock() panic, error is:%v", err)
	cancelFunc()
	// release/revoke the lease.
	t.Lease.Revoke(context.TODO(), leaseID)
	return
}

func (t *TaskLock) Unlock() (err error) {
	if t.IsLocked {
		t.CancelFunc()
		t.Lease.Revoke(context.TODO(), t.LeaseID)
	}
	return
}
