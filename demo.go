package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"net/http"
	"os"
	"time"
)

var db *gorm.DB

const DefaultDevelopmentEnv = "dev"

//数据库常量
type DbConfig struct {
	Db MysqlConf `yaml:"DB"`
}

type MysqlConf struct {
	AutoCreateTable bool   `yaml:"AutoCreateTable"`
	DriverName      string `yaml:"DriverName"`
	Url             string `yaml:"Url"`
	Username        string `yaml:"Username"`
	Password        string `yaml:"Password"`
	Dialect         string `yaml:"Dialect"`
	MaxIdle         int    `yaml:"maxIdle"`
	MaxOpen         int    `yaml:"maxOpen"`
}

func InitDB(activeEnv string, dbConfig *DbConfig) (err error) {
	db, err := gorm.Open(dbConfig.Db.Dialect, dbConfig.Db.Url)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	if len(activeEnv) != 0 && activeEnv == DefaultDevelopmentEnv {
		db.LogMode(true)
	}
	//SetMaxOpenConns用于设置最大打开的连接数
	//SetMaxIdleConns用于设置闲置的连接数

	//设置闲置的连接数
	db.DB().SetMaxIdleConns(dbConfig.Db.MaxIdle)
	//设置最大打开的连接数
	db.DB().SetMaxOpenConns(dbConfig.Db.MaxOpen)
	db.DB().SetConnMaxLifetime(time.Duration(30) * time.Minute) //心跳時間設置為MySQL的一半
	//注册 db回调钩子操作
	db.Callback().Create().Replace("gorm:create_time", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time", updateTimeStampForUpdateCallback)
	err = db.DB().Ping()
	/**
	 *禁用表名复数>
	 *!!!如不禁用则会出现表 y结尾边ies的问题
	 *!!!如果只是部分表需要使用源表名，请在实体类中声明TableName的构造函数
	 *
	 *func (实体名) TableName() string {
	 *   return "数据库表名"
	 *}
	 */
	db.SingularTable(true)
	return
}

// updateTimeStampForCreateCallback will set `CreatedOn`, `ModifiedOn` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().Unix()
		if createTimeField, ok := scope.FieldByName("CreatedAt"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("UpdatedAt"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// updateTimeStampForUpdateCallback will set `ModifyTime` when updating
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("UpdatedAt", time.Now().Unix())
	}
}

func GetDB() *gorm.DB {
	return db
}

func main() {
	app := gin.Default()
	GetDB()
	CloseDB()
	app.GET("/test", func(context *gin.Context) {
		context.String(http.StatusOK, "dadad")
	})
	app.Run(":8081")
}

// 关闭DB
func CloseDB() {
	fmt.Println("<<<<<<<<<<<<<DB.close.....>>>>>>")
	if db != nil {
		db.Close()
	}
}
