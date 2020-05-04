/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: main.go
 * Date: 2020/5/2 上午1:18
 * Author: sevth
 */

package main

import (
	"./core"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	"os"
	"time"
	_ "time"
)

var util *core.Util

func main() {
	// 初始化工具集(包含日志)
	util = core.LoadUtil()
	defer util.Log.Sync()

	args := os.Args
	fps, method := argsRoute(args)
	// 将使用方法挂载到util对象
	util.Method = method
	// 只有包含文件路径时会继续向下执行，否则直接中断，这里是防止在前面没有退出。
	if len(fps) < 1 {
		_ = util.SendNoFileNotify()
		util.Log.Fatal("未找到有效的要上传文件。")
	}

	util.Log.Info("============= 初始化 Log 完成,准备开始上传任务 =============")
	// 先查询配置文件是否存在，否则直接退出
	if !util.IsFileExist(util.ConfigPath + "/config.json") {
		util.Log.Fatal("未找到配置文件，请先初始化配置文件！")
		fmt.Println("未找到配置文件，请先初始化配置文件！")
	}
	// 加载配置文件 这里不需要返回的config对象，初始化后，内部可以直接调用
	_, _ = core.LoadConfig()
	// 初始化数据库 后面进行数据查询等，先初始化一次获得db对象，避免重复初始化占用资源
	db := core.InitDB()

	util.Log.Info("使用的 method 为 ", method)
	util.Log.Info("一共上传 ", len(fps), " 个文件")
	// 下面开始上传文件
	data := makeUpload(fps)
	// 处理返回的结果，根据不同的使用方法返回不同的数据
	makeResult(data, method)
	// 处理完成后，调用db对象，将数据存进去
	if err := db.Save(data); err != nil {
		util.Log.Error(err)
	}
	util.Log.Info("=======================事务处理完毕! =======================")
	os.Exit(1)
}

// 参数路由
func argsRoute(args []string) (fps []string, method string) {
	if len(args) < 3 {
		goto help
	}
	switch args[1] {
	case "-h", "--help":
		goto help
	case "-i", "--init":
		goto init
	case "-u", "--use":
		goto uses
	case "-d", "--db":
		goto db
	default:
		goto help
	}

help:
	{
		help1 := []string{
			`-h --help`,
			"-i --init  config all",
			"-u --use   console system typora",
			"-d --db  	dump query",
		}
		help2 := []string{
			"to show this help info",
			"init config file; like: -i config",
			"How to run this program; like: -u console",
			"dump all data from database",
		}
		fmt.Printf("    %-s  <option> [args]\n\n", os.Args[0])
		for i := 0; i < len(help1); i++ {
			fmt.Printf("    %-40s%-s\n", help1[i], help2[i])
		}
		os.Exit(1)
	}
init:
	{
		switch args[2] {
		case "all":
			err := util.MakeDIR(util.ConfigPath)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("mkdir config pass")
			}
			err = util.MakeDIR(util.LogPath)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("mkdir log pass")
			}
			err = util.MakeDIR(util.SavePath)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("mkdir save pass")
			}
			err = core.InitConfig()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("init config pass")
			}
		case "config":
			err := core.InitConfig()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("init config pass")
			}
		case "log":
			err := util.MakeDIR(util.LogPath)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("init log pass")
			}
		default:
			goto help
		}
		os.Exit(1)
	}
uses:
	{
		fps = args[3:]
		switch args[2] {
		case "console":
			method = "console"
		case "system":
			method = "system"
		case "typora":
			method = "typora"
		default:
			goto help
		}
		return
	}
db:
	{
		switch args[2] {
		case "dump":
			_, _ = core.LoadConfig()
			core.InitDB().Dump()
		case "query":
			if args[3] == "" {
				fmt.Println("没有要查询的文件路径")
			} else {
				_, _ = core.LoadConfig()
				if data := core.InitDB().Query(args[3]); data != nil {
					fmt.Printf("    %-22s%-10s%-s\n", "文件名", "服务商", "链接")
					for _, d := range data {
						fmt.Printf("    %-25s%-13s%-s\n", d.FileName, d.Uses, d.Link)
					}
				} else {
					fmt.Println("没有要查询的数据")
				}
			}
		default:
			goto help
		}
		os.Exit(1)
	}
	return
}

// 处理上传
func makeUpload(fps []string) [][]*core.DbData {
	var data [][]*core.DbData
	//将需要上传的文件上传
	if util.Method == "system" {
		_ = util.SendStartUploadNotify(len(fps))
	}
	for _, FP := range fps {
		if FP == "" { // 虽然不太可能存在为空，以防万一
			util.Log.Info("发现一条空地址，已跳过")
			continue
		}
		if !util.IsFileExist(FP) { //检查文件是否真实存在，防止后面出错
			util.Log.Info("发现一个不存在的文件，\"" + FP + "\"已跳过")
			continue
		}

		util.Log.Info("准备上传文件", util.GetFileNameWithoutExt(FP))
		result, EOne := core.Upload(FP)
		// 存在长度才会将返回数据加入data切片中
		if len(result) > 0 {
			data = append(data, result)
			util.Except = append(util.Except, EOne)
		}
		//fmt.Println(data)
		//fmt.Println(util.Except)
	}
	util.Log.Info("=========== 所有文件处理完毕 ===========")
	return data
}

// 处理返回结果
func makeResult(Data [][]*core.DbData, method string) {
	if len(Data) == 0 {
		_ = clipboard.WriteAll("没有上传任何文件，请查看日志！")
		if util.Method == "system" {
			_ = util.SendUploadFailedNotify()
		}
		os.Exit(-1)
	}
	switch method {
	case "console":
		fmt.Println("    上传结果如下：")
		fmt.Printf("    %-22s%-10s%-s\n", "文件名", "服务商", "链接")
		for _, v := range Data {
			for _, d := range v {
				fmt.Printf("    %-25s%-13s%-s\n", d.FileName, d.Uses, d.Link)
			}
		}
	case "system":
		if len(Data[0]) == 1 {
			_ = util.SendUploadSuccessNotify(false)
			var name, link []string
			for _, v := range Data {
				name = append(name, v[0].FileName)
				link = append(link, v[0].Link)
			}
			_ = util.SetClipboard(name, link)
		} else {
			_ = util.SendUploadSuccessNotify(true)
			if err := util.IsPathExists(util.SavePath); err != nil {
				_ = util.MakeDIR(util.SavePath)
			}
			f, err := util.OpenFile(util.SavePath + "/system-" + time.Now().Format("2006-01-02 15:04:05") + ".json")
			if err != nil {
				util.Log.Error("打开文件失败，错误如下：" + err.Error())
			}
			if f != nil {
				defer f.Close()
				for _, v := range Data {
					for _, d := range v {
						b, _ := json.Marshal(d)    // 序列化数据
						_, _ = f.Write(b)          // 写入数据
						_, _ = f.Write([]byte{10}) // 写入换行符
					}
				}
			}
		}
	case "typora":
		fmt.Println("Upload result:")
		for _, v := range Data {
			fmt.Println(v[0].Link)
		}
	}
}
