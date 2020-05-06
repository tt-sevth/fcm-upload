/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: qiniu.go
 * Date: 2020/5/4 下午5:31
 * Author: sevth
 */

package core

import (
	"context"
	qnQbox "github.com/qiniu/api.v7/auth/qbox"
	qnStorage "github.com/qiniu/api.v7/storage"
)

type Qiniu struct {
	Name         string `json:"name"`
	AK           string `json:"ak"`
	SK           string `json:"sk"`
	BucketName   string `json:"bucket_name"`
	Endpoint     string `json:"endpoint"`
	CustomDomain string `json:"custom_domain"`
}

func (q Qiniu)upload(info *fileInfo)(link string) {
	region, err := qnStorage.GetRegion(q.AK, q.BucketName)
	if err != nil {
		util.Log.Error("Qiniu SDK throw err ", err)
		return
	}
	cfg := &qnStorage.Config{
		Zone: region,
	}
	putPolicy := &qnStorage.PutPolicy{
		Scope: q.BucketName,
	}
	upToken := putPolicy.UploadToken(qnQbox.NewMac(q.AK, q.SK))

	if info.fileSize <= maxFileSize {
		formUploader := qnStorage.NewFormUploader(cfg)
		ret := &qnStorage.PutRet{}
		err = formUploader.PutFile(context.Background(), ret, upToken, info.fileKey, info.filePath, nil)
	} else {
		resumeUploader := qnStorage.NewResumeUploader(cfg)
		ret := &qnStorage.PutRet{}
		putExtra := &qnStorage.RputExtra{}
		err = resumeUploader.PutFile(context.Background(), ret, upToken, info.fileKey, info.filePath, putExtra)
	}

	if err != nil {
		util.Log.Error("Qiniu SDK throw err ", err)
		return
	}
	return util.MakeReturnLink(q.CustomDomain, q.BucketName, q.Endpoint, info.fileKey)
}
