/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: baidu.go
 * Date: 2020/5/4 上午11:03
 * Author: sevth
 */

package core

import (
	"github.com/baidubce/bce-sdk-go/services/bos"
)

type Baidu struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func baidu() (link string) {
	util.Log.Info("使用 Baidu SDK 上传")
	BConfig := config.StorageTypes.Baidu
	client, err := bos.NewClient(BConfig.AccessKeyId, BConfig.SecretAccessKey, BConfig.Endpoint)
	if err != nil {
		util.Log.Error("Baidu SDK throw err ", err)
		return
	}
	err = baiduUploadMethod(client, BConfig.BucketName)
	if err != nil {
		util.Log.Error("Baidu SDK throw err ", err)
		return
	}
	return util.MakeReturnLink(BConfig.CustomDomain, BConfig.BucketName, BConfig.Endpoint)
}

func baiduUploadMethod(client *bos.Client, bucket string) (err error) {
	if fileSize <= maxFileSize {
		_, err = client.PutObjectFromFile(bucket, fileKey, filePath, nil)
		return
	}
	_, err = client.ParallelUpload(bucket, fileKey, filePath, "", nil)
	return
}
