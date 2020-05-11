/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: ucloud.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import (
	"github.com/ufilesdk-dev/ufile-gosdk"
)

type Ucloud struct {
	Name            string `json:"name"`
	PublicKey       string `json:"public_key"`
	PrivateKey      string `json:"private_key"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func (u Ucloud)upload(info *fileInfo) (link string) {
	util.Log.Info("使用 uclod SDK 上传")
	client, err := ufsdk.NewFileRequest(&ufsdk.Config{
		PublicKey:       u.PublicKey,
		PrivateKey:      u.PrivateKey,
		BucketName:      u.BucketName,
		FileHost:        u.Endpoint,
		VerifyUploadMD5: true,
	}, nil)
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return
	}

	if info.fileSize <= maxFileSize {
		err = client.PutFile(info.filePath, info.fileKey, info.fileMime)
	} else {
		util.Log.Info("ucloud 使用分片上传")
		err = client.AsyncUpload(info.filePath, info.fileKey, info.fileMime, 8)
	}
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(u.CustomDomain, u.BucketName, u.Endpoint, info.fileKey)
}

func (u Ucloud)delete(info *fileInfo) bool {
	client, err := ufsdk.NewFileRequest(&ufsdk.Config{
		PublicKey:       u.PublicKey,
		PrivateKey:      u.PrivateKey,
		BucketName:      u.BucketName,
		FileHost:        u.Endpoint,
		VerifyUploadMD5: true,
	}, nil)
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return false
	}

	err = client.DeleteFile(info.fileKey)
	if err != nil {
		return false
	}
	return true
}

