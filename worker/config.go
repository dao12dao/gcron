package worker

import "github.com/go-ini/ini"

type Config struct {
	Base      *Base
	ApiConf   *ApiConf
	EtcdConf  *EtcdConf
	MongoConf *MongoConf
}
type Base struct {
	LogConfigPath string
	Mode          string
	WebRoot       string
}

type ApiConf struct {
	Port         []string `delim:"|"`
	ReadTimeout  int
	WriteTimeout int
}

type EtcdConf struct {
	EndPoints   []string `delim:"|"`
	DialTimeout int
}

type MongoConf struct {
	Url               string
	ConnectionTimeout int
}

var Conf *Config

func init() {
	Conf = &Config{}
}

func InitConfig(configPath string) (err error) {
	if err = ini.MapTo(&Conf, configPath); err != nil {
		return
	}

	return
}
