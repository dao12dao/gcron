package master

import "github.com/gin-gonic/gin"

func Route(r *gin.RouterGroup) {
	adminGroup := r.Group("/cron")
	adminGroup.Use()
	{
		adminGroup.GET("/tasks", listTask)
		adminGroup.POST("/tasks", saveTask)
		adminGroup.DELETE("/tasks/:name", removeTask)
		adminGroup.POST("/tasks/kill", killTask)
	}
}
