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
	//Directory       string `json:"directory"`
	CustomDomain    string `json:"custom_domain"`
}

func aliyun(FilePath string) (link string) {
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
	err = aliyunUploadMethod(FilePath, bucket)
	if err != nil {
		util.Log.Error("Aliyun SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(AliConfig.CustomDomain, AliConfig.BucketName, AliConfig.Endpoint)

}

func aliyunUploadMethod(FilePath string, bucket *oss.Bucket) (err error) {
	if fileSize <= maxFileSize {
		err = bucket.PutObjectFromFile(fileKey, FilePath)
		return
	}
	err = bucket.UploadFile(fileKey, FilePath, partSize, oss.Routines(8), oss.Checkpoint(true, ""))
	return
}

