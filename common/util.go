package common

import (
	"crontab/common/model"
	"encoding/json"
	"strings"
)

func UnpackJsonToTask(value []byte) (task *model.Task, err error) {
	task = &model.Task{}
	if err = json.Unmarshal(value, &task); err != nil {
		return
	}

	return
}

func ExtractNameFromPath(path string) (name string) {
	return strings.TrimPrefix(path, SAVE_TASK_PATH)
}
