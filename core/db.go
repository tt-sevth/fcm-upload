/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: db.go
 * Date: 2020/5/2 上午12:56
 * Author: sevth
 */

package core

import (
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var engine *xorm.Engine
var db *Db

type Db struct {
	Uses     string `json:"uses"`
	Protocol string `json:"protocol"`
	Username string `json:"username"`
	Password string `json:"password"`
	Dbname   string `json:"dbname"`
	DsnLink  string `json:"dsn_link"`
	Debug    bool   `json:"debug"`
}

type DbData struct {
	Id       int64
	FileName string    `xorm:"varchar(100)" json:"file_name"`
	FileMd5  string    `xorm:"varchar(32) index" json:"file_md5"`
	FileMime string    `xorm:"varchar(255)" json:"file_mime"`
	FileKey  string    `xorm:"varchar(255)" json:"file_key"`
	FilePath string    `xorm:"varchar(255)" json:"file_path"`
	Uses     string    `xorm:"varchar(15) index" json:"uses"`
	Link     string    `xorm:"varchar(255)" json:"link"`
	Created  time.Time `xorm:"created" json:"created"`
}

// DB 数据库的初始化
// sql 通用连接，需要复杂连接使用 dsn 语句
func InitDB() *Db {
	var err error
	db = config.Dsn
	if db.Uses == "mysql" {
		engine, err = xorm.NewEngine("mysql", db.Username+":"+
			db.Password+"@"+db.Protocol+"/"+db.Dbname)
		if err != nil {
			return nil
		}
	}
	if db.Uses == "mssql" {
		engine, err = xorm.NewEngine("mssql", db.Protocol+"*"+
			db.Dbname+"/"+db.Username+"/"+db.Password)
		if err != nil {
			return nil
		}
	}
	if db.Uses == "postgres" {
		engine, err = xorm.NewEngine("postgres", "postgres://"+db.Username+":"+
			db.Password+"@"+db.Protocol+"/"+db.Dbname)
		if err != nil {
			return nil
		}
	}
	if db.DsnLink != "" {
		engine, err = xorm.NewEngine(db.Uses, db.DsnLink)
		if err != nil {
			return nil
		}
	}
	if db.Uses == "sqlite3" || db.Uses == "" {
		engine, err = xorm.NewEngine("sqlite3", util.ConfigPath+"/sqlite3.db")
		if err != nil {
			return nil
		}
	}

	_ = engine.Sync2(new(DbData))

	db.logger(db.Debug)

	return db
}

// 日志记录 当配置文件开启时记录 (目前好像没什么用 =.=)
func (d Db) logger(b bool) {
	if b {
		engine.ShowSQL(true)
		engine.Logger().SetLevel(log.LOG_DEBUG)
		f, _ := os.Create(util.LogPath + "/sql.log")
		defer f.Close()
		engine.SetLogger(log.NewSimpleLogger(f))
	}
}

// 保存数据到数据库
func (d Db) Save(data [][]*DbData) error {
	for k, v := range data {
		for i, d := range v {
			if util.Except[k][i] {
				continue
			}
			_, err := engine.Insert(d)
			if err != nil {
				return err
			}
		}
	}
	//defer engine.Close()	这里不需要关闭，内部实现会自动关闭
	return nil
}

// 控制台查询 返回多条记录
func (d Db) Query(filePath string) []*DbData {
	md5 := util.GetFileMD5(filePath)
	data := make([]*DbData, 0)
	_ = engine.Where("file_md5 = ?", md5).Find(&data)
	return data
}

// 查询文件是否存在记录并返回一条记录
func (d Db) QueryOne(md5, uses string) *DbData {
	data := &DbData{}
	has, err := engine.Where("file_md5 = ? and uses = ?", md5, uses).Get(data)
	//defer engine.Close() 	// 查询会自动关闭，这里不需要
	if err != nil {
		util.Log.Error("查询数据发生了错误 - ", err)
	}
	if has {
		return data
	}
	return nil
}

// 导出所有数据
func (d Db) Dump() {
	_ = engine.DumpAllToFile(util.SavePath + "/dumpAll.sql")
}

// 同步删除云数据
func (d *Db) DeleteSync(filePath string)(int,int) {
	data := d.Query(filePath)
	if len(data) == 0 {
		return 0,0
	}
	a,b := Delete(data)
	d.delete(data[0].FileMd5)
	return a,b
}

// 删除某一个文件的数据
func (d *Db) delete(md5 string) {
	_, _ = engine.Where("file_md5 = ?", md5).Delete(&DbData{})
}
