/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: tencent.go
 * Date: 2020/5/4 上午12:54
 * Author: sevth
 */

package core

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
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

func (t Tencent)upload(info *fileInfo) (link string) {
	util.Log.Info("使用 tencent By Aws SDK 上传")

	var err error
	var region string
	if t.Endpoint != "" {
		region = strings.Split(t.Endpoint, ".")[1]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(t.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(t.SecretID, t.SecretKey, t.SessionToken),
	}))

	svc := s3.New(sess)

	// 普通上传
	if info.fileSize <= maxFileSize {
		fd, _ := util.OpenFileByReadOnly(info.filePath)
		defer fd.Close()
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:       aws.ReadSeekCloser(fd),
			Bucket:     aws.String(t.BucketName),
			//ContentMD5: aws.String(info.fileMD5),
			Key:        aws.String(info.fileKey),
		})
	} else {
		// 分片上传
		util.Log.Info("腾讯云使用分片上传文件：", info.fileName)
		upload := &AwsMultiPartUpload{
			Bucket:   t.BucketName,
			FilePath: info.filePath,
			FileSize: info.fileSize,
			FileKey:  info.fileKey,
			FileMime: info.fileMime,
			PartSize: partSize,
		}
		err = upload.AwsMultipartUpload(svc)
	}

	if err != nil {
		util.Log.Error("腾讯云 By AWS SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(t.CustomDomain, t.BucketName, t.Endpoint, info.fileKey)
}

func (t Tencent)delete(info *fileInfo) bool {
	var err error
	var region string
	if t.Endpoint != "" {
		region = strings.Split(t.Endpoint, ".")[1]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(t.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(t.SecretID, t.SecretKey, t.SessionToken),
	}))

	svc := s3.New(sess)
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(t.BucketName),
		Key:    aws.String(info.fileKey),
	})
	if err != nil {
		return false
	}
	return true
}
