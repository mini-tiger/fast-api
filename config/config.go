package config

import (
	"fmt"

	"github.com/mini-tiger/fast-api/core"
	"github.com/mini-tiger/fast-api/dError"
	"gopkg.in/ini.v1"
)

var instance *ini.File

func init() {
	configPath := fmt.Sprintf("%s/conf/%s.ini", core.AppPath, core.Mode)
	var err error
	instance, err = ini.Load(configPath)
	if err != nil {
		panic(dError.NewError("读取配置文件出错", err))
	}
}

func GetInstance() *ini.File {
	return instance
}
