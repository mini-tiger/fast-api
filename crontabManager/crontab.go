package crontabManager

import "github.com/robfig/cron/v3"

var server *cron.Cron

func init() {
	server = cron.New()
}

func GetInstance() *cron.Cron {
	return server
}
