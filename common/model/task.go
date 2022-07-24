package model

import (
	"context"
	"time"

	"github.com/gorhill/cronexpr"
)

type Task struct {
	Name     string `json:"name" valid:"required"`
	Command  string `json:"command" valid:"required"`
	CronExpr string `json:"cron_expr" valid:"required"`
}

type KillInputTask struct {
	Name string `json:"name" valid:"required"`
}

// task schedule plan
type TaskSchedulePlan struct {
	Task     *Task
	CronExpr *cronexpr.Expression
	NextTime time.Time
}

// task execution info
type TaskExecution struct {
	Task         *Task
	PlanTime     time.Time
	ScheduleTime time.Time
	CancelCtx    context.Context
	CancelFunc   context.CancelFunc
}

type TaskExecuteResult struct {
	Execution *TaskExecution
	Output    []byte
	Err       error
	StartTime time.Time
	EndTime   time.Time
}

// build the schedule plan of task.
func (t *Task) BuildSchedulePlan() (plan *TaskSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)

	if expr, err = cronexpr.Parse(t.CronExpr); err != nil {
		goto ERR
	}

	plan = &TaskSchedulePlan{
		Task:     t,
		CronExpr: expr,
		NextTime: expr.Next(time.Now()),
	}
	return

ERR:
	return
}

// build task execution of schedule plan
func (p *TaskSchedulePlan) BuildTaskExecution() (execution *TaskExecution) {
	execution = &TaskExecution{
		Task:         p.Task,
		PlanTime:     p.NextTime,
		ScheduleTime: time.Now(),
	}
	execution.CancelCtx, execution.CancelFunc = context.WithCancel(context.TODO())
	return
}
