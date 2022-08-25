package worker

import (
	"gcron/common/zap"
)

func InitLogger() (err error) {
	logger := zap.LoadConfigurationForZaplogger(Conf.Base.LogConfigPath)
	defer logger.Sync()

	zap.Zlogger = logger
	return
}
