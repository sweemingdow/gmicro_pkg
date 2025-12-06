package app

import (
	"fmt"
	"gmicro_pkg/pkg/config"
	"gmicro_pkg/pkg/parser/cmd"
	"gmicro_pkg/pkg/parser/json"
	"gmicro_pkg/pkg/utils"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	profileDev  = "dev"
	profileTest = "test"
	profileProd = "prod"
)

var (
	ta   *App
	once sync.Once
)

type App struct {
	appId     string
	appName   string
	localIp   string
	cfg       *config.Config
	startTime time.Time
	profile   string
	httpPort  int
	rpcPort   int
}

func NewApp(cp *cmd.CmdParser) *App {
	once.Do(func() {
		cfgPath := cp.GetString("config")
		cfg, err := config.New(cfgPath)
		if err != nil {
			panic(fmt.Sprintf("parse config file:%s failed:%v\n", cfgPath, err))
		}

		var appName = cfg.AppCfg.AppName
		if appName == "" {
			cfgFile := filepath.Base(cfgPath)
			ext := filepath.Ext(cfgFile)
			appName = cfgFile[:strings.Index(cfgFile, ext)]
		}

		ta = &App{
			appId:     fmt.Sprintf("%s#%s", appName, utils.RandStr(8)),
			appName:   appName,
			localIp:   utils.GetLocalIp(),
			cfg:       cfg,
			startTime: time.Now(),
			profile:   cfg.AppCfg.Profile,
			httpPort:  cp.GetInt("http_port"),
			rpcPort:   cp.GetInt("rpc_port"),
		}

	})

	return ta
}

func GetTheApp() *App {
	return ta
}

func (app *App) GetAppId() string {
	return app.appId
}

func (app *App) GetAppName() string {
	return app.appName
}

func (app *App) GetLocalIp() string {
	return app.localIp
}

func (app *App) GetHttpPort() int {
	return app.httpPort
}

func (app *App) GetRpcPort() int {
	return app.rpcPort
}

func (app *App) GetConfig() *config.Config {
	return app.cfg
}

func (app *App) GetProfile() string {
	return app.profile
}

func (app *App) IsDevProfile() bool {
	return app.profile == profileDev
}

func (app *App) IsTestProfile() bool {
	return app.profile == profileTest
}

func (app *App) IsProdProfile() bool {
	return app.profile == profileProd
}

func (app *App) String() string {
	mm := make(map[string]any)
	mm["appId"] = app.appId
	mm["appName"] = app.appName
	//mm["cfg"] = app.cfg
	mm["startTime"] = app.startTime
	mm["profile"] = app.profile
	mm["httpPort"] = app.httpPort
	mm["rpcPort"] = app.rpcPort

	data, err := json.Fmt(&mm)
	if err != nil {
		return ""
	}

	return string(data)
}
