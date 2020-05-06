/*
 * Copyright (c) 2020. sevth <sevthdev@gmail.com>
 * Project name: FCM, File name: JD.go
 * Date: 2020/5/5 下午12:52
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

type JD struct {
	Name            string `json:"name"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	CustomDomain    string `json:"custom_domain"`
}

func jd() (link string) {
	var region string
	jdConfig := config.StorageTypes.JD

	if jdConfig.Endpoint != "" {
		region = strings.Split(jdConfig.Endpoint, ".")[1]
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(jdConfig.Endpoint),
		Region:      aws.String(region),
		DisableSSL:  aws.Bool(false),
		Credentials: credentials.NewStaticCredentials(jdConfig.AccessKeyId, jdConfig.AccessKeySecret, ""),
	}))

	svc := s3.New(sess)

	err := jdUploadMethod(svc, jdConfig.BucketName)

	if err != nil {
		util.Log.Error("JD By AWS SDK throw err ", err)
		return
	}
	return util.MakeReturnLink(jdConfig.CustomDomain, jdConfig.BucketName, jdConfig.Endpoint)
}

func jdUploadMethod(svc *s3.S3, bucket string) (err error) {
	// 普通上传
	if fileSize <= maxFileSize {
		fd, _ := util.OpenFileByReadOnly(filePath)
		defer fd.Close()
		_, err = svc.PutObject(&s3.PutObjectInput{
			Body:       aws.ReadSeekCloser(fd),
			Bucket:     aws.String(bucket),
			ContentMD5: aws.String(fileMD5),
			Key:        aws.String(fileKey),
		})
		return
	}

	// 分片上传
	util.Log.Info("京东云使用分片上传文件：", fileName)
	//upload := NewAwsMultiPartUpload()
	upload := &AwsMultiPartUpload{
		Bucket:         bucket,
		FilePath:       filePath,
		FileSize:       fileSize,
		FileKey:        fileKey,
		FileMime:       fileMime,
		PartSize:       partSize,
	}
	return upload.AwsMultipartUpload(svc)
}
