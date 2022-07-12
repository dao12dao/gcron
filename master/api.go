package master

import (
	"crontab/common"
	"crontab/common/model"

	"github.com/gin-gonic/gin"
)

func saveTask(c *gin.Context) {
	var (
		oldTask *model.Task
		err     error
	)

	task := &model.Task{}
	if err = c.BindJSON(&task); err != nil {
		goto ERR
	}

	if oldTask, err = GlobalTaskMgr.SaveTask(task); err != nil {
		goto ERR
	}

	common.BuildResposne(c, 0, "success", oldTask)
	return

ERR:
	common.ChkApiErr(c, err)
}

func removeTask(c *gin.Context) {
	var (
		name    string
		isOk    bool
		err     error
		oldTask *model.Task
	)

	if name, isOk = c.Params.Get("name"); !isOk {
		goto ERR
	}

	if oldTask, err = GlobalTaskMgr.RemoveTask(name); err != nil {
		goto ERR
	}

	common.BuildResposne(c, 0, "success", oldTask)
	return

ERR:
	common.ChkApiErr(c, err)
}

func listTask(c *gin.Context) {
	var (
		err  error
		list []*model.Task
	)

	if list, err = GlobalTaskMgr.ListTask(); err != nil {
		goto ERR
	}

	common.BuildResposne(c, 0, "success", list)
	return

ERR:
	common.ChkApiErr(c, err)
}

func killTask(c *gin.Context) {
	var (
		err  error
		task *model.Task
	)

	if err = c.BindJSON(&task); err != nil {
		goto ERR
	}

	if err = GlobalTaskMgr.KillTask(task.Name); err != nil {
		goto ERR
	}

	common.BuildResposne(c, 0, "success", nil)
	return

ERR:
	common.ChkApiErr(c, err)
}
