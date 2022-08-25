//go:build doc
// +build doc

package master

import (
	_ "gcron/master/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	// docs.SwaggerInfo.BasePath = "/api"
	apiDocHandler = ginSwagger.WrapHandler(swaggerfiles.Handler)
}
