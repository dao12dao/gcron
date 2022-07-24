package common

import (
	"crontab/common/constant"
	"crontab/common/model"
	"encoding/json"
	"strings"
	"time"
)

func UnpackJsonToTask(value []byte) (task *model.Task, err error) {
	task = &model.Task{}
	if err = json.Unmarshal(value, &task); err != nil {
		return
	}

	return
}

func ExtractNameFromPath(path string, prefix string) (name string) {
	return strings.TrimPrefix(path, prefix)
}

func TimeNowWithFormat() string {
	return time.Now().Format(constant.DATE_TIME_FORMAT)
}
