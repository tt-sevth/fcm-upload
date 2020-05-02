/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: config.go
 * Date: 2020/5/1 下午14:16
 * Author: sevth
 */

package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type Config struct {
	Name          string   `json:"name"`
	StorageTypes  *Storage `json:"storage_types"`
	Directory	  string	`json:"directory"`
	PrimaryDomain string   `json:"primary_domain"`
	Uses          []string `json:"uses"`
	Dsn           *Db      `json:"dsn"`
}

var config *Config

//LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	file, err := os.Open(util.ConfigPath + "/config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	cb, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config = new(Config)
	err = json.Unmarshal(cb, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// 初始化配置文件
func InitConfig() error {
	if !util.IsFileExist(util.ConfigPath + "/config.json") {
		fmt.Println("=====================")
		fmt.Println("config 配置文件不存在")
		fmt.Println("尝试创建 config 配置文件")
		err := createConfigFile()
		if err != nil {
			fmt.Println("创建失败，错误如下：", err)
			return err
		}
	}
	var execOpen string
	if util.GetOS() == 1 {
		execOpen = "open"
	}
	if util.GetOS() == 2 {
		execOpen = "nano"
	}
	if util.GetOS() == 3 {
		execOpen = "start"
	}
	if util.GetOS() == 0 {
		fmt.Println("未知操作系统，请手动打开并编辑配置文件")
		return nil
	}
	cmd := exec.Command(execOpen, util.ConfigPath+"/config.json")
	err := cmd.Start()
	if err != nil {
		return err
	}
	fmt.Println("创建完成，已打开文件")
	fmt.Println("=====================")
	return nil
}

//createConfigFile 创建配置文件
func createConfigFile() (err error) {
	if err := util.IsPathExists(util.ConfigPath); err != nil {
		err = util.MakeDIR(util.ConfigPath)
		if err != nil {
			return err
		}
		return nil
	}
	f, err := util.OpenFile(util.ConfigPath + "/config.json")
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return
	}
	_, err = f.WriteString(configTemplate())
	return
}

// 配置文件模板
func configTemplate() string {
	return `
{
    "name": "FCM 配置文件",
    "storage_types": {
        "ucloud": {
            "name": "ucloud",
            "说明1":"管理 bucket 创建和删除必须要公私钥，如果只做文件上传和下载用TOEKN就够了，为了安全，强烈建议只使用 TOKEN 做文件管理",
            "public_key":"TOKEN***********************7042",
            "private_key":"cfc*************************7fc",

            "说明2":"以下两个参数是用来管理文件用的。对应的是 file.go 里面的接口，file_host 是不带 bucket 名字的。比如：北京地域的host填cn-bj.ufileos.com，而不是填 bucketname.cn-bj.ufileos.com。如果是自定义域名，请直接带上 http 开头的 URL。如：http://example.com，而不是填 example.com。",
            "bucket_name":"bucket",
            "file_host":"cn-gd.ufileos.com",
            "说明3": "存放的文件目录与自定义域名 {R} 根据文件后缀判断文件类型，使用对应的路径，时间格式 {Y} 2020 {y} 20 {M} Apr {m} 04 {d} 01",
            "directory": "test/{Y}",
            "custom_domain": "https://example.com"
        }
    },
    "primary_domain": "",
    "uses": ["ucloud"],
    "dsn": {
        "uses": "sqlite3",
        "protocol": "",
        "username": "",
        "password": "",
        "dbname": "",
        "dsn_link": "",
        "debug": true
    }
}`
}
