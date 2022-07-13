package worker

import "crontab/common/model"

type TaskEventType int8

const (
	TaskEventSave   TaskEventType = 10
	TaskEventDelete TaskEventType = 20
)

type Event struct {
	EventType TaskEventType
	Task      *model.Task
}

func NewEvent(t TaskEventType, task *model.Task) *Event {
	return &Event{
		EventType: t,
		Task:      task,
	}
}
