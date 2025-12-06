package config

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/yaml"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils"
)

type Config struct {
	AppCfg         AppConfig         `yaml:"app-config"`
	NacosCfg       NacosConfig       `yaml:"nacos-config"`
	NacosCenterCfg NacosCenterConfig `yaml:"nacos-center-config"`
	LogCfg         LogConfig         `yaml:"log-config"`
}

func New(cfgPath string) (*Config, error) {
	data, err := utils.ReadAll(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Parse(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type AppConfig struct {
	AppName                  string `yaml:"app-name"`
	Profile                  string `yaml:"profile"`
	GracefulExitTimeoutMills int    `yaml:"graceful-exit-timeout-mills"`
}

type NacosConfig struct {
	NamespaceId string `yaml:"namespace-id"`
	Addresses   string `yaml:"addresses"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	LogLevel    string `yaml:"log-level"`
	LogDir      string `yaml:"log-dir"`
	CacheDir    string `yaml:"cache-dir"`
}

type NacosCenterConfig struct {
	ConfigCfg           NacosConfigConfig           `yaml:"config"`
	RegistryDiscoverCfg NacosRegistryDiscoverConfig `yaml:"registry-discover"`
}

type NacosConfigConfigItem struct {
	Name      string
	GroupName string
}

type NacosConfigConfig struct {
	ClusterName string                  `yaml:"cluster-name"`
	GroupName   string                  `yaml:"group-name"`
	Static      []NacosConfigConfigItem `yaml:"static"`
	Dynamic     []NacosConfigConfigItem `yaml:"dynamic"`
}

type NacosRegistryDiscoverConfig struct {
	ClusterName              string `yaml:"cluster-name"`
	GroupName                string `yaml:"group-name"`
	DiscoverDialTimeoutMills int    `yaml:"discover-dial-timeout-mills"`
}

type LogConfig struct {
	Level        string          `yaml:"level"`
	FileLogCfg   FileLogConfig   `yaml:"file-log-config"`
	RemoteLogCfg RemoteLogConfig `yaml:"remote-log-config"`
}

type FileLogConfig struct {
	FilePath    string `yaml:"file-path"`
	MaxFileSize int    `yaml:"max-file-size"`
	MaxBackup   int    `yaml:"max-backup"`
	HistoryDays int    `yaml:"history-days"`
	Compress    bool   `yaml:"compress"`
}

type RemoteLogConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	ReconnectMaxDelayMills int    `yaml:"reconnect-max-delay-mills"`
	QueueSize              int    `yaml:"queue-size"`
	StopTimeoutMills       int    `yaml:"stop-timeout-mills"`
	MustConnectedInInit    bool   `yaml:"must-connected-in-init"`
	BatchTimingMills       int    `yaml:"batch-timing-mills"`
	BatchQuantitativeSize  int    `yaml:"batch-quantitative-size"`
	Debug                  bool   `yaml:"debug"`
}
