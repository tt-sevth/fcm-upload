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

	var ExceptUses = []string{
		"smms",
		"gitee",
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
		if res[0].String() == "" {
			err = errors.New("上传至空间'" + v + "'发生错误！")
		}
		if err != nil {
			util.Log.Error(err)
			//_ = util.SendUploadFailedNotify(method)
			continue // 上传错误的话跳过此条
		}
		temp := res[0].String()
		ExceptSave = append(ExceptSave, false)
		util.Log.Info(v, "返回结果为 - ", temp)
		result = append(result, makeData(v, temp, info))
	}
	util.Log.Info(`-------------------此文件已全部处理完成。-------------------`)
	return
}

func makeData(usesName, res string, info *fileInfo) *DbData {
	D := &DbData{
		FileName: info.fileName,
		FileMd5:  info.fileMD5,
		FileMime: info.fileMime,
		FileKey:  info.fileKey,
		FilePath: info.filePath,
		Uses:     usesName,
		Link:     res,
	}

	if config.PrimaryDomain != "" {
		D.Link = config.PrimaryDomain + "/" + info.fileKey
	}
	return D
}

// delete method
func Delete(data []*DbData) (int, int) {
	var storage = config.StorageTypes
	var DeleteStorage = map[string]interface{}{
		"ucloud": storage.Ucloud.delete,
		"aliyun": storage.Aliyun.delete,
		"baidu":  storage.Baidu.delete,
		"gitee":  storage.Gitee.delete,
		"jd":     storage.JD.delete,
		"qiniu":  storage.Qiniu.delete,
		"tencent": storage.Tencent.delete,
		"upyun": storage.Upyun.delete,
	}

	var success, fail int

	var ExceptUses = []string{
		"smms",
	}
	for _, v := range data {
		if collection.Collect(ExceptUses).Contains(v.Uses) {
			continue
		}
		res, err := call(DeleteStorage, v.Uses, &fileInfo{
			fileKey: v.FileKey,
		})

		if res[0].Interface() == false {
			err = errors.New("云存储空间'" + v.Uses + "'删除失败！")
		}
		if err != nil {
			fail++
			util.Log.Error("删除文件'"+v.FileKey+"'存在错误 ", err)
			continue
		}
		success++
	}
	return success, fail
}

func call(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}
