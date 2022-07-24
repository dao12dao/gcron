package task

import (
	"context"
	"crontab/common/zap"
	"crontab/worker"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskLog struct {
	TaskName     string `bson:"task_name" json:"task_name"`
	Command      string `bson:"command" json:"command"`
	Output       string `bson:"output" json:"output"`
	Err          string `bson:"err" json:"err"`
	PlanTime     string `bson:"plan_time" json:"plan_time"`
	ScheduleTime string `bson:"schedule_time" json:"schedule_time"`
	StartTime    string `bson:"start_time" json:"start_time"`
	EndTime      string `bson:"end_time" json:"end_time"`
}

type LogBatch struct {
	Logs []any
}

type TaskLogMgr struct {
	Client     *mongo.Client
	Collection *mongo.Collection
	LogChan    chan *TaskLog
}

var (
	GlobalTaskLogMgr *TaskLogMgr
)

func InitTaskLogManager(conf *worker.MongoConf) (err error) {
	var (
		client *mongo.Client
	)

	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.Url), options.Client().SetConnectTimeout(time.Duration(conf.ConnectionTimeout)*time.Millisecond)); err != nil {
		goto ERR
	}

	GlobalTaskLogMgr = &TaskLogMgr{
		Client:     client,
		Collection: client.Database("cron").Collection("log"),
		LogChan:    make(chan *TaskLog, 1000),
	}

	// goroutine, write task log loop.
	go GlobalTaskLogMgr.writeLogLoop()

	return
ERR:
	zap.Zlogger.Errorf("task.InitTaskLogManager() panic, error is:%v", err)
	return
}

func (t *TaskLogMgr) writeLogLoop() {
	var (
		log      *TaskLog
		logBatch *LogBatch
		idx      int
		err      error
	)

BREAK:
	for {
		select {
		case log = <-t.LogChan:
			if idx >= worker.Conf.MongoConf.BatchCount-1 {
				if err = t.batchSaveLogs(logBatch, idx); err != nil {
					break BREAK
				}

				logBatch = nil
				idx = 0
				continue
			}

			// batch insert into mongo, not insert one.
			if logBatch == nil {
				logBatch = &LogBatch{
					Logs: make([]any, worker.Conf.MongoConf.BatchCount),
				}
			}

			logBatch.Logs[idx] = log
			idx++
		case <-time.After(2 * time.Second):
			if logBatch != nil {
				if err = t.batchSaveLogs(logBatch, idx); err != nil {
					break BREAK
				}
				logBatch = nil
				idx = 0
			}
		}
	}
}

func (t *TaskLogMgr) AppendTaskLog(log *TaskLog) {
	t.LogChan <- log
}

func (t *TaskLogMgr) batchSaveLogs(batch *LogBatch, len int) (err error) {
	if _, err = t.Collection.InsertMany(context.TODO(), batch.Logs[:len]); err != nil {
		return
	}

	return
}
