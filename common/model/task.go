package model

type Task struct {
	Name     string `json:"name" valid:"required"`
	Command  string `json:"command" valid:"required"`
	CronExpr string `json:"cron_expr" valid:"required"`
}
