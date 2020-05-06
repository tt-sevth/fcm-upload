/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: qiniu.go
 * Date: 2020/5/4 下午5:31
 * Author: sevth
 */

package core

import (
	"context"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

type Qiniu struct {
	Name         string `json:"name"`
	AK           string `json:"ak"`
	SK           string `json:"sk"`
	BucketName   string `json:"bucket_name"`
	Endpoint     string `json:"endpoint"`
	CustomDomain string `json:"custom_domain"`
}

func qiniu()(link string) {
	QConfig := config.StorageTypes.Qiniu
	reg, err := storage.GetRegion(QConfig.AK, QConfig.BucketName)
	if err != nil {
		util.Log.Error("Qiniu SDK throw err ", err)
		return
	}
	cfg := storage.Config{
		Zone: reg,
	}
	putPolicy := storage.PutPolicy{
		Scope: QConfig.BucketName,
	}
	upToken := putPolicy.UploadToken(qbox.NewMac(QConfig.AK, QConfig.SK))

	err = qiniuUploadMethod(&cfg, upToken)
	if err != nil {
		util.Log.Error("Qiniu SDK throw err ", err)
		return
	}
	return util.MakeReturnLink(QConfig.CustomDomain, QConfig.BucketName, QConfig.Endpoint)
}

func qiniuUploadMethod(cfg *storage.Config, upToken string) (err error) {
	if fileSize <= maxFileSize {
		formUploader := storage.NewFormUploader(cfg)
		ret := storage.PutRet{}
		return formUploader.PutFile(context.Background(), &ret, upToken, fileKey, filePath, nil)
	}
	resumeUploader := storage.NewResumeUploader(cfg)
	ret := storage.PutRet{}
	putExtra := storage.RputExtra{}
	return resumeUploader.PutFile(context.Background(), &ret, upToken, fileKey, filePath, &putExtra)
}
