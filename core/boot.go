package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mini-tiger/fast-api/dError"
)

// AppPath 项目根目录 示例： /Users/project/go/src/go-server-image
var AppPath = "/app"

// Mode 项目运行环境 [local, dev, produce]
var Mode = "local"

func init() {
	// 初始化项目根目录
	initAppPth()
	// 初始化运行环境
	initMode()
}

func initAppPth() {
	dirString, err := os.Getwd()
	if nil != err {
		panic(dError.NewError("找不到项目根目录！", err))
	}
	for {
		mainPath := fmt.Sprintf("%s/main.go", dirString)
		if FileExist(mainPath) {
			AppPath = dirString
			return
		}
		// 直到根目录依然没有找到main.go
		if "/" == dirString {
			panic(dError.NewError("找不到项目根目录！", err))
		}
		dirString = filepath.Dir(dirString)
	}
}

func initMode() {
	if 1 == len(os.Args) {
		Mode = "local"
		return
	}
	modeTemp := os.Args[1]
	modeMap := map[string]struct{}{"local": {}, "dev": {}, "produce": {}}

	if _, ok := modeMap[modeTemp]; ok {
		Mode = modeTemp
		return
	}
}

// FileExist 判断文件是否存在
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
