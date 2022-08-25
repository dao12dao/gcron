package master

import "github.com/gin-gonic/gin"

func ApiRoute(r *gin.RouterGroup) {
	adminGroup := r.Group("/cron")
	adminGroup.Use()
	{
		adminGroup.GET("/tasks", listTask)
		adminGroup.POST("/tasks", saveTask)
		adminGroup.DELETE("/tasks/:name", removeTask)
		adminGroup.POST("/task/kill", killTask)
		adminGroup.GET("/task/log/:name", logTask)

		adminGroup.GET("/workers", listWorker)
	}
}
