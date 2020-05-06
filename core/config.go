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
	Directory     string   `json:"directory"`
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
	//if err := util.IsPathExists(util.ConfigPath); err != nil {
	err = util.MakeDIR(util.ConfigPath)
	if err != nil {
		return err
	}
	//}
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
      "name": "Ucloud SDK modify By sevth",
      "public_key": "TOKEN***********************042",
      "private_key": "cfc***********************7fc",
      "说明2": "以下两个参数是用来管理文件用的。对应的是 file.go 里面的接口，file_host 是不带 bucket 名字的。比如：北京地域的host填cn-bj.ufileos.com，而不是填 bucketname.cn-bj.ufileos.com。如果是自定义域名，请直接带上 http 开头的 URL。如：http://example.com，而不是填 example.com。",
      "bucket_name": "bucket",
      "endpoint": "cn-gd.ufileos.com",
      "custom_domain": "http://example.com"
    },
    "aliyun": {
      "name": "Aliyun oss SDK",
      "access_key_id": "LT***********************KS",
      "access_key_secret": "Tu***********************e0",
      "bucket_name": "sevth-test",
      "endpoint": "oss-cn-shenzhen.aliyuncs.com",
      "custom_domain": ""
    },
    "tencent": {
      "name": "Tencent cos SDK",
      "secret_id": "TOKEN***********************7042",
      "secret_key": "cfc*************************7fc",
      "session_token": "",
      "bucket_name": "bucket",
      "endpoint": "cos.COS_REGION.myqcloud.com",
      "custom_domain": "https://example.com"
    },
    "baidu": {
      "name": "baidu bos SDK",
      "access_key_id": "a8***********************38",
      "secret_access_key": "91a***********************c4e",
      "bucket_name": "sevth",
      "endpoint": "gz.bcebos.com",
      "custom_domain": ""
    },
    "jd": {
      "name": "JD oss SDK",
      "access_key_id": "8A***********************F5",
      "access_key_secret": "64***********************9A",
      "bucket_name": "sevth",
      "endpoint": "s3.cn-south-1.jdcloud-oss.com",
      "custom_domain": ""
    },
    "qiniu": {
      "name": "Qiniu oss SDK",
      "ak": "LN***********************f3",
      "sk": "IA***********************6F",
      "bucket_name": "sevth",
      "qiniu说明": "下面的endpoint没啥用处，写不写都无所谓,将测试域名或者自定义域名填写到 custom_domain 里面",
      "endpoint": "s3-cn-south-1.qiniucs.com",
      "custom_domain": "https://example.com"
    },
    "upyun": {
      "name": "Upyun oss SDK",
      "operator": "root",
      "password": "4n***********************7Tk",
      "bucket_name": "sevth",
      "endpoint": "test.upcdn.net",
      "custom_domain": ""
    },
    "smms": {
      "name": "smms",
      "access_token": "EVYkI2DGsBGcWnt8LK4AtGoGag3qcyQY",
      "proxy": ""
    },
    "gitee": {
      "name": "gitee",
      "owner": "sevth",
      "repo": "image",
      "access_token": "00***********************82"
    }
  },
  "dir说明": "存放的文件目录 {R} 根据文件后缀判断文件类型，使用对应的路径，时间格式 {Y}:2020 {y}:20 {M}:Apr {m}:04 {d}:01",
  "directory": "test/{Y}/{m}",
  "primary_domain": "",
  "uses": [
    "ucloud"
  ],
  "dsn": {
    "uses": "sqlite3",
    "protocol": "",
    "username": "",
    "password": "",
    "dbname": "",
    "dsn_link": "",
    "debug": false
  }
}
`
}
