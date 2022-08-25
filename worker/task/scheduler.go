package task

import (
	"gcron/common"
	"gcron/common/model"
	"gcron/common/zap"
	"time"
)

type Scheduler struct {
	taskEventChan      chan *TaskEvent
	taskPlanTable      map[string]*model.TaskSchedulePlan
	taskExecutingTable map[string]*model.TaskExecution
	taskResultChan     chan *model.TaskExecuteResult
}

var (
	GlobalScheduler *Scheduler
)

// Init Scheduler
func InitScheduler() (err error) {
	GlobalScheduler = &Scheduler{
		taskEventChan:      make(chan *TaskEvent, 200),
		taskPlanTable:      make(map[string]*model.TaskSchedulePlan, 200),
		taskExecutingTable: make(map[string]*model.TaskExecution, 200),
		taskResultChan:     make(chan *model.TaskExecuteResult, 200),
	}

	// start a goroutine to monitor event chan.
	go GlobalScheduler.scheduleLoop()
	return
}

// append taskevent into event chan
func (s *Scheduler) AppendTaskEvent(event *TaskEvent) {
	s.taskEventChan <- event
}

// start schedule monitor in loop
func (s *Scheduler) scheduleLoop() {
	var (
		taskEvent     *TaskEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		result        *model.TaskExecuteResult
	)

	scheduleAfter = s.TrySchedule()
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		// monitor the task event.
		case taskEvent = <-s.taskEventChan:
			s.HandleTaskEvent(taskEvent)
		case <-scheduleTimer.C:
		case result = <-s.taskResultChan:
			s.HandleTaskResult(result)
		}

		scheduleAfter = s.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}

// handle task event from etcd, include: save event and delete event.
func (s *Scheduler) HandleTaskEvent(event *TaskEvent) (err error) {
	var (
		plan          *model.TaskSchedulePlan
		taskExecution *model.TaskExecution
		isOk          bool
	)

	switch event.EventType {
	case TASK_EVENT_SAVE:
		if plan, err = event.Task.BuildSchedulePlan(); err != nil {
			return
		}
		s.taskPlanTable[event.Task.Name] = plan
	case TASK_EVENT_DELETE:
		if _, isOk = s.taskPlanTable[event.Task.Name]; isOk {
			delete(s.taskPlanTable, event.Task.Name)
		}
	case TASK_EVENT_KILL:
		if taskExecution, isOk = s.taskExecutingTable[event.Task.Name]; isOk {
			taskExecution.CancelFunc()
		}
	}
	return
}

// why is try? when nextTime isn't arrrived or task execution isn't completed.
// returns the duration to next schedule time.
func (s *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var (
		plan     *model.TaskSchedulePlan
		now      time.Time = time.Now()
		nearTime *time.Time
	)

	if len(s.taskPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	for _, plan = range s.taskPlanTable {
		// if nextTime is arrived, do task.
		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			// try to start the schedule plan
			s.TryStartSchedulePlan(plan)
			plan.NextTime = plan.CronExpr.Next(now)
		}

		if nearTime == nil || plan.NextTime.Before(*nearTime) {
			nearTime = &plan.NextTime
		}
	}
	scheduleAfter = nearTime.Sub(now)
	return
}

// try do the task if task is not running
func (s *Scheduler) TryStartSchedulePlan(plan *model.TaskSchedulePlan) (err error) {
	var (
		execution *model.TaskExecution
		isRunning bool
	)

	if _, isRunning = s.taskExecutingTable[plan.Task.Name]; isRunning {
		return
	}

	execution = plan.BuildTaskExecution()
	s.taskExecutingTable[plan.Task.Name] = execution

	GlobalTaskExecutor.ExecuteTask(execution)
	return
}

// append execution result into the result chan.
func (s *Scheduler) AppendExecutionResult(result *model.TaskExecuteResult) {
	s.taskResultChan <- result
}

// handle task execution result.
func (s *Scheduler) HandleTaskResult(result *model.TaskExecuteResult) {
	delete(s.taskExecutingTable, result.Execution.Task.Name)

	if result.Err == common.ErrorLockIsOccupied {
		return
	}
	timeNow := common.TimeNowWithFormat()
	taskLog := &TaskLog{
		TaskName:     result.Execution.Task.Name,
		Command:      result.Execution.Task.Command,
		Output:       string(result.Output),
		PlanTime:     timeNow,
		ScheduleTime: timeNow,
		StartTime:    timeNow,
		EndTime:      timeNow,
	}
	if result.Err != nil {
		taskLog.Err = result.Err.Error()
	}

	GlobalTaskLogMgr.AppendTaskLog(taskLog)
	zap.Logf(zap.INFO, "task is done. name:%+v, result:%+v, err:%+v\n", result.Execution.Task.Name, string(result.Output), result.Err)
}
