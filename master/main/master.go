package main

import (
	"crontab/common/zap"
	"crontab/master"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	configPath string
)

func initEnv() {
	runtime.GOMAXPROCS((runtime.NumCPU()))
}

func initArgs() {
	flag.StringVar(&configPath, "config", "./config.ini", "specify the config file path to load.")
	flag.Parse()
}

func main() {
	Init()
	defer Quit()
}

func Init() {
	var (
		err error
		c   chan os.Signal
	)

	initEnv()
	initArgs()

	if err = master.InitConfig(configPath); err != nil {
		goto ERR
	}

	if err = master.InitLogger(master.Conf.Base.LogConfigPath); err != nil {
		goto ERR
	}

	if err = master.InitWorkerManager(master.Conf.EtcdConf); err != nil {
		goto ERR
	}

	if err = master.InitTaskLogManager(master.Conf.MongoConf); err != nil {
		goto ERR
	}

	if err = master.InitController(master.Conf.ApiConf.Port); err != nil {
		goto ERR
	}

	if err = master.InitTaskManager(master.Conf.EtcdConf); err != nil {
		goto ERR
	}

	zap.Zlogger.Infof("master.Init() completed!")

	c = make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
			Init()
		default:
			return
		}
	}

ERR:
	zap.Zlogger.Errorf("master.Init() panic, error is:%v", err)
}

func Quit() {
	master.CloseController()
	zap.Zlogger.Infof("master.Quit() Completed!")
}
