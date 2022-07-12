package master

import (
	"crontab/common/zap"
)

func InitLogger(path string) (err error) {
	logger := zap.LoadConfigurationForZaplogger(Conf.Base.LogConfigPath)
	defer logger.Sync()

	zap.Zlogger = logger

	return
}
