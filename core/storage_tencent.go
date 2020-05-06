/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: tencent.go
 * Date: 2020/5/4 上午12:54
 * Author: sevth
 */

package core

import (
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
)

type Tencent struct {
	Name         string `json:"name"`
	SecretID     string `json:"secret_id"`
	SecretKey    string `json:"secret_key"`
	SessionToken string `json:"session_token"`
	BucketName   string `json:"bucket_name"`
	Endpoint     string `json:"endpoint"`
	CustomDomain string `json:"custom_domain"`
}

func tencent() (link string) {
	util.Log.Info("使用 tencent SDK 上传")
	TConfig := config.StorageTypes.Tencent
	u, _ := url.Parse("https://" + TConfig.BucketName + "." + TConfig.Endpoint)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     TConfig.SecretID,
			SecretKey:    TConfig.SecretKey,
			SessionToken: TConfig.SessionToken,
		},
	})
	if client == nil {
		util.Log.Error("Tencent SDK throw err ", "client 创建失败")
	}
	err := tencentUploadMethod(client)
	if err != nil {
		util.Log.Error("Tencent SDK throw err ", err)
	}
	return util.MakeReturnLink(TConfig.CustomDomain, TConfig.BucketName, TConfig.Endpoint)
}

func tencentUploadMethod(client *cos.Client) (err error) {
	if fileSize <= maxFileSize {
		_, err = client.Object.PutFromFile(context.Background(), fileKey, filePath, nil)
		return
	}
	_, _, err = client.Object.Upload(context.Background(), fileKey, filePath, &cos.MultiUploadOptions{
		PartSize:       4,	//腾讯的SDK设置问题，这里以M为单位 所以不使用 partSize
		ThreadPoolSize: 8,	//8个线程
	})
	return
}
