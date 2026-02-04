package dbManager

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mini-tiger/fast-api/core"
	gormLogger "gorm.io/gorm/logger"
)

type loggerType struct {
	LogLevel                            string
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *loggerType) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

// Info print info
func (l loggerType) Info(ctx context.Context, msg string, data ...interface{}) {
	spew.Dump(data)
}

// Warn print warn messages
func (l loggerType) Warn(ctx context.Context, msg string, data ...interface{}) {
	//spew.Dump(data)
}

// Error print error messages
func (l loggerType) Error(ctx context.Context, msg string, data ...interface{}) {
	//spew.Dump(data)
}

// Trace print sql message
func (l loggerType) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rowNum := fc()
	if "INSERT INTO `log`" == sql[:17] {
		return
	}
	if "SELECT FOUND_ROWS() as total" == sql {
		return
	}
	printString := ""

	if nil != err && "record not found" != err.Error() {
		printString = fmt.Sprintf(gormLogger.Red+"%s (%s)"+gormLogger.Reset, sql, err.Error())
	} else {
		// 仅本地环境打印sql正常日志
		if "local" != core.Mode {
			return
		}
		printString = fmt.Sprintf(gormLogger.Green+"%s (%d)"+gormLogger.Reset, sql, rowNum)
	}

	fmt.Println(printString)
}
