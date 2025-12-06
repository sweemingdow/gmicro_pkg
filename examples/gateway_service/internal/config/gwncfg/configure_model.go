package gwncfg

type GatewayStaticConfig struct {
	HttpClientCfg HttpClientConfig `yaml:"http-client-config"`
}

type GatewayDynamicConfig struct {
	LogLevel map[string]string `yaml:"log-level"`
}

type HttpClientConfig struct {
	MaxConnsPerHost          int `yaml:"max-conns-per-host"`
	MaxIdleConnDurationMills int `yaml:"max-idle-conn-duration-mills"`
	MaxConnDurationMills     int `yaml:"max-conn-duration-mills"`
	ReadTimeoutMills         int `yaml:"read-timeout-mills"`
	WriteTimeoutMills        int `yaml:"write-timeout-mills"`
	ShutdownTimeoutMills     int `yaml:"shutdown-timeout-mills"`
}
