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

func (b Baidu)upload(info *fileInfo) (link string) {
	util.Log.Info("使用 Baidu SDK 上传")
	client, err := bos.NewClient(b.AccessKeyId, b.SecretAccessKey, b.Endpoint)
	if err != nil {
		util.Log.Error("Baidu SDK throw err ", err)
		return
	}

	if info.fileSize <= maxFileSize {
		_, err = client.PutObjectFromFile(b.BucketName, info.fileKey, info.filePath, nil)
	} else {
		_, err = client.ParallelUpload(b.BucketName, info.fileKey, info.filePath, "", nil)

	}

	if err != nil {
		util.Log.Error("Baidu SDK throw err ", err)
		return
	}
	return util.MakeReturnLink(b.CustomDomain, b.BucketName, b.Endpoint, info.fileKey)
}
