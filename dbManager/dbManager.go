package dbManager

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mini-tiger/fast-api/config"
	"github.com/mini-tiger/fast-api/core"
	"github.com/mini-tiger/fast-api/dError"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {
	mysqlConfig := config.GetInstance().Section("mysql")
	username := mysqlConfig.Key("username").Value()
	password := mysqlConfig.Key("password").Value()
	host := mysqlConfig.Key("host").Value()
	port := mysqlConfig.Key("port").Value()
	dbname := mysqlConfig.Key("dbname").Value()
	timeout := mysqlConfig.Key("timeout").Value()

	//拼接下dsn参数, dsn格式可以参考上面的语法，这里使用Sprintf动态拼接dsn参数，因为一般数据库连接参数，我们都是保存在配置文件里面，需要从配置文件加载参数，然后拼接dsn。
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local&timeout=%s", username, password, host, port, dbname, timeout)

	var err error

	logLevel := logger.Silent
	if core.Mode == core.Dev {
		logLevel = logger.Info
	}

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logLevel,    // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				ParameterizedQueries:      false,       // Don't include params in the SQL log
				Colorful:                  true,
			},
		),
	})

	if nil != err {
		panic(dError.NewError("连接数据库出错", err))
	}
	// 控制数据库连接池
	sqlDB, err := db.DB()

	if nil != err {
		panic(dError.NewError("数据库连接池错误", err))
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
}

func GetInstance() *gorm.DB {
	return db
}
