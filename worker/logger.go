package worker

import (
	"crontab/common/zap"
)

func InitLogger() (err error) {
	logger := zap.LoadConfigurationForZaplogger(Conf.Base.LogConfigPath)
	defer logger.Sync()

	zap.Zlogger = logger

	return
}

func Info() {
	zap.Zlogger.Info()
}
