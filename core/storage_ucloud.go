/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: ucloud.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import "github.com/ufilesdk-dev/ufile-gosdk"

type Ucloud struct {
	Name            string `json:"name"`
	PublicKey       string `json:"public_key"`
	PrivateKey      string `json:"private_key"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func ucloud() (link string) {
	util.Log.Info("使用 uclod SDK 上传")
	UConfig := config.StorageTypes.Ucloud
	UC := &ufsdk.Config{
		PublicKey:       UConfig.PublicKey,
		PrivateKey:      UConfig.PrivateKey,
		BucketName:      UConfig.BucketName,
		FileHost:        UConfig.Endpoint,
		VerifyUploadMD5: true,
	}
	client, err := ufsdk.NewFileRequest(UC, nil)
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return
	}
	//fileKey := Util.MakeFileKey(c.Directory, FilePath)
	err = ucloudUploadMethod(client)
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(UConfig.CustomDomain, UConfig.BucketName, UConfig.Endpoint)
}

func ucloudUploadMethod(client *ufsdk.UFileRequest) (err error) {
	if err = client.UploadHit(filePath, fileKey); err == nil {
		util.Log.Info("文件秒传至Ucloud成功")
		return nil
	}
	if fileSize <= maxFileSize {
		err = client.PutFile(filePath, fileKey, fileMime)
		return
	}
	err = client.AsyncUpload(filePath, fileKey, fileMime, 8)
	return
}
