package dbManager

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mini-tiger/fast-api/core"
	"gorm.io/gorm"
)

type BaseModel struct {
	CreatedTime string `gorm:"column:create_time" json:"create_time"`
	UpdatedTime string `gorm:"column:update_time" json:"update_time"`
}

func (b *BaseModel) BeforeCreate(*gorm.DB) error {
	b.CreatedTime = time.Now().Format("2006-01-02 15:16:05")
	b.UpdatedTime = b.CreatedTime
	return nil
}

func (b BaseModel) BeforeUpdate(*gorm.DB) error {
	b.UpdatedTime = time.Now().Format("2006-01-02 15:16:05")
	return nil
}

type AdminModel struct {
	BaseModel
	Id        int    `gorm:"primaryKey" json:"id"`
	AdminName string `json:"admin_name"`
	Password  string `json:"-"`
	Mobile    string `json:"phone"`
}

func (a *AdminModel) TableName() string {
	return "v1_admin"
}

func TestCreate(t *testing.T) {
	core.Start()

	adminInfo := new(AdminModel)

	GetInstance().Debug().Create(adminInfo)
	spew.Dump(adminInfo.Id)
}

func TestGet(t *testing.T) {
	adminInfo := new(AdminModel)

	GetInstance().Debug().Where(10).First(adminInfo)
	spew.Dump(adminInfo.CreatedTime)
}
