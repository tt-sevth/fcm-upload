/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: ucloud.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import (
	"../SDK/ufile"
)

const (
	maxFileSize = 2 << 25
)

type Ucloud struct {
	Name            string `json:"name"`
	PublicKey       string `json:"public_key"`
	PrivateKey      string `json:"private_key"`
	BucketName      string `json:"bucket_name"`
	FileHost        string `json:"file_host"`
	Directory       string `json:"directory"`
	CustomDomain    string `json:"custom_domain"`
	VerifyUploadMD5 bool   `json:"verify_upload_md_5"`
}

func ucloud(FilePath, FileKey string) (link string) {
	util.Log.Info("使用 uclod SDK 上传")
	UConfig := config.StorageTypes.Ucloud
	UC := &ufile.Config{
		PublicKey:       UConfig.PublicKey,
		PrivateKey:      UConfig.PrivateKey,
		BucketName:      UConfig.BucketName,
		FileHost:        UConfig.FileHost,
		VerifyUploadMD5: UConfig.VerifyUploadMD5,
	}
	req, err := ufile.NewFileRequest(UC, nil)
	if err != nil {
		util.Log.Error(err)
		return
	}
	//fileKey := Util.MakeFileKey(c.Directory, FilePath)
	err = ucloudUploadMethod(FilePath, FileKey, req)
	if err != nil {
		util.Log.Error(err.Error())
		return
	}
	domain := UConfig.CustomDomain
	if domain == "" {
		domain = "http://" + UConfig.BucketName + "." + UConfig.FileHost
	}
	link = domain + "/" + FileKey
	return
}

func ucloudUploadMethod(filePath, keyName string, request *ufile.Request) (err error) {
	if err = request.UploadHit(filePath, keyName); err == nil {
		util.Log.Info("文件秒传至Ucloud成功")
		return nil
	}
	//mimeType := util.GetFileMimeType(filePath)
	fileSize := util.GetFileSize(filePath)
	if fileSize <= maxFileSize {
		err = request.PutFile(filePath, keyName, "")
		return
	}
	err = request.AsyncMPut(filePath, keyName, "")
	return
}
