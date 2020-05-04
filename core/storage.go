/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: storage.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import (
	"errors"
	"reflect"
)

const (
	partSize    int64 = 2 << 21
	maxFileSize       = 2 << 26
)

type Storage struct {
	Ucloud  *Ucloud  `json:"ucloud"`
	Aliyun  *Aliyun  `json:"aliyun"`
	Tencent *Tencent `json:"tencent"`
	Baidu   *Baidu   `json:"baidu"`
	Qiniu   *Qiniu   `json:"qiniu"`
	Upyun   *Upyun   `json:"upyun"`
}

var filePath, fileName, fileMD5, fileMime, fileKey string
var fileSize int64

func Upload(path string) (result []*DbData, EOne []bool) {
	util.Log.Info("初始化上传模块。")
	fileName, fileMD5, fileMime, filePath, fileSize = util.FileInfo(path)
	fileKey = util.MakeFileKey(config.Directory, filePath)

	if len(config.Uses) < 1 {
		util.Log.Fatal("未设置可用的存储服务商。")
	}
	util.Log.Info("将使用的所有上传服务商是 - ", config.Uses)
	funcMap := getStorageMethodMap()
	util.Log.Info("获取 funcMap 完成 - ", funcMap)

	//var filePath string = "./test.tar.gz"
	for _, v := range config.Uses {
		if v == "" {
			util.Log.Error("uses 设置不正确")
			continue //为空的选项跳过
		}
		// =======================================
		// 处理数据库已存在的记录，直接返回
		if res := db.QueryOne(fileMD5, v); res != nil {
			util.Log.Info(`数据库中已存在记录，跳过上传'`, fileName, `'到'`, v, `'`)
			EOne = append(EOne, true)
			result = append(result, res)
			continue
		}
		// =======================================
		util.Log.Info(v, "上传至 bucket 的全路径为 ", fileKey)
		res, err := call(funcMap, v)
		if err != nil {
			util.Log.Error(err)
			//_ = util.SendUploadFailedNotify(method)
			continue // 上传错误的话跳过此条
		}
		EOne = append(EOne, false)
		util.Log.Info(v, "返回结果为 - ", res)
		result = append(result, makeData(v, filePath, res))
	}
	util.Log.Info(`-------------------此文件已全部处理完成。-------------------`)
	return
}

func getStorageMethodMap() map[string]interface{} {
	return map[string]interface{}{
		"ucloud":  ucloud,
		"aliyun":  aliyun,
		"tencent": tencent,
		"baidu":   baidu,
		"qiniu":   qiniu,
		"upyun":   upyun,
	}
}

func makeData(usesName, filePath, res string) *DbData {
	D := new(DbData)
	D.Uses = usesName
	D.FileName = fileName
	D.FileMd5 = fileMD5
	D.FileMime = fileMime
	D.FilePath = filePath
	D.Link = res
	if config.PrimaryDomain != "" {
		D.Link = config.PrimaryDomain + "/" + fileKey
	}
	return D
}

func call(m map[string]interface{}, name string, params ...interface{}) (result string, err error) {
	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	res := f.Call(in)
	if res[0].String() == "" {
		err = errors.New("上传至空间'" + name + "'发生错误！")
	}
	result = res[0].String()
	return
}
