package database

import (
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Mysql *gorm.DB
)

type GVA_MODEL struct {
	ID        uint      `json:"id" gorm:"primarykey"`      // 主键ID
	CreatedAt time.Time `json:"ctime" gorm:"column:ctime"` // 创建时间
}

func (model *GVA_MODEL) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	model.CreatedAt = now
	return nil
}

/*
*

初始化mysql链接
设置连接池信息
返回 错误信息，如果没有返回则 创建连接成功
*/
func InitMysql() (err error) {
	// 建立连接
	databaseDsn := fmt.Sprintf("root:20010214@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local")

	Mysql, err = gorm.Open(mysql.Open(databaseDsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Error(" 创建数据库连接失败 ...")
	} else {
		log.Error(" 创建数据库连接成功 ...")
	}

	sqlDB, err := Mysql.DB()

	// 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(100)

	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(150)

	// 设置连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Second * 30)

	return err
}
