package zap

type LogLevel int

const (
	INFO LogLevel = iota
	DEBUG
	WARNING
	ERROR
	FATAL
)

func Logf(level LogLevel, template string, args ...any) {
	switch level {
	case INFO:
		Infof(template, args...)
	case DEBUG:
		Debugf(template, args...)
	case WARNING:
		Warnf(template, args...)
	case ERROR:
		Errorf(template, args...)
	case FATAL:
		Fatalf(template, args...)
	default:
		Infof(template, args...)
	}
}

func Infof(template string, args ...any) {
	Zlogger.Named("info").Infof(template, args...)
}

func Debugf(template string, args ...any) {
	Zlogger.Named("debug").Debugf(template, args...)
}

func Warnf(template string, args ...any) {
	Zlogger.Named("warn").Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	Zlogger.Named("err").Errorf(template, args...)
}

func Fatalf(template string, args ...any) {
	Zlogger.Named("fatal").Fatalf(template, args...)
}
