package master

import (
	"context"
	"gcron/common/zap"
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

type TaskLogFilter struct {
	TaskName string `bson:"task_name"`
}

type TaskLogSortBy struct {
	SortOrder int `bson:"start_time"`
}

type TaskLogMgr struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

var (
	GlobalTaskLogMgr *TaskLogMgr
)

func InitTaskLogManager(conf *MongoConf) (err error) {
	var (
		client *mongo.Client
	)

	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.Url), options.Client().SetConnectTimeout(time.Duration(conf.ConnectionTimeout)*time.Millisecond)); err != nil {
		goto ERR
	}

	GlobalTaskLogMgr = &TaskLogMgr{
		Client:     client,
		Collection: client.Database("cron").Collection("log"),
	}

	return
ERR:
	zap.Logf(zap.ERROR, "master.InitTaskLogManager() panic, error is:%+v", err)
	return
}

func (t *TaskLogMgr) ListLog(name string, offset, limit int) (logList []*TaskLog, err error) {
	var (
		cursor *mongo.Cursor
	)

	logList = make([]*TaskLog, 0)

	filter := &TaskLogFilter{TaskName: name}
	sort := &TaskLogSortBy{SortOrder: -1}

	if cursor, err = t.Collection.Find(context.TODO(), filter, options.Find().SetSort(sort), options.Find().SetSkip(int64(offset)), options.Find().SetLimit(int64(limit))); err != nil {
		goto ERR
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		taskLog := &TaskLog{}

		if err = cursor.Decode(taskLog); err != nil {
			continue
		}

		logList = append(logList, taskLog)
	}

	return
ERR:
	zap.Logf(zap.ERROR, "master.ListLog() panic, error is:%+v", err)
	return
}
