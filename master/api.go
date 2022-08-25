package master

import (
	"gcron/common"
	"gcron/common/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary save task to db
// @Schemes http
// @Description save task to db
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param params body model.Task true "request params"
// @Success 200 {object} common.Response{data=model.Task}
// @Router /api/cron/tasks [POST]
func saveTask(c *gin.Context) {
	var (
		oldTask *model.Task
		err     error
	)

	task := &model.Task{}
	if err = c.BindJSON(&task); err != nil {
		goto ERR
	}

	if len(task.Name) <= 0 || len(task.Command) <= 0 || len(task.CronExpr) <= 0 {
		err = common.ErrorTaskFieldIsNil
		goto ERR
	}

	if oldTask, err = GlobalTaskMgr.SaveTask(task); err != nil {
		goto ERR
	}

	common.BuildResposne(c, oldTask)
	return

ERR:
	common.ChkApiErr(c, err)
}

// @Summary delete task by name
// @Schemes http
// @Description delete task by name
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param name path string true "request params"
// @Success 200 {object} common.Response{data=model.Task}
// @Router /api/cron/tasks/{name} [DELETE]
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

	common.BuildResposne(c, oldTask)
	return

ERR:
	common.ChkApiErr(c, err)
}

// @Summary get task list
// @Schemes http
// @Description get task list
// @Tags 任务管理
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=model.Task}
// @Router /api/cron/tasks [GET]
func listTask(c *gin.Context) {
	var (
		err  error
		list []*model.Task
	)

	if list, err = GlobalTaskMgr.ListTask(); err != nil {
		goto ERR
	}

	common.BuildResposne(c, list)
	return

ERR:
	common.ChkApiErr(c, err)
}

// @Summary kill task by name
// @Schemes http
// @Description kill task by name
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param params body model.KillInputTask true "request params"
// @Success 200 {object} common.Response{}
// @Router /api/cron/task/kill [POST]
func killTask(c *gin.Context) {
	var (
		err  error
		task *model.KillInputTask
	)

	task = &model.KillInputTask{}
	if err = c.BindJSON(&task); err != nil {
		goto ERR
	}

	if err = GlobalTaskMgr.KillTask(task.Name); err != nil {
		goto ERR
	}

	common.BuildResposne(c, struct{}{})
	return

ERR:
	common.ChkApiErr(c, err)
}

// @Summary get log list by task name
// @Schemes http
// @Description get log list by task name
// @Tags 任务管理
// @Accept json
// @Produce json
// @Param name path string true "request params"
// @Success 200 {object} common.Response{data=TaskLog}
// @Router /api/cron/task/log/{name} [GET]
func logTask(c *gin.Context) {
	var (
		name    string
		isOk    bool
		err     error
		logList []*TaskLog
		offset  int
		limit   int
	)

	if name, isOk = c.Params.Get("name"); !isOk {
		goto ERR
	}

	if offset, err = strconv.Atoi(c.DefaultQuery("offset", "0")); err != nil {
		goto ERR
	}

	if limit, err = strconv.Atoi(c.DefaultQuery("limit", "20")); err != nil {
		goto ERR
	}

	if logList, err = GlobalTaskLogMgr.ListLog(name, offset, limit); err != nil {
		goto ERR
	}

	common.BuildResposne(c, logList)
	return

ERR:
	common.ChkApiErr(c, err)
}

// @Summary get worker list
// @Schemes http
// @Description get worker list
// @Tags 任务管理
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=string}
// @Router /api/cron/workers [GET]
func listWorker(c *gin.Context) {
	var (
		err  error
		list []string
	)

	if list, err = GlobalWorkerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	common.BuildResposne(c, list)
	return

ERR:
	common.ChkApiErr(c, err)
}
