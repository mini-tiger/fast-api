package dLogger

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mini-tiger/fast-api/config"
	"github.com/mini-tiger/fast-api/core"
)

// 错误等级
const (
	LeverInfo   LogLevelType = "info"
	LeverWaning LogLevelType = "warning"
	LeverError  LogLevelType = "error"
)

type LogLevelType string

var source string
var mode core.ModeType

// 初始化
func init() {
	source = config.GetInstance().Section("core").Key("serverName").Value()
	mode = core.Mode
}

// Write 写入日志到logStash
func Write(logLevel LogLevelType, typeString string, message any) {
	if core.Mode == core.Dev {
		toWrite(logLevel, typeString, message)
	} else {
		// 开启一个携程异步写入
		go func() {
			toWrite(logLevel, typeString, message)
		}()
	}
}

func toWrite(logLevel LogLevelType, typeString string, message any) {
	messageNew := ""
	switch v := message.(type) {
	case string:
		messageNew = v
		break
	case error:
		messageNew = v.Error()
	default:
		messageJson, err := json.Marshal(message)
		messageNew = string(messageJson)
		if nil != err {
			messageNew = fmt.Sprintf("日志记录错误：%s\n", err.Error())
		}
	}
	if core.Mode == core.Dev {
		fmt.Printf("[%s][%s][%s]\n", logLevel, typeString, messageNew)
	}
	logData := &LogModelType{
		Source:     source,
		Mode:       mode,
		LogLevel:   logLevel,
		Type:       typeString,
		Message:    messageNew,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	_, err := logData.Create()
	if nil != err {
		fmt.Printf("写入日志失败： %s", err.Error())
	}
}
