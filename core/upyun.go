/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: upyun.go
 * Date: 2020/5/4 下午8:14
 * Author: sevth
 */

package core

import upyun2 "github.com/upyun/go-sdk/upyun"

type Upyun struct {
	Name         string `json:"name"`
	Operator     string `json:"operator"`
	Password     string `json:"password"`
	BucketName   string `json:"bucket_name"`
	Endpoint     string `json:"endpoint"`
	CustomDomain string `json:"custom_domain"`
}

func upyun() (link string) {
	UConfig := config.StorageTypes.Upyun
	client := upyun2.NewUpYun(&upyun2.UpYunConfig{
		Bucket:   UConfig.BucketName,
		Operator: UConfig.Operator,
		Password: UConfig.Password,
	})

	err := upyunUploadMethod(client)
	if err != nil {
		util.Log.Error("Upyun SDK throw err ", err)
		return
	}
	return util.MakeReturnLink(UConfig.CustomDomain, UConfig.BucketName, UConfig.Endpoint)
}

func upyunUploadMethod(client *upyun2.UpYun) (err error) {
	// form 表单使用 post 接口，不是单独的 put ，所以不进行分别使用不同接口，直接用put 具体实现在内部，并没有相关接口，只有设置选项
	if fileSize <= maxFileSize {
		return client.Put(&upyun2.PutObjectConfig{
			Path:              fileKey,
			LocalPath:         filePath,
			UseMD5:            true,
			UseResumeUpload:   false,
			ResumePartSize:    partSize,
			MaxResumePutTries: 0,
		})
	}
	return client.Put(&upyun2.PutObjectConfig{
		Path:              fileKey,
		LocalPath:         filePath,
		UseMD5:            true,
		UseResumeUpload:   true,
		ResumePartSize:    partSize,
		MaxResumePutTries: 3,
	})
}
