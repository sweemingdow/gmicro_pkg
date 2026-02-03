package config

import (
	"github.com/sweemingdow/gmicro_pkg/pkg/parser/yaml"
	"github.com/sweemingdow/gmicro_pkg/pkg/utils"
	"time"
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
	ClusterName         string        `yaml:"cluster-name"`
	GroupName           string        `yaml:"group-name"`
	DiscoverDialTimeout time.Duration `yaml:"discover-dial-timeout"`
}

type LogConfig struct {
	Level           string             `yaml:"level"`
	LogAsyncCfg     LogAsyncConfig     `yaml:"log-async-config"`
	FileLogCfg      FileLogConfig      `yaml:"file-log-config"`
	TcpLogWriterCfg TcpLogWriterConfig `yaml:"tcp-log-writer-config"`
}

type FileLogConfig struct {
	FilePath    string `yaml:"file-path"`
	MaxFileSize int    `yaml:"max-file-size"`
	MaxBackup   int    `yaml:"max-backup"`
	HistoryDays int    `yaml:"history-days"`
	Compress    bool   `yaml:"compress"`
}

type LogAsyncConfig struct {
	QueueSize        int           `yaml:"queue-size"`
	QuantitativeSize int           `yaml:"quantitative-size"`
	Timing           time.Duration `yaml:"timing"`
	StopTimeout      time.Duration `yaml:"stop-timeout"`
	FlushWorkers     int           `yaml:"flush-workers"`
	Debug            bool          `yaml:"debug"`
}

type TcpLogWriterConfig struct {
	Host                string        `yaml:"host"`
	Port                int           `yaml:"port"`
	KeepAlive           time.Duration `yaml:"keep-alive"`
	ReconnectMaxDelay   time.Duration `yaml:"reconnect-max-delay"`
	DialTimeout         time.Duration `yaml:"dial-timeout"`
	WriteTimeout        time.Duration `yaml:"write-timeout"`
	Debug               bool          `yaml:"debug"`
	MustConnectedInInit bool          `yaml:"must-connected-in-init"`
}
