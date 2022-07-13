package master

import (
	"context"
	"crontab/common/middleware"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
)

var (
	ServerList []*http.Server
)

func InitController(ports []string) (err error) {
	var (
		sockets  []net.Listener
		listener net.Listener
	)

	sockets = make([]net.Listener, len(ports))
	for idx, port := range ports {
		if listener, err = net.Listen("tcp", port); err != nil {
			return
		}

		sockets[idx] = listener
	}

	ServerList = make([]*http.Server, len(sockets))
	for idx, socket := range sockets {
		s := &http.Server{
			Handler:        initHandler(),
			ReadTimeout:    time.Duration(Conf.ApiConf.ReadTimeout) * time.Millisecond,
			WriteTimeout:   time.Duration(Conf.ApiConf.WriteTimeout) * time.Millisecond,
			MaxHeaderBytes: 1 << 20,
		}
		ServerList[idx] = s

		go func(svr *http.Server, l net.Listener) {
			if err := svr.Serve(l); err != nil {
				if err == http.ErrServerClosed {
					return
				}

				panic(err)
			}
		}(s, socket)
	}

	return nil
}

func initHandler() http.Handler {
	gin.SetMode(Conf.Base.Mode)
	handler := gin.New()
	handler.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, POST, PUT, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		Credentials:     false,
		ValidateHeaders: false,
		MaxAge:          50 * time.Second,
	}))
	handler.Use(gin.Logger())
	handler.Use(gin.Recovery())
	handler.Use(middleware.LoggerWithWriter())
	handler.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, "test")
	})

	// static file mapping
	if len(Conf.Base.WebRoot) > 0 {
		handler.Static("/web", Conf.Base.WebRoot)
	}

	// api group
	Route(handler.Group("api"))

	return handler
}

func CloseController() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_ = cancel
	for _, svr := range ServerList {
		if err := svr.Shutdown(ctx); err != nil {
			fmt.Printf("shutdown err:%v", err)
		}
	}

	ServerList = []*http.Server{}
}
