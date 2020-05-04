/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: ucloud.go
 * Date: 2020/5/2 上午1:19
 * Author: sevth
 */

package core

import "../SDK/ufile"

type Ucloud struct {
	Name            string `json:"name"`
	PublicKey       string `json:"public_key"`
	PrivateKey      string `json:"private_key"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
	VerifyUploadMD5 bool   `json:"verify_upload_md5"`
}

func ucloud() (link string) {
	util.Log.Info("使用 uclod SDK 上传")
	UConfig := config.StorageTypes.Ucloud
	UC := &ufile.Config{
		PublicKey:       UConfig.PublicKey,
		PrivateKey:      UConfig.PrivateKey,
		BucketName:      UConfig.BucketName,
		FileHost:        UConfig.Endpoint,
		VerifyUploadMD5: UConfig.VerifyUploadMD5,
	}
	req, err := ufile.NewFileRequest(UC, nil)
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return
	}
	//fileKey := Util.MakeFileKey(c.Directory, FilePath)
	err = ucloudUploadMethod(req)
	if err != nil {
		util.Log.Error("Ucloud SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(UConfig.CustomDomain, UConfig.BucketName, UConfig.Endpoint)
}

func ucloudUploadMethod(request *ufile.Request) (err error) {
	if err = request.UploadHit(filePath, fileKey); err == nil {
		util.Log.Info("文件秒传至Ucloud成功")
		return nil
	}
	//mimeType := util.GetFileMimeType(filePath)
	//fileSize := util.GetFileSize(filePath)
	if fileSize <= maxFileSize {
		err = request.PutFile(filePath, fileKey, "")
		return
	}
	err = request.AsyncMPut(filePath, fileKey, "")
	return
}
