/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: aliyun.go
 * Date: 2020/5/3 上午3:44
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

type Aliyun struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func (a Aliyun) upload(info *fileInfo) (link string) {
	util.Log.Info("使用 aliyun SDK 上传")
	var err error
	var region string
	if a.Endpoint != "" {
		region = strings.Split(a.Endpoint, ".")[0]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(a.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(a.AccessKeyId, a.AccessKeySecret, ""),
	}))

	svc := s3.New(sess)

	// 普通上传
	if info.fileSize <= maxFileSize {
		fd, _ := util.OpenFileByReadOnly(info.filePath)
		defer fd.Close()
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:       aws.ReadSeekCloser(fd),
			Bucket:     aws.String(a.BucketName),
			//ContentMD5: aws.String(info.fileMD5),
			Key:        aws.String(info.fileKey),
		})
	} else {
		// 分片上传
		util.Log.Info("阿里云使用分片上传文件：", info.fileName)
		//upload := NewAwsMultiPartUpload()
		upload := &AwsMultiPartUpload{
			Bucket:   a.BucketName,
			FilePath: info.filePath,
			FileSize: info.fileSize,
			FileKey:  info.fileKey,
			FileMime: info.fileMime,
			PartSize: partSize,
		}
		err = upload.AwsMultipartUpload(svc)
	}

	if err != nil {
		util.Log.Error("a By AWS SDK throw err ", err)
		return
	}

	return util.MakeReturnLink(a.CustomDomain, a.BucketName, a.Endpoint, info.fileKey)
}

func (a Aliyun) delete(info *fileInfo) bool {
	var err error
	var region string
	if a.Endpoint != "" {
		region = strings.Split(a.Endpoint, ".")[0]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(a.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(a.AccessKeyId, a.AccessKeySecret, ""),
	}))

	svc := s3.New(sess)
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(info.fileKey),
	})
	if err != nil {
		util.Log.Error("Aliyun SDK throw err ", err)
		return false
	}
	return true
}
