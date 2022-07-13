package main

import (
	"crontab/worker"
	"flag"
	"fmt"
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

	if err = worker.InitLogger(worker.Conf.Base.LogConfigPath); err != nil {
		goto ERR
	}

	// if err = master.InitController(master.Conf.ApiConf.Port); err != nil {
	// 	goto ERR
	// }

	if err = worker.InitTaskManager(worker.Conf.EtcdConf); err != nil {
		goto ERR
	}

	println("worker.Init() run complete!")

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
	fmt.Printf("%+v", err)
}

func Quit() {
	// master.CloseController()
}
