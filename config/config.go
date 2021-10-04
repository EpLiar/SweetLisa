package config

import (
	"github.com/e14914c0-6759-480d-be89-66b7b7676451/SweetLisa/pkg/log"
	"github.com/stevenroose/gonfig"
	log2 "log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Params struct {
	Address             string `id:"address" short:"a" default:"0.0.0.0:2017" desc:"Listening address"`
	Config              string `id:"config" short:"c" desc:"SweetLisa configuration directory"`
	BotToken            string `id:"bot-token"`
	Host                string `id:"host"`
	LogLevel            string `id:"log-level" default:"info" desc:"Optional values: trace, debug, info, warn or error"`
	LogFile             string `id:"log-file" desc:"The path of log file"`
	LogMaxDays          int64  `id:"log-max-days" default:"3" desc:"Maximum number of days to keep log files"`
	LogDisableColor     bool   `id:"log-disable-color"`
	LogDisableTimestamp bool   `id:"log-disable-timestamp"`
}

var params Params

func initFunc() {
	err := gonfig.Load(&params, gonfig.Conf{
		FileDisable:       true,
		FlagIgnoreUnknown: false,
		EnvPrefix:         "LISA_",
	})
	if err != nil {
		if err.Error() != "unexpected word while parsing flags: '-test.v'" {
			log2.Fatal(err)
		}
	}
	if params.Config == "" {
		params.Config = "/etc/sweetLisa"
	}
	// replace all dots of the filename with underlines
	params.Config = filepath.Join(
		filepath.Dir(params.Config),
		strings.ReplaceAll(filepath.Base(params.Config), ".", "_"),
	)
	if strings.Contains(params.Config, "$HOME") {
		if h, err := os.UserHomeDir(); err == nil {
			params.Config = strings.ReplaceAll(params.Config, "$HOME", h)
		}
	}
	logWay := "console"
	if params.LogFile != "" {
		logWay = "file"
	}
	log.InitLog(logWay, params.LogFile, params.LogLevel, params.LogMaxDays, params.LogDisableColor, params.LogDisableTimestamp)
}

var once sync.Once

func GetConfig() *Params {
	once.Do(initFunc)
	return &params
}

func SetConfig(config Params) {
	params = config
}
