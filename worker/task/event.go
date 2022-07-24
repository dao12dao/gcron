package task

import "crontab/common/model"

type TaskEventType int8

const (
	TASK_EVENT_SAVE   TaskEventType = 10
	TASK_EVENT_DELETE TaskEventType = 20
	TASK_EVENT_KILL   TaskEventType = 30
)

type TaskEvent struct {
	EventType TaskEventType
	Task      *model.Task
}

func NewTaskEvent(t TaskEventType, task *model.Task) *TaskEvent {
	return &TaskEvent{
		EventType: t,
		Task:      task,
	}
}
