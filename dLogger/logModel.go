package dLogger

import (
	"github.com/mini-tiger/fast-api/core"
	"github.com/mini-tiger/fast-api/dbManager"
)

type LogModelType struct {
	Id         int           `gorm:"primaryKey"`
	Source     string        `json:"source"`
	Mode       core.ModeType `json:"mode"`
	LogLevel   LogLevelType  `json:"log_level"`
	Type       string        `json:"type"`
	Message    string        `json:"message"`
	CreateTime string        `json:"time"`
}

func (l *LogModelType) TableName() string {
	return "log"
}

func (l *LogModelType) Create() (int64, error) {
	db := dbManager.GetInstance().Create(l)
	return db.RowsAffected, db.Error
}
