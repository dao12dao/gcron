package zap

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// ExtFormat 文件后缀名格式
	ExtFormat string = ".2006-01-02-15"
	// CleanFormat 清理格式化
	CleanFormat string = "2006-01-02 15"
	// CleanCycle 清理周期
	CleanCycle int64      = 60 * 60 // 1h
	Zlogger    *Zaplogger = &Zaplogger{
		zap.NewNop().Sugar(),
	}
	logger *zap.Logger
	cleans []*clean
	files  []*writer
)

// Zaplogger 对于整个项目用一个logger对象，使用此对象
// 该对象在整个项目中都指向同一个logger
// 如果使用该对象，请调用LoadConfigurationForZaplogger做初始化，而非LoadConfiguration
type Zaplogger struct {
	*zap.SugaredLogger
}

// Config ...
type Config struct {
	zap.Config
	RetentionHours int `json:"retentionHours" yaml:"retentionHours"`
}

// Logger zap.Logger
type Logger = zap.Logger

type EncoderCFG struct {
	MessageKey   string `json:"messageKey"`
	LevelKey     string `json:"levelKey"`
	LevelEncoder string `json:"levelEncoder"`
}

type Conf struct {
	Level             string     `json:"level"`
	DisableCaller     bool       `json:"disableCaller"`
	DisableStacktrace bool       `json:"disableStacktrace"`
	Encoding          string     `json:"encoding"`
	OutputPaths       []string   `json:"outputPaths"`
	ErrorOutputPaths  []string   `json:"errorOutputPaths"`
	EncoderConfig     EncoderCFG `json:"encoderConfig"`
	RetentionHours    int        `json:"retentionHours"`
	//DisableCaller     bool     `json:"disableCaller"`
}

func Init(conf *Conf) error {
	conf.EncoderConfig.MessageKey = "message"
	conf.EncoderConfig.LevelKey = "level"
	conf.EncoderConfig.LevelEncoder = "lowercase"

	jsonByte, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}
	cfg := &Config{}
	if err := json.Unmarshal(jsonByte, cfg); err != nil {
		panic(err)
	}
	cfg.DisableCaller = true
	cfg.EncoderConfig.TimeKey = "datetime"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	default:
		panic("encoding error is " + cfg.Encoding)
	}
	logger = zap.New(zapcore.NewCore(encoder, open(cfg.OutputPaths, cfg.RetentionHours), cfg.Level), buildOptions(cfg)...)
	Zlogger = &Zaplogger{logger.Sugar()}

	return nil
}

// LoadConfiguration ...
func LoadConfiguration(filename string) *Logger {
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	cfg := &Config{}
	if err := json.Unmarshal(jsonData, cfg); err != nil {
		panic(err)
	}
	if cfg.EncoderConfig.TimeKey != "" {
		cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}
	}
	if cfg.EncoderConfig.EncodeDuration == nil {
		cfg.EncoderConfig.EncodeDuration = func(time.Duration, zapcore.PrimitiveArrayEncoder) {}
	}

	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	default:
		panic("encoding error is " + cfg.Encoding)
	}
	logger = zap.New(zapcore.NewCore(encoder, open(cfg.OutputPaths, cfg.RetentionHours), cfg.Level), buildOptions(cfg)...)
	return logger
}

// LoadConfigurationForZaplogger 返回Zaplogger指针
func LoadConfigurationForZaplogger(filename string) *Zaplogger {
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	cfg := &Config{}
	if err := json.Unmarshal(jsonData, cfg); err != nil {
		panic(err)
	}
	cfg.EncoderConfig.TimeKey = "datetime"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	var encoder zapcore.Encoder
	switch cfg.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
	default:
		panic("encoding error is " + cfg.Encoding)
	}
	logger = zap.New(zapcore.NewCore(encoder, open(cfg.OutputPaths, cfg.RetentionHours), cfg.Level), buildOptions(cfg)...)
	zapLogger := Zaplogger{logger.Sugar()}
	return &zapLogger
}

// Close ...
func Close() {
	logger.Sync()
	for _, f := range files {
		f.close()
	}
	for _, c := range cleans {
		c.Close()
	}
}

func buildOptions(cfg *Config) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(open(cfg.ErrorOutputPaths, cfg.RetentionHours))}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	stackLevel := zap.ErrorLevel
	if cfg.Development {
		stackLevel = zap.WarnLevel
	}

	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]zap.Field, 0, len(cfg.InitialFields))
		keys := make([]string, 0, len(cfg.InitialFields))
		for k := range cfg.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, cfg.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	return opts
}

func open(paths []string, retentionHours int) zapcore.WriteSyncer {
	writers := make([]zapcore.WriteSyncer, 0, len(paths))
	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, os.Stdout)
			// Don't close standard out.
			continue
		case "stderr":
			writers = append(writers, os.Stderr)
			// Don't close standard error.
			continue
		}
		f := newWriter(path)
		files = append(files, f)
		writers = append(writers, f)
		c := newClean(path, retentionHours)
		cleans = append(cleans, c)
	}
	return zap.CombineWriteSyncers(writers...)
}
