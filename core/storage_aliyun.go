/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: aliyun.go
 * Date: 2020/5/3 上午3:44
 * Author: sevth
 */

package core

import "github.com/aliyun/aliyun-oss-go-sdk/oss"

type Aliyun struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func (a Aliyun)upload(info *fileInfo) (link string) {
	util.Log.Info("使用 aliyun SDK 上传")
	AliConfig := config.StorageTypes.Aliyun
	client, err := oss.New(AliConfig.Endpoint, AliConfig.AccessKeyId, AliConfig.AccessKeySecret)
	if err != nil {
		util.Log.Error("Aliyun SDK throw err ", err)
		return
	}
	bucket, err := client.Bucket(AliConfig.BucketName)
	if err != nil {
		util.Log.Error("Aliyun SDK throw err ", err)
		return
	}
	if info.fileSize <= maxFileSize {	// 直接上传，最大5G文件
		err = bucket.PutObjectFromFile(info.fileKey, info.filePath)
	} else {	// 分片上传，支持断点续传
		err = bucket.UploadFile(info.fileKey, info.filePath, partSize, oss.Routines(8), oss.Checkpoint(true, ""))
	}
	if err != nil {
		util.Log.Error("Aliyun SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(a.CustomDomain, a.BucketName, a.Endpoint, info.fileKey)
}
