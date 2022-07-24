package main

import (
	"crontab/common/zap"
	"crontab/worker"
	"crontab/worker/task"
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

	if err = worker.InitConfig(configPath); err != nil {
		goto ERR
	}

	if err = worker.InitLogger(); err != nil {
		goto ERR
	}

	if err = worker.InitRegister(worker.Conf.EtcdConf); err != nil {
		goto ERR
	}

	if err = task.InitTaskLogManager(worker.Conf.MongoConf); err != nil {
		goto ERR
	}

	if err = task.InitTaskExecutor(); err != nil {
		goto ERR
	}

	if err = task.InitScheduler(); err != nil {
		goto ERR
	}

	if err = task.InitTaskManager(worker.Conf.EtcdConf); err != nil {
		goto ERR
	}

	zap.Zlogger.Infof("worker.Init() completed!")

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
	zap.Zlogger.Errorf("worker.Init() panic, error is:%+v", err)
}

func Quit() {
	zap.Zlogger.Infof("worker.Quit() completed!")
}
