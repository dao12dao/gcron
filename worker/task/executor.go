package task

import (
	"gcron/common/model"
	"math/rand"
	"os/exec"
	"time"
)

type TaskExecutor struct {
}

var (
	GlobalTaskExecutor *TaskExecutor
)

func InitTaskExecutor() (err error) {
	GlobalTaskExecutor = &TaskExecutor{}

	return
}

func (e *TaskExecutor) ExecuteTask(execution *model.TaskExecution) {
	go func() {
		var (
			err    error
			output []byte
		)
		// return the cmd exec result to scheduler, and delete from task executing table.
		result := &model.TaskExecuteResult{
			Execution: execution,
			Output:    make([]byte, 0),
		}

		taskLock := GlobalTaskMgr.CreateLock(execution.Task.Name)
		// use random to balance the distributed executor
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		err = taskLock.TryLock()
		defer taskLock.Unlock()

		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			result.StartTime = time.Now()
			cmd := exec.CommandContext(execution.CancelCtx, "/bin/bash", "-c", execution.Task.Command)
			output, err = cmd.CombinedOutput()

			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}

		GlobalScheduler.AppendExecutionResult(result)
	}()
}
