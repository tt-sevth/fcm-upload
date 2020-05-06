/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: storage.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import (
	"errors"
	"github.com/chenhg5/collection"
	"reflect"
)

const (
	partSize    int64 = 2 << 21
	maxFileSize int64 = 2 << 26
)

var ExceptUses = []string{
	"smms",
	"gitee",
}

type Storage struct {
	Ucloud  *Ucloud  `json:"ucloud"`
	Aliyun  *Aliyun  `json:"aliyun"`
	Tencent *Tencent `json:"tencent"`
	Baidu   *Baidu   `json:"baidu"`
	JD      *JD      `json:"jd"`
	Qiniu   *Qiniu   `json:"qiniu"`
	Upyun   *Upyun   `json:"upyun"`
	Smms    *Smms    `json:"smms"`
	Gitee   *Gitee   `json:"gitee"`
}

type fileInfo struct {
	filePath string
	fileName string
	fileMD5  string
	fileMime string
	fileKey  string
	fileSize int64
}

func Upload(path string) (result []*DbData, ExceptSave []bool) {
	util.Log.Info("初始化上传模块。")
	var storage = config.StorageTypes
	var UploadStorage = map[string]interface{}{
		"ucloud":  storage.Ucloud.upload,
		"aliyun":  storage.Aliyun.upload,
		"tencent": storage.Tencent.upload,
		"baidu":   storage.Baidu.upload,
		"jd":      storage.JD.upload,
		"qiniu":   storage.Qiniu.upload,
		"upyun":   storage.Upyun.upload,
		"smms":    storage.Smms.upload,
		"gitee":   storage.Gitee.upload,
	}

	info := &fileInfo{}
	info.fileName, info.fileMD5, info.fileMime, info.filePath, info.fileSize = util.FileInfo(path)
	info.fileKey = util.MakeFileKey(config.Directory, path)

	if len(config.Uses) < 1 {
		util.Log.Fatal("未设置可用的存储服务商。")
	}
	util.Log.Info("将使用的所有上传服务商是 - ", config.Uses)
	util.Log.Info("获取 uploadMap - ", UploadStorage)

	for _, v := range config.Uses {
		if v == "" {
			util.Log.Error("uses 设置不正确")
			continue //为空的选项跳过
		}
		// 添加一个列表，在列表中的uses只能上传图片
		if collection.Collect(ExceptUses).Contains(v) && util.GetArchiveDirName(path) != "picture" {
			util.Log.Error("跳过上传至'", v, "'")
			continue
		}
		// =======================================
		// 处理数据库已存在的记录，直接返回
		if res := db.QueryOne(info.fileMD5, v); res != nil {
			util.Log.Info(`数据库中已存在记录，跳过上传'`, info.fileName, `'到'`, v, `'`)
			ExceptSave = append(ExceptSave, true)
			result = append(result, res)
			continue
		}
		// =======================================
		util.Log.Info("空间'", v, "'fileKey is ", info.fileKey)
		res, err := call(UploadStorage, v, info)
		if err != nil {
			util.Log.Error(err)
			//_ = util.SendUploadFailedNotify(method)
			continue // 上传错误的话跳过此条
		}
		ExceptSave = append(ExceptSave, false)
		util.Log.Info(v, "返回结果为 - ", res)
		result = append(result, makeData(v, res, info))
	}
	util.Log.Info(`-------------------此文件已全部处理完成。-------------------`)
	return
}

func makeData(usesName, res string, info *fileInfo) *DbData {
	D := new(DbData)
	D.Uses = usesName
	D.FileName = info.fileName
	D.FileMd5 = info.fileMD5
	D.FileMime = info.fileMime
	D.FilePath = info.filePath
	D.Link = res
	if config.PrimaryDomain != "" {
		D.Link = config.PrimaryDomain + "/" + info.fileKey
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
